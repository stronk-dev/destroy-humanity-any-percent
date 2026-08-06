package harness

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"sort"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/routes"
)

const RelevancePolicySchemaVersion = 1
const relevanceMaxSafeInteger = int64(9_007_199_254_740_991)

var relevanceIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
var relevanceHashPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

type RelevanceWindow struct {
	FromGate string  `json:"from_gate"`
	ToGate   *string `json:"to_gate"`
}

type RelevancePolicyItem struct {
	PurchasableID    string          `json:"purchasable_id"`
	Availability     RelevanceWindow `json:"availability_window"`
	EpsilonMS        int64           `json:"epsilon_ms"`
	TrapExempt       bool            `json:"trap_exempt"`
	JustificationKey *string         `json:"justification_key"`
	GroupIDs         []string        `json:"group_ids"`
}

type RelevancePolicyGroup struct {
	GroupID   string   `json:"group_id"`
	Axis      string   `json:"axis"`
	MemberIDs []string `json:"member_ids"`
	EpsilonMS int64    `json:"epsilon_ms"`
}

type RelevancePolicy struct {
	SchemaVersion int                    `json:"schema_version"`
	Items         []RelevancePolicyItem  `json:"items"`
	Groups        []RelevancePolicyGroup `json:"groups"`
	Hash          string                 `json:"-"`
}

type rawRelevancePolicy struct {
	SchemaVersion *json.Number              `json:"schema_version"`
	Items         []rawRelevancePolicyItem  `json:"items"`
	Groups        []rawRelevancePolicyGroup `json:"groups"`
}

type rawRelevanceWindow struct {
	FromGate *string         `json:"from_gate"`
	ToGate   json.RawMessage `json:"to_gate"`
}

type rawRelevancePolicyItem struct {
	PurchasableID    *string             `json:"purchasable_id"`
	Availability     *rawRelevanceWindow `json:"availability_window"`
	EpsilonMS        *json.Number        `json:"epsilon_ms"`
	TrapExempt       *bool               `json:"trap_exempt"`
	JustificationKey json.RawMessage     `json:"justification_key"`
	GroupIDs         []string            `json:"group_ids"`
}

type rawRelevancePolicyGroup struct {
	GroupID   *string      `json:"group_id"`
	Axis      *string      `json:"axis"`
	MemberIDs []string     `json:"member_ids"`
	EpsilonMS *json.Number `json:"epsilon_ms"`
}

func LoadRelevancePolicy(data []byte, catalog *economy.Catalog, routeCatalog *routes.Catalog) (*RelevancePolicy, error) {
	if catalog == nil || routeCatalog == nil {
		return nil, errors.New("relevance policy requires economy and routes catalogs")
	}
	var raw rawRelevancePolicy
	if err := decodeRelevanceStrict(data, &raw); err != nil {
		return nil, fmt.Errorf("relevance policy: %w", err)
	}
	schemaVersion, schemaErr := relevanceSafeInteger(raw.SchemaVersion)
	if schemaErr != nil || schemaVersion != RelevancePolicySchemaVersion || raw.Items == nil || raw.Groups == nil {
		return nil, errors.New("relevance policy: missing or unsupported envelope")
	}
	policy := &RelevancePolicy{SchemaVersion: int(schemaVersion), Items: make([]RelevancePolicyItem, 0, len(raw.Items)), Groups: make([]RelevancePolicyGroup, 0, len(raw.Groups))}
	for index, source := range raw.Items {
		item, err := parseRelevanceItem(source)
		if err != nil {
			return nil, fmt.Errorf("relevance policy items[%d]: %w", index, err)
		}
		policy.Items = append(policy.Items, item)
	}
	for index, source := range raw.Groups {
		group, err := parseRelevanceGroup(source)
		if err != nil {
			return nil, fmt.Errorf("relevance policy groups[%d]: %w", index, err)
		}
		policy.Groups = append(policy.Groups, group)
	}
	if err := validateRelevancePolicy(policy, catalog, routeCatalog); err != nil {
		return nil, err
	}
	digest := sha256.Sum256(data)
	policy.Hash = "sha256:" + hex.EncodeToString(digest[:])
	return policy, nil
}

func parseRelevanceItem(source rawRelevancePolicyItem) (RelevancePolicyItem, error) {
	if source.PurchasableID == nil || source.Availability == nil || source.EpsilonMS == nil || source.TrapExempt == nil || source.JustificationKey == nil || source.GroupIDs == nil {
		return RelevancePolicyItem{}, errors.New("missing required key")
	}
	if source.Availability.FromGate == nil || source.Availability.ToGate == nil {
		return RelevancePolicyItem{}, errors.New("availability_window missing required key")
	}
	toGate, err := nullableString(source.Availability.ToGate)
	if err != nil {
		return RelevancePolicyItem{}, fmt.Errorf("to_gate: %w", err)
	}
	justification, err := nullableString(source.JustificationKey)
	if err != nil {
		return RelevancePolicyItem{}, fmt.Errorf("justification_key: %w", err)
	}
	epsilon, err := relevanceSafeInteger(source.EpsilonMS)
	if err != nil {
		return RelevancePolicyItem{}, fmt.Errorf("epsilon_ms: %w", err)
	}
	return RelevancePolicyItem{PurchasableID: *source.PurchasableID,
		Availability: RelevanceWindow{FromGate: *source.Availability.FromGate, ToGate: toGate},
		EpsilonMS:    epsilon, TrapExempt: *source.TrapExempt, JustificationKey: justification,
		GroupIDs: append([]string(nil), source.GroupIDs...)}, nil
}

func parseRelevanceGroup(source rawRelevancePolicyGroup) (RelevancePolicyGroup, error) {
	if source.GroupID == nil || source.Axis == nil || source.MemberIDs == nil || source.EpsilonMS == nil {
		return RelevancePolicyGroup{}, errors.New("missing required key")
	}
	epsilon, err := relevanceSafeInteger(source.EpsilonMS)
	if err != nil {
		return RelevancePolicyGroup{}, fmt.Errorf("epsilon_ms: %w", err)
	}
	return RelevancePolicyGroup{GroupID: *source.GroupID, Axis: *source.Axis,
		MemberIDs: append([]string(nil), source.MemberIDs...), EpsilonMS: epsilon}, nil
}

func relevanceSafeInteger(value *json.Number) (int64, error) {
	if value == nil {
		return 0, errors.New("missing integer")
	}
	rational, ok := new(big.Rat).SetString(value.String())
	if !ok || !rational.IsInt() || !rational.Num().IsInt64() {
		return 0, errors.New("expected exact integer")
	}
	result := rational.Num().Int64()
	if result < 0 || result > relevanceMaxSafeInteger {
		return 0, errors.New("integer outside exact range")
	}
	return result, nil
}

func validateRelevancePolicy(policy *RelevancePolicy, catalog *economy.Catalog, routeCatalog *routes.Catalog) error {
	purchasables := make(map[string]bool)
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		purchasables[generator.ID] = true
	}
	for _, upgrade := range catalog.Upgrades() {
		purchasables[upgrade.ID] = true
	}
	if len(policy.Items) != len(purchasables) {
		return fmt.Errorf("relevance policy item set is incomplete: got %d want %d", len(policy.Items), len(purchasables))
	}
	gateOrder := make(map[string]int)
	for index, gate := range routeCatalog.Gates() {
		gateOrder[gate.ID] = index
	}
	itemByID := make(map[string]RelevancePolicyItem, len(policy.Items))
	for index, item := range policy.Items {
		if !relevanceIDPattern.MatchString(item.PurchasableID) || !purchasables[item.PurchasableID] || item.EpsilonMS <= 0 || item.EpsilonMS > relevanceMaxSafeInteger ||
			index > 0 && policy.Items[index-1].PurchasableID >= item.PurchasableID {
			return fmt.Errorf("invalid or unsorted relevance item %q", item.PurchasableID)
		}
		if _, ok := gateOrder[item.Availability.FromGate]; !ok {
			return fmt.Errorf("item %q references unknown from_gate %q", item.PurchasableID, item.Availability.FromGate)
		}
		if item.Availability.ToGate != nil {
			to, ok := gateOrder[*item.Availability.ToGate]
			if !ok || gateOrder[item.Availability.FromGate] >= to {
				return fmt.Errorf("item %q has invalid to_gate", item.PurchasableID)
			}
		}
		if item.TrapExempt != (item.JustificationKey != nil) {
			return fmt.Errorf("item %q violates trap exemption biconditional", item.PurchasableID)
		}
		if item.JustificationKey != nil && !relevanceIDPattern.MatchString(*item.JustificationKey) {
			return fmt.Errorf("item %q has invalid justification key", item.PurchasableID)
		}
		if err := sortedUniqueIDs(item.GroupIDs); err != nil {
			return fmt.Errorf("item %q group_ids: %w", item.PurchasableID, err)
		}
		itemByID[item.PurchasableID] = item
	}
	groupByID := make(map[string]RelevancePolicyGroup, len(policy.Groups))
	axisByMember := make(map[string]map[string]bool)
	for index, group := range policy.Groups {
		if !relevanceIDPattern.MatchString(group.GroupID) || index > 0 && policy.Groups[index-1].GroupID >= group.GroupID ||
			(group.Axis != "tier" && group.Axis != "category" && group.Axis != "declared") || group.EpsilonMS <= 0 || group.EpsilonMS > relevanceMaxSafeInteger {
			return fmt.Errorf("invalid or unsorted relevance group %q", group.GroupID)
		}
		if err := sortedUniqueIDs(group.MemberIDs); err != nil || len(group.MemberIDs) == 0 {
			return fmt.Errorf("group %q has invalid members", group.GroupID)
		}
		for _, id := range group.MemberIDs {
			item, ok := itemByID[id]
			if !ok || !containsSorted(item.GroupIDs, group.GroupID) {
				return fmt.Errorf("group %q has dangling/asymmetric member %q", group.GroupID, id)
			}
			if axisByMember[id] == nil {
				axisByMember[id] = map[string]bool{}
			}
			if axisByMember[id][group.Axis] {
				return fmt.Errorf("item %q has multiple %s groups", id, group.Axis)
			}
			axisByMember[id][group.Axis] = true
		}
		groupByID[group.GroupID] = group
	}
	for _, item := range policy.Items {
		for _, groupID := range item.GroupIDs {
			group, ok := groupByID[groupID]
			if !ok || !containsSorted(group.MemberIDs, item.PurchasableID) {
				return fmt.Errorf("item %q references dangling/asymmetric group %q", item.PurchasableID, groupID)
			}
		}
	}
	for _, group := range policy.Groups {
		if group.Axis == "declared" {
			continue
		}
		var expected []string
		first, ok := catalog.GeneratorClass(group.MemberIDs[0])
		if !ok {
			return fmt.Errorf("derived group %q contains a non-generator", group.GroupID)
		}
		for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
			matches := group.Axis == "tier" && generator.Tier == first.Tier || group.Axis == "category" && generator.Category == first.Category
			if matches {
				expected = append(expected, generator.ID)
			}
		}
		sort.Strings(expected)
		if !equalStrings(expected, group.MemberIDs) {
			return fmt.Errorf("derived group %q is incomplete", group.GroupID)
		}
	}
	return nil
}

func nullableString(data json.RawMessage) (*string, error) {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return nil, nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil || value == "" {
		return nil, errors.New("expected non-empty string or null")
	}
	return &value, nil
}

func sortedUniqueIDs(values []string) error {
	for index, value := range values {
		if !relevanceIDPattern.MatchString(value) || index > 0 && values[index-1] >= value {
			return errors.New("values must be raw-byte sorted unique ids")
		}
	}
	return nil
}

func containsSorted(values []string, target string) bool {
	index := sort.SearchStrings(values, target)
	return index < len(values) && values[index] == target
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func decodeRelevanceStrict(data []byte, destination any) error {
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
