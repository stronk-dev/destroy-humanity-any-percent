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

// MinigameActivityResolver is the read-only MA-C12 active-session predicate
// (minigame.Repository.ActiveMinigame), the same one Exit freezes into replay.
type MinigameActivityResolver interface {
	ActiveMinigame(context.Context, string) (bool, error)
}

type Projector struct {
	store            *save.Store
	catalogs         CatalogResolver
	contributions    ContributionProvider
	minigameActivity MinigameActivityResolver
}

type Option func(*Projector) error

// WithMinigameActivity binds the active-session predicate. Bundles that pin a
// minigame_api artifact cannot be projected without it: an omitted resolver
// would silently preview a wind_down the server rejects.
func WithMinigameActivity(resolver MinigameActivityResolver) Option {
	return func(projector *Projector) error {
		if resolver == nil {
			return ErrInvalidProjection
		}
		projector.minigameActivity = resolver
		return nil
	}
}

func New(store *save.Store, catalogs CatalogResolver, contributions ContributionProvider, options ...Option) (*Projector, error) {
	if store == nil || catalogs == nil {
		return nil, ErrInvalidProjection
	}
	projector := &Projector{store: store, catalogs: catalogs, contributions: contributions}
	for _, option := range options {
		if err := option(projector); err != nil {
			return nil, err
		}
	}
	return projector, nil
}

// minigameActive mirrors production's Exit gate: the predicate applies only
// when the run's bundle pins minigame_api, and fails loud without a resolver.
func (projector *Projector) minigameActive(ctx context.Context, bundle production.CatalogBundle, founderID string) (bool, error) {
	if bundle.MinigameAPI == nil {
		return false, nil
	}
	if projector.minigameActivity == nil {
		return false, ErrInvalidProjection
	}
	return projector.minigameActivity.ActiveMinigame(ctx, founderID)
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
	GeneratorID      string  `json:"generator_id"`
	MaxAffordable    int64   `json:"max_affordable"`
	NextCost         string  `json:"next_cost"`
	NextCostResource string  `json:"next_cost_resource_id"`
	Owned            int64   `json:"owned"`
	Provisioned      int64   `json:"provisioned"`
	ProvisionCap     *intCap `json:"provision_cap"`
	RateContribution string  `json:"rate_contribution"`
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

type crossGateTransition struct {
	Eligible bool   `json:"eligible"`
	GateID   string `json:"gate_id"`
	RouteID  any    `json:"route_id"`
}

type eligibilityTransition struct {
	Eligible bool `json:"eligible"`
}

type transitionRows struct {
	CrossGate *crossGateTransition `json:"cross_gate"`
	// Incorporate is omitted unless incorporate can apply (Tier 2 content §E2);
	// an optional property keeps v3/v4 snapshots below Tier 2 byte-identical.
	Incorporate *incorporateTransition `json:"incorporate,omitempty"`
	WindDown    eligibilityTransition  `json:"wind_down"`
}

type incorporateTransition struct {
	Factions []incorporateFaction `json:"factions"`
}

type incorporateFaction struct {
	CopyKey   string `json:"copy_key"`
	FactionID string `json:"faction_id"`
}

type snapshot struct {
	ConstantsHash      string          `json:"constants_hash"`
	EvaluatedThroughMS int64           `json:"evaluated_through_ms"`
	Facts              []factRow       `json:"facts"`
	Features           featureRows     `json:"features"`
	FounderRevision    int64           `json:"founder_revision"`
	Generators         []generatorRow  `json:"generators"`
	ManualAction       manualActionRow `json:"manual_action"`
	Progress           []progressRow   `json:"progress"`
	Resources          []resourceRow   `json:"resources"`
	Revision           int64           `json:"revision"`
	Run                runRow          `json:"run"`
	SchemaVersion      int             `json:"schema_version"`
	ServerNowMS        int64           `json:"server_now_ms"`
	Transitions        transitionRows  `json:"transitions"`
	Upgrades           []upgradeRow    `json:"upgrades"`
}

// GameUISnapshot projects the latest committed Company revision and its pinned
// catalog. The supplied time resolves display clocks and the read-only active-
// play rate sample; it never evaluates or commits production.
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
	contributions := []multiplier.Contribution{}
	if projector.contributions != nil {
		revision := loaded.Revision
		revision.OwnerID = loaded.Key.OwnerID
		contributions, err = projector.contributions.Contributions(ctx, state, catalog, revision)
		if err != nil {
			return nil, err
		}
	}
	founder, err := projector.store.LoadSiblingLatest(ctx, streamID, economy.ScopeFounder)
	if err != nil || founder.State == nil {
		return nil, errors.Join(ErrInvalidProjection, err)
	}
	minigameActive, err := projector.minigameActive(ctx, bundle, loaded.Key.OwnerID)
	if err != nil {
		return nil, err
	}
	return projectSnapshot(bundle, loaded.Key.OwnerID, loaded.Revision.Number, founder.Revision.Number, state, founder.State, contributions, now, minigameActive)
}

// InitialGameUISnapshot projects transaction-local initial states without a
// database read. Bootstrap can therefore validate and persist the exact
// response in the account-creation transaction.
func (projector *Projector) InitialGameUISnapshot(_ context.Context, constantsHash, founderID string,
	company, founder *save.State, frozen []save.FrozenContribution, now time.Time) (json.RawMessage, error) {
	if projector == nil || constantsHash == "" || founderID == "" || company == nil || founder == nil || now.IsZero() {
		return nil, ErrInvalidProjection
	}
	bundle, ok := projector.catalogs.ResolveReplayCatalogs(constantsHash)
	if !ok || bundle.Economy == nil || bundle.Routes == nil {
		return nil, ErrInvalidProjection
	}
	contributions, err := production.ResolveFrozenContributions(bundle.Economy, frozen)
	if err != nil {
		return nil, err
	}
	// A Founder created in this transaction cannot own a minigame session.
	return projectSnapshot(bundle, founderID, 1, 1, company, founder, contributions, now, false)
}

func projectSnapshot(bundle production.CatalogBundle, founderID string, revision, founderRevision int64, state, founder *save.State,
	contributions []multiplier.Contribution, now time.Time, minigameActive bool) (json.RawMessage, error) {
	if state == nil || state.Ledger == nil || founder == nil || founder.Ledger == nil || founderID == "" || revision < 1 || founderRevision < 1 || now.IsZero() {
		return nil, ErrInvalidProjection
	}
	catalog := bundle.Economy
	attendedMS, err := production.ResolveRateProjectionAttendedMS(bundle, state, now)
	if err != nil {
		return nil, err
	}
	rates, err := production.ProjectRates(bundle, state, contributions, attendedMS)
	if err != nil {
		return nil, err
	}
	resources, err := resourceRows(catalog, state, rates.Resources)
	if err != nil {
		return nil, err
	}
	generators, err := generatorRows(catalog, state, rates.Generators)
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
	progress := []progressRow{}
	if _, exists := catalog.ProgressCoordinate(int(state.Tier)); exists {
		value, progressErr := production.SubProgressValue(catalog, state, int(state.Tier))
		if progressErr != nil {
			return nil, progressErr
		}
		progress = append(progress, progressRow{Current: value.String(), StageID: "progress.tier", Target: "1e0"})
	}
	features, err := projectFeatures(bundle, state, founder, now, minigameActive, attendedMS)
	if err != nil {
		return nil, err
	}
	if features.Reputation, err = projectReputation(bundle, founder, contributions); err != nil {
		return nil, err
	}
	if features.AxisStack, err = projectAxisStack(bundle, state); err != nil {
		return nil, err
	}
	facts := append([]factRow{{FactID: "bootstrap.needed", Value: false}, {FactID: "run.pre_timer", Value: state.RunPreTimer}}, featureFacts(features)...)
	for _, gate := range bundle.Routes.Gates() {
		facts = append(facts, factRow{FactID: gate.ID, Value: state.GatesCrossed[gate.ID]})
	}
	sort.Slice(facts, func(left, right int) bool { return facts[left].FactID < facts[right].FactID })
	transitionPreview, err := previewPhaseATransitions(bundle, state, founder, save.Revision{
		OwnerID: founderID, Number: revision, ConstantsHash: bundle.ConstantsHash,
	}, now, contributions, minigameActive)
	if err != nil {
		return nil, err
	}
	transitions := transitionRows{WindDown: eligibilityTransition{Eligible: transitionPreview.WindDown}}
	if transitionPreview.CrossGate != nil {
		transitions.CrossGate = &crossGateTransition{Eligible: transitionPreview.CrossGate.Eligible,
			GateID: transitionPreview.CrossGate.GateID, RouteID: nil}
	}
	if transitionPreview.Incorporate != nil {
		transitions.Incorporate = &incorporateTransition{Factions: make([]incorporateFaction, 0, len(transitionPreview.Incorporate))}
		for _, factionID := range transitionPreview.Incorporate {
			row, _ := bundle.Faction.Faction(factionID)
			transitions.Incorporate.Factions = append(transitions.Incorporate.Factions, incorporateFaction{CopyKey: row.IncorporationCopyKey, FactionID: row.ID})
		}
	}
	result := snapshot{
		ConstantsHash:   bundle.ConstantsHash,
		Facts:           facts,
		Features:        features,
		FounderRevision: founderRevision,
		Generators:      generators,
		ManualAction: manualActionRow{ActionID: manualActions[0].ID, BucketCapMilli: policy.BucketCapMilli,
			RefilledAtMS: state.ManualTokenRefilledAt.UnixMilli(), RefillMilliPerMS: policy.RefillMilliPerMS,
			TokensMilli: state.ManualTokenMilli},
		Progress: progress, Resources: resources, Revision: revision,
		Run: runRow{Category: "any_percent", ExitCount: int64(len(founder.ExitHistory)), FounderID: founderID,
			RunSeq: state.RunSeq, RunStartedAtMS: state.RunStartedAt.UnixMilli(), Tier: state.Tier},
		EvaluatedThroughMS: state.EvaluatedThrough.UnixMilli(), SchemaVersion: snapshotSchemaVersion,
		ServerNowMS: save.CanonicalServerTime(now).UnixMilli(), Transitions: transitions, Upgrades: upgrades,
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func resourceRows(catalog *economy.Catalog, state *save.State, rates []production.ResourceRate) ([]resourceRow, error) {
	byResource := make(map[string]decimal.Decimal, len(rates))
	for _, rate := range rates {
		byResource[rate.ResourceID] = rate.Rate
	}
	result := []resourceRow{}
	for _, resource := range catalog.Resources() {
		if resource.Scope != economy.ScopeCompany {
			continue
		}
		balance, ok := state.Ledger.Balance(resource.ID)
		if !ok {
			return nil, ErrInvalidProjection
		}
		rate, ok := byResource[resource.ID]
		if !ok {
			return nil, ErrInvalidProjection
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

func generatorRows(catalog *economy.Catalog, state *save.State, rates []production.GeneratorRate) ([]generatorRow, error) {
	definitions := catalog.GeneratorClassesForScope(economy.ScopeCompany)
	byGenerator := make(map[string]decimal.Decimal, len(rates))
	for _, rate := range rates {
		byGenerator[rate.GeneratorID] = rate.Rate
	}
	result := make([]generatorRow, 0, len(definitions))
	for _, generator := range definitions {
		owned, provisioned := state.GeneratorCounts[generator.ID], state.GeneratorProvisioned[generator.ID]
		rate, ok := byGenerator[generator.ID]
		if !ok {
			return nil, ErrInvalidProjection
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
		var provisionCap *intCap
		if generator.ProvisionedHardcap != nil {
			provisionCap = &intCap{Amount: generator.ProvisionedHardcap.Count, ReasonKey: generator.ProvisionedHardcap.ReasonKey}
		}
		result = append(result, generatorRow{GeneratorID: generator.ID, MaxAffordable: affordable.Count, ProvisionCap: provisionCap,
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
		if eligible && upgrade.AxisMinimum > 0 {
			x, _, axisErr := production.AxisInput(state, catalog)
			if axisErr != nil {
				return nil, axisErr
			}
			eligible = x >= upgrade.AxisMinimum
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
