package production

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

const reputationTaxonomySourceSHA = "f9b129e36af5b536f67c7eddb8d6088c76ce5cb0b170cfeec6a5172e8a009782"

var reputationTaxonomyRejections = map[string][2]string{
	"rejects-inactive-tree":           {"not_eligible", "reputation_tree_inactive"},
	"rejects-invalid-fields":          {"invalid", "purchase_reputation_node.fields"},
	"rejects-invalid-node-id":         {"invalid", "node_id"},
	"rejects-unknown-node":            {"unknown_id", "reputation.unlock.p99"},
	"rejects-missing-prerequisite":    {"not_eligible", "requires"},
	"rejects-ladder-prerequisite":     {"not_eligible", "requires"},
	"rejects-owned":                   {"not_eligible", "owned"},
	"rejects-unaffordable-starter":    {"unaffordable", "reputation"},
	"rejects-cost-one-over-available": {"unaffordable", "reputation"},
}

func validateReputationTaxonomySource(source reputationCorpus, raw []byte) error {
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != reputationTaxonomySourceSHA {
		return fmt.Errorf("taxonomy source SHA changed")
	}
	if source.Version != 1 || len(source.Cases) != 20 {
		return fmt.Errorf("taxonomy source needs exactly twenty cases")
	}
	names := map[string]bool{}
	applied := 0
	for _, row := range source.Cases {
		if names[row.Name] {
			return fmt.Errorf("duplicate taxonomy case %s", row.Name)
		}
		names[row.Name] = true
		var receipt struct {
			Outcome   string `json:"outcome"`
			Rejection struct {
				Category string `json:"category"`
				Detail   string `json:"detail"`
			} `json:"rejection"`
		}
		if json.Unmarshal([]byte(row.ReceiptJSON), &receipt) != nil || receipt.Outcome != row.Outcome {
			return fmt.Errorf("taxonomy outcome differs for %s", row.Name)
		}
		if want, rejected := reputationTaxonomyRejections[row.Name]; rejected {
			if row.Outcome != string(save.IntentRejected) || receipt.Rejection.Category != want[0] || receipt.Rejection.Detail != want[1] {
				return fmt.Errorf("taxonomy rejection differs for %s", row.Name)
			}
		} else if row.Outcome == string(save.IntentApplied) {
			applied++
		} else {
			return fmt.Errorf("unknown taxonomy outcome %s", row.Name)
		}
	}
	for name := range reputationTaxonomyRejections {
		if !names[name] {
			return fmt.Errorf("missing taxonomy case %s", name)
		}
	}
	if applied != 11 {
		return fmt.Errorf("taxonomy requires eleven applied controls")
	}
	return nil
}

type reputationTaxonomyProfile struct {
	name    string
	row     reputationCorpusCase
	bundle  CatalogBundle
	before  *save.State
	company *save.State
	request IntentRequest
}

func reputationTaxonomyProfiles(t *testing.T, now time.Time) []reputationTaxonomyProfile {
	t.Helper()
	source, raw := reputationInputShapeSource(t)
	if err := validateReputationTaxonomySource(source, raw); err != nil {
		t.Fatal(err)
	}
	bundles := reputationShapeCatalogs(t, source)
	rows := slices.Clone(source.Cases)
	overdue := source.Cases[slices.IndexFunc(source.Cases, func(row reputationCorpusCase) bool { return row.Name == "rejects-unknown-node" })]
	overdue.Name = "rejects-with-overdue-fiscal-rollback"
	rows = append(rows, overdue)
	profiles := make([]reputationTaxonomyProfile, 0, 21)
	for index, row := range rows {
		bundle := bundles[row.Bundle]
		before := reputationShapeRestore(t, row, bundle)
		before.FiscalPeriodOpenedWallMS = now.UnixMilli()
		if row.Name == overdue.Name {
			before.FiscalPeriodOpenedWallMS -= bundle.Fiscal.Clock.AutoMS
		}
		if err := bundle.ValidateFoundationState(before); err != nil {
			t.Fatal(err)
		}
		payload := reputationShapeObject(t, row.CanonicalPayload)
		payload["intent_id"] = reputationShapeJSON(t, fmt.Sprintf("01986666-9f01-7000-8000-%012d", index+1))
		payload["expected_revision"] = json.RawMessage("1")
		request, err := ParseIntent(reputationShapeJSON(t, payload))
		if err != nil {
			t.Fatal(err)
		}
		companyIndex := slices.IndexFunc(source.ExitCases, func(exit reputationExitCase) bool { return exit.Company.ConstantsHash == bundle.ConstantsHash })
		if companyIndex < 0 {
			t.Fatal("missing pinned Company fixture")
		}
		company := replayFixtureStateFromEncoded(t, bundle, source.ExitCases[companyIndex].Company.Case.PreState)
		if err := bundle.ValidateFoundationState(company); err != nil {
			t.Fatal(err)
		}
		profiles = append(profiles, reputationTaxonomyProfile{row.Name, row, bundle, before, company, request})
	}
	return profiles
}

// The source fixes purchase behavior; only clock-dependent Fiscal fields and
// independent stream coordinates are adapted. This is not a live-resolver oracle.
func assertReputationTaxonomyResult(t *testing.T, profile reputationTaxonomyProfile, serverTS int64, receipt []byte, state *save.State, events []save.EventWrite) {
	t.Helper()
	wantReceipt := reputationShapeObject(t, []byte(profile.row.ReceiptJSON))
	wantReceipt["intent_id"] = reputationShapeJSON(t, profile.request.IntentID)
	expected := reputationShapeRestore(t, profile.row, profile.bundle)
	expected.FiscalPeriodOpenedWallMS = profile.before.FiscalPeriodOpenedWallMS
	if profile.row.Outcome == string(save.IntentRejected) {
		wantReceipt["current_revision"] = json.RawMessage("1")
		expected = profile.before
		if len(events) != 0 {
			t.Fatalf("rejection committed %d events", len(events))
		}
	} else {
		var err error
		expected, err = save.RestoreState([]byte(profile.row.PostStateJSON), profile.row.StateVersion, profile.bundle.Economy, economy.ScopeFounder, time.Time{})
		if err != nil {
			t.Fatalf("invalid expected post-state: %v", err)
		}
		expected.FiscalPeriodOpenedWallMS = profile.before.FiscalPeriodOpenedWallMS
		fiscalState := fiscalStateFromSave(profile.before)
		sweep, err := profile.bundle.Fiscal.Sweep(&fiscalState, serverTS)
		if err != nil {
			t.Fatal(err)
		}
		fiscalStateToSave(expected, fiscalState)
		wantReceipt["founder_revision"] = json.RawMessage("2")
		wantReceipt["fiscal_sweep"] = json.RawMessage("null")
		wantEvents := []save.EventWrite{}
		if sweep != nil {
			wantReceipt["fiscal_sweep"] = reputationShapeJSON(t, fiscalSweepWire(sweep))
			payload := reputationShapeObject(t, reputationShapeJSON(t, fiscalSweepWire(sweep)))
			payload["source"] = json.RawMessage(`"automatic"`)
			wantEvents = append(wantEvents, save.EventWrite{Kind: save.EventFiscalPeriodHarvested, SchemaVersion: 1, IntentID: profile.request.IntentID, Payload: reputationShapeJSON(t, payload)})
		}
		var sourceEvents []fixtureEvent
		if err := json.Unmarshal([]byte(profile.row.EventsJSON), &sourceEvents); err != nil {
			t.Fatal(err)
		}
		for _, event := range sourceEvents {
			if event.Kind == string(save.EventReputationNodePurchased) {
				wantEvents = append(wantEvents, save.EventWrite{Kind: save.EventReputationNodePurchased, SchemaVersion: event.SchemaVersion, IntentID: profile.request.IntentID, Payload: event.Payload})
			}
		}
		if len(wantEvents) == 0 || wantEvents[len(wantEvents)-1].Kind != save.EventReputationNodePurchased {
			t.Fatal("missing applied source event")
		}
		if canonicalFixtureValue(t, fixtureEvents(events)) != canonicalFixtureValue(t, fixtureEvents(wantEvents)) {
			t.Fatal("ordered events differ from pinned purchase and calculated Fiscal sweep")
		}
	}
	if canonicalFixtureJSON(t, receipt) != canonicalFixtureValue(t, wantReceipt) {
		t.Fatalf("receipt differs: %s", receipt)
	}
	if !bytes.Equal(mustEncodeState(t, state), mustEncodeState(t, expected)) {
		t.Fatal("complete state differs from pinned purchase / rejected pre-state")
	}
}

func TestReputationPurchaseTaxonomyPopulation(t *testing.T) {
	source, raw := reputationInputShapeSource(t)
	if err := validateReputationTaxonomySource(source, raw); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"source-hash", "missing-row", "changed-outcome"} {
		t.Run(name, func(t *testing.T) {
			forged := source
			forged.Cases = slices.Clone(source.Cases)
			forgedRaw := raw
			switch name {
			case "source-hash":
				forgedRaw = append(slices.Clone(raw), ' ')
			case "missing-row":
				forged.Cases = forged.Cases[:len(forged.Cases)-1]
			case "changed-outcome":
				forged.Cases[0].Outcome = string(save.IntentApplied)
			}
			if validateReputationTaxonomySource(forged, forgedRaw) == nil {
				t.Fatal("forged taxonomy population accepted")
			}
		})
	}
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	profiles := reputationTaxonomyProfiles(t, now)
	if len(profiles) != 21 {
		t.Fatalf("prepared profiles=%d", len(profiles))
	}
	for _, profile := range profiles {
		t.Run(profile.name, func(t *testing.T) {
			state, err := save.RestoreState(mustEncodeState(t, profile.before), save.VersionForState(profile.before), profile.bundle.Economy, economy.ScopeFounder, time.Time{})
			if err != nil {
				t.Fatal(err)
			}
			command := save.FounderReplayCommand{IntentID: profile.request.IntentID, FounderStreamID: "01986666-9f00-4000-8000-000000000002", FounderID: "01986666-9f00-4000-8000-000000000003", Revision: 1, FounderLogSeq: 1, ServerTSMS: now.UnixMilli()}
			var resolved any = founderInvalidResolved{Kind: "invalid", Detail: profile.request.InvalidDetail}
			if profile.request.InvalidDetail == "" {
				resolved, err = resolveReputationPurchase(profile.bundle, state, profile.request.ReputationNodeID)
				if err != nil {
					t.Fatal(err)
				}
			}
			inputs, err := save.MarshalFounderReplayInputs(command, resolved)
			if err != nil {
				t.Fatal(err)
			}
			result, err := ApplyFounderLogged(state, profile.request.CanonicalPayload, profile.bundle, inputs)
			if err != nil || string(result.Outcome) != profile.row.Outcome || result.ResultConstantsHash != profile.bundle.ConstantsHash {
				t.Fatalf("prepared outcome=%s pin=%s err=%v", result.Outcome, result.ResultConstantsHash, err)
			}
			assertReputationTaxonomyResult(t, profile, command.ServerTSMS, result.Receipt, state, result.Events)
		})
	}
	t.Run("tree-present-v21-is-invalid-pinned-state", func(t *testing.T) {
		inactive := profiles[0]
		tree := profiles[1].bundle
		if save.VersionForState(inactive.before) != 21 || tree.ReputationTree == nil {
			t.Fatal("invalid mismatch control population")
		}
		if tree.ValidateFoundationState(inactive.before) == nil {
			t.Fatal("tree/v21 pinned mismatch accepted")
		}
	})
}

// Unlike the historical thin integration resolver, all writes enforce the
// same pinned foundation contract as real production catalog loading.
type reputationTaxonomyCatalogs struct {
	integrationCatalogs
	bundles ReplayCatalogSet
}

func (catalogs reputationTaxonomyCatalogs) ValidateState(hash string, state *save.State) error {
	if err := catalogs.integrationCatalogs.ValidateState(hash, state); err != nil {
		return err
	}
	bundle, ok := catalogs.bundles[hash]
	if !ok {
		return ErrInvalidEngineState
	}
	return bundle.ValidateFoundationState(state)
}

type reputationTaxonomyCounts struct{ Revisions, Logs, Intents, Events, Outbox int }

func reputationTaxonomyDBCounts(t *testing.T, db *sql.DB, stream string) reputationTaxonomyCounts {
	t.Helper()
	var result reputationTaxonomyCounts
	if err := db.QueryRow(`SELECT
		(SELECT count(*) FROM save_revisions WHERE stream_id=$1),
		(SELECT count(*) FROM founder_log WHERE founder_stream_id=$1),
		(SELECT count(*) FROM intent_records WHERE stream_id=$1),
		(SELECT count(*) FROM events WHERE stream_id=$1),
		(SELECT count(*) FROM transport_player_outbox WHERE stream_id=$1)`, stream).
		Scan(&result.Revisions, &result.Logs, &result.Intents, &result.Events, &result.Outbox); err != nil {
		t.Fatal(err)
	}
	return result
}

// RP-253: execution is mandatory on the declared disposable Postgres service.
// A host run without TEST_DATABASE_URL is explicitly preparation, not SQL proof.
func TestReputationPurchaseTaxonomyIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set; persisted taxonomy NOT EXECUTED")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE accounts,save_streams,catalog_sets,epochs RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	var preparationMS int64
	if err := db.QueryRowContext(ctx, `SELECT floor(extract(epoch FROM clock_timestamp())*1000)::bigint`).Scan(&preparationMS); err != nil {
		t.Fatal(err)
	}
	profiles := reputationTaxonomyProfiles(t, time.UnixMilli(preparationMS).UTC())
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{}, routes: map[string]*routes.Catalog{}, prestige: map[string]*prestigecore.Policy{}, factions: map[string]*faction.Catalog{}}
	set := ReplayCatalogSet{}
	for _, profile := range profiles {
		bundle := profile.bundle
		if _, exists := set[bundle.ConstantsHash]; exists {
			continue
		}
		set[bundle.ConstantsHash] = bundle
		resolver.economy[bundle.ConstantsHash], resolver.routes[bundle.ConstantsHash] = bundle.Economy, bundle.Routes
		resolver.prestige[bundle.ConstantsHash], resolver.factions[bundle.ConstantsHash] = bundle.Prestige, bundle.Faction
		seedProductionEpoch(t, db, bundle.ConstantsHash, bundle.Artifacts)
	}
	policyResolver := reputationTaxonomyCatalogs{resolver, set}
	store, err := save.NewStore(db, policyResolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService(store, resolver, nil, nil, nil, WithProgressionRuntime(resolver), WithReplayCatalogs(set), WithGuildSettlements(emptyGuildSettlements{}))
	if err != nil {
		t.Fatal(err)
	}
	for index, profile := range profiles {
		t.Run(profile.name, func(t *testing.T) {
			accountID := fmt.Sprintf("01986666-9f10-4000-8000-%012d", index+1)
			founderID := fmt.Sprintf("01986666-9f20-4000-8000-%012d", index+1)
			if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')`, accountID); err != nil {
				t.Fatal(err)
			}
			if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)`, accountID, founderID); err != nil {
				t.Fatal(err)
			}
			company := profile.company
			companyRev, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeCompany}, profile.bundle.ConstantsHash, company, save.WriteContext{Cause: "reputation.taxonomy"})
			if err != nil {
				t.Fatal(err)
			}
			founderRev, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder}, profile.bundle.ConstantsHash, profile.before, save.WriteContext{Cause: "reputation.taxonomy"})
			if err != nil {
				t.Fatal(err)
			}
			beforeCounts := reputationTaxonomyDBCounts(t, db, founderRev.StreamID)
			result, err := service.Handle(ctx, companyRev.StreamID, ModeOnline, time.Now().UTC(), profile.request.CanonicalPayload)
			if err != nil || result.Replay {
				t.Fatalf("initial Handle replay=%v err=%v", result.Replay, err)
			}
			loaded, err := store.LoadLatest(ctx, founderRev.StreamID)
			if err != nil {
				t.Fatal(err)
			}
			history, err := store.LoadFounderHistory(ctx, founderRev.StreamID)
			if err != nil || len(history.Entries) != 1 {
				t.Fatalf("recorded entries=%d err=%v", len(history.Entries), err)
			}
			entry := history.Entries[0]
			wantRevision := int64(1)
			if profile.row.Outcome == string(save.IntentApplied) {
				wantRevision = 2
			}
			if loaded.Revision.Number != wantRevision || loaded.Revision.ConstantsHash != profile.bundle.ConstantsHash || entry.ConstantsHash != profile.bundle.ConstantsHash || entry.Sequence != 1 || entry.IntentID != profile.request.IntentID || !bytes.Equal(entry.CanonicalPayload, profile.request.CanonicalPayload) || (entry.AppliedRevision != nil) != (wantRevision == 2) {
				t.Fatal("recorded command/revision/pin coordinates differ")
			}
			if entry.AppliedRevision != nil && *entry.AppliedRevision != 2 {
				t.Fatal("wrong applied revision")
			}
			if canonicalFixtureJSON(t, entry.Receipt) != canonicalFixtureJSON(t, result.Receipt) {
				t.Fatal("stored receipt differs")
			}
			assertReputationTaxonomyResult(t, profile, entry.ServerTSMS, result.Receipt, loaded.State, entry.Events)
			if verdict := VerifyFounderHistory(history, set); verdict != ReplayVerified {
				t.Fatalf("history verdict=%s", verdict)
			}
			afterCounts := reputationTaxonomyDBCounts(t, db, founderRev.StreamID)
			wantCounts := beforeCounts
			wantCounts.Revisions += int(wantRevision - 1)
			wantCounts.Logs++
			wantCounts.Intents++
			wantCounts.Events += len(entry.Events)
			wantCounts.Outbox += 1 + len(entry.Events)
			if afterCounts != wantCounts {
				t.Fatalf("persisted counts=%+v want=%+v", afterCounts, wantCounts)
			}
			var receiptRows int
			var outboxReceipt []byte
			if err := db.QueryRowContext(ctx, `SELECT count(*) FROM transport_player_outbox WHERE stream_id=$1 AND source_id=$2 AND message_kind='receipt'`, founderRev.StreamID, profile.request.IntentID).Scan(&receiptRows); err != nil || receiptRows != 1 {
				t.Fatalf("receipt outbox rows=%d err=%v", receiptRows, err)
			}
			if err := db.QueryRowContext(ctx, `SELECT payload::text FROM transport_player_outbox WHERE stream_id=$1 AND source_id=$2 AND message_kind='receipt'`, founderRev.StreamID, profile.request.IntentID).Scan(&outboxReceipt); err != nil || canonicalFixtureJSON(t, outboxReceipt) != canonicalFixtureJSON(t, result.Receipt) {
				t.Fatalf("outbox receipt differs err=%v", err)
			}
			retry, err := service.Handle(ctx, companyRev.StreamID, ModeOnline, time.Now().UTC(), profile.request.CanonicalPayload)
			if err != nil || !retry.Replay || !bytes.Equal(retry.Receipt, result.Receipt) {
				t.Fatalf("identical retry replay=%v err=%v", retry.Replay, err)
			}
			for _, category := range []string{"revision_conflict", "idempotency_conflict"} {
				payload := reputationShapeObject(t, profile.request.CanonicalPayload)
				detail := profile.request.IntentID
				if category == "revision_conflict" {
					payload["intent_id"] = reputationShapeJSON(t, fmt.Sprintf("01986666-9f02-7000-8000-%012d", index+1))
					payload["expected_revision"] = reputationShapeJSON(t, wantRevision+1)
					detail = "expected_revision"
				} else {
					alternate := "reputation.unlock.p99"
					if profile.request.ReputationNodeID == alternate {
						alternate = "reputation.unlock.p05"
					}
					payload["node_id"] = reputationShapeJSON(t, alternate)
				}
				conflict, err := service.Handle(ctx, companyRev.StreamID, ModeOnline, time.Now().UTC(), reputationShapeJSON(t, payload))
				if err != nil || conflict.Replay {
					t.Fatalf("%s replay=%v err=%v", category, conflict.Replay, err)
				}
				gotCategory, gotDetail := rejectionOf(t, conflict.Receipt)
				if gotCategory != category || gotDetail != detail {
					t.Fatalf("conflict=%s/%s want=%s/%s", gotCategory, gotDetail, category, detail)
				}
			}
			if counts := reputationTaxonomyDBCounts(t, db, founderRev.StreamID); counts != afterCounts {
				t.Fatalf("retry/conflicts wrote rows: %+v != %+v", counts, afterCounts)
			}
			final, err := store.LoadLatest(ctx, founderRev.StreamID)
			if err != nil || final.Revision.Number != wantRevision || !bytes.Equal(mustEncodeState(t, final.State), mustEncodeState(t, loaded.State)) {
				t.Fatalf("retry/conflicts changed Founder head err=%v", err)
			}
			companyAfter, err := store.LoadLatest(ctx, companyRev.StreamID)
			if err != nil || companyAfter.Revision.Number != 1 || !bytes.Equal(mustEncodeState(t, companyAfter.State), mustEncodeState(t, company)) {
				t.Fatalf("Founder purchase changed Company err=%v", err)
			}
		})
	}
}
