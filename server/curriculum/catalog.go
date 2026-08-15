// Package curriculum owns the declarative first-hour terminal curriculum.
package curriculum

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"sort"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

const SchemaVersion = 1

var (
	ErrInvalidCatalog = errors.New("invalid curriculum catalog")
	mechanicalID      = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

type StarterPackage struct {
	Kind        string `json:"kind"`
	ResourceID  string `json:"resource_id,omitempty"`
	Amount      string `json:"amount,omitempty"`
	GeneratorID string `json:"generator_id,omitempty"`
	Count       int64  `json:"count,omitempty"`
	UpgradeID   string `json:"upgrade_id,omitempty"`
}

type Branch struct {
	Branch                     string         `json:"branch"`
	MinimumPurchasedGenerators int64          `json:"minimum_purchased_generators,omitempty"`
	MinimumOwnedUpgrades       int64          `json:"minimum_owned_upgrades,omitempty"`
	CheapestPriceFactor        string         `json:"cheapest_price_factor,omitempty"`
	RouteKnowledgeBonus        int64          `json:"route_knowledge_bonus"`
	StarterPackage             StarterPackage `json:"starter_package"`
	TitleKey                   string         `json:"title_key"`
	BodyKey                    string         `json:"body_key"`
}

type FirstFailure struct {
	RunSeq                 int64    `json:"run_seq"`
	FounderExitCount       int64    `json:"founder_exit_count"`
	AttendedMS             int64    `json:"attended_ms"`
	GateID                 string   `json:"gate_id"`
	ExitType               string   `json:"exit_type"`
	Evaluation             string   `json:"evaluation"`
	RequestedCommandEffect string   `json:"requested_command_effect"`
	NextRunKey             string   `json:"next_run_key"`
	Branches               []Branch `json:"branches"`
}

type Catalog struct {
	SchemaVersion int          `json:"schema_version"`
	FirstFailure  FirstFailure `json:"first_failure"`
}

type Declarations struct {
	Economy  *economy.Catalog
	CopyKeys map[string]struct{}
	GateIDs  map[string]struct{}
}

func Load(data []byte, declarations Declarations) (*Catalog, error) {
	var catalog Catalog
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&catalog); err != nil {
		return nil, ErrInvalidCatalog
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, ErrInvalidCatalog
	}
	if err := catalog.validate(declarations); err != nil {
		return nil, err
	}
	return &catalog, nil
}

func (catalog *Catalog) validate(declarations Declarations) error {
	if catalog == nil || catalog.SchemaVersion != SchemaVersion || declarations.Economy == nil || len(declarations.CopyKeys) == 0 || len(declarations.GateIDs) == 0 {
		return ErrInvalidCatalog
	}
	failure := catalog.FirstFailure
	if failure.RunSeq != 1 || failure.FounderExitCount != 0 || failure.AttendedMS != 900_000 || failure.ExitType != "scripted_first" ||
		failure.Evaluation != "first_player_company_command_after_accrual" || failure.RequestedCommandEffect != "replaced_by_terminal_transition" ||
		failure.GateID != "gate.t0_to_t1" || !known(declarations.GateIDs, failure.GateID) || !known(declarations.CopyKeys, failure.NextRunKey) || len(failure.Branches) != 3 {
		return ErrInvalidCatalog
	}
	expected := []string{"acquihire", "burnout", "pivot"}
	for index := range failure.Branches {
		branch := failure.Branches[index]
		if branch.Branch != expected[index] || branch.RouteKnowledgeBonus < 0 || branch.RouteKnowledgeBonus > decimal.MaxExactInteger ||
			!known(declarations.CopyKeys, branch.TitleKey) || !known(declarations.CopyKeys, branch.BodyKey) {
			return ErrInvalidCatalog
		}
		switch branch.Branch {
		case "acquihire":
			if branch.MinimumPurchasedGenerators < 1 || branch.MinimumPurchasedGenerators > decimal.MaxExactInteger || branch.MinimumOwnedUpgrades < 1 || branch.MinimumOwnedUpgrades > decimal.MaxExactInteger || branch.CheapestPriceFactor != "" {
				return ErrInvalidCatalog
			}
		case "burnout":
			factor, err := decimal.ParseCanonical(branch.CheapestPriceFactor)
			if err != nil || !factor.Gt(decimal.Zero) || !factor.IsStateValue() || branch.MinimumPurchasedGenerators != 0 || branch.MinimumOwnedUpgrades != 0 {
				return ErrInvalidCatalog
			}
		case "pivot":
			if branch.MinimumPurchasedGenerators != 0 || branch.MinimumOwnedUpgrades != 0 || branch.CheapestPriceFactor != "" {
				return ErrInvalidCatalog
			}
		}
		if err := validateStarter(branch.StarterPackage, declarations.Economy); err != nil {
			return err
		}
	}
	return nil
}

func validateStarter(starter StarterPackage, catalog *economy.Catalog) error {
	switch starter.Kind {
	case "resource_grant":
		amount, err := decimal.ParseCanonical(starter.Amount)
		resource, ok := catalog.Resource(starter.ResourceID)
		if err != nil || !ok || resource.Scope != economy.ScopeCompany || !amount.Gt(decimal.Zero) || !amount.IsStateValue() || starter.GeneratorID != "" || starter.Count != 0 || starter.UpgradeID != "" {
			return ErrInvalidCatalog
		}
	case "generated_generators":
		generator, ok := catalog.GeneratorClass(starter.GeneratorID)
		priceResource, priceOK := catalog.Resource(generator.Price.ResourceID)
		if !ok || !priceOK || priceResource.Scope != economy.ScopeCompany || starter.Count < 1 || starter.Count > decimal.MaxExactInteger || starter.ResourceID != "" || starter.Amount != "" || starter.UpgradeID != "" {
			return ErrInvalidCatalog
		}
	case "preowned_upgrade":
		upgrade, ok := catalog.Upgrade(starter.UpgradeID)
		priceResource, priceOK := catalog.Resource(upgrade.Cost.ResourceID)
		if !ok || !priceOK || priceResource.Scope != economy.ScopeCompany || starter.ResourceID != "" || starter.Amount != "" || starter.GeneratorID != "" || starter.Count != 0 {
			return ErrInvalidCatalog
		}
	default:
		return ErrInvalidCatalog
	}
	return nil
}

func (catalog *Catalog) SelectBranch(state *save.State, economyCatalog *economy.Catalog) (Branch, error) {
	if catalog == nil || state == nil || economyCatalog == nil {
		return Branch{}, ErrInvalidCatalog
	}
	acquihire := catalog.FirstFailure.Branches[0]
	if state.GeneratorPurchasedTotal >= acquihire.MinimumPurchasedGenerators && ownedUpgradeCount(state) >= acquihire.MinimumOwnedUpgrades {
		return acquihire, nil
	}
	burnout := catalog.FirstFailure.Branches[1]
	price, err := cheapestGeneratorPrice(economyCatalog, state)
	if err != nil {
		return Branch{}, err
	}
	factor, _ := decimal.ParseCanonical(burnout.CheapestPriceFactor)
	threshold := price.Mul(factor).Quantize(decimal.CanonicalSignificantDigits)
	cash, ok := state.Ledger.Balance("company.cash")
	if !ok || !threshold.IsStateValue() {
		return Branch{}, ErrInvalidCatalog
	}
	if cash.Lt(threshold) {
		return burnout, nil
	}
	return catalog.FirstFailure.Branches[2], nil
}

func (catalog *Catalog) ApplyStarter(state *save.State, branch Branch) error {
	if catalog == nil || state == nil {
		return ErrInvalidCatalog
	}
	starter := branch.StarterPackage
	switch starter.Kind {
	case "resource_grant":
		amount, err := decimal.ParseCanonical(starter.Amount)
		if err != nil {
			return err
		}
		_, err = state.Ledger.Apply(economy.Transaction{Entries: []economy.Entry{{ResourceID: starter.ResourceID, Delta: amount}}})
		return err
	case "generated_generators":
		state.GeneratorProvisioned[starter.GeneratorID] = starter.Count
		return nil
	case "preowned_upgrade":
		state.UpgradesOwned[starter.UpgradeID] = true
		return nil
	default:
		return ErrInvalidCatalog
	}
}

func cheapestGeneratorPrice(catalog *economy.Catalog, state *save.State) (decimal.Decimal, error) {
	ids := make([]string, 0)
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		ids = append(ids, generator.ID)
	}
	sort.Strings(ids)
	price := decimal.NaN
	for _, id := range ids {
		cost, err := catalog.BulkCost(id, state.GeneratorCounts[id], 1)
		if err != nil {
			return decimal.NaN, err
		}
		if price.IsNaN() || cost.Lt(price) {
			price = cost
		}
	}
	if price.IsNaN() {
		return decimal.NaN, ErrInvalidCatalog
	}
	return price, nil
}

func ownedUpgradeCount(state *save.State) int64 {
	var count int64
	for _, owned := range state.UpgradesOwned {
		if owned {
			count++
		}
	}
	return count
}

func known(values map[string]struct{}, value string) bool {
	_, ok := values[value]
	return ok && mechanicalID.MatchString(value)
}
