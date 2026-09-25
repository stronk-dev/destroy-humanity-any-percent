package minigame

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// clockTenant records every ServerTimeMs it receives into its snapshot and
// stalls inside Apply, so a second clock read at insert time (the pre-TT-PA1
// behavior) would persist a different stamp than the tenant saw.
type clockTenant struct{}

type clockSnapshot struct {
	Stamps []int64 `json:"stamps"`
}

func (clockTenant) Descriptor() Descriptor {
	return Descriptor{EngineRef: "fixture.clock", EngineVersion: "1.0.0", CommandSchema: "fixture.clock.command.v1",
		SnapshotSchema: "fixture.clock.snapshot.v1", ResultSchema: "fixture.result.v1", Modes: []Mode{ModeSolo},
		ErrorTaxonomy: []string{"invalid_command"}, Destinations: map[string]DestinationClass{"era": DestinationPresentation}}
}
func (clockTenant) ValidateCommand(json.RawMessage) error { return nil }
func (clockTenant) ValidateSnapshot(data json.RawMessage) error {
	var snapshot clockSnapshot
	if json.Unmarshal(data, &snapshot) != nil || snapshot.Stamps == nil {
		return ErrInvalidTenant
	}
	return nil
}
func (clockTenant) ValidateResult(result *Result) error {
	if result != nil && (result.Outcome != "complete" || len(result.ScoreFacts) != 1) {
		return ErrInvalidTenant
	}
	return nil
}
func (clockTenant) Create(CreateInput) (json.RawMessage, error) {
	return json.RawMessage(`{"stamps":[]}`), nil
}
func (clockTenant) Apply(input ApplyInput) (ApplyOutput, error) {
	var snapshot clockSnapshot
	var command struct {
		Finish bool `json:"finish"`
	}
	if json.Unmarshal(input.Snapshot, &snapshot) != nil || json.Unmarshal(input.Command, &command) != nil {
		return ApplyOutput{}, ErrInvalidTenant
	}
	time.Sleep(25 * time.Millisecond)
	snapshot.Stamps = append(snapshot.Stamps, input.ServerTimeMs)
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return ApplyOutput{}, err
	}
	output := ApplyOutput{Snapshot: encoded}
	if command.Finish {
		output.Result = &Result{Outcome: "complete", ScoreFacts: []ScoreFact{{Kind: "score.total", Value: int64(len(snapshot.Stamps))}}}
	}
	return output, nil
}

// TT-PA1 / Typer AC3: the stamp the tenant receives is exactly the persisted
// server_ts_ms, terminal included, and verification replay reproduces the
// snapshot from the persisted stamps.
func TestServerTimeSampleReachesTenantAndLogIntegration(t *testing.T) {
	db := minigameIntegrationDB(t)
	seedMinigameRun(t, db)
	repository, err := NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	tenants, err := NewTenantRegistry(clockTenant{})
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService(repository, tenants)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := service.Start(ctx, StartRequest{SessionID: testSessionID, MinigameID: "combat.duel", FounderID: testFounderID,
		CompanyStreamID: testStreamID, RunSeq: 1, EngineRef: "fixture.clock", EngineVersion: "1.0.0", ConstantsHash: testHash,
		ScalingInputs: map[string]int64{"era": 1}, Seed: "8", Mode: ModeSolo}); err != nil {
		t.Fatal(err)
	}
	for revision, command := range []string{`{"finish":false}`, `{"finish":false}`} {
		if _, err := service.Play(ctx, PlayRequest{FounderID: testFounderID, SessionID: testSessionID,
			ExpectedRevision: int64(revision + 1), Command: json.RawMessage(command)}); err != nil {
			t.Fatal(err)
		}
	}
	terminal, err := service.Play(ctx, PlayRequest{FounderID: testFounderID, SessionID: testSessionID,
		ExpectedRevision: 3, Command: json.RawMessage(`{"finish":true}`)})
	if err != nil || terminal.Resolution == nil {
		t.Fatalf("terminal=%+v err=%v", terminal, err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.validateCertifiedTx(ctx, tx, terminal.Resolution); err != nil {
		_ = tx.Rollback()
		t.Fatalf("replay from persisted stamps diverged: %v", err)
	}
	resolved, err := resolveTx(ctx, tx, terminal.Resolution.identity, terminal.Resolution.command, terminal.Resolution.state,
		terminal.Resolution.bytes, json.RawMessage(`{"outcome":"applied"}`), 2, 2)
	if err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	var final clockSnapshot
	if err := json.Unmarshal(resolved.State, &final); err != nil || len(final.Stamps) != 3 {
		t.Fatalf("final=%s err=%v", resolved.State, err)
	}
	rows, err := db.QueryContext(ctx, "SELECT server_ts_ms FROM minigame_session_commands WHERE session_id=$1 ORDER BY seq", testSessionID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var persisted []int64
	for rows.Next() {
		var stamp int64
		if err := rows.Scan(&stamp); err != nil {
			t.Fatal(err)
		}
		persisted = append(persisted, stamp)
	}
	if len(persisted) != 3 {
		t.Fatalf("persisted stamps=%v", persisted)
	}
	for index, stamp := range persisted {
		if stamp != final.Stamps[index] || stamp < 1 {
			t.Fatalf("command %d: tenant saw %d, log persisted %d", index+1, final.Stamps[index], stamp)
		}
	}
}
