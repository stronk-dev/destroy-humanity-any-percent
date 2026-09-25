// Package reputation owns the Founder Reputation tree artifact and its pure
// accounting and bonus arithmetic (rfc/reputation-tree-v1.md R1–R3). It has no
// store or production imports.
package reputation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"sort"
	"strings"

	"cloud-clicker/server/curriculum"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
)

const (
	SchemaVersion = 1
	MaxNodes      = 64
	PPM           = int64(1_000_000)
	Provider      = "reputation_tree"

	KindBonusUnlock = "bonus_unlock"
	KindStarter     = "starter"
)

var (
	ErrInvalidTree   = errors.New("invalid reputation tree")
	ErrInvalidState  = errors.New("invalid reputation state")
	idPattern        = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
	bonusDenominator = decimal.FromFloat64(1e12)
)

type Bonus struct {
	PerLevelPPM int64
	SourceID    string
	Slot        multiplier.Slot
	Target      string
}

// Node is one tree row. UnlockPPM is set only for bonus_unlock rows; Starter
// only for starter rows (byte-for-byte the curriculum starter union).
type Node struct {
	NodeID    string
	Kind      string
	Cost      int64
	Requires  []string
	UnlockPPM int64
	Starter   *curriculum.StarterPackage
	TitleKey  string
	BodyKey   string
}

type Tree struct {
	Bonus Bonus
	nodes []Node
	byID  map[string]int
}

// Declarations are the pinned cross-references a tree is validated against.
// Curriculum may be nil (no curriculum artifact in the bundle).
type Declarations struct {
	Economy    *economy.Catalog
	Curriculum *curriculum.Catalog
	CopyKeys   map[string]struct{}
}

type rawTree struct {
	SchemaVersion *int              `json:"schema_version"`
	Bonus         *rawBonus         `json:"bonus"`
	Nodes         []json.RawMessage `json:"nodes"`
}

type rawBonus struct {
	PerLevelPPM *int64 `json:"per_level_ppm"`
	SourceID    string `json:"source_id"`
	Slot        string `json:"slot"`
	Target      string `json:"target"`
}

type rawNode struct {
	NodeID    string                     `json:"node_id"`
	Kind      string                     `json:"kind"`
	Cost      *int64                     `json:"cost"`
	Requires  []string                   `json:"requires"`
	UnlockPPM *int64                     `json:"unlock_ppm"`
	Starter   *curriculum.StarterPackage `json:"starter"`
	TitleKey  string                     `json:"title_key"`
	BodyKey   string                     `json:"body_key"`
}

var nodeKeys = map[string][]string{
	KindBonusUnlock: {"body_key", "cost", "kind", "node_id", "requires", "title_key", "unlock_ppm"},
	KindStarter:     {"body_key", "cost", "kind", "node_id", "requires", "starter", "title_key"},
}

// LoadTree strictly decodes and validates a reputation_tree artifact (R2
// rules 1–8). Every rejection wraps ErrInvalidTree and names its rule.
func LoadTree(data []byte, declarations Declarations) (*Tree, error) {
	if declarations.Economy == nil || len(declarations.CopyKeys) == 0 {
		return nil, fmt.Errorf("%w: missing declarations", ErrInvalidTree)
	}
	var raw rawTree
	if err := decodeStrict(data, &raw); err != nil {
		return nil, fmt.Errorf("%w: decode: %v", ErrInvalidTree, err)
	}
	if raw.SchemaVersion == nil || *raw.SchemaVersion != SchemaVersion || raw.Bonus == nil || raw.Bonus.PerLevelPPM == nil || raw.Nodes == nil {
		return nil, fmt.Errorf("%w: rule 1: schema", ErrInvalidTree)
	}
	bonus := Bonus{*raw.Bonus.PerLevelPPM, raw.Bonus.SourceID, multiplier.Slot(raw.Bonus.Slot), raw.Bonus.Target}
	if bonus.PerLevelPPM < 1 || bonus.PerLevelPPM > PPM || !validDeclaration(declarations.Economy, bonus) {
		return nil, fmt.Errorf("%w: rule 1: bonus declaration", ErrInvalidTree)
	}
	if len(raw.Nodes) < 1 || len(raw.Nodes) > MaxNodes {
		return nil, fmt.Errorf("%w: rule 2: node count", ErrInvalidTree)
	}
	tree := &Tree{Bonus: bonus, byID: make(map[string]int, len(raw.Nodes))}
	previousBonus := -1
	for index, source := range raw.Nodes {
		node, err := decodeNode(source)
		if err != nil {
			return nil, fmt.Errorf("%w: nodes[%d]: %v", ErrInvalidTree, index, err)
		}
		if !idPattern.MatchString(node.NodeID) || !strings.HasPrefix(node.NodeID, "reputation.") {
			return nil, fmt.Errorf("%w: rule 3: nodes[%d] id", ErrInvalidTree, index)
		}
		if _, duplicate := tree.byID[node.NodeID]; duplicate {
			return nil, fmt.Errorf("%w: rule 3: nodes[%d] duplicate id", ErrInvalidTree, index)
		}
		if node.Cost < 1 || node.Cost > decimal.MaxExactInteger {
			return nil, fmt.Errorf("%w: rule 4: nodes[%d] cost", ErrInvalidTree, index)
		}
		for position, requirement := range node.Requires {
			if position > 0 && node.Requires[position-1] >= requirement {
				return nil, fmt.Errorf("%w: rule 5: nodes[%d] requires unsorted", ErrInvalidTree, index)
			}
			if _, earlier := tree.byID[requirement]; !earlier || requirement == node.NodeID {
				return nil, fmt.Errorf("%w: rule 5: nodes[%d] requires a later or unknown row", ErrInvalidTree, index)
			}
		}
		switch node.Kind {
		case KindBonusUnlock:
			if node.UnlockPPM < 1 || node.UnlockPPM > PPM {
				return nil, fmt.Errorf("%w: rule 6: nodes[%d] unlock ppm", ErrInvalidTree, index)
			}
			if previousBonus >= 0 {
				prior := tree.nodes[previousBonus]
				if node.UnlockPPM <= prior.UnlockPPM || !contains(node.Requires, prior.NodeID) {
					return nil, fmt.Errorf("%w: rule 6: nodes[%d] unlock ladder", ErrInvalidTree, index)
				}
			}
			previousBonus = len(tree.nodes)
		case KindStarter:
			if err := curriculum.ValidateStarter(*node.Starter, declarations.Economy); err != nil {
				return nil, fmt.Errorf("%w: rule 7: nodes[%d] starter", ErrInvalidTree, index)
			}
		}
		if !known(declarations.CopyKeys, node.TitleKey) || !known(declarations.CopyKeys, node.BodyKey) {
			return nil, fmt.Errorf("%w: rule 8: nodes[%d] copy key", ErrInvalidTree, index)
		}
		tree.byID[node.NodeID] = len(tree.nodes)
		tree.nodes = append(tree.nodes, node)
	}
	if previousBonus < 0 || tree.nodes[previousBonus].UnlockPPM != PPM {
		return nil, fmt.Errorf("%w: rule 6: unlock ladder must end at 1000000", ErrInvalidTree)
	}
	if err := validateHeadroom(tree, declarations); err != nil {
		return nil, err
	}
	return tree, nil
}

func decodeNode(source json.RawMessage) (Node, error) {
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(source, &keys); err != nil {
		return Node{}, err
	}
	var kind string
	if err := json.Unmarshal(keys["kind"], &kind); err != nil {
		return Node{}, errors.New("kind")
	}
	expected, ok := nodeKeys[kind]
	if !ok || len(keys) != len(expected) {
		return Node{}, errors.New("keys")
	}
	for _, key := range expected {
		if _, present := keys[key]; !present {
			return Node{}, errors.New("keys")
		}
	}
	var raw rawNode
	if err := decodeStrict(source, &raw); err != nil {
		return Node{}, err
	}
	if raw.Cost == nil || raw.Requires == nil || kind == KindBonusUnlock && raw.UnlockPPM == nil || kind == KindStarter && raw.Starter == nil {
		return Node{}, errors.New("fields")
	}
	node := Node{NodeID: raw.NodeID, Kind: kind, Cost: *raw.Cost, Requires: append([]string{}, raw.Requires...), Starter: raw.Starter, TitleKey: raw.TitleKey, BodyKey: raw.BodyKey}
	if raw.UnlockPPM != nil {
		node.UnlockPPM = *raw.UnlockPPM
	}
	return node, nil
}

// validateHeadroom is R2 rule 7's aggregate: applying every starter (the
// curriculum's largest grant plus every tree grant) can never exceed a cap.
func validateHeadroom(tree *Tree, declarations Declarations) error {
	grants := map[string]decimal.Decimal{}
	generated := map[string]*big.Int{}
	upgrades := map[string]bool{}
	for _, node := range tree.nodes {
		if node.Kind != KindStarter {
			continue
		}
		starter := node.Starter
		switch starter.Kind {
		case "resource_grant":
			amount, _ := decimal.ParseCanonical(starter.Amount)
			if prior, ok := grants[starter.ResourceID]; ok {
				amount = prior.Add(amount)
			}
			grants[starter.ResourceID] = amount
		case "generated_generators":
			if generated[starter.GeneratorID] == nil {
				generated[starter.GeneratorID] = new(big.Int)
			}
			generated[starter.GeneratorID].Add(generated[starter.GeneratorID], big.NewInt(starter.Count))
		case "preowned_upgrade":
			if upgrades[starter.UpgradeID] {
				return fmt.Errorf("%w: rule 7: duplicate preowned upgrade %s", ErrInvalidTree, starter.UpgradeID)
			}
			upgrades[starter.UpgradeID] = true
		}
	}
	curriculumGrant := map[string]decimal.Decimal{}
	curriculumCount := map[string]int64{}
	if declarations.Curriculum != nil {
		for _, branch := range declarations.Curriculum.FirstFailure.Branches {
			starter := branch.StarterPackage
			switch starter.Kind {
			case "resource_grant":
				amount, _ := decimal.ParseCanonical(starter.Amount)
				if prior, ok := curriculumGrant[starter.ResourceID]; !ok || amount.Gt(prior) {
					curriculumGrant[starter.ResourceID] = amount
				}
			case "generated_generators":
				if starter.Count > curriculumCount[starter.GeneratorID] {
					curriculumCount[starter.GeneratorID] = starter.Count
				}
			}
		}
	}
	for _, resourceID := range sortedKeys(grants) {
		resource, _ := declarations.Economy.Resource(resourceID)
		total := resource.Initial.Add(grants[resourceID])
		if extra, ok := curriculumGrant[resourceID]; ok {
			total = total.Add(extra)
		}
		if resource.Hardcap != nil && total.Gt(resource.Hardcap.Amount) || !total.IsStateValue() {
			return fmt.Errorf("%w: rule 7: resource %s exceeds its hardcap in aggregate", ErrInvalidTree, resourceID)
		}
	}
	for _, generatorID := range sortedKeys(generated) {
		generator, _ := declarations.Economy.GeneratorClass(generatorID)
		limit := decimal.MaxExactInteger
		if generator.ProvisionedHardcap != nil {
			limit = generator.ProvisionedHardcap.Count
		}
		total := new(big.Int).Add(generated[generatorID], big.NewInt(curriculumCount[generatorID]))
		if total.Cmp(big.NewInt(limit)) > 0 {
			return fmt.Errorf("%w: rule 7: generator %s exceeds its provisioned hardcap in aggregate", ErrInvalidTree, generatorID)
		}
	}
	return nil
}

func validDeclaration(catalog *economy.Catalog, bonus Bonus) bool {
	if !idPattern.MatchString(bonus.SourceID) || bonus.Slot != multiplier.Slot("prestige") || bonus.Target != "all" {
		return false
	}
	declaration, ok := catalog.MultiplierSource(bonus.SourceID)
	return ok && multiplier.Slot(declaration.Slot) == bonus.Slot && declaration.Target == bonus.Target && declaration.Provider == Provider
}

func (tree *Tree) Nodes() []Node {
	if tree == nil {
		return nil
	}
	result := make([]Node, len(tree.nodes))
	for index, node := range tree.nodes {
		result[index] = node
		result[index].Requires = append([]string{}, node.Requires...)
	}
	return result
}

func (tree *Tree) Node(id string) (Node, bool) {
	if tree == nil {
		return Node{}, false
	}
	index, ok := tree.byID[id]
	if !ok {
		return Node{}, false
	}
	node := tree.nodes[index]
	node.Requires = append([]string{}, node.Requires...)
	return node, true
}

// Available is R1's derived balance: level − spent, never persisted.
func Available(level, spent int64) (int64, error) {
	if level < 0 || level > decimal.MaxExactInteger || spent < 0 || spent > level {
		return 0, ErrInvalidState
	}
	return level - spent, nil
}

// UnlockPPM is R1's derived unlock: the largest unlock_ppm among owned
// bonus_unlock nodes known to this tree, or 0. Owned ids must be byte-sorted
// and unique; ids unknown to the tree contribute nothing (OD-7).
func (tree *Tree) UnlockPPM(owned []string) (int64, error) {
	if tree == nil || !SortedUnique(owned) {
		return 0, ErrInvalidState
	}
	var result int64
	for _, id := range owned {
		if node, ok := tree.Node(id); ok && node.Kind == KindBonusUnlock && node.UnlockPPM > result {
			result = node.UnlockPPM
		}
	}
	return result, nil
}

// BonusFactor is R3: 1 + level × per_level_ppm × unlock_ppm / 1e12, quantized
// once. It is computed from the earned level, never from the available balance
// (level − spent); spent is validated only so a caller cannot pass available.
func BonusFactor(level, spent, perLevelPPM, unlockPPM int64) (decimal.Decimal, error) {
	if _, err := Available(level, spent); err != nil || perLevelPPM < 1 || perLevelPPM > PPM || unlockPPM < 0 || unlockPPM > PPM {
		return decimal.NaN, ErrInvalidState
	}
	numerator := new(big.Int).Mul(big.NewInt(level), big.NewInt(perLevelPPM))
	numerator.Mul(numerator, big.NewInt(unlockPPM))
	value := decimal.FromString(numerator.String()).Div(bonusDenominator).Add(decimal.One).Quantize(decimal.CanonicalSignificantDigits)
	if !value.IsStateValue() || value.Lt(decimal.One) {
		return decimal.NaN, ErrInvalidState
	}
	return value, nil
}

func (tree *Tree) BonusFactor(level, spent, unlockPPM int64) (decimal.Decimal, error) {
	if tree == nil {
		return decimal.NaN, ErrInvalidState
	}
	return BonusFactor(level, spent, tree.Bonus.PerLevelPPM, unlockPPM)
}

// SortedUnique reports whether ids are mechanical, byte-sorted, and unique.
func SortedUnique(ids []string) bool {
	for index, id := range ids {
		if !idPattern.MatchString(id) || index > 0 && ids[index-1] >= id {
			return false
		}
	}
	return true
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func known(values map[string]struct{}, value string) bool {
	_, ok := values[value]
	return ok && idPattern.MatchString(value)
}

func sortedKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func decodeStrict(data []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

// Rejection is a typed, recorded purchase refusal (R5 steps 4–7).
type Rejection struct {
	Category string
	Detail   string
}

// PurchaseResult is an applied purchase (R5 step 8).
type PurchaseResult struct {
	Node           Node
	SpentAfter     int64
	OwnedAfter     []string
	UnlockPPMAfter int64
}

// Purchase evaluates R5 steps 4–8 against Founder accounting. Exactly one of
// the result and the rejection is meaningful when err is nil; err reports an
// invalid input state, never a player-facing refusal.
func (tree *Tree) Purchase(level, spent int64, owned []string, nodeID string) (PurchaseResult, *Rejection, error) {
	available, err := Available(level, spent)
	if tree == nil || err != nil || !SortedUnique(owned) {
		return PurchaseResult{}, nil, ErrInvalidState
	}
	node, ok := tree.Node(nodeID)
	if !ok {
		return PurchaseResult{}, &Rejection{Category: "unknown_id", Detail: nodeID}, nil
	}
	if contains(owned, nodeID) {
		return PurchaseResult{}, &Rejection{Category: "not_eligible", Detail: "owned"}, nil
	}
	for _, requirement := range node.Requires {
		if !contains(owned, requirement) {
			return PurchaseResult{}, &Rejection{Category: "not_eligible", Detail: "requires"}, nil
		}
	}
	if node.Cost > available {
		return PurchaseResult{}, &Rejection{Category: "unaffordable", Detail: "reputation"}, nil
	}
	after := append(append([]string{}, owned...), nodeID)
	sort.Strings(after)
	unlock, err := tree.UnlockPPM(after)
	if err != nil {
		return PurchaseResult{}, nil, err
	}
	return PurchaseResult{Node: node, SpentAfter: spent + node.Cost, OwnedAfter: after, UnlockPPMAfter: unlock}, nil, nil
}

// OwnedStarters returns the owned starter nodes known to this tree, in tree
// array order (R4 step 4). Unknown owned ids are ignored (OD-7).
func (tree *Tree) OwnedStarters(owned []string) ([]Node, error) {
	if tree == nil || !SortedUnique(owned) {
		return nil, ErrInvalidState
	}
	result := []Node{}
	for _, node := range tree.nodes {
		if node.Kind == KindStarter && contains(owned, node.NodeID) {
			result = append(result, node)
		}
	}
	return result, nil
}
