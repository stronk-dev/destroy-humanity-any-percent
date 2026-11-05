package production

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"reflect"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

var updateRunEndAge = flag.Bool("update-run-end-age", false, "write the v13 run-end age vectors")

type runEndAgeVector struct {
	Name      string          `json:"name"`
	PreState  json.RawMessage `json:"pre_state"`
	Payload   json.RawMessage `json:"canonical_payload"`
	Inputs    json.RawMessage `json:"replay_inputs"`
	AgeMS     int64           `json:"expected_founder_age_ms"`
	Schema    int             `json:"expected_schema_version"`
	EndedJSON string          `json:"ended_event_json"`
}
type runEndAgeVectors struct {
	Version        int               `json:"version"`
	ConstantsHash  string            `json:"constants_hash"`
	ArtifactSource string            `json:"artifact_source"`
	Cases          []runEndAgeVector `json:"cases"`
}

func TestRunEndedFounderAgeVectors(t *testing.T) {
	bundle := activeContentBundle(t)
	founderFloor, _ := bundle.versionFloors()
	now := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	out := runEndAgeVectors{Version: 1, ConstantsHash: bundle.ConstantsHash, ArtifactSource: "apply-logged-v1.json:first_content_exit", Cases: []runEndAgeVector{}}
	appendCase := func(name string, input crossRuntimeTerminalCase, version int, age int64, schema int) {
		t.Helper()
		payload := []byte(canonicalFixtureJSON(t, input.CanonicalPayload))
		var wire replayInputsWire
		if err := json.Unmarshal(input.ReplayInputs, &wire); err != nil {
			t.Fatal(err)
		}
		wire.Version = version
		inputs, err := json.Marshal(wire)
		if err != nil {
			t.Fatal(err)
		}
		state := replayFixtureStateFromEncoded(t, bundle, input.PreState)
		result, err := ApplyLoggedExit(state, payload, bundle, inputs)
		if err != nil || result.Decision.Outcome != save.IntentApplied {
			t.Fatalf("%s outcome=%s err=%v", name, result.Decision.Outcome, err)
		}
		if result.Founder.AgeMS != age {
			t.Fatalf("%s private age=%d want=%d", name, result.Founder.AgeMS, age)
		}
		var ended *save.EventWrite
		for i := range result.Decision.CompanyEndedEvents {
			if result.Decision.CompanyEndedEvents[i].Kind == save.EventRunEnded {
				if ended != nil {
					t.Fatal("duplicate ended")
				}
				ended = &result.Decision.CompanyEndedEvents[i]
			}
		}
		if ended == nil || ended.SchemaVersion != schema {
			t.Fatalf("%s wrong ended version: %+v", name, ended)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(ended.Payload, &fields); err != nil {
			t.Fatal(err)
		}
		ageBytes, present := fields["founder_age_ms"]
		if schema == 4 {
			var got int64
			if !present || json.Unmarshal(ageBytes, &got) != nil || got != age {
				t.Fatalf("%s missing/wrong frozen age", name)
			}
		} else if present {
			t.Fatalf("%s changed historical age shape", name)
		}
		// Changing only the format discriminator must preserve all game outcomes.
		legacyWire := wire
		legacyWire.Version = 12
		legacyInputs, _ := json.Marshal(legacyWire)
		prior, err := ApplyLoggedExit(replayFixtureStateFromEncoded(t, bundle, input.PreState), payload, bundle, legacyInputs)
		if err != nil || !bytes.Equal(result.Decision.Receipt, prior.Decision.Receipt) || !bytes.Equal(mustEncodeState(t, result.Company), mustEncodeState(t, prior.Company)) || !bytes.Equal(mustEncodeState(t, result.Decision.NewCompanyState), mustEncodeState(t, prior.Decision.NewCompanyState)) || !reflect.DeepEqual(result.Founder, prior.Founder) {
			t.Fatalf("%s format changed state/receipt: %v", name, err)
		}
		encoded := canonicalFixtureValue(t, fixtureEvents([]save.EventWrite{*ended})[0])
		out.Cases = append(out.Cases, runEndAgeVector{Name: name, PreState: input.PreState, Payload: payload, Inputs: inputs, AgeMS: age, Schema: schema, EndedJSON: encoded})
	}
	for index, age := range []int64{3_600_000, 7_200_000} {
		row := makeReputationPlanExitCase(t, "age", bundle, bundle, founderFloor, 0, nil, now, func(company, founder *save.State) {
			founder.AgeMS = age
			company.RunStartedAt = now.Add(-time.Second)
			company.EvaluatedThrough = now.Add(-time.Second)
		})
		name := []string{"one-hour-Founder", "two-hour-Founder"}[index]
		appendCase(name, row.Company.Case, 13, age+1_000, 4)
		if index == 0 {
			appendCase("retained-v12-unbranched", row.Company.Case, 12, age+1_000, 2)
		}
	}
	data, err := os.ReadFile("../../testdata/replay/apply-logged-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var historical crossRuntimeFixture
	if err = json.Unmarshal(data, &historical); err != nil {
		t.Fatal(err)
	}
	if historical.CurriculumExit.ConstantsHash != bundle.ConstantsHash {
		t.Fatal("historical source pin differs")
	}
	appendCase("v13-curriculum-branch", historical.CurriculumExit.Case, 13, 900_000, 4)
	appendCase("retained-v12-curriculum", historical.CurriculumExit.Case, 12, 900_000, 3)
	// Existing historical output bytes, not a regenerated legacy expectation.
	if out.Cases[len(out.Cases)-1].EndedJSON != canonicalFixtureValue(t, func() fixtureEvent {
		for _, e := range historical.CurriculumExit.Case.CompanyEndedEvents {
			if e.Kind == "run_ended" {
				return e
			}
		}
		t.Fatal("missing historical event")
		return fixtureEvent{}
	}()) {
		t.Fatal("historical v3 event changed")
	}
	encoded, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	path := "../../testdata/replay/run-ended-age-v1.json"
	if *updateRunEndAge {
		if err = os.WriteFile(path, encoded, 0644); err != nil {
			t.Fatal(err)
		}
		return
	}
	expected, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, expected) {
		t.Fatal("run-end age vectors differ; inspect before explicit generation")
	}
}
