package production

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

// RP-301 continues actual RP-299 activation heads. Never create/reseed later
// streams, pins, genesis or frozen rows. Initial earned6/run2 remain diagnostic.
func testReputationActivatedCareer(t *testing.T, ctx context.Context, db *sql.DB, store *save.Store, service *Service, companyID, founderID string, old, tree CatalogBundle, now time.Time, command string) {
	t.Helper()
	load := func(id string) save.Loaded {
		t.Helper()
		value, err := store.LoadLatest(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	intentIndex := 0
	handle := func(at time.Time, kind, wantOutcome string, fields map[string]any) ([]byte, HandleResult) {
		t.Helper()
		owner, player := load(founderID), load(companyID)
		intentIndex++
		body := map[string]any{"intent_id": fmt.Sprintf("01986666-e300-7000-8000-%012d", intentIndex), "kind": kind, "expected_revision": player.Revision.Number}
		if kind == IntentPurchaseReputationNode {
			body["expected_revision"] = owner.Revision.Number
		}
		if kind == IntentWindDown {
			body["expected_founder_revision"] = owner.Revision.Number
		}
		for key, value := range fields {
			body[key] = value
		}
		request := reputationShapeJSON(t, body)
		result, err := service.Handle(ctx, companyID, ModeOnline, at, request)
		var receipt struct {
			Outcome string `json:"outcome"`
		}
		if err != nil || result.Replay || json.Unmarshal(result.Receipt, &receipt) != nil || receipt.Outcome != wantOutcome {
			t.Fatalf("activated career %s receipt=%s replay=%v err=%v", kind, result.Receipt, result.Replay, err)
		}
		return request, result
	}
	retry := func(at time.Time, request []byte, result HandleResult) {
		t.Helper()
		before := reputationPlanDBSnapshot(t, ctx, db)
		got, err := service.Handle(ctx, companyID, ModeOnline, at, request)
		if err != nil || !got.Replay || !bytes.Equal(got.Receipt, result.Receipt) {
			t.Fatalf("career retry differs: %v", err)
		}
		for table, rows := range reputationPlanDBSnapshot(t, ctx, db) {
			if rows != before[table] {
				t.Fatalf("career retry rewrote %s", table)
			}
		}
	}
	beforeOwner, beforePlayer := load(founderID), load(companyID)
	ownedRequest, ownedResult := handle(now.Add(time.Second), IntentPurchaseReputationNode, "rejected", map[string]any{"node_id": "reputation.unlock.p05"})
	category, detail := rejectionOf(t, ownedResult.Receipt)
	if category != "not_eligible" || detail != "owned" {
		t.Fatal("activated owned purchase taxonomy differs")
	}
	owner, player := load(founderID), load(companyID)
	if owner.Revision != beforeOwner.Revision || player.Revision != beforePlayer.Revision || !bytes.Equal(mustEncodeState(t, owner.State), mustEncodeState(t, beforeOwner.State)) || !bytes.Equal(mustEncodeState(t, player.State), mustEncodeState(t, beforePlayer.State)) {
		t.Fatal("owned purchase changed activated full heads")
	}
	retry(now.Add(2*time.Second), ownedRequest, ownedResult)
	manual := map[string]any{"action_id": "manual.click", "count": 1, "window_ms": 1}
	factor, err := decimal.ParseCanonical("1.003e0")
	if err != nil {
		t.Fatal(err)
	}
	handle(now.Add(20000*time.Second), IntentPerformManualBatch, "applied", manual)
	player = load(companyID)
	wantCash := decimal.FromFloat64(5).Mul(factor).Mul(decimal.FromFloat64(20000)).Add(decimal.New(1, 3)).Add(decimal.One)
	cash, ok := player.State.Ledger.Balance("company.cash")
	if !ok || !cash.Eq(wantCash) || player.State.RunSeq != 3 {
		t.Fatalf("activated run3 production cash=%s want=%s run=%d", cash, wantCash, player.State.RunSeq)
	}
	handle(now.Add(20001*time.Second), IntentCrossGate, "applied", map[string]any{"gate_id": "gate.t0_to_t1", "route_id": nil})
	player = load(companyID)
	if player.State.Tier != 1 || !player.State.GatesCrossed["gate.t0_to_t1"] {
		t.Fatal("activated career did not cross actual Garage gate")
	}
	oldRows := map[string]string{}
	readOldRows := func(table string) string {
		t.Helper()
		var rows string
		query := fmt.Sprintf(`SELECT coalesce(jsonb_agg(to_jsonb(row) ORDER BY run_seq,to_jsonb(row)::text),'[]'::jsonb)::text FROM %s row WHERE company_stream_id=$1 AND run_seq<=3`, table)
		if err := db.QueryRowContext(ctx, query, companyID).Scan(&rows); err != nil {
			t.Fatal(err)
		}
		return rows
	}
	for _, table := range []string{"run_epochs", "run_genesis", "run_frozen_contributions"} {
		oldRows[table] = readOldRows(table)
	}
	secondRequest, secondResult := handle(now.Add(20002*time.Second), IntentWindDown, "applied", nil)
	owner, player = load(founderID), load(companyID)
	firstKind := command
	if firstKind == "wind_down" {
		firstKind = "collapse"
	}
	wantOwned := []string{"reputation.starter.cash_small", "reputation.starter.generated_beige_tower", "reputation.unlock.p05"}
	if owner.Revision.Number != 3 || owner.Revision.Version != 22 || owner.Revision.ConstantsHash != tree.ConstantsHash || owner.State.ReputationLevel != 6 || owner.State.ReputationSpent != 6 || owner.State.ReputationUnlockPPM != 50000 || !slices.Equal(owner.State.ReputationNodesOwned, wantOwned) || len(owner.State.ExitHistory) != 3 || owner.State.ExitHistory[1].ExitType != firstKind || owner.State.ExitHistory[2].ExitType != "collapse" || player.Revision.Number != 7 || player.State.RunSeq != 4 || player.Revision.ConstantsHash != tree.ConstantsHash {
		t.Fatal("second Exit lost activated Founder accounting/pin/history")
	}
	cash, ok = player.State.Ledger.Balance("company.cash")
	if !ok || cash.String() != "1e3" || player.State.GeneratorProvisioned["generator.beige_tower"] != 5 || player.State.GeneratorCounts["generator.beige_tower"] != 0 || player.State.GeneratorPurchasedTotal != 0 {
		t.Fatal("second Exit starter must repeat exactly once, not accumulate generated inventory")
	}
	for table, before := range oldRows {
		if got := readOldRows(table); got != before {
			t.Fatalf("second Exit rewrote old %s", table)
		}
	}
	var frozen, pin, summary string
	var epoch, purchases int
	if err := db.QueryRowContext(ctx, `SELECT
        (SELECT factor FROM run_frozen_contributions WHERE company_stream_id=$1 AND run_seq=4 AND source_id='reputation.founder_bonus'),
        (SELECT constants_hash FROM run_epochs WHERE company_stream_id=$1 AND run_seq=4),
        (SELECT epoch_id FROM run_epochs WHERE company_stream_id=$1 AND run_seq=4),
        (SELECT payload->'reputation_tree'->>'bonus_factor' FROM events WHERE stream_id=$1 AND kind='run_started' AND payload->'run_id'->>'run_seq'='4'),
        (SELECT count(*) FROM events WHERE stream_id=$2 AND kind='reputation_node_purchased.v1')`, companyID, founderID).Scan(&frozen, &pin, &epoch, &summary, &purchases); err != nil || frozen != "1.003e0" || summary != "1.003e0" || pin != tree.ConstantsHash || epoch != 2 || purchases != 3 {
		t.Fatalf("second Exit repeated frozen/pin/summary/purchases differ: %s/%s %s/%d %d %v", frozen, summary, pin, epoch, purchases, err)
	}
	for _, run := range []int64{2, 3} {
		genesis, version, entries := reputationCareerReplay(t, db, companyID, founderID, run, &tree)
		bundle, wantEntries := old, 1
		if run == 3 {
			bundle, wantEntries = tree, 3
		}
		if len(entries) != wantEntries || VerifyReplayRun(genesis, version, bundle, entries, bundle.ConstantsHash, false) != ReplayVerified {
			t.Fatalf("cross-pin career run%d did not verify", run)
		}
		if run == 2 && VerifyReplayRun(genesis, version, bundle, entries, tree.ConstantsHash, false) != ReplayConstantsMismatch {
			t.Fatal("old completed run accepted deploy-current pin")
		}
	}
	set := ReplayCatalogSet{old.ConstantsHash: old, tree.ConstantsHash: tree}
	history, err := store.LoadFounderHistory(ctx, founderID)
	if err != nil || len(history.Entries) != 4 || history.Genesis.ConstantsHash != old.ConstantsHash || history.HeadConstants != tree.ConstantsHash || VerifyFounderHistory(history, set) != ReplayVerified {
		t.Fatalf("complete four-entry cross-pin history did not verify: %v", err)
	}
	if history.Entries[1].Source == nil || history.Entries[1].Source.RunSeq != 2 || history.Entries[3].Source == nil || history.Entries[3].Source.RunSeq != 3 || history.Entries[3].Source.RunLogSeq != 3 {
		t.Fatal("both Exit source coordinates are not actual completed runs")
	}
	forgedHead := cloneReputationHistory(t, history)
	head := reputationShapeObject(t, forgedHead.HeadState)
	head["reputation_spent"] = json.RawMessage("5")
	forgedHead.HeadState = reputationShapeJSON(t, head)
	if VerifyFounderHistory(forgedHead, set) != ReplayStateDivergence {
		t.Fatal("cross-pin career verifier accepted forged final spent")
	}
	forgedSource := cloneReputationHistory(t, history)
	forgedSource.Entries[3].Source.RunSeq++
	if VerifyFounderHistory(forgedSource, set) != ReplayStateDivergence {
		t.Fatal("cross-pin career verifier accepted mismatched second Exit source")
	}
	retry(now.Add(20003*time.Second), secondRequest, secondResult)
	handle(now.Add(21002*time.Second), IntentPerformManualBatch, "applied", manual)
	player = load(companyID)
	wantCash = decimal.FromFloat64(5).Mul(factor).Mul(decimal.FromFloat64(1000)).Add(decimal.New(1, 3)).Add(decimal.One)
	cash, ok = player.State.Ledger.Balance("company.cash")
	if !ok || !cash.Eq(wantCash) || player.State.RunSeq != 4 {
		t.Fatalf("repeated run4 production cash=%s want=%s", cash, wantCash)
	}
	genesis, version, entries := reputationCareerReplay(t, db, companyID, founderID, 4, &tree)
	if len(entries) != 1 || entries[0].Terminal {
		t.Fatal("ongoing run4 must not claim completed verification")
	}
	state, err := save.RestoreState(genesis, version, tree.Economy, economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	continued, err := ApplyLogged(state, entries[0].CanonicalPayload, tree, entries[0].ReplayInputs)
	if err != nil || !canonicalJSONEqual(continued.Receipt, entries[0].ReceiptJSON) || !canonicalJSONEqual(marshalReplayEvents(continued.Events), entries[0].EventsJSON) || !bytes.Equal(mustEncodeState(t, continued.State), mustEncodeState(t, player.State)) {
		t.Fatalf("ongoing run4 complete recorded replay differs: %v", err)
	}
	t.Log("activated cross-pin career: two completed Company runs/four-entry Founder history, repeated starter/frozen production, source/head negatives and exact retries verified; run4 stays unfinished")
}
