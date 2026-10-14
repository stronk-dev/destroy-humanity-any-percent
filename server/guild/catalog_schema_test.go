package guild

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"cloud-clicker/server/publicapi"
)

func guildSchemaDefinitions(t *testing.T) map[string]*publicapi.Schema {
	t.Helper()
	definitions, err := publicapi.ValidateSchemaDefinitions(GuildSchemas())
	if err != nil {
		t.Fatal(err)
	}
	return definitions
}

func TestGuildOwnerSchemaAdmitsCurrentStoredAndExplicitZero(t *testing.T) {
	definitions := guildSchemaDefinitions(t)
	for _, path := range []string{
		"../../balance/guilds/phase0.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789/artifacts/guilds.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6/artifacts/guilds.json",
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		before := bytes.Clone(data)
		if err := publicapi.ValidateJSON(GuildCatalogSchemaName, data, definitions); err != nil {
			t.Fatal(err)
		}
		catalog, err := LoadCatalog(data)
		if err != nil || catalog.ConsumptionBonusPPMPerUnit != 0 {
			t.Fatalf("explicit-zero policy: %v", err)
		}
		if !bytes.Equal(before, data) {
			t.Fatal("validation changed artifact bytes")
		}
	}
}

func TestGuildOwnerSchemaRequiresEachLiteralAndTypedMember(t *testing.T) {
	definitions := guildSchemaDefinitions(t)
	var original map[string]json.RawMessage
	if err := json.Unmarshal([]byte(phase0Catalog), &original); err != nil {
		t.Fatal(err)
	}
	if len(original) != 11 {
		t.Fatal("changed policy population")
	}
	for key := range original {
		for _, variant := range []string{"missing", "null", "drift", "string", "false", "array"} {
			t.Run(key+"/"+variant, func(t *testing.T) {
				candidate := make(map[string]json.RawMessage, len(original))
				for key, value := range original {
					candidate[key] = value
				}
				switch variant {
				case "missing":
					delete(candidate, key)
				case "null":
					candidate[key] = json.RawMessage("null")
				case "string":
					candidate[key] = json.RawMessage(`"0"`)
				case "false":
					candidate[key] = json.RawMessage("false")
				case "array":
					candidate[key] = json.RawMessage("[]")
				case "drift":
					var value int64
					if err := json.Unmarshal(candidate[key], &value); err != nil {
						t.Fatal(err)
					}
					encoded, err := json.Marshal(value + 1)
					if err != nil {
						t.Fatal(err)
					}
					candidate[key] = encoded
				}
				data, err := json.Marshal(candidate)
				if err != nil {
					t.Fatal(err)
				}
				if err := publicapi.ValidateJSON(GuildCatalogSchemaName, data, definitions); err == nil {
					t.Fatal("descriptor admitted invalid policy")
				}
				if _, err := LoadCatalog(data); err == nil {
					t.Fatal("owner admitted invalid policy")
				}
			})
		}
	}
	private := append(bytes.Clone([]byte(phase0Catalog[:len(phase0Catalog)-1])), []byte(`,"member_accounts":[]}`)...)
	if err := publicapi.ValidateJSON(GuildCatalogSchemaName, private, definitions); err == nil {
		t.Fatal("open/private policy field admitted")
	}
	if _, err := LoadCatalog(private); err == nil {
		t.Fatal("owner admitted private policy field")
	}
}

func TestGuildOwnerSchemaPreservesMemberControlsAndFreshExports(t *testing.T) {
	definitions := guildSchemaDefinitions(t)
	data := []byte(phase0Catalog)
	for _, candidate := range [][]byte{
		bytes.Replace(data, []byte(`"schema_version"`), []byte(`"schema_versi\u006fn"`), 1),
		append(bytes.Clone(data[:len(data)-1]), []byte(`,"guild_tithe_ppm":30000,"guild_tithe_ppm":20000}`)...),
	} {
		if bytes.Equal(data, candidate) {
			t.Fatal("control did not change bytes")
		}
		if err := publicapi.ValidateJSON(GuildCatalogSchemaName, candidate, definitions); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCatalog(candidate); err != nil {
			t.Fatal(err)
		}
	}
	before, fresh := GuildSchemas(), GuildSchemas()
	if !reflect.DeepEqual(before, fresh) {
		t.Fatal("nondeterministic exports")
	}
	*fresh[0].Schema.Fields[0].Schema.Maximum = 0
	if !reflect.DeepEqual(before, GuildSchemas()) {
		t.Fatal("caller changed owner schema")
	}
}
