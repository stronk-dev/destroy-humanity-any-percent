// Package formulas owns the generated production-formula artifact model and
// its API descriptor. It describes existing math; it does not perform replay,
// regenerate historical artifacts or choose balance values.
package formulas

import (
	"cloud-clicker/server/commons"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
)

const SchemaVersion = 14

type Artifact struct {
	SchemaVersion       int               `json:"schema_version"`
	ProductionRate      string            `json:"production_rate"`
	MultiplierSlotOrder []multiplier.Slot `json:"multiplier_slot_order"`
	WithinSlotOrder     string            `json:"within_slot_order"`
	SourceFingerprint   string            `json:"source_fingerprint"`
	Commons             CommonsFormula    `json:"commons"`
	Guild               GuildFormula      `json:"guild"`
	PurchasableContent  ContentFormula    `json:"purchasable_content"`
	Meters              MeterFormula      `json:"meters"`
	MinigameScaling     MinigameFormula   `json:"minigame_scaling"`
	AxisStack           AxisStackFormula  `json:"axis_stack"`
}

// AxisStackFormula publishes the Clout v1 run-local axis stack (CV3). Pinned
// is null until an epoch's economy declares axis_stack.
type AxisStackFormula struct {
	Input  string            `json:"input"`
	Factor string            `json:"factor"`
	Stack  string            `json:"stack"`
	Slot   string            `json:"slot"`
	Timing string            `json:"timing"`
	Pinned *AxisStackPinning `json:"pinned"`
}

type AxisStackPinning struct {
	Input           string `json:"input"`
	InputCap        int64  `json:"input_cap"`
	CapReasonKey    string `json:"cap_reason_key"`
	MaximumAtCapPPM string `json:"maximum_product_at_cap"`
}

type MinigameFormula struct {
	Grammar             string   `json:"grammar"`
	OperationOrder      []string `json:"operation_order"`
	Rounding            string   `json:"rounding"`
	FairnessGate        string   `json:"fairness_gate"`
	FallbackArms        []string `json:"fallback_arms"`
	FallbackRule        string   `json:"fallback_rule"`
	OfflineGradeGrammar string   `json:"offline_grade_grammar"`
	OfflineGradeRule    string   `json:"offline_grade_rule"`
	PayoutGrammar       string   `json:"payout_grammar"`
	PayoutOrder         []string `json:"payout_order"`
	PayoutMath          string   `json:"payout_math"`
}

type MeterFormula struct {
	HookOrder         string `json:"hook_order"`
	AttendedStep      string `json:"attended_step"`
	Decay             string `json:"decay"`
	LedgerFactInput   string `json:"ledger_fact_input"`
	ContributionInput string `json:"contribution_input"`
	BandEvents        string `json:"band_events"`
}

type ContentFormula struct {
	Provisioning        string               `json:"provisioning"`
	ManualOutput        string               `json:"manual_output"`
	StockRate           string               `json:"stock_rate"`
	SynergyLinear       string               `json:"synergy_linear"`
	SynergyLog          string               `json:"synergy_log"`
	Ladders             string               `json:"ladders"`
	ProvisionTickMS     int64                `json:"provision_tick_ms"`
	ProvisionedHardcaps []ContentHardcap     `json:"provisioned_hardcaps"`
	SynergyPools        []ContentSynergyPool `json:"synergy_pools"`
}

type ContentHardcap struct {
	GeneratorID string `json:"generator_id"`
	Count       int64  `json:"count"`
	ReasonKey   string `json:"reason_key"`
}

type ContentSynergyPool struct {
	ID      string                 `json:"id"`
	Curve   economy.SynergyCurve   `json:"curve"`
	Slot    multiplier.Slot        `json:"slot"`
	Target  string                 `json:"target"`
	Sources []ContentSynergySource `json:"sources"`
}

type ContentSynergySource struct {
	Kind        economy.SynergySourceKind `json:"kind"`
	ID          string                    `json:"id_or_class"`
	PerCountPPM int64                     `json:"per_count_ppm"`
}

type GuildFormula struct {
	Tithe                      string `json:"tithe"`
	Health                     string `json:"health"`
	Clearing                   string `json:"clearing"`
	StockConsumption           string `json:"stock_consumption"`
	GuildTithePPM              int64  `json:"guild_tithe_ppm"`
	GuildXPTargetPerFounder    int64  `json:"guild_xp_target_per_founder"`
	ClearingRatePPM            int64  `json:"clearing_rate_ppm"`
	NPCExchangePPM             int64  `json:"npc_exchange_ppm"`
	StockIntakeCap             int64  `json:"stock_intake_cap"`
	ConsumptionBonusPPMPerUnit int64  `json:"consumption_bonus_ppm_per_unit"`
	ClearingIntervalMS         int64  `json:"clearing_interval_ms"`
}

type CommonsFormula struct {
	Enclosure                string                 `json:"enclosure"`
	Compliance               string                 `json:"compliance"`
	Health                   string                 `json:"health"`
	EffectiveHealth          string                 `json:"effective_health"`
	Modifier                 string                 `json:"modifier"`
	Solidarity               string                 `json:"solidarity"`
	EntryParticipationWeight string                 `json:"entry_participation_weight"`
	SourceWeights            []commons.SourceWeight `json:"source_weights"`
	DefaultTithePPM          int64                  `json:"default_tithe_ppm"`
	MinimumTithePPM          int64                  `json:"minimum_tithe_ppm"`
	MaximumTithePPM          int64                  `json:"maximum_tithe_ppm"`
	GuildHealthWeightPPM     int64                  `json:"guild_health_weight_ppm"`
	CohortHealthWeightPPM    int64                  `json:"cohort_health_weight_ppm"`
	ServerHealthWeightPPM    int64                  `json:"server_health_weight_ppm"`
	CollectiveWeightPPM      int64                  `json:"collective_weight_ppm"`
	CollectiveExponentPPM    int64                  `json:"collective_exponent_ppm"`
	CollapseHealthPPM        int64                  `json:"collapse_health_ppm"`
	HealthyHealthPPM         int64                  `json:"healthy_health_ppm"`
	MaximumBonus             string                 `json:"maximum_bonus"`
	HealthRecoveryPPMPerHour int64                  `json:"health_recovery_ppm_per_hour"`
	HealthDecayPPMPerHour    int64                  `json:"health_decay_ppm_per_hour"`
	SolidarityWindowMS       int64                  `json:"solidarity_window_ms"`
	CohortTargetSize         int                    `json:"cohort_target_size"`
	CohortMergeFloor         int                    `json:"cohort_merge_floor"`
	NPCPopulationFloor       int                    `json:"npc_population_floor"`
	NPCWeightPPM             int64                  `json:"npc_weight_ppm"`
	NPCCompliancePPM         int64                  `json:"npc_compliance_ppm"`
	PopulationTolerancePPM   int64                  `json:"population_tolerance_ppm"`
}
