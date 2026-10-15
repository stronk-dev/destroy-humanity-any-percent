package pitch

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"cloud-clicker/server/minigame"
)

// An admitted retuned catalog/snapshot, not a published content change or proof
// that this hack was acquired in a reachable session. It characterizes engine
// 1.0.0; the proposed aggregation overlay must not silently redefine that name.
func TestAggregationCompatibilityResearchPitch(t *testing.T) {
	source, err := os.ReadFile("../../testdata/pitch/aggregation-compatibility-research-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var vector struct {
		Version int
		Seed    uint64
		Engine  string `json:"engine_version"`
		Cards   []struct {
			ID       string `json:"card_id"`
			Instance string
			Metric   string `json:"base_metric"`
		}
		Hack      string `json:"hack_id"`
		Factor    string `json:"card_factor"`
		Target    string `json:"funding_target"`
		Valuation string `json:"current_valuation"`
		Phase     string `json:"current_phase"`
	}
	if json.Unmarshal(source, &vector) != nil || vector.Version != 1 || len(vector.Cards) != 4 || vector.Engine != EngineVersion {
		t.Fatal("invalid compatibility vector or changed engine identity")
	}
	_, catalog := loadFixture(t)
	selected := make([]string, len(vector.Cards))
	for index, row := range vector.Cards {
		found := false
		for at := range catalog.MetricCards {
			if catalog.MetricCards[at].CardID == row.ID {
				catalog.MetricCards[at].BaseMetric, found = row.Metric, true
			}
		}
		if !found {
			t.Fatal("unknown vector card", row.ID)
		}
		selected[index] = row.Instance
	}
	found := false
	for at := range catalog.GrowthHacks {
		if catalog.GrowthHacks[at].HackID == vector.Hack {
			catalog.GrowthHacks[at].Effect.Factor, found = vector.Factor, true
		}
	}
	if !found {
		t.Fatal("unknown vector hack")
	}
	for index := range catalog.FundingCurve {
		catalog.FundingCurve[index].FundingTarget = "1e" + strconv.Itoa(index+4)
	}
	catalog.FundingCurve[0].FundingTarget = vector.Target
	content, err := json.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(content, pitchCopyDeclarations()); err != nil {
		t.Fatal("research catalog not admitted", err)
	}
	registry, err := minigame.NewTenantRegistry(NewTenant())
	if err != nil {
		t.Fatal(err)
	}
	input := minigame.CreateInput{Mode: minigame.ModeSolo, Seed: vector.Seed,
		ScalingInputs: map[string]int64{ScalingDestination: 1}, Content: content,
		ContentHash: ContentHash(content), ContentSchemaVersion: SchemaVersion}
	genesis, err := registry.Create(EngineRef, vector.Engine, input)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := decodeSnapshot(genesis)
	if err != nil {
		t.Fatal(err)
	}
	// Keep the real seeded hand. Only hack ownership is synthetic/admitted.
	snapshot.SlottedHacks = []string{vector.Hack}
	encoded, err := encodeSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	command, err := json.Marshal(map[string]any{"kind": "play_hand", "card_ids": selected})
	if err != nil {
		t.Fatal(err)
	}
	output, err := registry.Apply(EngineRef, vector.Engine, minigame.ApplyInput{
		Mode: input.Mode, Seed: input.Seed, Revision: 1, Snapshot: encoded, Command: command,
		ScalingInputs: input.ScalingInputs, Content: content, ContentHash: input.ContentHash,
		ContentSchemaVersion: input.ContentSchemaVersion})
	if err != nil {
		t.Fatal(err)
	}
	after, err := decodeSnapshot(output.Snapshot)
	if err != nil || after.RoundBestValuation != vector.Valuation || after.Phase != vector.Phase || output.Result != nil {
		t.Fatalf("engine%s valuation%s phase%s, want%s/%s; err%v", vector.Engine,
			after.RoundBestValuation, after.Phase, vector.Valuation, vector.Phase, err)
	}
}
