// Package gameui projects replay-owned state into the closed, presentation-only
// Game UI snapshot. It never mutates a save and never resolves deploy-current data.
package gameui

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/production"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

var ErrInvalidProjection = errors.New("invalid game UI projection")

type CatalogResolver interface {
	ResolveReplayCatalogs(string) (production.CatalogBundle, bool)
}

type ContributionProvider interface {
	Contributions(context.Context, *save.State, *economy.Catalog, save.Revision) ([]multiplier.Contribution, error)
}

type Projector struct {
	store         *save.Store
	catalogs      CatalogResolver
	contributions ContributionProvider
}

func New(store *save.Store, catalogs CatalogResolver, contributions ContributionProvider) (*Projector, error) {
	if store == nil || catalogs == nil {
		return nil, ErrInvalidProjection
	}
	return &Projector{store: store, catalogs: catalogs, contributions: contributions}, nil
}

type resourceCap struct {
	Amount    string `json:"amount"`
	ReasonKey string `json:"reason_key"`
}

type resourceRow struct {
	Amount        string       `json:"amount"`
	Cap           *resourceCap `json:"cap"`
	RatePerSecond string       `json:"rate_per_second"`
	ResourceID    string       `json:"resource_id"`
}

type factRow struct {
	FactID string `json:"fact_id"`
	Value  any    `json:"value"`
}

type generatorRow struct {
	GeneratorID      string `json:"generator_id"`
	MaxAffordable    int64  `json:"max_affordable"`
	NextCost         string `json:"next_cost"`
	NextCostResource string `json:"next_cost_resource_id"`
	Owned            int64  `json:"owned"`
	Provisioned      int64  `json:"provisioned"`
	RateContribution string `json:"rate_contribution"`
}

type upgradeRow struct {
	CostAmount     string `json:"cost_amount"`
	CostResourceID string `json:"cost_resource_id"`
	Eligible       bool   `json:"eligible"`
	Owned          bool   `json:"owned"`
	UpgradeID      string `json:"upgrade_id"`
}

type manualActionRow struct {
	ActionID         string `json:"action_id"`
	BucketCapMilli   int64  `json:"bucket_cap_milli"`
	RefilledAtMS     int64  `json:"refilled_at_ms"`
	RefillMilliPerMS int64  `json:"refill_milli_per_ms"`
	TokensMilli      int64  `json:"tokens_milli"`
}

type progressRow struct {
	Current string `json:"current"`
	StageID string `json:"stage_id"`
	Target  string `json:"target"`
}

type runRow struct {
	Category       string `json:"category"`
	ExitCount      int64  `json:"exit_count"`
	FounderID      string `json:"founder_id"`
	RunSeq         int64  `json:"run_seq"`
	RunStartedAtMS int64  `json:"run_started_at_ms"`
	Tier           int64  `json:"tier"`
}

type snapshot struct {
	ConstantsHash      string          `json:"constants_hash"`
	EvaluatedThroughMS int64           `json:"evaluated_through_ms"`
	Facts              []factRow       `json:"facts"`
	Generators         []generatorRow  `json:"generators"`
	ManualAction       manualActionRow `json:"manual_action"`
	Progress           []progressRow   `json:"progress"`
	Resources          []resourceRow   `json:"resources"`
	Revision           int64           `json:"revision"`
	Run                runRow          `json:"run"`
	SchemaVersion      int             `json:"schema_version"`
	ServerNowMS        int64           `json:"server_now_ms"`
	Upgrades           []upgradeRow    `json:"upgrades"`
}

// GameUISnapshot projects the latest committed Company revision and its pinned
// catalog. The supplied time is a display sample only; it never evaluates or
// commits production.
func (projector *Projector) GameUISnapshot(ctx context.Context, streamID string, now time.Time) (json.RawMessage, error) {
	if projector == nil || now.IsZero() {
		return nil, ErrInvalidProjection
	}
	loaded, err := projector.store.LoadLatest(ctx, streamID)
	if err != nil || loaded.Key.Scope != economy.ScopeCompany || loaded.State == nil || loaded.State.Ledger == nil {
		return nil, errors.Join(ErrInvalidProjection, err)
	}
	bundle, ok := projector.catalogs.ResolveReplayCatalogs(loaded.Revision.ConstantsHash)
	if !ok || bundle.Economy == nil || bundle.Routes == nil {
		return nil, ErrInvalidProjection
	}
	state, catalog := loaded.State, bundle.Economy
	// Schema-v4 content and active-play rates are assembled inside the replay
	// kernel from catalog-owned and attended-time inputs. Until that kernel-owned
	// read model is exposed, rejecting is safer than publishing a plausible but
	// incomplete rate contribution from the external-provider subset.
	if requiresKernelRateProjection(catalog, state) {
		return nil, ErrInvalidProjection
	}
	contributions := []multiplier.Contribution{}
	if projector.contributions != nil {
		revision := loaded.Revision
		revision.OwnerID = loaded.Key.OwnerID
		contributions, err = projector.contributions.Contributions(ctx, state, catalog, revision)
		if err != nil {
			return nil, err
		}
	}
	counts, err := totalGeneratorCounts(catalog, state)
	if err != nil {
		return nil, err
	}
	rates, err := production.Rates(catalog, counts, contributions)
	if err != nil {
		return nil, err
	}
	resources, err := resourceRows(catalog, state, rates)
	if err != nil {
		return nil, err
	}
	generators, err := generatorRows(catalog, state, contributions)
	if err != nil {
		return nil, err
	}
	upgrades, err := upgradeRows(catalog, bundle.Routes, state)
	if err != nil {
		return nil, err
	}
	manualActions := catalog.ManualActions()
	if len(manualActions) != 1 {
		return nil, ErrInvalidProjection
	}
	policy := catalog.ManualPolicy()
	founder, err := projector.store.LoadSiblingLatest(ctx, streamID, economy.ScopeFounder)
	if err != nil || founder.State == nil {
		return nil, errors.Join(ErrInvalidProjection, err)
	}
	progress := []progressRow{}
	if _, exists := catalog.ProgressCoordinate(int(state.Tier)); exists {
		value, progressErr := production.SubProgressValue(catalog, state, int(state.Tier))
		if progressErr != nil {
			return nil, progressErr
		}
		progress = append(progress, progressRow{Current: value.String(), StageID: "progress.tier", Target: "1e0"})
	}
	facts := []factRow{{FactID: "bootstrap.needed", Value: false}, {FactID: "run.pre_timer", Value: state.RunPreTimer}}
	for _, gate := range bundle.Routes.Gates() {
		facts = append(facts, factRow{FactID: gate.ID, Value: state.GatesCrossed[gate.ID]})
	}
	sort.Slice(facts, func(left, right int) bool { return facts[left].FactID < facts[right].FactID })
	result := snapshot{
		ConstantsHash: loaded.Revision.ConstantsHash,
		Facts:         facts,
		Generators:    generators,
		ManualAction: manualActionRow{ActionID: manualActions[0].ID, BucketCapMilli: policy.BucketCapMilli,
			RefilledAtMS: state.ManualTokenRefilledAt.UnixMilli(), RefillMilliPerMS: policy.RefillMilliPerMS,
			TokensMilli: state.ManualTokenMilli},
		Progress: progress, Resources: resources, Revision: loaded.Revision.Number,
		Run: runRow{Category: "any_percent", ExitCount: int64(len(founder.State.ExitHistory)), FounderID: loaded.Key.OwnerID,
			RunSeq: state.RunSeq, RunStartedAtMS: state.RunStartedAt.UnixMilli(), Tier: state.Tier},
		EvaluatedThroughMS: state.EvaluatedThrough.UnixMilli(), SchemaVersion: 1,
		ServerNowMS: save.CanonicalServerTime(now).UnixMilli(), Upgrades: upgrades,
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func requiresKernelRateProjection(catalog *economy.Catalog, state *save.State) bool {
	if state != nil && save.VersionForState(state) >= 18 {
		return true
	}
	if catalog == nil || len(catalog.Upgrades()) != 0 || len(catalog.SynergyPools()) != 0 {
		return true
	}
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		if len(generator.Ladder) != 0 {
			return true
		}
	}
	return false
}

func totalGeneratorCounts(catalog *economy.Catalog, state *save.State) (map[string]int64, error) {
	result := make(map[string]int64)
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		owned, ownedOK := state.GeneratorCounts[generator.ID]
		provisioned := int64(0)
		provisionedOK := state.GeneratorProvisioned == nil
		if state.GeneratorProvisioned != nil {
			provisioned, provisionedOK = state.GeneratorProvisioned[generator.ID]
		}
		if !ownedOK || !provisionedOK || owned < 0 || provisioned < 0 || owned > decimal.MaxExactInteger-provisioned {
			return nil, ErrInvalidProjection
		}
		result[generator.ID] = owned + provisioned
	}
	return result, nil
}

func resourceRows(catalog *economy.Catalog, state *save.State, rates map[string][]decimal.Decimal) ([]resourceRow, error) {
	result := []resourceRow{}
	for _, resource := range catalog.Resources() {
		if resource.Scope != economy.ScopeCompany {
			continue
		}
		balance, ok := state.Ledger.Balance(resource.ID)
		if !ok {
			return nil, ErrInvalidProjection
		}
		rate := decimal.Zero
		for _, part := range rates[resource.ID] {
			rate = rate.Add(part)
		}
		var cap *resourceCap
		if resource.Hardcap != nil {
			cap = &resourceCap{Amount: resource.Hardcap.Amount.String(), ReasonKey: resource.Hardcap.ReasonKey}
		}
		result = append(result, resourceRow{Amount: balance.String(), Cap: cap, RatePerSecond: rate.String(), ResourceID: resource.ID})
	}
	sort.Slice(result, func(left, right int) bool { return result[left].ResourceID < result[right].ResourceID })
	return result, nil
}

func generatorRows(catalog *economy.Catalog, state *save.State, contributions []multiplier.Contribution) ([]generatorRow, error) {
	definitions := catalog.GeneratorClassesForScope(economy.ScopeCompany)
	result := make([]generatorRow, 0, len(definitions))
	for _, generator := range definitions {
		owned, provisioned := state.GeneratorCounts[generator.ID], state.GeneratorProvisioned[generator.ID]
		isolated := make(map[string]int64, len(definitions))
		for _, row := range definitions {
			isolated[row.ID] = 0
		}
		isolated[generator.ID] = owned + provisioned
		rates, err := production.Rates(catalog, isolated, contributions)
		if err != nil {
			return nil, err
		}
		rate := decimal.Zero
		if generator.Production != nil {
			for _, part := range rates[generator.Production.ResourceID] {
				rate = rate.Add(part)
			}
		}
		balance, exists := state.Ledger.Balance(generator.Price.ResourceID)
		if !exists {
			return nil, ErrInvalidProjection
		}
		next, err := catalog.BulkCost(generator.ID, owned, 1)
		if err != nil {
			return nil, err
		}
		affordable, err := catalog.MaxAffordableDetailed(generator.ID, balance, owned)
		if err != nil {
			return nil, err
		}
		result = append(result, generatorRow{GeneratorID: generator.ID, MaxAffordable: affordable.Count,
			NextCost: next.String(), NextCostResource: generator.Price.ResourceID, Owned: owned,
			Provisioned: provisioned, RateContribution: rate.String()})
	}
	sort.Slice(result, func(left, right int) bool { return result[left].GeneratorID < result[right].GeneratorID })
	return result, nil
}

func upgradeRows(catalog *economy.Catalog, routeCatalog *routes.Catalog, state *save.State) ([]upgradeRow, error) {
	context := routes.Context{ContextVersion: routeCatalog.ContextVersion(), Resources: map[string]decimal.Decimal{},
		DoctrinesByTransition: state.DoctrinesByTransition, StructureID: state.StructureID,
		LedgerFactKinds: state.LedgerFactKinds, MeterBands: state.MeterBands, RegionTraits: state.RegionTraits}
	for _, resource := range catalog.Resources() {
		if value, ok := state.Ledger.Balance(resource.ID); ok {
			context.Resources[resource.ID] = value
		}
	}
	result := make([]upgradeRow, 0, len(catalog.Upgrades()))
	for _, upgrade := range catalog.Upgrades() {
		owned := state.UpgradesOwned[upgrade.ID]
		inWindow := (upgrade.Window.FromGate == "" || state.GatesCrossed[upgrade.Window.FromGate]) &&
			(upgrade.Window.ToGate == "" || !state.GatesCrossed[upgrade.Window.ToGate])
		eligible, err := routes.EvaluatePredicate(upgrade.Requires, context)
		if err != nil {
			return nil, err
		}
		balance, exists := state.Ledger.Balance(upgrade.Cost.ResourceID)
		if !exists {
			return nil, ErrInvalidProjection
		}
		result = append(result, upgradeRow{CostAmount: upgrade.Cost.Amount.String(), CostResourceID: upgrade.Cost.ResourceID,
			Eligible: !owned && inWindow && eligible && balance.Gte(upgrade.Cost.Amount), Owned: owned, UpgradeID: upgrade.ID})
	}
	sort.Slice(result, func(left, right int) bool { return result[left].UpgradeID < result[right].UpgradeID })
	return result, nil
}
