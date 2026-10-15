package production

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

// This measures the existing event-bearing hook/wait boundary. It is not a
// new calculation API, event publisher, persisted settlement or AC6 repair.
func TestAxisCalculationEventOwnershipResearch(t *testing.T) {
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	commonsCatalog, err := commons.LoadCatalog(bundle.Artifacts["commons"])
	if err != nil {
		t.Fatal(err)
	}
	external := []multiplier.Contribution{{SourceID: guild.StockConsumptionSourceID,
		Slot: multiplier.SlotFaction, Target: "all", Factor: decimal.One}}
	weight := int64(500_000)
	for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
		t.Run(string(mode), func(t *testing.T) {
			initial := axisTimingState(t, bundle, start)
			initial.CompactMember, initial.CompactTithePPM = true, commonsCatalog.DefaultTithePPM
			initial.GuildTitheCarryPPM = 999_999
			if err := bundle.ValidateFoundationState(initial); err != nil {
				t.Fatal(err)
			}
			origin := mustEncodeState(t, initial)
			projected := activationRestore(t, bundle, origin, 19, start)
			contributions, err := assembleContributions(projected, bundle.Economy, external)
			if err != nil {
				t.Fatal(err)
			}
			at := start.Add(2001 * time.Millisecond)
			evaluation, err := Evaluate(projected, bundle.Economy, at, mode, contributions)
			if err != nil {
				t.Fatal(err)
			}
			hook := closedReplayAccrualHook(bundle, &weight)
			revision := save.Revision{Number: 1, ConstantsHash: bundle.ConstantsHash,
				StreamID: "01986666-1000-7000-8000-000000000001", OwnerID: "01986666-2000-7000-8000-000000000001"}
			// No invented intent or event envelope: returned events are private values.
			events, err := runAccrualHook(hook, "", projected, bundle.Economy, revision, evaluation, contributions)
			if err != nil {
				t.Fatal(err)
			}
			if len(events) != 2 || events[0].Kind != save.EventGuildTitheAccrued ||
				events[1].Kind != save.EventCompactSampled {
				t.Fatalf("closed hook event population/order=%+v", events)
			}
			for _, event := range events {
				if event.IntentID != "" || event.SchemaVersion != 1 || len(event.Payload) == 0 {
					t.Fatalf("projected event acquired an intent or lost its payload/schema: %+v", event)
				}
			}
			if evaluation.ProgressDeltaPPM <= 0 || !projected.LifetimeValue.Gt(initial.LifetimeValue) || len(projected.CompactSamples) == 0 ||
				!bytes.Equal(origin, mustEncodeState(t, initial)) {
				t.Fatal("projection lost complete hook effects or mutated the origin")
			}
			waiting := activationRestore(t, bundle, origin, 19, start)
			if _, err := SimulateAdvance(waiting, bundle.Economy, SimulationDependencies{Routes: bundle.Routes, Hook: hook},
				revision, mode, at, external, AblationMask{}); !errors.Is(err, ErrInvalidEngineState) {
				t.Fatalf("legacy event-bearing wait error=%v want ErrInvalidEngineState", err)
			}
			control := activationRestore(t, bundle, origin, 19, start)
			if _, err := SimulateAdvance(control, bundle.Economy, SimulationDependencies{Routes: bundle.Routes},
				revision, mode, at, external, AblationMask{}); err != nil || !control.EvaluatedThrough.Equal(at) {
				t.Fatalf("no-hook wait control failed: %v", err)
			}
			if control.LifetimeValue.Gt(initial.LifetimeValue) || len(control.CompactSamples) != 0 ||
				!bytes.Equal(origin, mustEncodeState(t, initial)) {
				t.Fatal("no-hook control gained hook effects or a scratch wait consumed the origin")
			}
			t.Logf("closed chain returns %s then %s; legacy wait refuses; no-hook control advances", events[0].Kind, events[1].Kind)
		})
	}
}
