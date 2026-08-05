package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/multiplier"
)

type formulaArtifact struct {
	SchemaVersion       int               `json:"schema_version"`
	ProductionRate      string            `json:"production_rate"`
	MultiplierSlotOrder []multiplier.Slot `json:"multiplier_slot_order"`
	WithinSlotOrder     string            `json:"within_slot_order"`
	SourceFingerprint   string            `json:"source_fingerprint"`
	Commons             commonsFormula    `json:"commons"`
	Guild               guildFormula      `json:"guild"`
	PurchasableContent  contentFormula    `json:"purchasable_content"`
	Meters              meterFormula      `json:"meters"`
	MinigameScaling     minigameFormula   `json:"minigame_scaling"`
}

type minigameFormula struct {
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

type meterFormula struct {
	HookOrder         string `json:"hook_order"`
	AttendedStep      string `json:"attended_step"`
	Decay             string `json:"decay"`
	LedgerFactInput   string `json:"ledger_fact_input"`
	ContributionInput string `json:"contribution_input"`
	BandEvents        string `json:"band_events"`
}

type contentFormula struct {
	Provisioning        string               `json:"provisioning"`
	ManualOutput        string               `json:"manual_output"`
	StockRate           string               `json:"stock_rate"`
	SynergyLinear       string               `json:"synergy_linear"`
	SynergyLog          string               `json:"synergy_log"`
	Ladders             string               `json:"ladders"`
	ProvisionTickMS     int64                `json:"provision_tick_ms"`
	ProvisionedHardcaps []contentHardcap     `json:"provisioned_hardcaps"`
	SynergyPools        []contentSynergyPool `json:"synergy_pools"`
}

type contentHardcap struct {
	GeneratorID string `json:"generator_id"`
	Count       int64  `json:"count"`
	ReasonKey   string `json:"reason_key"`
}

type contentSynergyPool struct {
	ID      string                 `json:"id"`
	Curve   economy.SynergyCurve   `json:"curve"`
	Slot    multiplier.Slot        `json:"slot"`
	Target  string                 `json:"target"`
	Sources []contentSynergySource `json:"sources"`
}

type contentSynergySource struct {
	Kind        economy.SynergySourceKind `json:"kind"`
	ID          string                    `json:"id_or_class"`
	PerCountPPM int64                     `json:"per_count_ppm"`
}

type guildFormula struct {
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

type commonsFormula struct {
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

type authorityKind int

const (
	authorityFunction authorityKind = iota
	authorityMethod
	authorityValue
)

type authoritySpec struct {
	label  string
	path   string
	kind   authorityKind
	symbol string
}

var formulaAuthorities = []authoritySpec{
	{label: "production.ratesWithProvisionedAndPolicy", path: "production/engine.go", kind: authorityFunction, symbol: "ratesWithProvisionedAndPolicy"},
	{label: "production.accrueContent", path: "production/engine.go", kind: authorityFunction, symbol: "accrueContent"},
	{label: "production.contentContributionsWithPolicy", path: "production/content.go", kind: authorityFunction, symbol: "contentContributionsWithPolicy"},
	{label: "production.countPPMFactor", path: "production/content.go", kind: authorityFunction, symbol: "countPPMFactor"},
	{label: "production.synergyFactor", path: "production/content.go", kind: authorityFunction, symbol: "synergyFactor"},
	{label: "production.materializeProvisionBoundaryWithPolicy", path: "production/content.go", kind: authorityFunction, symbol: "materializeProvisionBoundaryWithPolicy"},
	{label: "production.applyFoundationTransition", path: "production/foundation_transition.go", kind: authorityFunction, symbol: "applyFoundationTransition"},
	{label: "faction.AccrualHook.AfterAccrual", path: "faction/hook.go", kind: authorityMethod, symbol: "AfterAccrual"},
	{label: "multiplier.Order", path: "multiplier/contribution.go", kind: authorityValue, symbol: "Order"},
	{label: "multiplier.OrderedSourceIDs", path: "multiplier/contribution.go", kind: authorityFunction, symbol: "OrderedSourceIDs"},
	{label: "commons.EnclosureIndex", path: "commons/formula.go", kind: authorityFunction, symbol: "EnclosureIndex"},
	{label: "commons.EntryParticipationWeightPPM", path: "commons/formula.go", kind: authorityFunction, symbol: "EntryParticipationWeightPPM"},
	{label: "commons.EffectiveHealthPPM", path: "commons/formula.go", kind: authorityFunction, symbol: "EffectiveHealthPPM"},
	{label: "commons.Modifier", path: "commons/formula.go", kind: authorityFunction, symbol: "Modifier"},
	{label: "commons.AggregateHealth", path: "commons/health.go", kind: authorityFunction, symbol: "AggregateHealth"},
	{label: "commons.SmoothHealthPPM", path: "commons/health.go", kind: authorityFunction, symbol: "SmoothHealthPPM"},
	{label: "guild.HealthPPM", path: "guild/projector.go", kind: authorityFunction, symbol: "HealthPPM"},
	{label: "guild.Clear", path: "guild/exchange.go", kind: authorityFunction, symbol: "Clear"},
	{label: "guild.ApplySettlements", path: "guild/clearing_store.go", kind: authorityFunction, symbol: "ApplySettlements"},
	{label: "meters.Advance", path: "meters/transition.go", kind: authorityFunction, symbol: "Advance"},
	{label: "meters.wholeSteps", path: "meters/transition.go", kind: authorityFunction, symbol: "wholeSteps"},
	{label: "minigame.LoadScalingPolicy", path: "minigame/scaling.go", kind: authorityFunction, symbol: "LoadScalingPolicy"},
	{label: "minigame.ScalingPolicy.Resolve", path: "minigame/scaling.go", kind: authorityMethod, symbol: "Resolve"},
	{label: "minigame.floorBigInt", path: "minigame/scaling.go", kind: authorityFunction, symbol: "floorBigInt"},
	{label: "minigame.LoadFallbackPolicy", path: "minigame/fallback.go", kind: authorityFunction, symbol: "LoadFallbackPolicy"},
	{label: "minigame.uniqueJSONKeys", path: "minigame/fallback.go", kind: authorityFunction, symbol: "uniqueJSONKeys"},
	{label: "minigame.LoadOfflineQualityPolicy", path: "minigame/offline_quality.go", kind: authorityFunction, symbol: "LoadOfflineQualityPolicy"},
	{label: "minigame.OfflineGradeForScore", path: "minigame/offline_quality.go", kind: authorityFunction, symbol: "OfflineGradeForScore"},
	{label: "minigame.LoadPayoutPolicy", path: "minigame/payout.go", kind: authorityFunction, symbol: "LoadPayoutPolicy"},
	{label: "minigame.SelectPayoutScore", path: "minigame/payout.go", kind: authorityFunction, symbol: "SelectPayoutScore"},
	{label: "minigame.ConvertPayout", path: "minigame/payout.go", kind: authorityFunction, symbol: "ConvertPayout"},
	{label: "minigame.applyFaucetWindowTx", path: "minigame/faucet.go", kind: authorityFunction, symbol: "applyFaucetWindowTx"},
}

func main() {
	output := flag.String("output", "", "output JSON filename")
	flag.Parse()
	if *output == "" {
		fmt.Fprintln(os.Stderr, "-output is required")
		os.Exit(2)
	}
	root, err := moduleRoot()
	if err != nil {
		panic(err)
	}
	fingerprint, err := sourceFingerprint(root)
	if err != nil {
		panic(err)
	}
	commonsBytes, err := os.ReadFile(filepath.Join(root, "..", "balance", "commons", "phase0.json"))
	if err != nil {
		panic(err)
	}
	commonsCatalog, err := commons.LoadCatalog(commonsBytes)
	if err != nil {
		panic(err)
	}
	guildBytes, err := os.ReadFile(filepath.Join(root, "..", "balance", "guilds", "phase0.json"))
	if err != nil {
		panic(err)
	}
	guildCatalog, err := guild.LoadCatalog(guildBytes)
	if err != nil {
		panic(err)
	}
	economyBytes, err := os.ReadFile(filepath.Join(root, "..", "balance", "catalogs", "phase0.json"))
	if err != nil {
		panic(err)
	}
	economyCatalog, err := economy.LoadCatalog(economyBytes)
	if err != nil {
		panic(err)
	}
	artifact := formulaArtifact{
		SchemaVersion:       13,
		ProductionRate:      "sum_generators((purchased_count + provisioned_count) * base_rate * product(multiplier_slots))",
		MultiplierSlotOrder: append([]multiplier.Slot(nil), multiplier.Order[:]...),
		WithinSlotOrder:     multiplier.WithinSlotOrder,
		SourceFingerprint:   fingerprint,
		Commons: commonsFormula{
			Enclosure:                "clamp(1 - product(clean weighted factors) / product(all weighted factors), 0, 1)",
			Compliance:               "clamp(tithe_ppm / target_tithe_ppm, 0, 1) * (1 - enclosure)",
			Health:                   "sum(weight_ppm * compliance_ppm) / sum(weight_ppm)",
			EffectiveHealth:          "(guild_health_weight_ppm * guild_health_ppm + cohort_health_weight_ppm * cohort_health_ppm + server_health_weight_ppm * server_health_ppm) / 1000000; guildless substitutes cohort for guild",
			Modifier:                 "1 + maximum_bonus * (collective_weight * max(0, ((health - collapse_health) / (1 - collapse_health))^collective_exponent) + (1 - collective_weight) * solidarity)",
			Solidarity:               "sum(hourly_compliance_ppm * covered_ms) / 2592000000",
			EntryParticipationWeight: "clamp(floor(tithe_ppm * 1000000 / default_tithe_ppm), floor(minimum_tithe_ppm * 1000000 / default_tithe_ppm), min(1000000, floor(maximum_tithe_ppm * 1000000 / default_tithe_ppm)))",
			SourceWeights:            append([]commons.SourceWeight{}, commonsCatalog.SourceWeights...),
			DefaultTithePPM:          commonsCatalog.DefaultTithePPM,
			MinimumTithePPM:          commonsCatalog.MinimumTithePPM,
			MaximumTithePPM:          commonsCatalog.MaximumTithePPM,
			GuildHealthWeightPPM:     commonsCatalog.GuildHealthWeightPPM,
			CohortHealthWeightPPM:    commonsCatalog.CohortHealthWeightPPM,
			ServerHealthWeightPPM:    commonsCatalog.ServerHealthWeightPPM,
			CollectiveWeightPPM:      commonsCatalog.CollectiveWeightPPM,
			CollectiveExponentPPM:    commonsCatalog.CollectiveExponentPPM,
			CollapseHealthPPM:        commonsCatalog.CollapseHealthPPM,
			HealthyHealthPPM:         commonsCatalog.HealthyHealthPPM,
			MaximumBonus:             commonsCatalog.MaximumBonus.String(),
			HealthRecoveryPPMPerHour: commonsCatalog.HealthRecoveryPPMPerHour,
			HealthDecayPPMPerHour:    commonsCatalog.HealthDecayPPMPerHour,
			SolidarityWindowMS:       commonsCatalog.SolidarityWindowMS,
			CohortTargetSize:         commonsCatalog.CohortTargetSize,
			CohortMergeFloor:         commonsCatalog.CohortMergeFloor,
			NPCPopulationFloor:       commonsCatalog.NPCPopulationFloor,
			NPCWeightPPM:             commonsCatalog.NPCWeightPPM,
			NPCCompliancePPM:         commonsCatalog.NPCCompliancePPM,
			PopulationTolerancePPM:   commonsCatalog.PopulationTolerancePPM,
		},
		Guild: guildFormula{
			Tithe:            "xp_delta = floor((progress_delta_ppm * guild_tithe_ppm + carry_ppm) / 1000000); persist remainder",
			Health:           "clamp(1000000 * window_xp / (active_founders * guild_xp_target_per_founder), 0, 1000000)",
			Clearing:         "ordered one-pass allocation of floor(stock_units * clearing_rate_ppm / 1000000), capped by intake and stock headroom, no redistribution",
			StockConsumption: "1 + consumed_this_window * consumption_bonus_ppm_per_unit / 1000000",
			GuildTithePPM:    guildCatalog.GuildTithePPM, GuildXPTargetPerFounder: guildCatalog.GuildXPTargetPerFounder,
			ClearingRatePPM: guildCatalog.ClearingRatePPM, NPCExchangePPM: guildCatalog.NPCExchangePPM,
			StockIntakeCap: guildCatalog.StockIntakeCap, ConsumptionBonusPPMPerUnit: guildCatalog.ConsumptionBonusPPMPerUnit,
			ClearingIntervalMS: guildCatalog.ClearingIntervalMS,
		},
		PurchasableContent: contentFormulaFor(economyCatalog),
		Meters: meterFormula{
			HookOrder:         "decay, then newly emitted ledger facts, then active non-neutral contribution inputs; clamp once; derive one prior-to-final band event; achievements evaluate after meters",
			AttendedStep:      fmt.Sprintf("attended_ms = attended_total(after) - attended_total(before), using the canonical offline-span ledger; whole = floor((rate_per_attended_hour * attended_ms + remainder) / %d); remainder = modulus", meters.MillisPerHour),
			Decay:             "move linearly toward declared target by whole attended steps; clear remainder on target saturation",
			LedgerFactInput:   "apply declared delta once when fact_kind enters the committed Company fact set in this transition",
			ContributionInput: "integrate declared delta_per_attended_hour only while (slot, source_id) is committed and non-neutral",
			BandEvents:        "derive bands from numeric values; emit at most one meter_band_changed.v1 per changed meter in meter_id byte order",
		},
		MinigameScaling: minigameFormula{
			Grammar:             "one exact-key row per unique destination: destination, destination_class, source_kind, source_ref, op, operand, clamp_min, clamp_max",
			OperationOrder:      []string{"resolve_source", "apply_" + string(minigame.ScalingIdentity) + "|" + string(minigame.ScalingAdd) + "|" + string(minigame.ScalingMul) + "|" + string(minigame.ScalingFloorDiv), "clamp"},
			Rounding:            "floordiv uses mathematical floor, including negative non-integral quotients",
			FairnessGate:        "ranked && destination_class == power rejects the catalog",
			FallbackArms:        []string{string(minigame.FallbackSolo), string(minigame.FallbackBot), string(minigame.FallbackNPCPartner)},
			FallbackRule:        "every policy is one exact-key arm; bot_ref/npc_profile identity and semantic version are frozen; rate_reduction_ppm is within [0,1000000]",
			OfflineGradeGrammar: "grade_curve rows are exact {score_threshold,grade_ppm}; score thresholds strictly ascend, grades never descend, and the lowest grade equals neutral_floor_ppm",
			OfflineGradeRule:    "scores use the grade at the greatest threshold they meet; scores below the first threshold use neutral_floor_ppm",
			PayoutGrammar:       "exact keys: credited_resource_id, sends_per_day, per_send_cap, conversion_ppm, payout_score_fact_id, cap_reason_key; resource, score fact, and copy key must exist in their owning declarations",
			PayoutOrder:         []string{"select the one nonnegative certified payout_score_fact_id", "floor(score * (1000000 - rate_reduction_ppm) / 1000000)", "floor((reduced_score * conversion_ppm + prior_remainder_ppm) / 1000000)", "persist modulo remainder on Founder/minigame/attended-day window", "apply per_send_cap when quota_used < sends_per_day", "increment quota_used for each admitted send"},
			PayoutMath:          "exact integer intermediates; legal exact-domain inputs cannot overflow the converted output",
		},
	}
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		panic(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(*output, data, 0o644); err != nil {
		panic(err)
	}
}

func contentFormulaFor(catalog *economy.Catalog) contentFormula {
	result := contentFormula{
		Provisioning:        "at each run-aligned tick: staged = floor(((purchased + provisioned) * rate_ppm + remainder_ppm) / 1000000); persist modulus; saturate at declared cap; new units produce next bucket",
		ManualOutput:        "1 + purchased_count * per_purchased_ppm / 1000000",
		StockRate:           "scaled_ms = floor((elapsed_ms * (1000000 + sum(purchased_count * per_purchased_ppm)) + remainder_ppm) / 1000000); persist modulus",
		SynergyLinear:       "1 + sum_ppm / 1000000",
		SynergyLog:          "1 + log10(1 + sum_ppm / 1000000)",
		Ladders:             "product(reached purchased-only rung multiplier_ppm / 1000000) in raw-byte source order",
		ProvisionedHardcaps: []contentHardcap{}, SynergyPools: []contentSynergyPool{},
	}
	if catalog == nil {
		return result
	}
	result.ProvisionTickMS = catalog.ProvisionTickMS()
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		if generator.ProvisionedHardcap != nil {
			result.ProvisionedHardcaps = append(result.ProvisionedHardcaps, contentHardcap{GeneratorID: generator.ID, Count: generator.ProvisionedHardcap.Count, ReasonKey: generator.ProvisionedHardcap.ReasonKey})
		}
	}
	for _, pool := range catalog.SynergyPools() {
		declaration, _ := catalog.MultiplierSource(pool.ID)
		published := contentSynergyPool{ID: pool.ID, Curve: pool.Curve, Slot: pool.Slot, Target: declaration.Target, Sources: []contentSynergySource{}}
		for _, source := range pool.Sources {
			published.Sources = append(published.Sources, contentSynergySource{Kind: source.Kind, ID: source.ID, PerCountPPM: source.PerCountPPM})
		}
		result.SynergyPools = append(result.SynergyPools, published)
	}
	return result
}

func moduleRoot() (string, error) {
	directory, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return "", fmt.Errorf("go.mod not found from working directory")
		}
		directory = parent
	}
}

func sourceFingerprint(root string) (string, error) {
	return sourceFingerprintFrom(func(path string) ([]byte, error) {
		return os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	})
}

func sourceFingerprintFrom(readSource func(string) ([]byte, error)) (string, error) {
	hash := sha256.New()
	for _, authority := range formulaAuthorities {
		source, err := readSource(authority.path)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", authority.label, err)
		}
		canonical, err := canonicalAuthority(source, authority)
		if err != nil {
			return "", err
		}
		_, _ = hash.Write([]byte(authority.label))
		_, _ = hash.Write([]byte{'\n'})
		_, _ = hash.Write(canonical)
		_, _ = hash.Write([]byte{'\n'})
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func canonicalAuthority(source []byte, authority authoritySpec) ([]byte, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, authority.path, source, 0)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", authority.label, err)
	}
	var matches []ast.Node
	for _, declaration := range file.Decls {
		switch authority.kind {
		case authorityFunction:
			function, ok := declaration.(*ast.FuncDecl)
			if ok && function.Recv == nil && function.Name.Name == authority.symbol {
				matches = append(matches, function)
			}
		case authorityMethod:
			function, ok := declaration.(*ast.FuncDecl)
			if ok && function.Recv != nil && function.Name.Name == authority.symbol {
				matches = append(matches, function)
			}
		case authorityValue:
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.VAR && general.Tok != token.CONST {
				continue
			}
			for _, raw := range general.Specs {
				value, ok := raw.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range value.Names {
					if name.Name == authority.symbol {
						matches = append(matches, value)
					}
				}
			}
		default:
			return nil, fmt.Errorf("unsupported authority kind for %s", authority.label)
		}
	}
	if len(matches) != 1 {
		return nil, fmt.Errorf("%s: found %d declarations, want exactly 1", authority.label, len(matches))
	}
	var canonical bytes.Buffer
	if err := format.Node(&canonical, fileSet, matches[0]); err != nil {
		return nil, fmt.Errorf("format %s: %w", authority.label, err)
	}
	return []byte(strings.TrimSpace(canonical.String())), nil
}
