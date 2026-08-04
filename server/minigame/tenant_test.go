package minigame

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

type fixtureTenant struct {
	version       string
	invalidOutput bool
	invalidResult bool
	unknownError  bool
}

type fixtureSnapshot struct {
	Done  bool
	Total int64
}

type fixtureCommand struct {
	Add    int64
	Finish bool
}

func (tenant fixtureTenant) Descriptor() Descriptor {
	version := tenant.version
	if version == "" {
		version = "1.0.0"
	}
	return Descriptor{
		EngineRef: "fixture.counter", EngineVersion: version, CommandSchema: "fixture.command.v1",
		SnapshotSchema: "fixture.snapshot.v1", ResultSchema: "fixture.result.v1",
		Modes: []Mode{ModeSolo, ModeAsyncSnapshot}, ErrorTaxonomy: []string{"invalid_command"},
		Destinations: map[string]DestinationClass{"era": DestinationPresentation, "trust_ppm": DestinationBreadth},
	}
}

func (tenant fixtureTenant) ValidateCommand(data json.RawMessage) error {
	var command fixtureCommand
	if strictDecodeFixture(data, &command) != nil {
		return &Rejection{Code: "invalid_command", Detail: "command schema mismatch"}
	}
	return nil
}

func (tenant fixtureTenant) ValidateSnapshot(data json.RawMessage) error {
	var snapshot fixtureSnapshot
	if strictDecodeFixture(data, &snapshot) != nil {
		return ErrInvalidTenant
	}
	return nil
}

func (tenant fixtureTenant) ValidateResult(result *Result) error {
	if result == nil {
		return nil
	}
	if result.Outcome != "complete" || result.RatingDelta != nil || len(result.ScoreFacts) != 1 || result.ScoreFacts[0].Kind != "score.total" {
		return ErrInvalidTenant
	}
	return nil
}

func (tenant fixtureTenant) Create(input CreateInput) (json.RawMessage, error) {
	if input.ScalingInputs["era"] < 0 || input.ScalingInputs["trust_ppm"] < 0 {
		return nil, &Rejection{Code: "invalid_command", Detail: "negative fixture scaling"}
	}
	return json.Marshal(map[string]any{"done": false, "total": int64(input.Seed % 7)})
}

func (tenant fixtureTenant) Apply(input ApplyInput) (ApplyOutput, error) {
	if tenant.unknownError {
		return ApplyOutput{}, &Rejection{Code: "undeclared", Detail: "not in descriptor"}
	}
	var snapshot fixtureSnapshot
	var command fixtureCommand
	if strictDecodeFixture(input.Snapshot, &snapshot) != nil || strictDecodeFixture(input.Command, &command) != nil ||
		command.Add < 0 || snapshot.Done {
		return ApplyOutput{}, &Rejection{Code: "invalid_command", Detail: "illegal fixture command"}
	}
	snapshot.Total += command.Add
	snapshot.Done = command.Finish
	encoded, _ := json.Marshal(map[string]any{"done": snapshot.Done, "total": snapshot.Total})
	output := ApplyOutput{Snapshot: encoded}
	if command.Finish {
		output.Result = &Result{Outcome: "complete", ScoreFacts: []ScoreFact{{Kind: "score.total", Value: snapshot.Total}}}
		if tenant.invalidResult {
			output.Result.ScoreFacts[0].Kind = "score.wrong"
		}
	}
	if tenant.invalidOutput {
		output.Snapshot = json.RawMessage("{\"bogus\":true}")
	}
	return output, nil
}

func TestTenantRegistryConformance(t *testing.T) {
	registry, err := NewTenantRegistry(fixtureTenant{})
	if err != nil {
		t.Fatal(err)
	}
	descriptor, ok := registry.Descriptor("fixture.counter")
	if !ok || descriptor.EngineVersion != "1.0.0" || descriptor.Destinations["trust_ppm"] != DestinationBreadth {
		t.Fatalf("descriptor=%+v ok=%v", descriptor, ok)
	}
	descriptor.Destinations["trust_ppm"] = DestinationPower
	reloaded, _ := registry.Descriptor("fixture.counter")
	if reloaded.Destinations["trust_ppm"] != DestinationBreadth {
		t.Fatal("descriptor map escaped registry ownership")
	}

	scaling := map[string]int64{"era": 1, "trust_ppm": 500_000}
	snapshot, err := registry.Create("fixture.counter", "1.0.0", CreateInput{Mode: ModeSolo, Seed: 8, ScalingInputs: scaling})
	if err != nil || !bytes.Equal(snapshot, []byte("{\"done\":false,\"total\":1}")) {
		t.Fatalf("create snapshot=%s err=%v", snapshot, err)
	}
	output, err := registry.Apply("fixture.counter", "1.0.0", ApplyInput{Mode: ModeSolo, Revision: 1, Snapshot: snapshot,
		Command: json.RawMessage("{\"add\":4,\"finish\":false}"), ScalingInputs: scaling})
	if err != nil || !bytes.Equal(output.Snapshot, []byte("{\"done\":false,\"total\":5}")) || output.Result != nil {
		t.Fatalf("play output=%+v err=%v", output, err)
	}
	output, err = registry.Apply("fixture.counter", "1.0.0", ApplyInput{Mode: ModeSolo, Revision: 2, Snapshot: output.Snapshot,
		Command: json.RawMessage("{\"add\":2,\"finish\":true}"), ScalingInputs: scaling})
	if err != nil || !bytes.Equal(output.Snapshot, []byte("{\"done\":true,\"total\":7}")) || output.Result == nil ||
		output.Result.Outcome != "complete" || len(output.Result.ScoreFacts) != 1 || output.Result.ScoreFacts[0].Value != 7 {
		t.Fatalf("finish output=%+v err=%v", output, err)
	}
}

func TestTenantRegistryFailsClosed(t *testing.T) {
	if _, err := NewTenantRegistry(); !errors.Is(err, ErrInvalidTenant) {
		t.Fatalf("empty registry error=%v", err)
	}
	if _, err := NewTenantRegistry(fixtureTenant{}, fixtureTenant{}); !errors.Is(err, ErrInvalidTenant) {
		t.Fatalf("duplicate registry error=%v", err)
	}
	registry, err := NewTenantRegistry(fixtureTenant{})
	if err != nil {
		t.Fatal(err)
	}
	scaling := map[string]int64{"era": 1, "trust_ppm": 500_000}
	validSnapshot := json.RawMessage("{\"done\":false,\"total\":0}")
	tests := []struct {
		name  string
		input ApplyInput
	}{
		{"noncanonical snapshot", ApplyInput{Mode: ModeSolo, Revision: 1, Snapshot: json.RawMessage("{ \"done\":false,\"total\":0}"), Command: json.RawMessage("{\"add\":1,\"finish\":false}"), ScalingInputs: scaling}},
		{"unknown scaling", ApplyInput{Mode: ModeSolo, Revision: 1, Snapshot: validSnapshot, Command: json.RawMessage("{\"add\":1,\"finish\":false}"), ScalingInputs: map[string]int64{"era": 1, "unknown": 2}}},
		{"live pvp", ApplyInput{Mode: "live_pvp", Revision: 1, Snapshot: validSnapshot, Command: json.RawMessage("{\"add\":1,\"finish\":false}"), ScalingInputs: scaling}},
		{"zero revision", ApplyInput{Mode: ModeSolo, Revision: 0, Snapshot: validSnapshot, Command: json.RawMessage("{\"add\":1,\"finish\":false}"), ScalingInputs: scaling}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := registry.Apply("fixture.counter", "1.0.0", test.input); err == nil {
				t.Fatal("invalid boundary input accepted")
			}
		})
	}
	if _, err := registry.Apply("fixture.counter", "1.0.0", ApplyInput{Mode: ModeSolo, Revision: 1, Snapshot: validSnapshot,
		Command: json.RawMessage("{\"add\":1,\"finish\":false,\"score\":99}"), ScalingInputs: scaling}); !errors.Is(err, ErrTenantRejected) {
		t.Fatalf("tenant schema rejection error=%v", err)
	}

	badOutputRegistry, err := NewTenantRegistry(fixtureTenant{invalidOutput: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := badOutputRegistry.Apply("fixture.counter", "1.0.0", ApplyInput{Mode: ModeSolo, Revision: 1, Snapshot: validSnapshot,
		Command: json.RawMessage("{\"add\":1,\"finish\":false}"), ScalingInputs: scaling}); !errors.Is(err, ErrTenantDivergence) {
		t.Fatalf("noncanonical output error=%v", err)
	}
	unknownErrorRegistry, err := NewTenantRegistry(fixtureTenant{unknownError: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := unknownErrorRegistry.Apply("fixture.counter", "1.0.0", ApplyInput{Mode: ModeSolo, Revision: 1, Snapshot: validSnapshot,
		Command: json.RawMessage("{\"add\":1,\"finish\":false}"), ScalingInputs: scaling}); !errors.Is(err, ErrTenantDivergence) {
		t.Fatalf("undeclared rejection error=%v", err)
	}
	invalidResultRegistry, err := NewTenantRegistry(fixtureTenant{invalidResult: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := invalidResultRegistry.Apply("fixture.counter", "1.0.0", ApplyInput{Mode: ModeSolo, Revision: 1, Snapshot: validSnapshot,
		Command: json.RawMessage("{\"add\":1,\"finish\":true}"), ScalingInputs: scaling}); !errors.Is(err, ErrTenantDivergence) {
		t.Fatalf("wrong-schema result error=%v", err)
	}
}

func strictDecodeFixture(data []byte, destination any) error {
	var object map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&object); err != nil || len(object) != 2 {
		return ErrInvalidTenant
	}
	switch value := destination.(type) {
	case *fixtureSnapshot:
		if object["done"] == nil || object["total"] == nil {
			return ErrInvalidTenant
		}
		return json.Unmarshal(data, value)
	case *fixtureCommand:
		if object["add"] == nil || object["finish"] == nil {
			return ErrInvalidTenant
		}
		return json.Unmarshal(data, value)
	default:
		return ErrInvalidTenant
	}
}
