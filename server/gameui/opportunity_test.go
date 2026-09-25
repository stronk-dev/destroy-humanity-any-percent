package gameui

import (
	"testing"
	"time"

	"cloud-clicker/server/save"
)

func TestOpportunityArmProjectsLiveActivePlayOnly(t *testing.T) {
	bundle := pinnedBundle(t)
	company, founder := featureStates(bundle)
	target := "generator.beige_tower"
	company.GeneratorCounts = map[string]int64{target: 1}
	company.PendingOpportunity = &save.PendingOpportunity{OpportunityID: "01986666-0000-7000-8000-000000000001", SpawnedAttendedMS: 900,
		ExpiresAttendedMS: 5_000, EffectRowID: "active.building", SelectedGeneratorID: &target}
	company.ActiveBuffs = []save.ActiveBuff{
		{BuffInstanceID: "01986666-0000-7000-8000-00000000000b", EffectRowID: "active.production", ActivatedAttendedMS: 100, ExpiresAttendedMS: 4_000},
		{BuffInstanceID: "01986666-0000-7000-8000-00000000000a", EffectRowID: "active.click", ActivatedAttendedMS: 100, ExpiresAttendedMS: 1_000},
	}
	now := time.UnixMilli(1_800_000_000_000).UTC()
	features, err := projectFeatures(bundle, company, founder, now, false, 2_000)
	if err != nil {
		t.Fatal(err)
	}
	arm := features.Opportunity
	if arm == nil || arm.AttendedNowMS != 2_000 || arm.Pending == nil || arm.Pending.OpportunityID != company.PendingOpportunity.OpportunityID ||
		arm.Pending.SelectedGeneratorID == nil || *arm.Pending.SelectedGeneratorID != target {
		t.Fatalf("pending opportunity was not projected: %+v", arm)
	}
	// The click buff expired at 1000 <= 2000: it is omitted; the live one stays.
	if len(arm.Buffs) != 1 || arm.Buffs[0].EffectRowID != "active.production" {
		t.Fatalf("expired buffs must be omitted: %+v", arm.Buffs)
	}
	if arm.Combo.Cap != "1e4" || arm.Combo.ReasonKey != "cap.active_combo" || arm.Combo.Saturated {
		t.Fatalf("unexpected combo row %+v", arm.Combo)
	}
	facts := featureFacts(features)
	for _, fact := range facts {
		if fact.FactID == "feature.active_play" && fact.Value != true {
			t.Fatal("feature.active_play must follow the opportunity arm")
		}
	}

	// At the expiry coordinate the opportunity is no longer claimable: omit it.
	features, err = projectFeatures(bundle, company, founder, now, false, 5_000)
	if err != nil || features.Opportunity.Pending != nil {
		t.Fatalf("an expired opportunity must not be projected: %+v %v", features.Opportunity, err)
	}
}

func TestOpportunityArmReportsComboSaturationFromTheKernelClamp(t *testing.T) {
	bundle := pinnedBundle(t)
	company, founder := featureStates(bundle)
	target := "generator.beige_tower"
	// building_special at 100000 ppm per owned: 1 + 0.1 * 200000 = 20001 > combo cap 1e4.
	company.GeneratorCounts = map[string]int64{target: 200_000}
	company.ActiveBuffs = []save.ActiveBuff{{BuffInstanceID: "01986666-0000-7000-8000-00000000000c", EffectRowID: "active.building",
		SelectedTarget: &target, ActivatedAttendedMS: 0, ExpiresAttendedMS: 10_000}}
	features, err := projectFeatures(bundle, company, founder, time.UnixMilli(1_800_000_000_000).UTC(), false, 1_000)
	if err != nil {
		t.Fatal(err)
	}
	if !features.Opportunity.Combo.Saturated {
		t.Fatal("a buff product above the combo cap must project saturated")
	}
}

func TestOpportunityArmIsNullBelowActivePlayVersion(t *testing.T) {
	bundle := pinnedBundle(t)
	company, founder := featureStates(bundle)
	company.WireVersion = 17
	features, err := projectFeatures(bundle, company, founder, time.UnixMilli(1_800_000_000_000).UTC(), false, 0)
	if err != nil {
		t.Fatal(err)
	}
	if features.Opportunity != nil {
		t.Fatal("a pre-v18 Company has no active play to project")
	}
}
