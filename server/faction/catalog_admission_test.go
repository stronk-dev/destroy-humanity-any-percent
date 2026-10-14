package faction

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"cloud-clicker/server/commons"
)

func factionFixture(t *testing.T) ([]byte, CompactTitheBand) {
	t.Helper()
	data, err := os.ReadFile("../../balance/factions/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	policy, err := os.ReadFile("../../balance/commons/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	compact, err := commons.LoadCatalog(policy)
	if err != nil {
		t.Fatal(err)
	}
	return data, CompactTitheBand{compact.MinimumTithePPM, compact.DefaultTithePPM, compact.MaximumTithePPM}
}

func editFactionJSON(t *testing.T, source []byte, edit func(map[string]any)) []byte {
	t.Helper()
	var root map[string]any
	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.UseNumber()
	if err := decoder.Decode(&root); err != nil {
		t.Fatal(err)
	}
	edit(root)
	result, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestFactionCatalogRequiresLiteralKeysAndPresentNullableCompact(t *testing.T) {
	data, band := factionFixture(t)
	type field struct {
		path, key string
		row       map[string]any
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	var fields []field
	add := func(path string, row map[string]any) {
		for key := range row {
			fields = append(fields, field{path, key, row})
		}
	}
	add("root", root)
	for _, source := range root["factions"].([]any) {
		row := source.(map[string]any)
		add(row["id"].(string), row)
		if row["compact"] != nil {
			add(row["id"].(string)+"/compact", row["compact"].(map[string]any))
		}
	}
	if len(fields) != 30 {
		t.Fatalf("unexpected literal-key population: %d", len(fields))
	}
	for _, member := range fields {
		t.Run("alias/"+member.path+"/"+member.key, func(t *testing.T) {
			original := member.row[member.key]
			delete(member.row, member.key)
			member.row[strings.ToUpper(member.key)] = original
			defer func() { delete(member.row, strings.ToUpper(member.key)); member.row[member.key] = original }()
			candidate, err := json.Marshal(root)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(candidate, band); err == nil {
				t.Fatal("nonliteral catalog member admitted")
			}
		})
	}
	for _, source := range root["factions"].([]any) {
		row := source.(map[string]any)
		if row["compact"] != nil {
			continue
		}
		t.Run("missing compact/"+row["id"].(string), func(t *testing.T) {
			delete(row, "compact")
			defer func() { row["compact"] = nil }()
			candidate, err := json.Marshal(root)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(candidate, band); err == nil {
				t.Fatal("missing compact defaulted to null")
			}
		})
	}
	if _, err := LoadCatalog(data, band); err != nil {
		t.Fatal("explicit nullable compact rejected", err)
	}
}
