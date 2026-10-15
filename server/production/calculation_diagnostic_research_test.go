package production

import (
	"bytes"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

func TestAxisCalculationDiagnosticSamplingResearch(t *testing.T) {
	bundle := activeContentBundle(t)
	input := ContentDynamicsActivePlayInput{FounderID: "018f6b7c-9abc-7def-8abc-000000000001",
		RunSeq: 1, EffectRowID: "active.production", TargetGeneratorID: "generator.beige_tower"}
	result, err := SimulateContentDynamicsActivePlay(bundle, input)
	if err != nil || result.DurationMS != 5000 || result.ClaimedOutput.String() != "3.5035e1" ||
		result.ControlOutput.String() != "5.005e0" || result.BonusOutput.String() != "3.003e1" {
		t.Fatalf("registered current diagnostic changed: %+v error=%v", result, err)
	}
	state, err := newContentDynamicsCompany(bundle.Economy, 18, input.RunSeq)
	if err != nil {
		t.Fatal(err)
	}
	state.GeneratorCounts[input.TargetGeneratorID] = 1
	spawn, err := bundle.Opportunities.Spawn(input.FounderID, input.RunSeq, 0, 0)
	if err != nil || spawn.EffectRowID != input.EffectRowID || spawn.SpawnedAttendedMS != result.SpawnedAttendedMS {
		t.Fatalf("natural first draw changed: %+v %v", spawn, err)
	}
	if _, err := initializeActivePlayState(state, bundle.Opportunities, input.FounderID); err != nil {
		t.Fatal(err)
	}
	schedule := activePlayScheduleEvidence{AttendedNowMS: spawn.SpawnedAttendedMS, BeforeSequence: 0,
		BeforeNextOpportunityMS: spawn.SpawnedAttendedMS, ExpiredBuffs: []activePlayExpiredBuff{}}
	if err := advanceActivePlaySchedule(state, bundle.Opportunities, input.FounderID, spawn.SpawnedAttendedMS, &schedule, true); err != nil {
		t.Fatal(err)
	}
	started := contentDynamicsEpoch.Add(time.Duration(spawn.SpawnedAttendedMS) * time.Millisecond)
	state.EvaluatedThrough, state.ManualTokenRefilledAt = started, started
	control, err := cloneReplayState(state, bundle.Economy)
	if err != nil {
		t.Fatal(err)
	}
	request := IntentRequest{IntentID: "018f6b7c-9abc-7def-8abc-0123456789ab", Kind: IntentClaimOpportunity,
		ExpectedRevision: 1, OpportunityID: spawn.OpportunityID}
	evidence := activePlayClaimEvidence{}
	decision, err := applyClaimOpportunity(request, state, bundle, save.Revision{Number: 1, OwnerID: input.FounderID},
		ModeOnline, started, nil, nil, spawn.SpawnedAttendedMS, nil, &evidence, true)
	if err != nil || decision.Outcome != save.IntentApplied || evidence.EffectRowID != input.EffectRowID {
		t.Fatalf("real claim did not produce the registered buff: %v %s", err, decision.Outcome)
	}
	for _, profile := range []struct {
		name   string
		origin *save.State
		frozen string
	}{
		{"claimed", state, "3.5035e1"},
		{"control", control, "5.005e0"},
	} {
		t.Run(profile.name, func(t *testing.T) {
			before := mustEncodeState(t, profile.origin)
			window, err := contentDynamicsAccrueWindow(bundle, profile.origin, input.TargetGeneratorID,
				spawn.SpawnedAttendedMS, 5000, false)
			if err != nil || window.String() != profile.frozen {
				t.Fatalf("start-sampled window=%s want=%s error=%v", window, profile.frozen, err)
			}
			endpoint, err := cloneReplayState(profile.origin, bundle.Economy)
			if err != nil {
				t.Fatal(err)
			}
			at := started.Add(5 * time.Second)
			resolved, err := resolveActivePlaySchedule(endpoint, bundle.Opportunities, bundle.Prestige, input.FounderID, at)
			if err != nil || resolved.AttendedNowMS != spawn.SpawnedAttendedMS+5000 {
				t.Fatalf("endpoint schedule: %v", err)
			}
			if _, err := applyActivePlaySchedule(endpoint, bundle.Opportunities, bundle.Prestige, input.FounderID, at, resolved); err != nil {
				t.Fatal(err)
			}
			active, err := activePlayContributions(endpoint, bundle.Opportunities, resolved.AttendedNowMS)
			if err != nil {
				t.Fatal(err)
			}
			contributions, err := assembleContributions(endpoint, bundle.Economy, active)
			if err != nil {
				t.Fatal(err)
			}
			prior, _ := endpoint.Ledger.Balance("company.cash")
			if _, err := Evaluate(endpoint, bundle.Economy, at, ModeOnline, contributions); err != nil {
				t.Fatal(err)
			}
			after, _ := endpoint.Ledger.Balance("company.cash")
			if after.Sub(prior).String() != "5.005e0" || !bytes.Equal(before, mustEncodeState(t, profile.origin)) {
				t.Fatal("endpoint sampling did not remove the expired bonus or mutated the diagnostic origin")
			}
			t.Logf("start-sampled=%s endpoint-sampled=%s; diagnostic origin unchanged", window, after.Sub(prior))
		})
	}
}
