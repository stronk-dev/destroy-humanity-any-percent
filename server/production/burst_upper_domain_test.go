package production

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

func TestBurstUpperDomainReplayParity(t *testing.T) {
	data, err := os.ReadFile("../../testdata/replay/burst-upper-domain-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases struct {
		Version int `json:"version"`
		Cases   []struct {
			Name        string         `json:"name"`
			Speed       string         `json:"burst_speed"`
			Rate        string         `json:"rate"`
			Efficiency  string         `json:"efficiency"`
			Mode        EvaluationMode `json:"mode"`
			ElapsedMS   int64          `json:"elapsed_ms"`
			BurstMS     int64          `json:"burst_ms"`
			ExpectError bool           `json:"expect_error"`
			ExpectCash  string         `json:"expect_cash"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if cases.Version != 1 || len(cases.Cases) != 6 {
		t.Fatal("incomplete burst population")
	}
	data, err = os.ReadFile("../../testdata/replay/apply-logged-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Run crossRuntimeFullRun `json:"doctrine_run"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, row := range cases.Cases {
		t.Run(row.Name, func(t *testing.T) {
			catalogs := doctrineReplayBundle(t)
			var root map[string]any
			if err := json.Unmarshal([]byte(fixture.Run.Artifacts["economy"]), &root); err != nil {
				t.Fatal(err)
			}
			policy := root["offline_policy"].(map[string]any)
			policy["burst_speed"], policy["efficiency"] = row.Speed, row.Efficiency
			root["resources"].([]any)[0].(map[string]any)["hardcap"].(map[string]any)["amount"] = "1e8999999999999999"
			root["generator_classes"].([]any)[0].(map[string]any)["production"].(map[string]any)["base_rate"] = row.Rate
			catalogs.Artifacts["economy"], err = json.Marshal(root)
			if err != nil {
				t.Fatal(err)
			}
			catalogs.Economy, err = economy.LoadCatalog(catalogs.Artifacts["economy"])
			if err != nil {
				t.Fatal("research catalog not admitted", err)
			}
			catalogs.ConstantsHash, err = save.ConstantsHashArtifacts(catalogs.Artifacts)
			if err != nil || !catalogs.valid(catalogs.ConstantsHash) {
				t.Fatal("invalid bundle", err)
			}
			var genesis map[string]any
			if err := json.Unmarshal(fixture.Run.Genesis, &genesis); err != nil {
				t.Fatal(err)
			}
			genesis["balances"].(map[string]any)["company.cash"] = "0"
			genesis["compute_burst_remaining_ms"], genesis["generators_purchased_total"] = row.BurstMS, 1
			wire, err := json.Marshal(genesis)
			if err != nil {
				t.Fatal(err)
			}
			state, err := save.RestoreState(wire, 17, catalogs.Economy, economy.ScopeCompany, time.Time{})
			if err != nil {
				t.Fatal("save not admitted", err)
			}
			before := mustEncodeState(t, state)
			entry := fixture.Run.Entries[5]
			var inputs map[string]any
			if err := json.Unmarshal(entry.ReplayInputs, &inputs); err != nil {
				t.Fatal(err)
			}
			at := state.EvaluatedThrough.Add(time.Duration(row.ElapsedMS) * time.Millisecond)
			inputs["evaluated_at_ms"], inputs["evaluation_mode"] = at.UnixMilli(), row.Mode
			inputs["resolved"].(map[string]any)["founder_carry"].(map[string]any)["founder_constants_hash"] = catalogs.ConstantsHash
			encoded, err := json.Marshal(inputs)
			if err != nil {
				t.Fatal(err)
			}
			result, err := ApplyLogged(state, []byte(canonicalFixtureJSON(t, entry.CanonicalPayload)), catalogs, encoded)
			if row.ExpectError {
				if err == nil || !strings.Contains(err.Error(), "accru") {
					t.Fatalf("expected actual accrual refusal, got %+v / %v", result, err)
				}
				if !bytes.Equal(before, mustEncodeState(t, state)) || string(result.Receipt) != "null" && len(result.Receipt) != 0 || len(result.Events) != 0 {
					t.Fatal("refusal mutated state or returned output")
				}
			} else {
				cash, ok := state.Ledger.Balance("company.cash")
				if err != nil || result.Outcome != save.IntentApplied || !ok || cash.String() != row.ExpectCash || !state.EvaluatedThrough.Equal(at) || state.ComputeBurstRemainingMS != 0 {
					t.Fatalf("got %s / %+v / %v", cash, result, err)
				}
			}
		})
	}
}
