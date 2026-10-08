package pet

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestIdentityJSONRequiresEveryNonNullField(t *testing.T) {
	want := Identity{SpeciesID: "pet_species.server_room_cat", Temperament: "curious", PaletteID: "pet_palette.fur_03",
		NameKey: "pet.name.server_room_cat.n07", AdoptedAtMS: 0, AdoptedAtAttendedMS: 0}
	encoded, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var restored Identity
	if err := json.Unmarshal(encoded, &restored); err != nil || restored != want {
		t.Fatalf("explicit zero coordinates must round-trip: %+v %v", restored, err)
	}
	var source map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &source); err != nil {
		t.Fatal(err)
	}
	for key := range source {
		for _, null := range []bool{false, true} {
			name := key + "/omitted"
			if null {
				name = key + "/null"
			}
			t.Run(name, func(t *testing.T) {
				candidate := make(map[string]json.RawMessage, len(source))
				for name, value := range source {
					candidate[name] = value
				}
				if null {
					candidate[key] = json.RawMessage(`null`)
				} else {
					delete(candidate, key)
				}
				data, err := json.Marshal(candidate)
				if err != nil {
					t.Fatal(err)
				}
				value := want
				if err := json.Unmarshal(data, &value); !errors.Is(err, ErrInvalidIdentity) || value != want {
					t.Fatalf("invalid identity admitted or mutated destination: %+v %v", value, err)
				}
			})
		}
	}
	for name, raw := range map[string]string{
		"null object": `null`, "array": `[]`, "scalar": `1`,
		"wrong type":  `{"species_id":false,"temperament":"curious","palette_id":"pet_palette.fur_03","name_key":"pet.name.server_room_cat.n07","adopted_at_ms":0,"adopted_at_attended_ms":0}`,
		"wrong case":  `{"Species_id":"pet_species.server_room_cat","temperament":"curious","palette_id":"pet_palette.fur_03","name_key":"pet.name.server_room_cat.n07","adopted_at_ms":0,"adopted_at_attended_ms":0}`,
		"extra field": string(encoded[:len(encoded)-1]) + `,"extra":1}`,
	} {
		t.Run(name, func(t *testing.T) {
			value := want
			if err := json.Unmarshal([]byte(raw), &value); !errors.Is(err, ErrInvalidIdentity) || value != want {
				t.Fatalf("invalid identity admitted or mutated destination: %+v %v", value, err)
			}
		})
	}
}
