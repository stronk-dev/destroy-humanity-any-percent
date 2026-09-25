package gameui

import (
	"sort"
	"time"

	"cloud-clicker/server/achievements"
	"cloud-clicker/server/fiscal"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

// Garage Player Surfaces GS0.1: the v4 `features` object. Every arm is a
// persisted field or a read-only kernel derivation on a discarded clone; an
// arm is null when the pinned bundle lacks its artifact or the save predates
// the version that activates it.
const snapshotSchemaVersion = 4

type featureRows struct {
	ActivePlay   *struct{}        `json:"active_play"`
	Achievements *achievementsArm `json:"achievements"`
	Fiscal       *fiscalArm       `json:"fiscal"`
	Meters       *metersArm       `json:"meters"`
	Minigames    *minigamesArm    `json:"minigames"`
	Pets         *struct{}        `json:"pets"`
	Reputation   *reputationArm   `json:"reputation"`
}

type achievementsArm struct {
	Rows  []achievementRow `json:"rows"`
	Score achievementScore `json:"score"`
}

type achievementScore struct {
	Lifetime int64 `json:"lifetime"`
	Run      int64 `json:"run"`
}

type achievementRow struct {
	AchievementID  string  `json:"achievement_id"`
	ConditionScope string  `json:"condition_scope"`
	CopyKey        string  `json:"copy_key"`
	Earned         *string `json:"earned"`
	ProofKind      string  `json:"proof_kind"`
	ScoreGrant     int64   `json:"score_grant"`
}

type metersArm struct {
	Meters []meterRow `json:"meters"`
}

type meterRow struct {
	BandID  string      `json:"band_id"`
	Bands   []meterBand `json:"bands"`
	Max     int         `json:"max"`
	MeterID string      `json:"meter_id"`
	Min     int         `json:"min"`
	Value   int         `json:"value"`
}

type meterBand struct {
	BandID     string `json:"band_id"`
	FloorValue int    `json:"floor_value"`
}

type fiscalArm struct {
	Credit          int64              `json:"credit"`
	CreditCap       intCap             `json:"credit_cap"`
	CreditPerPeriod int64              `json:"credit_per_period"`
	GeneratorLevels []fiscalLevelRow   `json:"generator_levels"`
	Hoard           fiscalHoard        `json:"hoard"`
	Period          fiscalPeriod       `json:"period"`
	SweepPreview    fiscalSweepPreview `json:"sweep_preview"`
	Unlocks         []fiscalUnlockRow  `json:"unlocks"`
}

type intCap struct {
	Amount    int64  `json:"amount"`
	ReasonKey string `json:"reason_key"`
}

type fiscalPeriod struct {
	AutoMS          int64 `json:"auto_ms"`
	EarlyMS         int64 `json:"early_ms"`
	EarlySuccessPPM int64 `json:"early_success_ppm"`
	GuaranteedMS    int64 `json:"guaranteed_ms"`
	OpenedWallMS    int64 `json:"opened_wall_ms"`
	Seq             int64 `json:"seq"`
}

type fiscalSweepPreview struct {
	CreditAfter int64 `json:"credit_after"`
	Credited    int64 `json:"credited"`
	Periods     int64 `json:"periods"`
	Saturated   bool  `json:"saturated"`
}

type fiscalHoard struct {
	CapCredits int64  `json:"cap_credits"`
	PreviewPPM int64  `json:"preview_ppm"`
	ReasonNote string `json:"reason_note"`
}

type fiscalLevelRow struct {
	GeneratorID   string `json:"generator_id"`
	Level         int64  `json:"level"`
	LevelCap      intCap `json:"level_cap"`
	NextLevelCost *int64 `json:"next_level_cost"`
	PPMPerLevel   int64  `json:"ppm_per_level"`
}

type fiscalUnlockRow struct {
	Cost     int64  `json:"cost"`
	Owned    bool   `json:"owned"`
	UnlockID string `json:"unlock_id"`
}

type minigamesArm struct {
	Rows []minigameAvailability `json:"rows"`
}

type minigameAvailability struct {
	ActiveSession      bool   `json:"active_session"`
	HumanContentLocked bool   `json:"human_content_locked"`
	MinigameID         string `json:"minigame_id"`
	Unlocked           bool   `json:"unlocked"`
}

func projectFeatures(bundle production.CatalogBundle, state, founder *save.State, now time.Time, minigameActive bool) (featureRows, error) {
	var result featureRows
	if bundle.Achievements != nil && save.VersionForState(state) >= 16 {
		arm, err := projectAchievements(bundle.Achievements, state, founder)
		if err != nil {
			return featureRows{}, err
		}
		result.Achievements = arm
	}
	if bundle.Meters != nil && save.VersionForState(state) >= 16 {
		arm, err := projectMeters(bundle.Meters, state)
		if err != nil {
			return featureRows{}, err
		}
		result.Meters = arm
	}
	if bundle.Fiscal != nil && save.VersionForState(founder) >= 19 {
		arm, err := projectFiscal(bundle.Fiscal, founder, save.CanonicalServerTime(now).UnixMilli())
		if err != nil {
			return featureRows{}, err
		}
		result.Fiscal = arm
	}
	if bundle.MinigameAPI != nil && bundle.Minigames != nil && save.VersionForState(founder) >= 21 {
		arm, err := projectMinigames(bundle, founder, minigameActive)
		if err != nil {
			return featureRows{}, err
		}
		result.Minigames = arm
	}
	return result, nil
}

func projectAchievements(catalog *achievements.Catalog, state, founder *save.State) (*achievementsArm, error) {
	rows := make([]achievementRow, 0, len(catalog.Definitions))
	for _, definition := range catalog.Definitions {
		var earned *string
		run, lifetime := state.AchievementsEarnedRun[definition.ID], founder.AchievementsEarnedLifetime[definition.ID]
		if run && lifetime {
			return nil, ErrInvalidProjection
		}
		if run {
			value := "run"
			earned = &value
		} else if lifetime {
			value := "lifetime"
			earned = &value
		}
		rows = append(rows, achievementRow{AchievementID: definition.ID, ConditionScope: string(definition.ConditionScope),
			CopyKey: definition.CopyKey, Earned: earned, ProofKind: string(definition.Proof.Kind), ScoreGrant: definition.ScoreGrant})
	}
	sort.Slice(rows, func(left, right int) bool { return rows[left].AchievementID < rows[right].AchievementID })
	return &achievementsArm{Rows: rows, Score: achievementScore{Lifetime: founder.AchievementScoreLifetime, Run: state.AchievementScoreRun}}, nil
}

func projectMeters(catalog *meters.Catalog, state *save.State) (*metersArm, error) {
	rows := []meterRow{}
	for _, id := range catalog.MeterIDs() {
		meter, ok := catalog.Meter(id)
		value, present := state.MeterValues[id]
		if !ok || !present || value < meters.MinimumValue || value > meters.MaximumValue {
			return nil, ErrInvalidProjection
		}
		bands := make([]meterBand, 0, len(meter.Bands))
		for _, band := range meter.Bands {
			bands = append(bands, meterBand{BandID: band.ID, FloorValue: band.FloorValue})
		}
		rows = append(rows, meterRow{BandID: meters.BandFor(meter, value), Bands: bands, Max: meters.MaximumValue,
			MeterID: id, Min: meters.MinimumValue, Value: value})
	}
	sort.Slice(rows, func(left, right int) bool { return rows[left].MeterID < rows[right].MeterID })
	return &metersArm{Meters: rows}, nil
}

func founderFiscalState(founder *save.State) fiscal.State {
	levels := make(map[string]int64, len(founder.FiscalGeneratorLevels))
	for id, level := range founder.FiscalGeneratorLevels {
		levels[id] = level
	}
	unlocks := make([]string, 0, len(founder.FiscalUnlocks))
	for id, owned := range founder.FiscalUnlocks {
		if owned {
			unlocks = append(unlocks, id)
		}
	}
	sort.Strings(unlocks)
	return fiscal.State{Credit: founder.FiscalCredit, PeriodOpenedWallMS: founder.FiscalPeriodOpenedWallMS,
		PeriodSequence: founder.FiscalPeriodSequence, GeneratorLevels: levels, Unlocks: unlocks}
}

func projectFiscal(catalog *fiscal.Catalog, founder *save.State, nowWallMS int64) (*fiscalArm, error) {
	// The auto-sweep runs on a discarded clone: the preview is what the next
	// Fiscal command would credit before acting, never a committed value.
	clone := founderFiscalState(founder)
	sweep, err := catalog.Sweep(&clone, nowWallMS)
	if err != nil {
		return nil, err
	}
	preview := fiscalSweepPreview{CreditAfter: founder.FiscalCredit}
	if sweep != nil {
		preview = fiscalSweepPreview{CreditAfter: sweep.CreditAfter, Credited: sweep.Credited, Periods: sweep.Periods, Saturated: sweep.Saturated}
	}
	levelRows := []fiscalLevelRow{}
	for _, row := range catalog.GeneratorLevelRows() {
		level := founder.FiscalGeneratorLevels[row.GeneratorID]
		var next *int64
		if level < row.LevelHardcap {
			cost, costErr := catalog.GeneratorLevelCost(row.GeneratorID, level, 1)
			if costErr != nil {
				return nil, costErr
			}
			next = &cost
		}
		levelRows = append(levelRows, fiscalLevelRow{GeneratorID: row.GeneratorID, Level: level,
			LevelCap: intCap{Amount: row.LevelHardcap, ReasonKey: row.HardcapReasonKey}, NextLevelCost: next, PPMPerLevel: row.PPMPerLevel})
	}
	sort.Slice(levelRows, func(left, right int) bool { return levelRows[left].GeneratorID < levelRows[right].GeneratorID })
	unlockRows := []fiscalUnlockRow{}
	for _, row := range catalog.UnlockRows() {
		unlockRows = append(unlockRows, fiscalUnlockRow{Cost: row.Cost, Owned: founder.FiscalUnlocks[row.UnlockID], UnlockID: row.UnlockID})
	}
	sort.Slice(unlockRows, func(left, right int) bool { return unlockRows[left].UnlockID < unlockRows[right].UnlockID })
	// Hoard preview: the published ppm contribution (design law 9) the next
	// run would freeze from the post-sweep credit, min(credit, cap) * ppm.
	hoardCredits := preview.CreditAfter
	if hoardCredits > catalog.Hoard.CapCredits {
		hoardCredits = catalog.Hoard.CapCredits
	}
	return &fiscalArm{
		Credit: founder.FiscalCredit, CreditCap: intCap{Amount: catalog.Credit.Hardcap, ReasonKey: catalog.Credit.HardcapReasonKey},
		CreditPerPeriod: catalog.Credit.CreditPerPeriod, GeneratorLevels: levelRows,
		Hoard: fiscalHoard{CapCredits: catalog.Hoard.CapCredits, PreviewPPM: hoardCredits * catalog.Hoard.PPMPerCredit, ReasonNote: "next_run"},
		Period: fiscalPeriod{AutoMS: catalog.Clock.AutoMS, EarlyMS: catalog.Clock.EarlyMS, EarlySuccessPPM: catalog.Clock.EarlySuccessPPM,
			GuaranteedMS: catalog.Clock.GuaranteedMS, OpenedWallMS: founder.FiscalPeriodOpenedWallMS, Seq: founder.FiscalPeriodSequence},
		SweepPreview: preview, Unlocks: unlockRows,
	}, nil
}

// projectMinigames restates, read-only, the create gates that
// production.StartMinigameAPISession evaluates (fiscal_unlock rule and
// human_hobby Soul gate) for every tenant the pinned minigame_api supports.
// The composed parity test pins the restatement to the server's answer.
func projectMinigames(bundle production.CatalogBundle, founder *save.State, active bool) (*minigamesArm, error) {
	rows := []minigameAvailability{}
	for _, id := range bundle.Minigames.MinigameIDs() {
		definition, ok := bundle.Minigames.Definition(id)
		if !ok {
			return nil, ErrInvalidProjection
		}
		if !bundle.MinigameAPI.SupportsTenant(id, definition.EngineRef, definition.EngineVersion) {
			continue
		}
		unlocked := true
		if definition.Unlock.Kind == "fiscal_unlock" {
			if bundle.Fiscal == nil {
				return nil, ErrInvalidProjection
			}
			_, declared := bundle.Fiscal.Unlock(definition.Unlock.UnlockID)
			unlocked = declared && founder.FiscalUnlocks[definition.Unlock.UnlockID]
		}
		locked := false
		if definition.SoulGate == "human_hobby" {
			if bundle.Soul == nil || save.VersionForState(founder) < 20 {
				return nil, ErrInvalidProjection
			}
			value, err := bundle.Soul.HumanContentLocked(founder.Soul)
			if err != nil {
				return nil, err
			}
			locked = value
		}
		rows = append(rows, minigameAvailability{ActiveSession: active, HumanContentLocked: locked, MinigameID: id, Unlocked: unlocked})
	}
	sort.Slice(rows, func(left, right int) bool { return rows[left].MinigameID < rows[right].MinigameID })
	return &minigamesArm{Rows: rows}, nil
}

// featureFacts are the GS0.5 surface unlock facts derived from the arms.
func featureFacts(features featureRows) []factRow {
	pitch := false
	if features.Minigames != nil {
		for _, row := range features.Minigames.Rows {
			pitch = pitch || row.MinigameID == "pitch"
		}
	}
	return []factRow{
		{FactID: "feature.achievements", Value: features.Achievements != nil},
		{FactID: "feature.active_play", Value: features.ActivePlay != nil},
		{FactID: "feature.fiscal", Value: features.Fiscal != nil},
		{FactID: "feature.meters", Value: features.Meters != nil},
		{FactID: "feature.minigame.pitch", Value: pitch},
		{FactID: "feature.pets", Value: features.Pets != nil},
		{FactID: "feature.reputation_tree", Value: features.Reputation != nil},
	}
}
