package production

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/garden"
	"cloud-clicker/server/save"
)

type gardenRefusalSnapshot struct {
	State                      gardenClockSnapshot
	ReceiptOutbox, EventOutbox int64
}

func (fixture gardenHarvestFixture) refusalSnapshot(t *testing.T) gardenRefusalSnapshot {
	t.Helper()
	snapshot := gardenRefusalSnapshot{State: fixture.clockSnapshot(t)}
	if err := fixture.db.QueryRowContext(fixture.ctx, `SELECT
		count(*) FILTER (WHERE message_kind='receipt'), count(*) FILTER (WHERE message_kind='event')
		FROM transport_player_outbox WHERE stream_id IN ($1,$2)`, fixture.founderStreamID, fixture.companyStreamID).Scan(&snapshot.ReceiptOutbox, &snapshot.EventOutbox); err != nil {
		t.Fatal(err)
	}
	return snapshot
}

// AC7: actually due growth, refusal persistence, recorded-row replay and a
// matched applied control. A no-work fixture cannot satisfy this instrument.
func TestGardenDueGrowthRefusalIntegration(t *testing.T) {
	arms := []struct {
		name, kind, fields, rejection string
		coldSalt, applied             bool
	}{
		{"occupied_plant", IntentGardenPlant, `"row":0,"col":0,"species_id":"strain_a"`, "not_eligible/plot_occupied", false, false},
		{"empty_uproot", IntentGardenUproot, `"row":5,"col":5`, "unknown_id/garden_plot", false, false},
		{"unchanged_substrate", IntentGardenSetSubstrate, `"substrate_id":"bare_metal"`, "not_eligible/substrate_unchanged", false, false},
		{"mixed_mature_empty_harvest", IntentGardenHarvest, `"plots":[{"row":0,"col":0},{"row":5,"col":5}]`, "unknown_id/garden_plot", false, false},
		{"null_salt_above_cap_unknown_substrate", IntentGardenSetSubstrate, `"substrate_id":"quantum"`, "unknown_id/garden_substrate", true, false},
		{"matched_applied_plant", IntentGardenPlant, `"row":5,"col":5,"species_id":"strain_b"`, "", false, true},
	}
	if len(arms) != 6 {
		t.Fatal("due-growth population changed")
	}
	for _, arm := range arms {
		t.Run(arm.name, func(t *testing.T) {
			fixture := newGardenHarvestFixtureWithState(t, func(state *save.State) {
				anchor := time.Now().Add(-750 * time.Second).UnixMilli()
				if arm.coldSalt {
					anchor = time.Now().Add(-25 * time.Hour).UnixMilli()
					state.ServerGarden.SaltHex = nil
				}
				state.ServerGarden.TickAnchorWallMS = &anchor
			})
			before := fixture.refusalSnapshot(t)
			if before.State.CompanyLogs != 0 {
				t.Fatal("instrument must start with Company genesis and no credit rows")
			}
			const intentID = "01986666-7f10-7000-8000-000000002001"
			body := []byte(fmt.Sprintf(`{"intent_id":%q,"kind":%q,"expected_revision":1,%s}`, intentID, arm.kind, arm.fields))
			result, err := fixture.service.Handle(fixture.ctx, fixture.companyStreamID, ModeOnline, time.UnixMilli(fixture.databaseMS(t)), body)
			if err != nil {
				t.Fatal(err)
			}
			after := fixture.refusalSnapshot(t)
			var payload, inputs []byte
			var envelope struct {
				Resolved struct {
					Advance garden.Advance `json:"advance"`
					Salt    string         `json:"garden_salt_hex"`
				} `json:"resolved"`
			}
			if err := fixture.db.QueryRowContext(fixture.ctx, `SELECT canonical_payload,replay_inputs FROM founder_log WHERE founder_stream_id=$1 AND intent_id=$2`, fixture.founderStreamID, intentID).Scan(&payload, &inputs); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(inputs, &envelope); err != nil {
				t.Fatal(err)
			}
			advance := envelope.Resolved.Advance
			if advance.TicksApplied < 2 || len(advance.Matured) == 0 {
				t.Fatalf("vacuous growth fixture: %+v", advance)
			}
			if arm.coldSalt && (len(envelope.Resolved.Salt) != 16 || advance.CatchupForfeitedMS <= 0 || advance.CatchupReasonKey == nil) {
				t.Fatalf("vacuous salt/cap fixture: %+v", envelope.Resolved)
			}
			if !arm.coldSalt && (advance.CatchupForfeitedMS != 0 || envelope.Resolved.Salt != "") {
				t.Fatalf("ordinary fixture drifted: %+v", envelope.Resolved)
			}
			// Replay the row actually emitted by Service, not a test-side resolved clone.
			working, err := save.RestoreState([]byte(before.State.FounderState), 25, fixture.bundle.Economy, economy.ScopeFounder, time.Time{})
			if err != nil {
				t.Fatal(err)
			}
			replayed, err := ApplyFounderLogged(working, payload, fixture.bundle, inputs)
			if err != nil {
				t.Fatal(err)
			}
			if arm.applied {
				if !strings.Contains(string(result.Receipt), `"outcome":"applied"`) || after.State.FounderRevision != before.State.FounderRevision+1 || after.State.CompanyState != before.State.CompanyState ||
					after.State.CompanyRevision != before.State.CompanyRevision || after.State.FounderState == before.State.FounderState || after.State.CompanyLogs != before.State.CompanyLogs ||
					after.State.Quota != before.State.Quota || after.State.Windows != before.State.Windows || after.ReceiptOutbox != before.ReceiptOutbox+1 || after.EventOutbox <= before.EventOutbox {
					t.Fatalf("matched applied control did not persist growth: receipt=%s before=%+v after=%+v", result.Receipt, before, after)
				}
				if replayed.Outcome != save.IntentApplied || string(mustEncodeState(t, working)) != after.State.FounderState {
					t.Fatal("applied row replay differs from committed full state")
				}
				plot, _, ok := working.ServerGarden.Plot(2, 2)
				if !ok || !plot.Mature() || plot.AgeTicks != 3 || working.ServerGarden.TickSeq < 2 {
					t.Fatalf("applied control failed to mature retained plant: %+v", plot)
				}
			} else {
				category, detail := rejectionOf(t, result.Receipt)
				if category+"/"+detail != arm.rejection || replayed.Outcome != save.IntentRejected {
					t.Fatalf("rejection=%s/%s want=%s", category, detail, arm.rejection)
				}
				want := before
				want.State.FounderLogs++
				want.State.IntentRecords++
				want.ReceiptOutbox++
				if after != want {
					t.Fatalf("refusal crossed persistence boundary: before=%+v after=%+v want=%+v", before, after, want)
				}
				t.Log("Store refusal boundary preserved full saved state/revisions and isolated the rejection receipt/log")
				if !bytes.Equal(mustEncodeState(t, working), []byte(before.State.FounderState)) {
					t.Fatal("recorded rejection failed in-memory rollback of due growth")
				}
			}
			retry, err := fixture.service.Handle(fixture.ctx, fixture.companyStreamID, ModeOnline, time.UnixMilli(fixture.databaseMS(t)), body)
			if err != nil || !retry.Replay || string(retry.Receipt) != string(result.Receipt) || fixture.refusalSnapshot(t) != after {
				t.Fatalf("retry altered receipt/state: replay=%v err=%v", retry.Replay, err)
			}
			conflictBody := []byte(fmt.Sprintf(`{"intent_id":%q,"kind":"garden_plant","expected_revision":1,"row":3,"col":4,"species_id":"strain_b"}`, intentID))
			conflict, err := fixture.service.Handle(fixture.ctx, fixture.companyStreamID, ModeOnline, time.UnixMilli(fixture.databaseMS(t)), conflictBody)
			if err != nil || !strings.Contains(string(conflict.Receipt), `"category":"idempotency_conflict"`) || fixture.refusalSnapshot(t) != after {
				t.Fatalf("changed-request retry wrote state: %s err=%v", conflict.Receipt, err)
			}
			followup := []byte(fmt.Sprintf(`{"intent_id":"01986666-7f10-7000-8000-000000002002","kind":"garden_uproot","expected_revision":%d,"row":0,"col":0}`, after.State.FounderRevision))
			continued, err := fixture.service.Handle(fixture.ctx, fixture.companyStreamID, ModeOnline, time.UnixMilli(fixture.databaseMS(t)), followup)
			if err != nil || !strings.Contains(string(continued.Receipt), `"outcome":"applied"`) {
				t.Fatalf("ordinary continuation after result failed: %s err=%v", continued.Receipt, err)
			}
			history, err := fixture.store.LoadFounderHistory(fixture.ctx, fixture.founderStreamID)
			if err != nil || VerifyFounderHistory(history, ReplayCatalogSet{fixture.bundle.ConstantsHash: fixture.bundle}) != ReplayVerified {
				t.Fatalf("stored Founder history failed: %v", err)
			}
			if fixture.refusalSnapshot(t).State.CompanyLogs != 0 {
				t.Fatal("Founder-only command wrote a Company log")
			}
			replayGardenRunLogCount(t, fixture, 0)
			t.Logf("ticks=%d matured=%d forfeited_ms=%d salt_initialized=%t applied=%t; direct replay, persistence, retries, outbox and continuation checked", advance.TicksApplied, len(advance.Matured), advance.CatchupForfeitedMS, envelope.Resolved.Salt != "", arm.applied)
		})
	}
}
