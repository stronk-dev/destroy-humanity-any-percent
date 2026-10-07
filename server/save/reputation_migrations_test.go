package save_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/production"
	"cloud-clicker/server/replaycatalog"
	"cloud-clicker/server/save"
)

type founderMigrationReputation struct {
	Level     int64    `json:"level"`
	Spent     int64    `json:"spent"`
	Owned     []string `json:"owned"`
	UnlockPPM int64    `json:"unlock_ppm"`
}

type founderMigrationCase struct {
	Name               string                      `json:"name"`
	SourceCase         string                      `json:"source_case"`
	FromVersion        int                         `json:"from_version"`
	Scope              string                      `json:"scope"`
	Operation          string                      `json:"operation"`
	InputPatch         map[string]json.RawMessage  `json:"input_patch"`
	ErrorStage         string                      `json:"error_stage"`
	ExpectedVersion    int                         `json:"expected_version"`
	ExpectedReputation *founderMigrationReputation `json:"expected_reputation"`
}

type founderMigrationSourceCase struct {
	Name    string `json:"name"`
	Company struct {
		ConstantsHash     string            `json:"constants_hash"`
		NextConstantsHash string            `json:"next_constants_hash"`
		Artifacts         map[string]string `json:"artifacts"`
		NextArtifacts     map[string]string `json:"next_artifacts"`
		Case              struct {
			PreState         json.RawMessage `json:"pre_state"`
			CanonicalPayload json.RawMessage `json:"canonical_payload"`
			ReplayInputs     json.RawMessage `json:"replay_inputs"`
			NewCompanyJSON   string          `json:"new_company_json"`
			ReceiptJSON      string          `json:"receipt_json"`
		} `json:"case"`
	} `json:"company"`
	Founder *struct {
		StateVersion     int             `json:"state_version"`
		PreState         json.RawMessage `json:"pre_state"`
		CanonicalPayload json.RawMessage `json:"canonical_payload"`
		ReplayInputs     json.RawMessage `json:"replay_inputs"`
		PostStateJSON    string          `json:"post_state_json"`
		ReceiptJSON      string          `json:"receipt_json"`
	} `json:"founder"`
}

// The external package can use the actual public boundary without introducing
// a save -> production cycle or a production export created just for testing.
func TestFounderReputationMigrationCorpus(t *testing.T) {
	data, err := os.ReadFile("../../testdata/save-migrations.json")
	if err != nil {
		t.Fatal(err)
	}
	var table struct {
		Version int               `json:"corpus_version"`
		Legacy  []json.RawMessage `json:"cases"`
		Source  struct {
			Path   string `json:"path"`
			SHA256 string `json:"sha256"`
		} `json:"founder_source"`
		Cases         []founderMigrationCase `json:"founder_cases"`
		CompanySource struct {
			Path   string `json:"path"`
			SHA256 string `json:"sha256"`
		} `json:"company_source"`
		CompanyCases []axisMigrationCase `json:"company_cases"`
	}
	strictFounderMigrationJSON(t, data, &table)
	var baseline struct {
		Version  int      `json:"schema_version"`
		Count    int      `json:"minimum_case_count"`
		Required []string `json:"required_case_names"`
	}
	baselineBytes, err := os.ReadFile("../../testdata/save-migrations-baseline.json")
	if err != nil {
		t.Fatal(err)
	}
	strictFounderMigrationJSON(t, baselineBytes, &baseline)
	if table.Version != 10 || len(table.Legacy) != 11 || len(table.Cases) != 4 || len(table.CompanyCases) != 5 || baseline.Version != 1 || baseline.Count != 20 || len(baseline.Required) != 20 {
		t.Fatal("incomplete migration population/ratchet")
	}
	names := []string{}
	for _, raw := range table.Legacy {
		var row struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(raw, &row); err != nil {
			t.Fatal(err)
		}
		names = append(names, row.Name)
	}
	modernNames := []string{}
	for _, row := range table.Cases {
		names = append(names, row.Name)
		modernNames = append(modernNames, row.Name)
	}
	for _, row := range table.CompanyCases {
		names = append(names, row.Name)
	}
	sort.Strings(names)
	sort.Strings(baseline.Required)
	sort.Strings(modernNames)
	if !reflect.DeepEqual(names, baseline.Required) || !reflect.DeepEqual(modernNames, []string{"founder-v21-nonzero-unlock", "founder-v21-to-v22", "founder-v22-spent-over-level", "founder-v22-unlock-mismatch"}) {
		t.Fatal("migration names do not cover the required population")
	}
	if table.Source.Path != "replay/reputation-tree-v1.json" {
		t.Fatal("unexpected migration source path")
	}
	sourceBytes, err := os.ReadFile("../../testdata/" + table.Source.Path)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(sourceBytes)
	if hex.EncodeToString(digest[:]) != table.Source.SHA256 {
		t.Fatal("Founder migration source SHA mismatch")
	}
	var source struct {
		ExitCases []founderMigrationSourceCase `json:"exit_cases"`
	}
	if err := json.Unmarshal(sourceBytes, &source); err != nil {
		t.Fatal(err)
	}
	for _, row := range table.Cases {
		t.Run(row.Name, func(t *testing.T) {
			var selected *founderMigrationSourceCase
			for i := range source.ExitCases {
				if source.ExitCases[i].Name == row.SourceCase {
					if selected != nil {
						t.Fatal("duplicate migration source case")
					}
					selected = &source.ExitCases[i]
				}
			}
			if selected == nil || selected.Founder == nil || row.Scope != "founder" || row.InputPatch == nil {
				t.Fatal("missing Founder migration source/patch")
			}
			if row.Operation == "decode" && row.ExpectedVersion != row.FromVersion {
				t.Fatal("decode cannot claim a migrated result version")
			}
			current := loadFounderMigrationBundle(t, selected.Company.ConstantsHash, selected.Company.Artifacts)
			next := loadFounderMigrationBundle(t, selected.Company.NextConstantsHash, selected.Company.NextArtifacts)
			current.Next = &next
			var input map[string]json.RawMessage
			if err := json.Unmarshal(selected.Founder.PreState, &input); err != nil {
				t.Fatal(err)
			}
			for key, value := range row.InputPatch {
				if key != "reputation_level" && key != "reputation_spent" && key != "reputation_nodes_owned" && key != "reputation_unlock_ppm" {
					t.Fatalf("undeclared migration patch key %s", key)
				}
				input[key] = value
			}
			patched, err := json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}
			bundle := current
			if row.FromVersion == 22 {
				bundle = next
			} else if row.FromVersion != 21 {
				t.Fatal("unsupported migration source version")
			}
			state, err := save.RestoreState(patched, row.FromVersion, bundle.Economy, economy.ScopeFounder, time.Time{})
			if row.ErrorStage == "structural" {
				if row.Operation != "decode" || row.ExpectedReputation != nil || !errors.Is(err, save.ErrInvalidState) {
					t.Fatalf("structural corruption was not rejected: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if row.ErrorStage == "pinned_mirror" {
				if row.Operation != "decode" || row.ExpectedReputation != nil || !errors.Is(bundle.ValidateFoundationState(state), production.ErrInvalidEngineState) {
					t.Fatal("structurally legal but false pinned mirror admitted")
				}
				return
			}
			if row.ErrorStage != "none" || row.Operation != "new_run_activation" || row.ExpectedReputation == nil || row.FromVersion != selected.Founder.StateVersion || save.VersionForState(state) != 21 {
				t.Fatal("invalid activation row or activation happened during load")
			}
			if err := current.ValidateFoundationState(state); err != nil {
				t.Fatal(err)
			}
			founder, err := production.ApplyFounderLogged(state, canonicalFounderMigrationJSON(t, selected.Founder.CanonicalPayload), current, selected.Founder.ReplayInputs)
			if err != nil || founder.Outcome != save.IntentApplied {
				t.Fatalf("Founder activation: outcome=%s err=%v", founder.Outcome, err)
			}
			assertFounderMigrationReputation(t, founder.State, row)
			encoded, err := save.EncodeState(founder.State)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(canonicalFounderMigrationJSON(t, encoded), []byte(selected.Founder.PostStateJSON)) || !bytes.Equal(canonicalFounderMigrationJSON(t, founder.Receipt), []byte(selected.Founder.ReceiptJSON)) {
				t.Fatal("Founder activation changed complete canonical result bytes")
			}
			company, err := save.RestoreState(selected.Company.Case.PreState, 18, current.Economy, economy.ScopeCompany, time.Time{})
			if err != nil {
				t.Fatal(err)
			}
			exit, err := production.ApplyLoggedExit(company, canonicalFounderMigrationJSON(t, selected.Company.Case.CanonicalPayload), current, selected.Company.Case.ReplayInputs)
			if err != nil || exit.Decision.Outcome != save.IntentApplied {
				t.Fatalf("Company activation: outcome=%s err=%v", exit.Decision.Outcome, err)
			}
			assertFounderMigrationReputation(t, exit.Founder, row)
			newCompany, err := save.EncodeState(exit.Decision.NewCompanyState)
			if err != nil || !bytes.Equal(canonicalFounderMigrationJSON(t, newCompany), []byte(selected.Company.Case.NewCompanyJSON)) || !bytes.Equal(canonicalFounderMigrationJSON(t, exit.Decision.Receipt), []byte(selected.Company.Case.ReceiptJSON)) {
				t.Fatalf("Company activation changed canonical genesis/receipt: %v", err)
			}
		})
	}
}

func assertFounderMigrationReputation(t *testing.T, state *save.State, row founderMigrationCase) {
	t.Helper()
	got := founderMigrationReputation{state.ReputationLevel, state.ReputationSpent, state.ReputationNodesOwned, state.ReputationUnlockPPM}
	if save.VersionForState(state) != row.ExpectedVersion || !reflect.DeepEqual(got, *row.ExpectedReputation) {
		t.Fatalf("activation version/accounting=%d/%+v want=%d/%+v", save.VersionForState(state), got, row.ExpectedVersion, row.ExpectedReputation)
	}
}

func loadFounderMigrationBundle(t *testing.T, hash string, artifacts map[string]string) production.CatalogBundle {
	t.Helper()
	values := map[string][]byte{}
	for name, data := range artifacts {
		values[name] = []byte(data)
	}
	bundle, err := replaycatalog.Load(hash, values)
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

func strictFounderMigrationJSON(t *testing.T, data []byte, target any) {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		t.Fatal(err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		t.Fatal("trailing migration JSON")
	}
}

func canonicalFounderMigrationJSON(t *testing.T, data []byte) []byte {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
