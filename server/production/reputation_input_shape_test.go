package production

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

const reputationInputShapePath = "../../testdata/reputation/purchase-input-shape-v1.json"

var reputationInputShapeFields = []string{"kind", "node_id", "resolved_cost", "reputation_level", "reputation_spent_before", "owned_before"}
var reputationInputShapeMutations = []string{"missing", "null", "alias", "duplicate", "alias_extra"}

type reputationInputShapeControl struct {
	Name       string `json:"name"`
	SourceCase string `json:"source_case"`
	ZeroEarned bool   `json:"zero_earned"`
}

type reputationInputShapeCase struct {
	Name        string `json:"name"`
	Control     string `json:"control"`
	Field       string `json:"field"`
	Mutation    string `json:"mutation"`
	RawResolved string `json:"raw_resolved"`
}

type reputationInputShapePopulation struct {
	Version      int                           `json:"schema_version"`
	SourceSHA256 string                        `json:"source_sha256"`
	Controls     []reputationInputShapeControl `json:"controls"`
	Cases        []reputationInputShapeCase    `json:"cases"`
}

func reputationInputShapeSource(t *testing.T) (reputationCorpus, []byte) {
	t.Helper()
	data, err := os.ReadFile(reputationCorpusPath)
	if err != nil {
		t.Fatal(err)
	}
	var source reputationCorpus
	if err := json.Unmarshal(data, &source); err != nil || len(source.Cases) != 20 {
		t.Fatalf("purchase source population: %v, cases=%d", err, len(source.Cases))
	}
	return source, data
}

func reputationShapeJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func reputationShapeObject(t *testing.T, data []byte) map[string]json.RawMessage {
	t.Helper()
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil || object == nil {
		t.Fatalf("invalid fixture object: %v", err)
	}
	return object
}

func reputationShapeControlCase(t *testing.T, source reputationCorpus, control reputationInputShapeControl) reputationCorpusCase {
	t.Helper()
	index := slices.IndexFunc(source.Cases, func(row reputationCorpusCase) bool { return row.Name == control.SourceCase })
	if index < 0 {
		t.Fatalf("unknown source case %q", control.SourceCase)
	}
	row := source.Cases[index]
	row.Name = control.Name
	// The fixture file is indented; the public replay entry requires the
	// original canonical command bytes, not its presentation whitespace.
	row.CanonicalPayload = json.RawMessage(canonicalFixtureJSON(t, row.CanonicalPayload))
	if control.ZeroEarned {
		if control.Name != "zero-earned-unaffordable" || control.SourceCase != "rejects-cost-one-over-available" {
			t.Fatal("unexpected zero-earned fixture profile")
		}
		pre := reputationShapeObject(t, row.PreState)
		pre["reputation_level"] = json.RawMessage("0")
		row.PreState = reputationShapeJSON(t, pre)
		row.PostStateJSON = canonicalFixtureJSON(t, row.PreState)
		wire := reputationShapeObject(t, row.ReplayInputs)
		resolved := reputationShapeObject(t, wire["resolved"])
		resolved["reputation_level"] = json.RawMessage("0")
		wire["resolved"] = reputationShapeJSON(t, resolved)
		row.ReplayInputs = reputationShapeJSON(t, wire)
	}
	return row
}

func reputationShapeMutation(t *testing.T, values map[string]json.RawMessage, field, mutation string) string {
	t.Helper()
	parts := []string{}
	for _, key := range reputationInputShapeFields {
		value := string(reputationShapeJSON(t, values[key]))
		entry := fmt.Sprintf("%q:%s", key, value)
		if key == field {
			switch mutation {
			case "missing":
				continue
			case "null":
				entry = fmt.Sprintf("%q:null", key)
			case "alias":
				entry = fmt.Sprintf("%q:%s", strings.ToUpper(key), value)
			case "duplicate":
				parts = append(parts, entry)
			case "alias_extra":
				parts = append(parts, fmt.Sprintf("%q:%s", strings.ToUpper(key), value))
			default:
				t.Fatalf("unknown shape mutation %q", mutation)
			}
		}
		parts = append(parts, entry)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func buildReputationInputShapePopulation(t *testing.T, source reputationCorpus, sourceBytes []byte) reputationInputShapePopulation {
	t.Helper()
	population := reputationInputShapePopulation{Version: 1, SourceSHA256: fmt.Sprintf("%x", sha256.Sum256(sourceBytes))}
	for _, row := range source.Cases {
		resolved := reputationShapeObject(t, reputationShapeObject(t, row.ReplayInputs)["resolved"])
		if string(resolved["kind"]) == `"purchase_reputation_node"` {
			population.Controls = append(population.Controls, reputationInputShapeControl{Name: row.Name, SourceCase: row.Name})
		}
	}
	if len(population.Controls) != 18 {
		t.Fatal("expected eighteen original purchase-arm controls; invalid-request arms are excluded")
	}
	population.Controls = append(population.Controls, reputationInputShapeControl{Name: "zero-earned-unaffordable", SourceCase: "rejects-cost-one-over-available", ZeroEarned: true})
	for _, control := range population.Controls {
		row := reputationShapeControlCase(t, source, control)
		resolved := reputationShapeObject(t, reputationShapeObject(t, row.ReplayInputs)["resolved"])
		if len(resolved) != len(reputationInputShapeFields) {
			t.Fatal("source resolved shape changed")
		}
		for _, field := range reputationInputShapeFields {
			if _, ok := resolved[field]; !ok {
				t.Fatalf("source lacks %s", field)
			}
			for _, mutation := range reputationInputShapeMutations {
				population.Cases = append(population.Cases, reputationInputShapeCase{
					Name: control.Name + "/" + field + "/" + mutation, Control: control.Name, Field: field, Mutation: mutation,
					RawResolved: reputationShapeMutation(t, resolved, field, mutation),
				})
			}
		}
	}
	if len(population.Controls) != 19 || len(population.Cases) != 570 {
		t.Fatal("incomplete shape population")
	}
	return population
}

func reputationShapeCatalogs(t *testing.T, source reputationCorpus) map[string]CatalogBundle {
	t.Helper()
	catalogs := map[string]CatalogBundle{"tree": reputationContentBundle(t), "live": activeContentBundle(t)}
	for name, bundle := range catalogs {
		if bundle.ConstantsHash != source.Bundles[name].ConstantsHash || canonicalFixtureValue(t, stringArtifacts(bundle.Artifacts)) != canonicalFixtureValue(t, source.Bundles[name].Artifacts) {
			t.Fatalf("%s source bundle changed", name)
		}
	}
	return catalogs
}

func reputationShapeRestore(t *testing.T, row reputationCorpusCase, catalogs CatalogBundle) *save.State {
	t.Helper()
	state, err := save.RestoreState(row.PreState, row.StateVersion, catalogs.Economy, economy.ScopeFounder, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if err := catalogs.ValidateFoundationState(state); err != nil {
		t.Fatalf("fixture pinned state: %v", err)
	}
	return state
}

func reputationShapeAssertControl(t *testing.T, row reputationCorpusCase, catalogs CatalogBundle) {
	t.Helper()
	state := reputationShapeRestore(t, row, catalogs)
	result, err := ApplyFounderLogged(state, row.CanonicalPayload, catalogs, row.ReplayInputs)
	if err != nil || string(result.Outcome) != row.Outcome || result.ResultConstantsHash != catalogs.ConstantsHash ||
		canonicalFixtureJSON(t, result.Receipt) != row.ReceiptJSON || canonicalFixtureValue(t, fixtureEvents(result.Events)) != row.EventsJSON ||
		canonicalFixtureJSON(t, mustEncodeState(t, state)) != row.PostStateJSON {
		t.Fatalf("%s control changed: outcome=%s err=%v", row.Name, result.Outcome, err)
	}
}

func TestReputationPurchaseInputShapeCorpus(t *testing.T) {
	source, data := reputationInputShapeSource(t)
	population := buildReputationInputShapePopulation(t, source, data)
	catalogs := reputationShapeCatalogs(t, source)
	for _, control := range population.Controls {
		t.Run(control.Name, func(t *testing.T) {
			row := reputationShapeControlCase(t, source, control)
			reputationShapeAssertControl(t, row, catalogs[row.Bundle])
		})
	}
	if t.Failed() {
		return // Invalid control output must never authorize fixture rewriting.
	}
	encoded, err := json.MarshalIndent(population, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if *updateReplayFixture {
		if err := os.WriteFile(reputationInputShapePath, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	pinned, err := os.ReadFile(reputationInputShapePath)
	if err != nil || !bytes.Equal(pinned, encoded) {
		t.Fatalf("input-shape corpus/source/population drift: %v; author only with make reputation-input-shape-corpus", err)
	}
}

func TestReputationPurchaseInputShapeAdmission(t *testing.T) {
	source, data := reputationInputShapeSource(t)
	var population reputationInputShapePopulation
	fixture, err := os.ReadFile(reputationInputShapePath)
	if err != nil || json.Unmarshal(fixture, &population) != nil {
		t.Fatalf("input-shape fixture: %v", err)
	}
	expected := buildReputationInputShapePopulation(t, source, data)
	if canonicalFixtureValue(t, population) != canonicalFixtureValue(t, expected) {
		t.Fatal("input-shape source/population/row drift")
	}
	catalogs := reputationShapeCatalogs(t, source)
	controls := map[string]reputationCorpusCase{}
	for _, control := range population.Controls {
		row := reputationShapeControlCase(t, source, control)
		reputationShapeAssertControl(t, row, catalogs[row.Bundle])
		controls[control.Name] = row
	}
	admitted := map[string]int{}
	totalAdmitted := 0
	for _, testCase := range population.Cases {
		row := controls[testCase.Control]
		bundle := catalogs[row.Bundle]
		state := reputationShapeRestore(t, row, bundle)
		before := mustEncodeState(t, state)
		wire := reputationShapeObject(t, row.ReplayInputs)
		wire["resolved"] = json.RawMessage(testCase.RawResolved)
		result, err := ApplyFounderLogged(state, row.CanonicalPayload, bundle, reputationShapeJSON(t, wire))
		if err == nil {
			admitted[testCase.Mutation]++
			totalAdmitted++
			continue
		}
		if !errors.Is(err, ErrInvalidReplayInputs) || len(result.Receipt) != 0 || len(result.Events) != 0 || !bytes.Equal(before, mustEncodeState(t, state)) {
			t.Errorf("%s refusal/rollback incorrect: outcome=%s receipt=%s events=%d err=%v", testCase.Name, result.Outcome, result.Receipt, len(result.Events), err)
		}
	}
	t.Logf("raw negatives=%d refused=%d admitted=%d groups=%v", len(population.Cases), len(population.Cases)-totalAdmitted, totalAdmitted, admitted)
	if totalAdmitted != 0 {
		t.Errorf("malformed frozen purchase inputs admitted: %d/%d %v", totalAdmitted, len(population.Cases), admitted)
	}
}
