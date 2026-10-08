package pet

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cloud-clicker/server/combat"
	"cloud-clicker/server/copykeys"
)

// Preserve literal spelling: JSON decode/re-encode would erase duplicate keys
// and the number tokens on which Go and the TS raw replay loader must agree.
func TestSpeciesRawCatalogAdmission(t *testing.T) {
	var corpus struct {
		Version int    `json:"version"`
		Valid   string `json:"valid"`
		Cases   []struct {
			Name        string `json:"name"`
			Reject      bool   `json:"reject"`
			From        string `json:"from"`
			To          string `json:"to"`
			Occurrences int    `json:"occurrences"`
			Prefix      string `json:"prefix"`
			Suffix      string `json:"suffix"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(readRepo(t, "testdata/pet/species-raw-fixtures-v1.json"), &corpus); err != nil || corpus.Version != 1 || len(corpus.Cases) != 19 {
		t.Fatalf("raw species corpus is incomplete or invalid: %v", err)
	}
	valid := string(readRepo(t, corpus.Valid))
	for _, test := range corpus.Cases {
		t.Run(test.Name, func(t *testing.T) {
			input := valid
			if test.From != "" {
				if count := strings.Count(input, test.From); count != test.Occurrences {
					t.Fatalf("replacement has %d matches, expected %d", count, test.Occurrences)
				}
				input = strings.Replace(input, test.From, test.To, 1)
			}
			catalog, err := LoadSpeciesCatalog([]byte(test.Prefix+input+test.Suffix), speciesDeclarations())
			if test.Reject {
				if err == nil || catalog != nil {
					t.Fatal("invalid raw pet_species artifact admitted")
				}
			} else if err != nil || catalog == nil {
				t.Fatalf("valid raw pet_species artifact refused: %v", err)
			}
		})
	}
}

func speciesDeclarations() SpeciesDeclarations {
	declarations := SpeciesDeclarations{CopyKeys: map[string]struct{}{}, CompanionKeys: map[string]struct{}{}}
	for _, key := range copykeys.All() {
		declarations.CopyKeys[key] = struct{}{}
	}
	for _, key := range copykeys.CompanionKeys() {
		declarations.CompanionKeys[key] = struct{}{}
	}
	return declarations
}

func readRepo(t *testing.T, relative string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", relative))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// AC1: one shared fixture set loads identically in Go and TS.
func TestSpeciesCatalogSharedFixtures(t *testing.T) {
	var corpus struct {
		Cases []struct {
			Name     string          `json:"name"`
			Artifact json.RawMessage `json:"artifact"`
			Valid    bool            `json:"valid"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(readRepo(t, "testdata/pet/species-fixtures-v1.json"), &corpus); err != nil {
		t.Fatal(err)
	}
	if len(corpus.Cases) < 17 {
		t.Fatalf("fixture corpus shrank to %d cases", len(corpus.Cases))
	}
	for _, testCase := range corpus.Cases {
		_, err := LoadSpeciesCatalog(testCase.Artifact, speciesDeclarations())
		if (err == nil) != testCase.Valid {
			t.Errorf("%s: valid=%v err=%v", testCase.Name, testCase.Valid, err)
		}
	}
}

func TestSpeciesFixtureArtifactLoads(t *testing.T) {
	catalog, err := LoadSpeciesCatalog(readRepo(t, "balance/testdata/pet-species/fixture-v1.json"), speciesDeclarations())
	if err != nil {
		t.Fatal(err)
	}
	row, ok := catalog.Row("pet_species.server_room_cat")
	if !ok || catalog.MaxPetsPerFounder != 1 || len(row.NameKeys) != 12 || len(row.PaletteIDs) != 10 || !row.HasName("pet.name.server_room_cat.n07") {
		t.Fatalf("unexpected fixture catalog %#v", catalog)
	}
}

func TestTemperamentOrderIsTheCombatEnum(t *testing.T) {
	for _, value := range TemperamentOrder {
		if _, err := combat.Chart(combat.Temperament(value), combat.Temperament(value)); err != nil {
			t.Fatalf("temperament %q is not a combat temperament: %v", value, err)
		}
	}
	if _, err := combat.Chart(combat.Temperament("grumpy"), combat.Lazy); err == nil {
		t.Fatal("combat accepted an unknown temperament; the parity probe is vacuous")
	}
	want := []combat.Temperament{combat.Lazy, combat.Playful, combat.Curious, combat.Sassy, combat.Shy, combat.Chaotic}
	for index, value := range want {
		if string(value) != TemperamentOrder[index] {
			t.Fatalf("temperament order diverges at %d", index)
		}
	}
}

// AC3: draws are nonce-derived; intent_id never enters any hash input.
func TestAdoptionDrawVectors(t *testing.T) {
	var corpus struct {
		Row struct {
			AllowedTemperaments []string `json:"allowed_temperaments"`
			PaletteIDs          []string `json:"palette_ids"`
		} `json:"row"`
		Vectors []struct {
			FounderID    string `json:"founder_id"`
			Nonce        string `json:"nonce"`
			ServerTimeMS int64  `json:"server_time_ms"`
			IntentID     string `json:"intent_id"`
			PetID        string `json:"pet_id"`
			Temperament  string `json:"temperament"`
			PaletteID    string `json:"palette_id"`
		} `json:"vectors"`
		RestrictedRow struct {
			AllowedTemperaments []string `json:"allowed_temperaments"`
			PaletteIDs          []string `json:"palette_ids"`
			Vector              struct {
				FounderID    string `json:"founder_id"`
				Nonce        string `json:"nonce"`
				ServerTimeMS int64  `json:"server_time_ms"`
				PetID        string `json:"pet_id"`
				Temperament  string `json:"temperament"`
				PaletteID    string `json:"palette_id"`
			} `json:"vector"`
		} `json:"restricted_row"`
	}
	if err := json.Unmarshal(readRepo(t, "testdata/pet/adoption-draw-vectors-v1.json"), &corpus); err != nil {
		t.Fatal(err)
	}
	row := SpeciesRow{AllowedTemperaments: corpus.Row.AllowedTemperaments, PaletteIDs: corpus.Row.PaletteIDs}
	temperaments, palettes := map[string]bool{}, map[string]bool{}
	for _, vector := range corpus.Vectors {
		draws, err := DrawAdoption(vector.FounderID, vector.Nonce, vector.ServerTimeMS, row)
		if err != nil || draws != (AdoptionDraws{PetID: vector.PetID, Temperament: vector.Temperament, PaletteID: vector.PaletteID}) {
			t.Fatalf("vector %s/%d: %#v %v", vector.Nonce, vector.ServerTimeMS, draws, err)
		}
		temperaments[draws.Temperament], palettes[draws.PaletteID] = true, true
	}
	if len(temperaments) != 6 || len(palettes) != 10 {
		t.Fatalf("vectors cover %d temperaments and %d palettes", len(temperaments), len(palettes))
	}
	restricted := corpus.RestrictedRow
	draws, err := DrawAdoption(restricted.Vector.FounderID, restricted.Vector.Nonce, restricted.Vector.ServerTimeMS,
		SpeciesRow{AllowedTemperaments: restricted.AllowedTemperaments, PaletteIDs: restricted.PaletteIDs})
	if err != nil || draws.PetID != restricted.Vector.PetID || draws.Temperament != restricted.Vector.Temperament || draws.PaletteID != restricted.Vector.PaletteID {
		t.Fatalf("restricted vector: %#v %v", draws, err)
	}
}

func TestAdoptionDrawRejectsMalformedInputs(t *testing.T) {
	row := SpeciesRow{AllowedTemperaments: []string{"lazy"}, PaletteIDs: []string{"pet_palette.fur_00"}}
	for _, testCase := range []struct{ founder, nonce string }{
		{"01986666-1111-7111-8111-111111111111", "0123456789ABCDEF0123456789ABCDEF"},
		{"01986666-1111-7111-8111-111111111111", "0123"},
		{"01986666-1111-7111-8111-11111111111", "0123456789abcdef0123456789abcdef"},
		{"01986666-1111-7111-8111-11111111111G", "0123456789abcdef0123456789abcdef"},
	} {
		if _, err := DrawAdoption(testCase.founder, testCase.nonce, 1, row); err == nil {
			t.Fatalf("accepted %q/%q", testCase.founder, testCase.nonce)
		}
	}
}
