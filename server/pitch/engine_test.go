package pitch

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"cloud-clicker/server/minigame"
)

func TestScoreUsesRuledOrderAndSingleQuantize(t *testing.T) {
	_, catalog := loadFixture(t)
	value, err := score([]string{"api_call#1", "api_call#2", "demo_day#1", "demo_day#2"},
		[]string{"ab_test", "dark_pattern", "pivot", "stealth_mode"}, catalog)
	if err != nil || value.String() != "9.45e3" {
		t.Fatalf("valuation=%s err=%v", value.String(), err)
	}
	withoutPartner, err := score([]string{"api_call#1", "api_call#2"}, []string{"stealth_mode"}, catalog)
	if err != nil || withoutPartner.String() != "3e1" {
		t.Fatalf("partner-absent valuation=%s err=%v", withoutPartner.String(), err)
	}
}

func TestTenantDeterministicFailureAndContentIdentity(t *testing.T) {
	data, _ := loadFixture(t)
	hash := ContentHash(data)
	tenant := NewTenant()
	input := minigame.CreateInput{Mode: minigame.ModeSolo, Seed: 7, ScalingInputs: map[string]int64{ScalingDestination: 1},
		Content: data, ContentHash: hash, ContentSchemaVersion: SchemaVersion}
	first, err := tenant.Create(input)
	second, secondErr := tenant.Create(input)
	if err != nil || secondErr != nil || !bytes.Equal(first, second) {
		t.Fatalf("create mismatch first=%s second=%s err=%v/%v", first, second, err, secondErr)
	}
	snapshot := first
	for revision := int64(1); revision <= 3; revision++ {
		output, applyErr := tenant.Apply(minigame.ApplyInput{Mode: minigame.ModeSolo, Seed: 7, Revision: revision,
			Snapshot: snapshot, Command: json.RawMessage(`{"kind":"play_hand","card_ids":[]}`),
			ScalingInputs: map[string]int64{ScalingDestination: 1}, Content: data, ContentHash: hash, ContentSchemaVersion: SchemaVersion})
		if applyErr != nil {
			t.Fatal(applyErr)
		}
		snapshot = output.Snapshot
		if revision < 3 && output.Result != nil {
			t.Fatalf("terminal at revision %d", revision)
		}
		if revision == 3 && (output.Result == nil || output.Result.Outcome != OutcomeFundingFailed || output.Result.ScoreFacts[1].Value != 1) {
			t.Fatalf("result=%+v", output.Result)
		}
	}
	input.ContentHash = "sha256:" + string(bytes.Repeat([]byte{'0'}, 64))
	if _, err := tenant.Create(input); !errors.Is(err, minigame.ErrInvalidTenant) {
		t.Fatalf("identity mismatch err=%v", err)
	}
}

func TestTenantRegistryAcceptsCanonicalPitchTransitions(t *testing.T) {
	data, _ := loadFixture(t)
	hash := ContentHash(data)
	registry, err := minigame.NewTenantRegistry(NewTenant())
	if err != nil {
		t.Fatal(err)
	}
	input := minigame.CreateInput{Mode: minigame.ModeSolo, Seed: 2,
		ScalingInputs: map[string]int64{ScalingDestination: 1}, Content: data,
		ContentHash: hash, ContentSchemaVersion: SchemaVersion}
	snapshot, err := registry.Create(EngineRef, EngineVersion, input)
	if err != nil {
		t.Fatal(err)
	}
	command := json.RawMessage(`{"card_ids":["demo_day#1","patch_release#1","testimonial#2","uptime_nines#2"],"kind":"play_hand"}`)
	applyInput := minigame.ApplyInput{Mode: minigame.ModeSolo,
		Seed: 2, Revision: 1, Snapshot: snapshot, Command: command,
		ScalingInputs: map[string]int64{ScalingDestination: 1}, Content: data,
		ContentHash: hash, ContentSchemaVersion: SchemaVersion}
	if _, err := registry.Apply(EngineRef, EngineVersion, applyInput); err != nil {
		direct, directErr := NewTenant().Apply(applyInput)
		validateErr := NewTenant().ValidateSnapshot(direct.Snapshot)
		t.Fatalf("registry rejected canonical Pitch transition: %v; direct_err=%v validate_err=%v output=%s snapshot=%s", err, directErr, validateErr, direct.Snapshot, snapshot)
	}
}

func TestDealUsesFixedPerRoundSlices(t *testing.T) {
	_, catalog := loadFixture(t)
	seen := map[string]bool{}
	for handNumber := int64(1); handNumber <= 3; handNumber++ {
		hand, remaining := deal(catalog, 19, 1, handNumber)
		if len(hand) != 7 || remaining != 24-handNumber*7 {
			t.Fatalf("hand %d len=%d remaining=%d", handNumber, len(hand), remaining)
		}
		for _, instance := range hand {
			if seen[instance] {
				t.Fatalf("instance %s repeated across deal slices", instance)
			}
			seen[instance] = true
		}
	}
}
