package production

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

// This is a diagnostic of the known G5 clock mismatch, NOT an acceptance
// assertion that skew-induced errors or future commits are valid. The fix must
// replace the broken-behavior expectations in its own reviewed range.
type gardenClockSnapshot struct {
	FounderState, CompanyState       string
	FounderRevision, CompanyRevision int64
	FounderLogs, CompanyLogs         int64
	Events, IntentRecords, Windows   int64
	Quota                            int64
}

func (fixture gardenHarvestFixture) clockSnapshot(t *testing.T) gardenClockSnapshot {
	t.Helper()
	founder, err := fixture.store.LoadLatest(fixture.ctx, fixture.founderStreamID)
	if err != nil {
		t.Fatal(err)
	}
	company, err := fixture.store.LoadLatest(fixture.ctx, fixture.companyStreamID)
	if err != nil {
		t.Fatal(err)
	}
	founderBytes, err := save.EncodeState(founder.State)
	if err != nil {
		t.Fatal(err)
	}
	companyBytes, err := save.EncodeState(company.State)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := gardenClockSnapshot{FounderState: string(founderBytes), CompanyState: string(companyBytes),
		FounderRevision: founder.Revision.Number, CompanyRevision: company.Revision.Number}
	err = fixture.db.QueryRowContext(fixture.ctx, `SELECT
		(SELECT count(*) FROM founder_log WHERE founder_stream_id=$1),
		(SELECT count(*) FROM run_log WHERE company_stream_id=$2),
		(SELECT count(*) FROM events WHERE stream_id IN ($1,$2)),
		(SELECT count(*) FROM intent_records WHERE stream_id IN ($1,$2)),
		(SELECT count(*) FROM minigame_faucet_window WHERE founder_id=$3),
		(SELECT COALESCE(sum(quota_used),0) FROM minigame_faucet_window WHERE founder_id=$3)`,
		fixture.founderStreamID, fixture.companyStreamID, fixture.founderID).Scan(&snapshot.FounderLogs,
		&snapshot.CompanyLogs, &snapshot.Events, &snapshot.IntentRecords, &snapshot.Windows, &snapshot.Quota)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func (fixture gardenHarvestFixture) databaseMS(t *testing.T) int64 {
	t.Helper()
	var now int64
	if err := fixture.db.QueryRowContext(fixture.ctx, `SELECT (extract(epoch FROM clock_timestamp())*1000)::bigint`).Scan(&now); err != nil {
		t.Fatal(err)
	}
	return now
}

func TestGardenHarvestClockIntegrationDiagnostic(t *testing.T) {
	for _, arm := range []struct {
		name     string
		offset   time.Duration
		immature bool
	}{
		{"mature_database_time", 0, false},
		{"mature_handler_ten_seconds_behind", -10 * time.Second, false},
		{"mature_handler_twenty_four_hours_ahead", 24 * time.Hour, false},
		{"immature_database_time", 0, true},
		{"immature_handler_ten_seconds_behind", -10 * time.Second, true},
	} {
		t.Run(arm.name, func(t *testing.T) {
			fixture := newGardenHarvestFixture(t)
			// Actual ordinary command establishes the Postgres-stamped head.
			plant := []byte(`{"intent_id":"01986666-7f10-7000-8000-000000001001","kind":"garden_plant","expected_revision":1,"row":3,"col":3,"species_id":"strain_a"}`)
			lower := fixture.databaseMS(t)
			planted, err := fixture.service.Handle(fixture.ctx, fixture.companyStreamID, ModeOnline, time.UnixMilli(lower), plant)
			upper := fixture.databaseMS(t)
			if err != nil || !strings.Contains(string(planted.Receipt), `"outcome":"applied"`) {
				t.Fatalf("ordinary clock control receipt=%s error=%v", planted.Receipt, err)
			}
			var stamp int64
			if err := fixture.db.QueryRowContext(fixture.ctx, `SELECT server_ts_ms FROM founder_log WHERE founder_stream_id=$1 ORDER BY seq DESC LIMIT 1`, fixture.founderStreamID).Scan(&stamp); err != nil || stamp < lower || stamp > upper {
				t.Fatalf("ordinary command outside database bounds: stamp=%d [%d,%d] error=%v", stamp, lower, upper, err)
			}
			before := fixture.clockSnapshot(t)
			company, err := fixture.store.LoadLatest(fixture.ctx, fixture.companyStreamID)
			if err != nil {
				t.Fatal(err)
			}
			lower = fixture.databaseMS(t)
			handlerMS := lower + arm.offset.Milliseconds()
			if handlerMS < company.State.EvaluatedThrough.UnixMilli() {
				t.Fatal("invalid instrument: handler clock falls behind Company attendance cursor")
			}
			target := `{"row":0,"col":0}`
			if arm.immature {
				target = `{"row":2,"col":2}`
			}
			body := []byte(fmt.Sprintf(`{"intent_id":"01986666-7f10-7000-8000-000000001002","kind":"garden_harvest","expected_revision":2,"plots":[%s]}`, target))
			result, handleErr := fixture.service.Handle(fixture.ctx, fixture.companyStreamID, ModeOnline, time.UnixMilli(handlerMS), body)
			upper = fixture.databaseMS(t)
			after := fixture.clockSnapshot(t)
			if arm.offset < 0 {
				if handleErr == nil || !strings.Contains(handleErr.Error(), "fiscal wall clock regressed") || before != after || len(result.Receipt) != 0 {
					t.Fatalf("lag diagnostic error=%v receipt=%s persistent snapshots equal=%v", handleErr, result.Receipt, before == after)
				}
				t.Logf("KNOWN DEFECT: handler=%d behind database [%d,%d], error=%v; full states/revisions/logs/events/records/window unchanged", handlerMS, lower, upper, handleErr)
				return
			}
			if handleErr != nil {
				t.Fatal(handleErr)
			}
			var envelope struct {
				EvaluatedAtMS int64 `json:"evaluated_at_ms"`
				Command       struct {
					ServerTSMS int64 `json:"server_ts_ms"`
				} `json:"command"`
				Resolved struct {
					ServerMS int64 `json:"server_ms"`
				} `json:"resolved"`
			}
			var inputs []byte
			if err := fixture.db.QueryRowContext(fixture.ctx, `SELECT server_ts_ms,replay_inputs FROM founder_log WHERE founder_stream_id=$1 ORDER BY seq DESC LIMIT 1`, fixture.founderStreamID).Scan(&stamp, &inputs); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(inputs, &envelope); err != nil || stamp != envelope.EvaluatedAtMS || stamp != envelope.Command.ServerTSMS || stamp != envelope.Resolved.ServerMS {
				t.Fatalf("log/envelope clock disagrees: stamp=%d inputs=%s error=%v", stamp, inputs, err)
			}
			if arm.immature {
				if !strings.Contains(string(result.Receipt), `"detail":"plant_not_mature"`) || stamp < lower || stamp > upper ||
					after.FounderState != before.FounderState || after.CompanyState != before.CompanyState ||
					after.FounderRevision != before.FounderRevision || after.CompanyRevision != before.CompanyRevision ||
					after.FounderLogs != before.FounderLogs+1 || after.CompanyLogs != before.CompanyLogs ||
					after.Events != before.Events || after.IntentRecords != before.IntentRecords+1 || after.Windows != before.Windows || after.Quota != before.Quota {
					t.Fatalf("immature database-time control receipt=%s stamp=%d [%d,%d] before=%+v after=%+v", result.Receipt, stamp, lower, upper, before, after)
				}
				t.Logf("immature control is Founder-only at database stamp %d", stamp)
				return
			}
			var receipt gardenHarvestAPIReceipt
			if err := json.Unmarshal(result.Receipt, &receipt); err != nil || receipt.Outcome != "applied" || receipt.Credited != "2e1" ||
				after.FounderRevision != before.FounderRevision+1 || after.CompanyRevision != before.CompanyRevision+1 || after.Quota != before.Quota+1 ||
				after.FounderLogs != before.FounderLogs+1 || after.CompanyLogs != before.CompanyLogs+1 || after.IntentRecords != before.IntentRecords+1 {
				t.Fatalf("mature arm did not commit expected credit: receipt=%s error=%v", result.Receipt, err)
			}
			// The contract-facing database-bound assertion was executed red
			// before retaining this diagnostic. A future commit is a finding,
			// not a valid wall-clock acceptance result.
			if arm.offset > 0 {
				if stamp <= upper || stamp != handlerMS || receipt.GardenAdvance.TicksApplied != 288 || receipt.GardenAdvance.CatchupForfeitedMS <= 0 {
					t.Fatalf("future clock diagnostic not reproduced: stamp=%d handler=%d receipt=%s", stamp, handlerMS, result.Receipt)
				}
				persisted, err := fixture.store.LoadLatest(fixture.ctx, fixture.founderStreamID)
				if err != nil {
					t.Fatal(err)
				}
				for _, coordinate := range [][2]int64{{2, 2}, {3, 3}} {
					plot, _, ok := persisted.State.ServerGarden.Plot(coordinate[0], coordinate[1])
					if !ok || !plot.Mature() || plot.AgeTicks != 3 {
						t.Fatalf("future clock did not mature retained starter: coordinate=%v plot=%+v", coordinate, plot)
					}
				}
				followup := []byte(`{"intent_id":"01986666-7f10-7000-8000-000000001003","kind":"garden_uproot","expected_revision":3,"row":3,"col":3}`)
				_, err = fixture.service.Handle(fixture.ctx, fixture.companyStreamID, ModeOnline, time.UnixMilli(fixture.databaseMS(t)), followup)
				if err == nil || !strings.Contains(err.Error(), "fiscal wall clock regressed") || fixture.clockSnapshot(t) != after {
					t.Fatalf("future-persisted cursor did not refuse next ordinary command without writes: %v", err)
				}
				t.Logf("KNOWN DEFECT: future stamp=%d database=[%d,%d], advance=%+v; next ordinary command error=%v without writes", stamp, lower, upper, receipt.GardenAdvance, err)
			} else {
				if stamp < lower || stamp > upper || stamp != handlerMS || receipt.GardenAdvance.TicksApplied != 0 {
					t.Fatalf("matched handler-time mature control differs: stamp=%d handler=%d advance=%+v", stamp, handlerMS, receipt.GardenAdvance)
				}
				t.Logf("mature control credits once, stamp=%d database=[%d,%d]", stamp, lower, upper)
			}
		})
	}
}
