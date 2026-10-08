package save

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"strings"
	"testing"
)

const gardenEventIntent = "01986666-9902-7000-8000-000000000001"

func gardenEventJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func gardenEventDecision(kind EventKind, payload []byte) IntentDecision {
	return IntentDecision{Outcome: IntentApplied, Receipt: []byte(`{"outcome":"applied"}`), Events: []EventWrite{{
		Kind: kind, SchemaVersion: 1, IntentID: gardenEventIntent, Payload: payload,
	}}}
}

// Exercise the actual pre-write validator, including nested SG8 records. These
// are malformed internal producer records, not client-submitted event payloads.
func TestGardenEventFieldAdmission(t *testing.T) {
	hash := "sha256:" + strings.Repeat("0", 64)
	fixtures := []struct {
		kind    EventKind
		payload string
		nested  []string
	}{
		{EventGardenAdvanced, `{"ticks_applied":1,"tick_seq_after":1,"catchup_forfeited_ms":0,"catchup_reason_key":null,"matured":[{"row":0,"col":0,"species_id":"strain_a"}],"spawned":[{"row":0,"col":1,"species_id":"strain_b","recipe_id":"r_b_spread"}]}`, []string{"matured", "spawned"}},
		{EventGardenPlanted, `{"row":0,"col":0,"species_id":"strain_a"}`, nil},
		{EventGardenUprooted, `{"row":0,"col":0,"species_id":"strain_a","was_mature":false}`, nil},
		{EventGardenSubstrateSet, `{"from_substrate_id":"bare_metal","to_substrate_id":"containerized","partial_tick_forfeited_ms":0}`, nil},
		{EventGardenHarvested, `{"plots":[{"row":0,"col":0,"species_id":"strain_a","units":0}],"total_units":0,"seeds_discovered":[],"harvest_hash":"` + hash + `"}`, []string{"plots"}},
		{EventGardenHarvestCredited, `{"harvest_hash":"` + hash + `","selected_units":0,"credited":"0","forfeited_units":0,"cap_reason_key":null,"faucet_applied":false}`, nil},
	}
	for _, fixture := range fixtures {
		t.Run(string(fixture.kind), func(t *testing.T) {
			check := func(name string, raw []byte, valid bool) {
				t.Helper()
				t.Run(name, func(t *testing.T) {
					err := validateIntentDecision(gardenEventDecision(fixture.kind, raw), gardenEventIntent)
					if valid && err != nil || !valid && !errors.Is(err, ErrInvalidStream) {
						t.Fatalf("valid=%v error=%v payload=%s", valid, err, raw)
					}
				})
			}
			check("valid", []byte(fixture.payload), true)
			// JSON escaping changes encoding, not the decoded field name.
			if strings.Contains(fixture.payload, `"row":`) {
				check("valid escaped field", []byte(strings.Replace(fixture.payload, `"row":`, `"r\u006fw":`, 1)), true)
			}
			var root map[string]json.RawMessage
			if err := json.Unmarshal([]byte(fixture.payload), &root); err != nil {
				t.Fatal(err)
			}
			// Mutate a single exact object, then restore it into its parent.
			fields := func(prefix string, object map[string]json.RawMessage, wrap func([]byte) []byte) {
				keys := make([]string, 0, len(object))
				for key := range object {
					keys = append(keys, key)
				}
				sort.Strings(keys)
				for _, key := range keys {
					original := object[key]
					delete(object, key)
					check(prefix+"missing/"+key, wrap(gardenEventJSON(t, object)), false)
					object[key] = json.RawMessage("null")
					check(prefix+"null/"+key, wrap(gardenEventJSON(t, object)), key == "catchup_reason_key" || key == "cap_reason_key")
					delete(object, key)
					object[strings.ToUpper(key)] = original
					check(prefix+"alias/"+key, wrap(gardenEventJSON(t, object)), false)
					delete(object, strings.ToUpper(key))
					object[key] = original
					raw := gardenEventJSON(t, object)
					duplicate := string(raw[:len(raw)-1]) + "," + string(gardenEventJSON(t, key)) + ":" + string(original) + "}"
					check(prefix+"duplicate/"+key, wrap([]byte(duplicate)), false)
				}
			}
			fields("", root, func(raw []byte) []byte { return raw })
			for _, array := range fixture.nested {
				original := root[array]
				var items []map[string]json.RawMessage
				if err := json.Unmarshal(original, &items); err != nil || len(items) != 1 {
					t.Fatalf("nested fixture: %v", err)
				}
				fields(array+"/", items[0], func(raw []byte) []byte {
					root[array] = append(append([]byte("["), raw...), ']')
					return gardenEventJSON(t, root)
				})
				root[array] = original
			}
		})
	}
}

// Preserve producer compatibility using the actual checked-in Go-authored
// receipt/event corpus, including the harvest's Company arm.
func TestGardenProducerEventsRemainAdmitted(t *testing.T) {
	data, err := os.ReadFile("../../testdata/replay/garden-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	type corpusRow struct {
		Name   string `json:"name"`
		Events string `json:"events_json"`
	}
	var corpus struct {
		Cases        []corpusRow `json:"cases"`
		CompanyCases []corpusRow `json:"company_cases"`
	}
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	counts := map[EventKind]int{}
	for _, row := range append(corpus.Cases, corpus.CompanyCases...) {
		var events []struct {
			Kind          EventKind       `json:"kind"`
			SchemaVersion int             `json:"schema_version"`
			IntentID      string          `json:"intent_id"`
			Payload       json.RawMessage `json:"payload"`
		}
		if err := json.Unmarshal([]byte(row.Events), &events); err != nil {
			t.Fatal(err)
		}
		for _, event := range events {
			if !isGardenEvent(event.Kind) {
				continue
			}
			counts[event.Kind]++
			t.Run(row.Name+"/"+string(event.Kind), func(t *testing.T) {
				decision := gardenEventDecision(event.Kind, event.Payload)
				decision.Events[0].SchemaVersion = event.SchemaVersion
				decision.Events[0].IntentID = event.IntentID
				if err := validateIntentDecision(decision, event.IntentID); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
	for kind := range gardenEventKeys {
		if counts[kind] == 0 {
			t.Fatalf("producer corpus never exercises %s", kind)
		}
	}
}
