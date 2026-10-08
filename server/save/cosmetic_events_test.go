package save

import (
	"strings"
	"testing"
)

// AC9 (Go half): the three cosmetic payloads are strict; an extra field, a
// missing replaced_cosmetic_id, or a non-positive order number rejects.
func TestCosmeticEventPayloadsAreStrict(t *testing.T) {
	pet := "01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa"
	for name, row := range map[string]struct {
		kind    EventKind
		payload string
		valid   bool
	}{
		"acquired":                 {EventCosmeticAcquired, `{"cosmetic_id":"horse_armor","order_number":1}`, true},
		"acquired with price":      {EventCosmeticAcquired, `{"cosmetic_id":"horse_armor","order_number":1,"price":0}`, false},
		"acquired order zero":      {EventCosmeticAcquired, `{"cosmetic_id":"horse_armor","order_number":0}`, false},
		"acquired without order":   {EventCosmeticAcquired, `{"cosmetic_id":"horse_armor"}`, false},
		"equipped fresh":           {EventCosmeticEquipped, `{"cosmetic_id":"horse_armor","pet_id":"` + pet + `","replaced_cosmetic_id":null}`, true},
		"equipped replace":         {EventCosmeticEquipped, `{"cosmetic_id":"horse_armor","pet_id":"` + pet + `","replaced_cosmetic_id":"zebra_armor"}`, true},
		"equipped missing replace": {EventCosmeticEquipped, `{"cosmetic_id":"horse_armor","pet_id":"` + pet + `"}`, false},
		"equipped bad pet":         {EventCosmeticEquipped, `{"cosmetic_id":"horse_armor","pet_id":"cat","replaced_cosmetic_id":null}`, false},
		"unequipped":               {EventCosmeticUnequipped, `{"cosmetic_id":"horse_armor","pet_id":"` + pet + `"}`, true},
		"unequipped extra":         {EventCosmeticUnequipped, `{"cosmetic_id":"horse_armor","pet_id":"` + pet + `","amount":0}`, false},
	} {
		err := validateEventPayload(EventWrite{Kind: row.kind, SchemaVersion: 1, IntentID: "01986666-6c01-7000-8000-000000000001", Payload: []byte(row.payload)})
		if row.valid != (err == nil) {
			t.Errorf("%s: valid=%v err=%v", name, row.valid, err)
		}
	}
}

func TestCosmeticEventPayloadKeySpelling(t *testing.T) {
	const pet = "01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa"
	for _, row := range []struct {
		name string
		kind EventKind
		raw  string
		keys []string
	}{
		{"acquired", EventCosmeticAcquired, `{"cosmetic_id":"horse_armor","order_number":1}`, []string{"cosmetic_id", "order_number"}},
		{"equipped", EventCosmeticEquipped, `{"cosmetic_id":"horse_armor","pet_id":"` + pet + `","replaced_cosmetic_id":null}`, []string{"cosmetic_id", "pet_id", "replaced_cosmetic_id"}},
		{"unequipped", EventCosmeticUnequipped, `{"cosmetic_id":"horse_armor","pet_id":"` + pet + `"}`, []string{"cosmetic_id", "pet_id"}},
	} {
		validate := func(raw string) error {
			return validateEventPayload(EventWrite{Kind: row.kind, SchemaVersion: 1,
				IntentID: "01986666-6c01-7000-8000-000000000001", Payload: []byte(raw)})
		}
		t.Run(row.name+"/canonical", func(t *testing.T) {
			if err := validate(row.raw); err != nil {
				t.Fatal(err)
			}
		})
		for _, key := range row.keys {
			t.Run(row.name+"/wrong-case/"+key, func(t *testing.T) {
				raw := strings.Replace(row.raw, `"`+key+`":`, `"`+strings.ToUpper(key)+`":`, 1)
				if err := validate(raw); err == nil {
					t.Fatalf("accepted wrong-case field: %s", raw)
				}
			})
			t.Run(row.name+"/whitespace/"+key, func(t *testing.T) {
				raw := strings.Replace(row.raw, `"`+key+`":`, `"`+key+"\" \n\t: ", 1)
				if err := validate(raw); err != nil {
					t.Fatalf("rejected equivalent JSON whitespace: %v", err)
				}
			})
			t.Run(row.name+"/escaped-key/"+key, func(t *testing.T) {
				raw := strings.Replace(row.raw, `"`+key+`":`, `"`+strings.ReplaceAll(key, "_", `\u005f`)+`":`, 1)
				if err := validate(raw); err != nil {
					t.Fatalf("rejected equivalent decoded key: %v", err)
				}
			})
		}
	}
}
