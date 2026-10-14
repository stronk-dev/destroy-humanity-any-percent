package routes

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"
	"testing"
)

func TestSharedConditionsRequireLiteralActiveArm(t *testing.T) {
	for _, source := range []string{
		`{"kind":"resource_at_least","resource_id":"company.cash","value":"0"}`,
		`{"kind":"resource_at_most","resource_id":"company.cash","value":"1e3"}`,
		`{"kind":"meter_band","meter_id":"meter.standing","min":0,"max":100}`,
		`{"kind":"doctrine_is","transition":"transition.t0_to_t1","doctrine_id":"doctrine.a"}`,
		`{"kind":"doctrine_is_not","transition":"transition.t0_to_t1","doctrine_id":"doctrine.a"}`,
		`{"kind":"ledger_fact_present","fact_kind":"fact.example"}`,
		`{"kind":"structure_is","structure_id":"structure.example"}`,
		`{"kind":"region_trait","trait_id":"trait.example"}`,
	} {
		var row map[string]any
		if err := json.Unmarshal([]byte(source), &row); err != nil {
			t.Fatal(err)
		}
		kind := row["kind"].(string)
		if _, err := ParseConditions([]byte("[" + source + "]")); err != nil {
			t.Fatal(kind, err)
		}
		keys := make([]string, 0, len(row))
		for key := range row {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			t.Run(kind+"/alias/"+key, func(t *testing.T) {
				value := row[key]
				delete(row, key)
				row[strings.ToUpper(key)] = value
				defer func() { delete(row, strings.ToUpper(key)); row[key] = value }()
				data, err := json.Marshal([]any{row})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := ParseConditions(data); err == nil {
					t.Fatal("case alias admitted")
				}
			})
		}
		for _, key := range []string{"resource_id", "value", "meter_id", "min", "max", "transition", "doctrine_id", "fact_kind", "structure_id", "trait_id"} {
			if _, present := row[key]; present {
				continue
			}
			t.Run(kind+"/wrong-arm-null/"+key, func(t *testing.T) {
				row[key] = nil
				defer delete(row, key)
				data, err := json.Marshal([]any{row})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := ParseConditions(data); err == nil {
					t.Fatal("wrong-arm null treated as absent")
				}
			})
		}
	}
}

func TestSharedConditionAdmissionAlsoGovernsRouteCatalogAndPreservesDecodedNames(t *testing.T) {
	_, data := loadPhase0(t)
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	changed := false
	for _, gate := range root["gates"].([]any) {
		for _, route := range gate.(map[string]any)["routes"].([]any) {
			for _, predicate := range route.(map[string]any)["predicate"].([]any) {
				row := predicate.(map[string]any)
				row["KIND"] = row["kind"]
				delete(row, "kind")
				changed = true
			}
		}
	}
	if !changed {
		t.Fatal("catalog control had no predicates")
	}
	candidate, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(candidate); err == nil {
		t.Fatal("catalog bypassed shared raw admission")
	}
	original := []byte(`[{"kind":"meter_band","meter_id":"meter.example","min":0,"max":100}]`)
	for _, candidate := range [][]byte{
		bytes.Replace(original, []byte(`"kind"`), []byte(`"ki\u006ed"`), 1),
		bytes.Replace(original, []byte(`"min":0`), []byte(`"min":50,"min":0`), 1),
	} {
		if _, err := ParseConditions(candidate); err != nil {
			t.Fatal("decoded-name/last-value behavior changed", err)
		}
	}
}
