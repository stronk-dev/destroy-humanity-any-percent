package gameui

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/production"
	"cloud-clicker/server/replaycatalog"
	"cloud-clicker/server/save"
)

// pinnedBundle loads the current epoch's exact artifact set through the same
// loader replay uses, so the arms are tested against the live catalogs.
func pinnedBundle(t *testing.T) production.CatalogBundle {
	t.Helper()
	raw, err := os.ReadFile("../../balance/epochs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	var declaration struct {
		Artifacts []struct {
			Name string `json:"name"`
			Path string `json:"path"`
		} `json:"artifacts"`
	}
	if err := json.Unmarshal(raw, &declaration); err != nil {
		t.Fatal(err)
	}
	artifacts := map[string][]byte{}
	for _, row := range declaration.Artifacts {
		data, err := os.ReadFile("../../" + row.Path)
		if err != nil {
			t.Fatal(err)
		}
		artifacts[row.Name] = data
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := replaycatalog.Load(hash, artifacts)
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

func featureStates(bundle production.CatalogBundle) (*save.State, *save.State) {
	company := &save.State{WireVersion: 18, MeterValues: map[string]int{}, AchievementsEarnedRun: map[string]bool{}}
	for _, id := range bundle.Meters.MeterIDs() {
		meter, _ := bundle.Meters.Meter(id)
		company.MeterValues[id] = meter.InitialValue
	}
	levels := map[string]int64{}
	for _, row := range bundle.Fiscal.GeneratorLevelRows() {
		levels[row.GeneratorID] = 0
	}
	founder := &save.State{WireVersion: 21, AchievementsEarnedLifetime: map[string]bool{}, FiscalGeneratorLevels: levels,
		FiscalUnlocks: map[string]bool{}, FiscalPeriodOpenedWallMS: 1_800_000_000_000, Soul: 100}
	return company, founder
}

func TestProjectFeaturesProjectsTheLiveArms(t *testing.T) {
	bundle := pinnedBundle(t)
	company, founder := featureStates(bundle)
	now := time.UnixMilli(1_800_000_000_000).UTC()
	features, err := projectFeatures(bundle, company, founder, now, false)
	if err != nil {
		t.Fatal(err)
	}
	if features.Meters == nil || len(features.Meters.Meters) != len(bundle.Meters.MeterIDs()) || features.Achievements == nil ||
		features.Fiscal == nil || features.Minigames == nil || features.ActivePlay != nil || features.Pets != nil {
		t.Fatalf("arms=%+v", features)
	}
	for _, row := range features.Meters.Meters {
		if row.Min != 0 || row.Max != 100 || row.BandID == "" {
			t.Fatalf("meter row %+v", row)
		}
	}
	pitch := features.Minigames.Rows
	if len(pitch) != 1 || pitch[0].MinigameID != "pitch" || pitch[0].Unlocked || pitch[0].HumanContentLocked {
		t.Fatalf("minigames=%+v", pitch)
	}
	facts := map[string]any{}
	for _, fact := range featureFacts(features) {
		facts[fact.FactID] = fact.Value
	}
	if facts["feature.fiscal"] != true || facts["feature.minigame.pitch"] != true || facts["feature.pets"] != false || facts["feature.active_play"] != false {
		t.Fatalf("facts=%v", facts)
	}
}

func TestProjectFeaturesNullsArmsBelowTheirActivatingVersions(t *testing.T) {
	bundle := pinnedBundle(t)
	company, founder := featureStates(bundle)
	founder.WireVersion = 18
	features, err := projectFeatures(bundle, company, founder, time.UnixMilli(1_800_000_000_000).UTC(), false)
	if err != nil {
		t.Fatal(err)
	}
	if features.Fiscal != nil || features.Minigames != nil || features.Meters == nil {
		t.Fatalf("pre-v19 Founder projected Fiscal or minigames: %+v", features)
	}
	withoutFiscal := bundle
	withoutFiscal.Fiscal = nil
	founder.WireVersion = 21
	if _, err := projectFeatures(withoutFiscal, company, founder, time.UnixMilli(1_800_000_000_000).UTC(), false); !errors.Is(err, ErrInvalidProjection) {
		t.Fatalf("a fiscal_unlock minigame without its Fiscal artifact projected: %v", err)
	}
}

// GS1-A1: the sweep preview is the auto-sweep the next Fiscal command would
// apply first. It must equal the credit a real Harvest on the same clone sees.
func TestFiscalArmSweepPreviewMatchesTheNextHarvest(t *testing.T) {
	bundle := pinnedBundle(t)
	_, founder := featureStates(bundle)
	founder.FiscalCredit = 4
	nowMS := founder.FiscalPeriodOpenedWallMS + 2*bundle.Fiscal.Clock.AutoMS + bundle.Fiscal.Clock.GuaranteedMS
	arm, err := projectFiscal(bundle.Fiscal, founder, nowMS)
	if err != nil {
		t.Fatal(err)
	}
	if arm.Credit != 4 || arm.SweepPreview.Periods != 2 || arm.SweepPreview.Credited != 2*bundle.Fiscal.Credit.CreditPerPeriod {
		t.Fatalf("sweep=%+v credit=%d", arm.SweepPreview, arm.Credit)
	}
	clone := founderFiscalState(founder)
	harvest, err := bundle.Fiscal.Harvest(&clone, "01985555-1111-7111-8111-111111111111", nowMS)
	if err != nil {
		t.Fatal(err)
	}
	if harvest.Sweep == nil || harvest.Sweep.CreditAfter != arm.SweepPreview.CreditAfter {
		t.Fatalf("harvest sweep=%+v preview=%+v", harvest.Sweep, arm.SweepPreview)
	}
	if founder.FiscalCredit != 4 || founder.FiscalPeriodSequence != 0 {
		t.Fatal("projection mutated the Founder")
	}
	for _, row := range arm.GeneratorLevels {
		if row.NextLevelCost == nil || row.Level != 0 {
			t.Fatalf("level row %+v", row)
		}
	}
}

func TestMinigamesArmRestatesTheCreateGates(t *testing.T) {
	bundle := pinnedBundle(t)
	_, founder := featureStates(bundle)
	definition, _ := bundle.Minigames.Definition("pitch")
	founder.FiscalUnlocks[definition.Unlock.UnlockID] = true
	arm, err := projectMinigames(bundle, founder, true)
	if err != nil {
		t.Fatal(err)
	}
	if !arm.Rows[0].Unlocked || !arm.Rows[0].ActiveSession {
		t.Fatalf("unlocked row=%+v", arm.Rows[0])
	}
	if definition.SoulGate == "human_hobby" {
		founder.Soul = 0
		arm, err = projectMinigames(bundle, founder, false)
		if err != nil || !arm.Rows[0].HumanContentLocked {
			t.Fatalf("near-zero Soul row=%+v err=%v", arm, err)
		}
	}
}

func TestAchievementsArmRejectsRunAndLifetimeOverlap(t *testing.T) {
	bundle := pinnedBundle(t)
	company, founder := featureStates(bundle)
	id := bundle.Achievements.Definitions[0].ID
	company.AchievementsEarnedRun[id] = true
	company.AchievementScoreRun = bundle.Achievements.Definitions[0].ScoreGrant
	arm, err := projectAchievements(bundle.Achievements, company, founder)
	if err != nil {
		t.Fatal(err)
	}
	earned := 0
	for _, row := range arm.Rows {
		if row.Earned != nil {
			earned++
			if *row.Earned != "run" || row.AchievementID != id {
				t.Fatalf("row=%+v", row)
			}
		}
	}
	if earned != 1 || arm.Score.Run != company.AchievementScoreRun {
		t.Fatalf("arm=%+v", arm)
	}
	founder.AchievementsEarnedLifetime[id] = true
	if _, err := projectAchievements(bundle.Achievements, company, founder); !errors.Is(err, ErrInvalidProjection) {
		t.Fatalf("overlap accepted: %v", err)
	}
}
