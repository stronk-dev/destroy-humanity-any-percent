package production

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/kernel"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

var worldProducerOutput = flag.String("world-producer-output", "", "manual World producer research JSON output (requires Postgres)")

type worldProducerObservation struct {
	Name              string          `json:"name"`
	Mode              EvaluationMode  `json:"mode"`
	ElapsedMS         int64           `json:"elapsed_ms"`
	ReferencePassive  string          `json:"reference_passive_delta"`
	ReferenceProgress int64           `json:"reference_progress_delta_ppm"`
	PersistedDelta    string          `json:"persisted_balance_delta"`
	Receipt           json.RawMessage `json:"receipt_summary"`
	RevisionBefore    int64           `json:"revision_before"`
	RevisionAfter     int64           `json:"revision_after"`
	CursorAdvanced    bool            `json:"cursor_advanced"`
	Counts            [5]int64        `json:"revision_event_intent_log_outbox_counts"`
	AppliedLogRows    int64           `json:"applied_log_rows"`
	Events            json.RawMessage `json:"events"`
	ErrorClass        string          `json:"error_class"`
	RetryUnchanged    bool            `json:"retry_unchanged"`
	RetryRedelivery   bool            `json:"retry_event_redelivery"`
}

type worldResearchProjector struct{ batches [][]save.EventRecord }

func (observer *worldResearchProjector) Project(_ context.Context, events []save.EventRecord) error {
	if len(events) > 0 {
		observer.batches = append(observer.batches, append([]save.EventRecord(nil), events...))
	}
	return nil
}

// This manual instrument records existing behavior. No World producer, event
// kind, balance row or acceptance criterion is introduced by the test.
func TestWorldOutputProducerIntegrationResearch(t *testing.T) {
	if *worldProducerOutput == "" {
		t.Skip("manual research requires -world-producer-output")
	}
	if _, err := os.Stat(*worldProducerOutput); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("research output must be a new path: %v", err)
	}
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Fatal("selected World research requires actual Postgres")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, os.Getenv("TEST_DATABASE_URL"))
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
	bundle := epoch5TestBundle(t)
	seedProductionEpoch(t, db, bundle.ConstantsHash, bundle.Artifacts)
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{bundle.ConstantsHash: bundle.Economy},
		routes:   map[string]*routes.Catalog{bundle.ConstantsHash: bundle.Routes},
		prestige: map[string]*prestigecore.Policy{bundle.ConstantsHash: bundle.Prestige},
		factions: map[string]*faction.Catalog{bundle.ConstantsHash: bundle.Faction}}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	observer := &worldResearchProjector{}
	service, err := NewService(store, resolver, nil, nil, nil, WithEventProjector(observer), WithProgressionRuntime(resolver),
		WithCurrentConstantsHash(bundle.ConstantsHash), WithReplayCatalogs(ReplayCatalogSet{bundle.ConstantsHash: bundle}),
		WithGuildSettlements(emptyGuildSettlements{}))
	if err != nil {
		t.Fatal(err)
	}
	var observations []worldProducerObservation
	for index, row := range []struct {
		name, cash, passive, delta string
		mode                       EvaluationMode
		elapsed                    time.Duration
		manual, reject, rollback   bool
	}{
		{"online_purchase", "1e2", "1e0", "-1.03e1", ModeOnline, time.Second, false, false, false},
		{"same_time_manual", "1e2", "0", "1e0", ModeOnline, 0, true, false, false},
		{"online_manual", "1e2", "1e0", "2e0", ModeOnline, time.Second, true, false, false},
		{"offline_purchase", "1e2", "9e-1", "-1.04e1", ModeOffline, time.Second, false, false, false},
		{"saturated_progress", "1e3", "1e0", "-1.03e1", ModeOnline, time.Second, false, false, false},
		{"unaffordable_purchase", "1e2", "1e0", "0", ModeOnline, time.Second, false, true, false},
		{"revision_write_failure", "1e2", "1e0", "0", ModeOnline, time.Second, false, false, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			observer.batches = nil
			cursor := time.Date(2026, 11, 5, 8, 0, 0, 0, time.UTC)
			founderID := fmt.Sprintf("01986666-f100-4000-8000-%012d", index+1)
			state := replayFixtureState(t, bundle.Economy, cursor)
			state.GeneratorCounts["generator.beige_tower"], state.GeneratorPurchasedTotal = 1, 1
			setCash(t, state, row.cash)
			created, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeCompany},
				bundle.ConstantsHash, state, save.WriteContext{Cause: "world-producer.research"})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := store.PinRunToCurrentEpoch(ctx, created.StreamID, founderID, 1, bundle.ConstantsHash); err != nil {
				t.Fatal(err)
			}
			before, err := store.LoadLatest(ctx, created.StreamID)
			if err != nil {
				t.Fatal(err)
			}
			reference, err := cloneReplayState(before.State, bundle.Economy)
			if err != nil {
				t.Fatal(err)
			}
			contributions, err := assembleContributions(reference, bundle.Economy, nil)
			if err != nil {
				t.Fatal(err)
			}
			evaluated, err := Evaluate(reference, bundle.Economy, cursor.Add(row.elapsed), row.mode, contributions)
			if err != nil {
				t.Fatal(err)
			}
			passive := decimal.Zero
			for _, change := range evaluated.Receipt.Changes {
				if change.ResourceID != "company.cash" {
					t.Fatalf("unexpected reference resource %s", change.ResourceID)
				}
				passive = passive.Add(decimal.FromString(change.Delta))
			}
			if passive.String() != row.passive || row.name == "saturated_progress" && evaluated.ProgressDeltaPPM != 0 {
				t.Fatalf("reference control: passive=%s progress=%d", passive.String(), evaluated.ProgressDeltaPPM)
			}
			countsBefore := worldProducerCounts(t, ctx, db, created.StreamID)
			if countsBefore != [5]int64{1, 0, 0, 0, 0} {
				t.Fatalf("unexpected genesis records: %v", countsBefore)
			}
			intentID := fmt.Sprintf("01986666-f101-7000-8000-%012d", index+1)
			body := fmt.Sprintf(`{"intent_id":%q,"kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":1}}`, intentID)
			if row.manual {
				body = fmt.Sprintf(`{"intent_id":%q,"kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`, intentID)
			} else if row.reject {
				body = strings.Replace(body, `"value":1`, `"value":1000`, 1)
			}
			if row.rollback {
				// This generated stream UUID is embedded only in the disposable test
				// DB's trigger. It is not input from a player or external source.
				_, err := db.ExecContext(ctx, fmt.Sprintf(`CREATE FUNCTION world_research_revision_fail() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
					IF NEW.stream_id='%s'::uuid THEN RAISE EXCEPTION 'world research injected revision write failure'; END IF;
					RETURN NEW; END $$`, created.StreamID))
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if _, err := db.ExecContext(ctx, `DROP TRIGGER IF EXISTS world_research_revision_fail ON save_revisions; DROP FUNCTION world_research_revision_fail()`); err != nil {
						t.Error(err)
					}
				})
				if _, err := db.ExecContext(ctx, `CREATE TRIGGER world_research_revision_fail BEFORE INSERT ON save_revisions FOR EACH ROW EXECUTE FUNCTION world_research_revision_fail()`); err != nil {
					t.Fatal(err)
				}
			}
			result, handleErr := service.Handle(ctx, created.StreamID, row.mode, cursor.Add(row.elapsed), []byte(body))
			errorClass := "none"
			if row.rollback {
				if handleErr == nil || !strings.Contains(handleErr.Error(), "world research injected revision write failure") {
					t.Fatalf("injected SQL failure not reached: %v", handleErr)
				}
				errorClass = "injected_revision_write_failure"
			} else if handleErr != nil {
				t.Fatal(handleErr)
			}
			after, err := store.LoadLatest(ctx, created.StreamID)
			if err != nil {
				t.Fatal(err)
			}
			balanceBefore, _ := before.State.Ledger.Balance("company.cash")
			balanceAfter, _ := after.State.Ledger.Balance("company.cash")
			delta := balanceAfter.Sub(balanceBefore).Quantize(decimal.CanonicalSignificantDigits).String()
			counts := worldProducerCounts(t, ctx, db, created.StreamID)
			wantCounts, wantRevision := [5]int64{2, 1, 1, 1, 2}, int64(2)
			wantOutcome := "applied"
			if row.manual {
				wantCounts[1], wantCounts[4] = 0, 1 // Manual batches deliberately emit no generic event.
			}
			if row.reject {
				wantCounts, wantRevision, wantOutcome = [5]int64{1, 0, 1, 1, 1}, 1, "rejected"
			}
			if row.rollback {
				wantCounts, wantRevision = countsBefore, 1
			} else {
				var receipt struct {
					Outcome string `json:"outcome"`
				}
				if json.Unmarshal(result.Receipt, &receipt) != nil || receipt.Outcome != wantOutcome {
					t.Fatalf("outcome control: %s", result.Receipt)
				}
			}
			if delta != row.delta || counts != wantCounts || after.Revision.Number != wantRevision {
				t.Fatalf("persisted control: delta=%s counts=%v revision=%d", delta, counts, after.Revision.Number)
			}
			if (row.reject || row.rollback) && !reflect.DeepEqual(after, before) {
				t.Fatal("refused/rolled-back operation changed saved state")
			}
			var events []byte
			var applied int64
			if err := db.QueryRowContext(ctx, `SELECT COALESCE(jsonb_agg(jsonb_build_object('kind',kind,'payload',payload) ORDER BY event_id),'[]'::jsonb) FROM events WHERE stream_id=$1`, created.StreamID).Scan(&events); err != nil {
				t.Fatal(err)
			}
			if err := db.QueryRowContext(ctx, `SELECT count(*) FROM run_log WHERE company_stream_id=$1 AND applied_revision IS NOT NULL`, created.StreamID).Scan(&applied); err != nil || applied != wantRevision-1 {
				t.Fatalf("applied log control: rows=%d err=%v", applied, err)
			}
			retryUnchanged, retryRedelivery := false, false
			if row.name == "online_purchase" {
				retry, err := service.Handle(ctx, created.StreamID, row.mode, cursor.Add(time.Hour), []byte(body))
				latest, loadErr := store.LoadLatest(ctx, created.StreamID)
				if err != nil || loadErr != nil || !retry.Replay || !bytes.Equal(retry.Receipt, result.Receipt) ||
					!reflect.DeepEqual(latest, after) || worldProducerCounts(t, ctx, db, created.StreamID) != counts {
					t.Fatalf("retry changed durable result: err=%v load=%v replay=%v", err, loadErr, retry.Replay)
				}
				if len(observer.batches) != 2 || len(observer.batches[0]) != 1 ||
					!reflect.DeepEqual(observer.batches[0], observer.batches[1]) {
					t.Fatalf("retry event delivery control: %+v", observer.batches)
				}
				retryUnchanged, retryRedelivery = true, true
			}
			var receiptSummary json.RawMessage
			if !row.rollback {
				var receipt struct {
					Outcome string          `json:"outcome"`
					Receipt json.RawMessage `json:"receipt,omitempty"`
				}
				if err := json.Unmarshal(result.Receipt, &receipt); err != nil {
					t.Fatal(err)
				}
				receiptSummary, err = json.Marshal(receipt)
				if err != nil {
					t.Fatal(err)
				}
			}
			observations = append(observations, worldProducerObservation{Name: row.name, Mode: row.mode, ElapsedMS: row.elapsed.Milliseconds(),
				ReferencePassive: passive.String(), ReferenceProgress: evaluated.ProgressDeltaPPM, PersistedDelta: delta, Receipt: receiptSummary,
				RevisionBefore: before.Revision.Number, RevisionAfter: after.Revision.Number, CursorAdvanced: after.State.EvaluatedThrough.After(before.State.EvaluatedThrough),
				Counts: counts, AppliedLogRows: applied, Events: events, ErrorClass: errorClass, RetryUnchanged: retryUnchanged, RetryRedelivery: retryRedelivery})
			t.Logf("passive reference=%s persisted=%s progress=%d counts=%v applied_log=%d", passive.String(), delta, evaluated.ProgressDeltaPPM, counts, applied)
		})
	}
	if t.Failed() || len(observations) != 7 {
		t.Fatal("incomplete/invalid World producer observation; no artifact written")
	}
	sources := map[string]string{}
	for _, name := range []string{"server/production/engine.go", "server/production/intents.go", "server/production/replay.go", "server/save/intent.go", "server/accrualhook/contract.go", "server/guild/hook.go", "server/gameserver/world_source.go", "server/production/world_producer_research_test.go"} {
		data, err := os.ReadFile("../../" + name)
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(data)
		sources[name] = hex.EncodeToString(hash[:])
	}
	var postgresVersion string
	if err := db.QueryRowContext(ctx, `SHOW server_version`).Scan(&postgresVersion); err != nil {
		t.Fatal(err)
	}
	artifact := map[string]any{"schema_version": 1, "kernel_version": kernel.Version, "fixture_constants_hash": bundle.ConstantsHash,
		"runtime":       map[string]string{"go": runtime.Version(), "os": runtime.GOOS, "arch": runtime.GOARCH, "postgres": postgresVersion},
		"source_sha256": sources, "observations": observations}
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(*worldProducerOutput, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

func worldProducerCounts(t *testing.T, ctx context.Context, db *sql.DB, streamID string) [5]int64 {
	t.Helper()
	var counts [5]int64
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT count(*) FROM save_revisions WHERE stream_id=$1),
		(SELECT count(*) FROM events WHERE stream_id=$1),
		(SELECT count(*) FROM intent_records WHERE stream_id=$1),
		(SELECT count(*) FROM run_log WHERE company_stream_id=$1),
		(SELECT count(*) FROM transport_player_outbox WHERE stream_id=$1)`, streamID).
		Scan(&counts[0], &counts[1], &counts[2], &counts[3], &counts[4]); err != nil {
		t.Fatal(err)
	}
	return counts
}
