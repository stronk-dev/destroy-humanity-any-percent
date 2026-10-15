package gameserver

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

// R-012 observation over the existing real HTTP/Guild/clearing population.
// Replay receives only its saved pre-head, pinned bundle and persisted inputs;
// no live membership, clearing, clock or other provider is passed to ApplyLogged.
func observeComposedGuildSettlementReplay(t *testing.T, db *sql.DB, catalogs *runtimeCatalogs, before, after account.FounderState, intentID, body string) {
	t.Helper()
	ctx := context.Background()
	if before.StreamID != after.StreamID || before.ConstantsHash != after.ConstantsHash || after.Revision != before.Revision+1 || before.Version != after.Version {
		t.Fatal("Guild replay capture is not one unchanged-pin Company transition")
	}
	bundle, ok := catalogs.ResolveReplayCatalogs(before.ConstantsHash)
	if !ok {
		t.Fatal("captured Company pin unavailable")
	}
	var payload, inputs, receipt []byte
	var revision int64
	if err := db.QueryRowContext(ctx, `SELECT canonical_payload,replay_inputs,receipt,applied_revision
		FROM run_log WHERE company_stream_id=$1 AND intent_id=$2`, before.StreamID, intentID).
		Scan(&payload, &inputs, &receipt, &revision); err != nil {
		t.Fatal(err)
	}
	request, err := production.ParseIntent([]byte(body))
	if err != nil || revision != after.Revision || !bytes.Equal(payload, request.CanonicalPayload) {
		t.Fatal("stored command/revision does not bind the actual HTTP action")
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(inputs, &envelope); err != nil {
		t.Fatal(err)
	}
	var mode string
	if err := json.Unmarshal(envelope["evaluation_mode"], &mode); err != nil || mode != "online" {
		t.Fatal("unexpected live evaluation mode", err)
	}
	var resolved map[string]json.RawMessage
	if err := json.Unmarshal(envelope["resolved"], &resolved); err != nil {
		t.Fatal(err)
	}
	var accrual map[string]json.RawMessage
	if err := json.Unmarshal(resolved["accrual"], &accrual); err != nil {
		t.Fatal(err)
	}
	var batch struct {
		GuildID     string `json:"guild_id"`
		BaseSeq     int64  `json:"base_seq"`
		Settlements []struct {
			BoundarySeq int64 `json:"boundary_seq"`
			CreditUnits int64 `json:"credit_units"`
		} `json:"settlements"`
	}
	if err := json.Unmarshal(accrual["guild_settlement_batch"], &batch); err != nil {
		t.Fatal(err)
	}
	credit := int64(0)
	for index, row := range batch.Settlements {
		if row.BoundarySeq != int64(index+1) || row.CreditUnits < 0 {
			t.Fatal("captured clearing order/credit invalid")
		}
		credit += row.CreditUnits
	}
	if batch.GuildID == "" || batch.BaseSeq != 0 || len(batch.Settlements) != 3 || credit <= 0 {
		t.Fatal("live nonempty consumer-clearing population absent")
	}
	restore := func() *save.State {
		t.Helper()
		state, err := save.RestoreState(before.State, before.Version, bundle.Economy, economy.ScopeCompany, time.Unix(0, 0).UTC())
		if err != nil {
			t.Fatal(err)
		}
		if err := bundle.ValidateFoundationState(state); err != nil {
			t.Fatal(err)
		}
		return state
	}
	encode := func(state *save.State) []byte {
		t.Helper()
		encoded, err := save.EncodeStateVersion(state, before.Version)
		if err != nil {
			t.Fatal(err)
		}
		return encoded
	}
	state := restore()
	originState := restore()
	owned := int64(0)
	for _, count := range originState.GeneratorCounts {
		owned += count
	}
	if owned != 1 || originState.GeneratorCounts["generator.beige_tower"] != 1 {
		t.Fatal("controlled single-producer population absent")
	}
	transition, err := production.ApplyLogged(state, payload, bundle, inputs)
	if err != nil || transition.Outcome != save.IntentApplied {
		t.Fatal("stored live input cannot replay", err)
	}
	if !state.LifetimeValue.Gt(originState.LifetimeValue) {
		t.Fatal("capture did not settle positive idle production through the Prestige accrual hook")
	}
	if !reflect.DeepEqual(guildReplayJSON(t, encode(state)), guildReplayJSON(t, after.State)) ||
		!reflect.DeepEqual(guildReplayJSON(t, transition.Receipt), guildReplayJSON(t, receipt)) {
		t.Fatal("stored live replay lost complete Company or receipt content")
	}
	var storedEvents []byte
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(jsonb_agg(jsonb_build_object(
		'kind',kind,'schema_version',schema_version,'intent_id',intent_id,'payload',payload) ORDER BY event_seq),'[]'::jsonb)
		FROM events WHERE stream_id=$1 AND intent_id=$2`, before.StreamID, intentID).Scan(&storedEvents); err != nil {
		t.Fatal(err)
	}
	events := make([]map[string]any, 0, len(transition.Events))
	for _, event := range transition.Events {
		events = append(events, map[string]any{"kind": event.Kind, "schema_version": event.SchemaVersion, "intent_id": event.IntentID, "payload": event.Payload})
	}
	replayedEvents, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 || !reflect.DeepEqual(guildReplayJSON(t, replayedEvents), guildReplayJSON(t, storedEvents)) {
		t.Fatal("stored live replay lost ordered Company events")
	}
	if state.GuildBoundarySeq != 3 || state.ConsumedStockUnits < credit {
		t.Fatal("positive replay did not apply actual clearing")
	}
	zeroProducer := restore()
	zeroProducer.GeneratorCounts["generator.beige_tower"] = 0
	zeroLifetime := zeroProducer.LifetimeValue
	zeroResult, err := production.ApplyLogged(zeroProducer, payload, bundle, inputs)
	if err != nil || zeroResult.Outcome != save.IntentApplied || !zeroProducer.LifetimeValue.Eq(zeroLifetime) ||
		zeroProducer.GuildBoundarySeq != state.GuildBoundarySeq || zeroProducer.ConsumedStockUnits != state.ConsumedStockUnits ||
		reflect.DeepEqual(guildReplayJSON(t, encode(zeroProducer)), guildReplayJSON(t, after.State)) {
		t.Fatal("zero-producer control did not isolate idle accrual from clearing", err)
	}

	// Alter only the retained batch. Keep every other captured input unchanged.
	mutateBatch := func(settlements json.RawMessage) []byte {
		t.Helper()
		var changed map[string]json.RawMessage
		if err := json.Unmarshal(accrual["guild_settlement_batch"], &changed); err != nil {
			t.Fatal(err)
		}
		changed["settlements"] = settlements
		mutated := map[string]json.RawMessage{}
		for key, value := range accrual {
			mutated[key] = value
		}
		mutated["guild_settlement_batch"], err = json.Marshal(changed)
		if err != nil {
			t.Fatal(err)
		}
		resolvedCopy := map[string]json.RawMessage{}
		for key, value := range resolved {
			resolvedCopy[key] = value
		}
		resolvedCopy["accrual"], err = json.Marshal(mutated)
		if err != nil {
			t.Fatal(err)
		}
		envelopeCopy := map[string]json.RawMessage{}
		for key, value := range envelope {
			envelopeCopy[key] = value
		}
		envelopeCopy["resolved"], err = json.Marshal(resolvedCopy)
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(envelopeCopy)
		if err != nil {
			t.Fatal(err)
		}
		return encoded
	}
	omitted := restore()
	omittedResult, err := production.ApplyLogged(omitted, payload, bundle, mutateBatch(json.RawMessage(`[]`)))
	if err != nil || omittedResult.Outcome != save.IntentApplied || omitted.GuildBoundarySeq != 0 ||
		omitted.ConsumedStockUnits == state.ConsumedStockUnits || reflect.DeepEqual(guildReplayJSON(t, encode(omitted)), guildReplayJSON(t, after.State)) {
		t.Fatal("batch omission did not distinguish lost settlement state", err)
	}
	var originalBatch map[string]json.RawMessage
	if err := json.Unmarshal(accrual["guild_settlement_batch"], &originalBatch); err != nil {
		t.Fatal(err)
	}
	var rows []json.RawMessage
	if err := json.Unmarshal(originalBatch["settlements"], &rows); err != nil {
		t.Fatal(err)
	}
	duplicated, err := json.Marshal(append(rows, rows[len(rows)-1]))
	if err != nil {
		t.Fatal(err)
	}
	refused := restore()
	refusedBefore := encode(refused)
	result, err := production.ApplyLogged(refused, payload, bundle, mutateBatch(duplicated))
	if !errors.Is(err, production.ErrInvalidReplayInputs) || !bytes.Equal(refusedBefore, encode(refused)) || result.Outcome != "" || len(result.Receipt) != 0 || len(result.Events) != 0 {
		t.Fatal("duplicate clearing did not refuse atomically", err)
	}
	t.Log("one live online Guild profile: three clearings, positive idle lifetime and zero-producer control; complete SQL state/receipt/events replay; omission loses settlement and duplicate refuses; conservation NOT_PROVEN")
	if os.Getenv("TEST_PUBLIC_REPLAY_CLIENT") != "1" {
		t.Log("TypeScript Guild capture seam not requested (root composed lane owns it)")
		return
	}
	clientInput, err := json.Marshal(struct {
		Kind                                            string            `json:"kind"`
		ConstantsHash                                   string            `json:"constants_hash"`
		Artifacts                                       map[string][]byte `json:"artifacts"`
		Version                                         int               `json:"version"`
		Before, After, Payload, Inputs, Receipt, Events json.RawMessage
		Omitted, Duplicate                              json.RawMessage
	}{Kind: "stored_guild_transition", ConstantsHash: before.ConstantsHash, Artifacts: bundle.Artifacts, Version: before.Version,
		Before: before.State, After: after.State, Payload: payload, Inputs: inputs, Receipt: receipt, Events: storedEvents,
		Omitted: mutateBatch(json.RawMessage(`[]`)), Duplicate: mutateBatch(duplicated)})
	if err != nil {
		t.Fatal(err)
	}
	clientCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	command := exec.CommandContext(clientCtx, "node", "client/tools/public-replay-composed.mjs")
	command.Dir = filepathRoot(t)
	command.Stdin = bytes.NewReader(clientInput)
	output, err := command.CombinedOutput()
	const passed = "stored live Guild inputs → TypeScript replay: complete state/receipt/events + positive idle/zero-producer + omission/duplicate controls: PASS"
	if err != nil || !bytes.Contains(output, []byte(passed)) {
		t.Fatalf("TypeScript Guild replay seam failed: %v (private capture/output withheld)", err)
	}
	t.Log(passed)
}

func guildReplayJSON(t *testing.T, data []byte) any {
	t.Helper()
	if !json.Valid(data) {
		t.Fatal("invalid captured JSON")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		t.Fatal(err)
	}
	return value
}
