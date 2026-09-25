// Package cosmetic is Cosmetic Shop v1 (rfc/cosmetic-shop-v1.md): the
// `cosmetics` catalog, the Founder ownership/equip shape, and the three pure
// transitions. It deliberately imports no ledger, multiplier, production,
// faction, fiscal, achievement, Soul, pet-care-transition, or metrics code
// (§5, N3, N8); `make verify-cosmetic-boundary` enforces that.
package cosmetic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
)

const (
	SchemaVersion = 1
	// SlotPet is the only v1 slot (§2); a new slot is a schema change.
	SlotPet = "pet"
	// UnlockActiveCompanyTierAtLeast is the only v1 unlock kind (§2, OD-3).
	UnlockActiveCompanyTierAtLeast = "active_company_tier_at_least"
	maxTier                        = 8
	maxItems                       = 256
)

var (
	ErrInvalidCatalog    = errors.New("invalid cosmetics artifact")
	ErrInvalidTransition = errors.New("invalid cosmetics epoch transition")

	mechanicalPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

// Unlock is the closed v1 unlock predicate.
type Unlock struct {
	Kind string
	Tier int64
}

// Item is one catalog row. There is no price field, by construction (I1, N1).
type Item struct {
	CosmeticID string
	Slot       string
	Unlock     Unlock
}

// Catalog is the loaded `cosmetics` artifact, rows in raw-byte order.
type Catalog struct {
	Items []Item
	byID  map[string]int
}

// Item returns the row for id.
func (catalog *Catalog) Item(id string) (Item, bool) {
	if catalog == nil {
		return Item{}, false
	}
	index, ok := catalog.byID[id]
	if !ok {
		return Item{}, false
	}
	return catalog.Items[index], true
}

// Load decodes and validates a `cosmetics` artifact. Unknown keys at any level
// reject the whole artifact, which is how price/currency/sku/product_id are
// refused: as unknown fields, never by a special case (§2).
func Load(data []byte) (*Catalog, error) {
	var top map[string]json.RawMessage
	if err := exactObject(data, &top, "schema_version", "items"); err != nil {
		return nil, err
	}
	var version int
	if err := strictDecode(top["schema_version"], &version); err != nil || version != SchemaVersion {
		return nil, fmt.Errorf("%w: schema_version must be %d", ErrInvalidCatalog, SchemaVersion)
	}
	var rawItems []json.RawMessage
	if err := strictDecode(top["items"], &rawItems); err != nil || len(rawItems) == 0 || len(rawItems) > maxItems {
		return nil, fmt.Errorf("%w: items must be a non-empty bounded array", ErrInvalidCatalog)
	}
	catalog := &Catalog{Items: make([]Item, 0, len(rawItems)), byID: map[string]int{}}
	for index, rawItem := range rawItems {
		var fields map[string]json.RawMessage
		if err := exactObject(rawItem, &fields, "cosmetic_id", "slot", "unlock"); err != nil {
			return nil, err
		}
		var item Item
		if err := strictDecode(fields["cosmetic_id"], &item.CosmeticID); err != nil || !mechanicalPattern.MatchString(item.CosmeticID) {
			return nil, fmt.Errorf("%w: cosmetic_id is not mechanical", ErrInvalidCatalog)
		}
		if index > 0 && catalog.Items[index-1].CosmeticID >= item.CosmeticID {
			return nil, fmt.Errorf("%w: cosmetic ids must be unique and byte-sorted", ErrInvalidCatalog)
		}
		if err := strictDecode(fields["slot"], &item.Slot); err != nil || item.Slot != SlotPet {
			return nil, fmt.Errorf("%w: unknown slot", ErrInvalidCatalog)
		}
		var unlock map[string]json.RawMessage
		if err := exactObject(fields["unlock"], &unlock, "kind", "tier"); err != nil {
			return nil, err
		}
		if err := strictDecode(unlock["kind"], &item.Unlock.Kind); err != nil || item.Unlock.Kind != UnlockActiveCompanyTierAtLeast {
			return nil, fmt.Errorf("%w: unknown unlock kind", ErrInvalidCatalog)
		}
		if err := strictDecode(unlock["tier"], &item.Unlock.Tier); err != nil || item.Unlock.Tier < 0 || item.Unlock.Tier > maxTier {
			return nil, fmt.Errorf("%w: unlock tier must be an exact integer in [0,%d]", ErrInvalidCatalog, maxTier)
		}
		catalog.byID[item.CosmeticID] = index
		catalog.Items = append(catalog.Items, item)
	}
	return catalog, nil
}

// ValidateTransition is §2's permanent-ID rule: a next epoch's catalog keeps
// every current cosmetic_id with its slot; only unlock may be retuned.
func ValidateTransition(current, next *Catalog) error {
	if current == nil {
		return nil
	}
	if next == nil {
		return fmt.Errorf("%w: cosmetics cannot disappear between epochs", ErrInvalidTransition)
	}
	for _, item := range current.Items {
		successor, ok := next.Item(item.CosmeticID)
		if !ok {
			return fmt.Errorf("%w: cosmetic %q was dropped", ErrInvalidTransition, item.CosmeticID)
		}
		if successor.Slot != item.Slot {
			return fmt.Errorf("%w: cosmetic %q changed slot", ErrInvalidTransition, item.CosmeticID)
		}
	}
	return nil
}

func exactObject(data []byte, target *map[string]json.RawMessage, keys ...string) error {
	if err := strictDecode(data, target); err != nil || *target == nil {
		return fmt.Errorf("%w: expected an object", ErrInvalidCatalog)
	}
	if !uniqueKeys(data) || len(*target) != len(keys) {
		return fmt.Errorf("%w: object keys are not exact", ErrInvalidCatalog)
	}
	for _, key := range keys {
		if _, ok := (*target)[key]; !ok {
			return fmt.Errorf("%w: object keys are not exact", ErrInvalidCatalog)
		}
	}
	return nil
}

func strictDecode(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("%w: trailing data", ErrInvalidCatalog)
	}
	return nil
}

// uniqueKeys rejects duplicate object keys at any depth, which encoding/json
// would otherwise silently collapse.
func uniqueKeys(data []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var walk func() bool
	walk = func() bool {
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		delim, ok := token.(json.Delim)
		if !ok {
			return true
		}
		switch delim {
		case '{':
			seen := map[string]bool{}
			for decoder.More() {
				keyToken, err := decoder.Token()
				key, isString := keyToken.(string)
				if err != nil || !isString || seen[key] {
					return false
				}
				seen[key] = true
				if !walk() {
					return false
				}
			}
			_, err = decoder.Token()
			return err == nil
		case '[':
			for decoder.More() {
				if !walk() {
					return false
				}
			}
			_, err = decoder.Token()
			return err == nil
		}
		return false
	}
	return walk()
}
