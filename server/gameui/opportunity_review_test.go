package gameui

import (
	"reflect"
	"testing"
	"time"

	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

// These are isolated projection inputs, not codec-admitted saves or player
// workflow evidence. Construct the expected input independently so an in-place
// mutation cannot also mutate the comparison oracle through a shared pointer.
func opportunityReviewState() *save.State {
	pendingTarget, buffTarget := "generator.beige_tower", "generator.beige_tower"
	return &save.State{
		WireVersion: 18, GeneratorCounts: map[string]int64{buffTarget: 1},
		PendingOpportunity: &save.PendingOpportunity{
			OpportunityID:     "01986666-0000-7000-8000-000000000001",
			SpawnedAttendedMS: 900, ExpiresAttendedMS: 5_000,
			EffectRowID: "active.building", SelectedGeneratorID: &pendingTarget,
		},
		ActiveBuffs: []save.ActiveBuff{
			{BuffInstanceID: "01986666-0000-7000-8000-00000000000b", EffectRowID: "active.production", ActivatedAttendedMS: 100, ExpiresAttendedMS: 4_000},
			{BuffInstanceID: "01986666-0000-7000-8000-00000000000a", EffectRowID: "active.building", SelectedTarget: &buffTarget, ActivatedAttendedMS: 100, ExpiresAttendedMS: 3_000},
		},
	}
}

func TestOpportunityReviewRepeatedProjectionIsReadOnlyAndSorted(t *testing.T) {
	bundle := pinnedBundle(t)
	state, expected := opportunityReviewState(), opportunityReviewState()
	first, err := projectOpportunity(bundle, state, 2_000)
	if err != nil {
		t.Fatal(err)
	}
	second, err := projectOpportunity(bundle, state, 2_000)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state, expected) {
		t.Fatal("projection changed its input state")
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("repeated read-only projection changed its output")
	}
	if len(first.Buffs) != 2 || first.Buffs[0].BuffInstanceID != expected.ActiveBuffs[1].BuffInstanceID ||
		first.Buffs[1].BuffInstanceID != expected.ActiveBuffs[0].BuffInstanceID {
		t.Fatalf("live buff IDs were not byte sorted: %+v", first.Buffs)
	}
}

func TestOpportunityReviewOutputTargetsDoNotAliasState(t *testing.T) {
	bundle := pinnedBundle(t)
	state, expected := opportunityReviewState(), opportunityReviewState()
	arm, err := projectOpportunity(bundle, state, 2_000)
	if err != nil {
		t.Fatal(err)
	}
	if arm.Pending == nil || arm.Pending.SelectedGeneratorID == nil || len(arm.Buffs) != 2 || arm.Buffs[0].SelectedTarget == nil {
		t.Fatal("fixture did not exercise both target pointer paths")
	}
	*arm.Pending.SelectedGeneratorID = "generator.changed_pending"
	*arm.Buffs[0].SelectedTarget = "generator.changed_buff"
	arm.Buffs[1].EffectRowID = "active.changed"
	if !reflect.DeepEqual(state, expected) {
		t.Fatal("projected target or buff aliases the input state")
	}
}

func TestOpportunityReviewActivationFactMatchesOptionalArm(t *testing.T) {
	for _, tc := range []struct {
		name    string
		version int
		catalog bool
		active  bool
	}{
		{"pre_v18", 17, true, false},
		{"v18", 18, true, true},
		{"later_version", 19, true, true},
		{"missing_artifact", 18, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bundle := pinnedBundle(t)
			company, founder := featureStates(bundle)
			company.WireVersion = tc.version
			if !tc.catalog {
				bundle.Opportunities = nil
			}
			features, err := projectFeatures(bundle, company, founder, time.UnixMilli(1_800_000_000_000).UTC(), false, 0)
			if err != nil {
				t.Fatal(err)
			}
			if (features.Opportunity != nil) != tc.active || features.ActivePlay != nil {
				t.Fatalf("unexpected current optional-arm activation: %+v", features)
			}
			matches := 0
			for _, fact := range featureFacts(features) {
				if fact.FactID == "feature.active_play" {
					matches++
					if fact.Value != tc.active {
						t.Fatalf("activation fact=%v, want %v", fact.Value, tc.active)
					}
				}
			}
			if matches != 1 {
				t.Fatalf("activation fact count=%d, want 1", matches)
			}
		})
	}
}

func TestOpportunityReviewKernelWrapperRejectsInvalidInputs(t *testing.T) {
	bundle := pinnedBundle(t)
	state := opportunityReviewState()
	if _, err := production.ProjectActiveCombo(nil, bundle.Opportunities, 0); err == nil {
		t.Fatal("nil state accepted")
	}
	if _, err := production.ProjectActiveCombo(state, nil, 0); err == nil {
		t.Fatal("nil catalog accepted")
	}
	if _, err := production.ProjectActiveCombo(state, bundle.Opportunities, -1); err == nil {
		t.Fatal("negative attended clock accepted")
	}
}
