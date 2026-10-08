package minigame

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestReleaseResolutionClaimRequiresCertification(t *testing.T) {
	ctx := context.Background()
	for _, service := range []*Service{nil, {}} {
		for _, resolution := range []*CertifiedResolution{nil, {}} {
			if err := service.ReleaseResolutionClaim(ctx, resolution); !errors.Is(err, ErrInvalidSession) {
				t.Fatalf("invalid certification error=%v", err)
			}
		}
	}
}

func TestReleaseResolutionClaimIntegration(t *testing.T) {
	db := minigameIntegrationDB(t)
	seedMinigameRun(t, db)
	repository, err := NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	tenants, err := NewTenantRegistry(fixtureTenant{})
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService(repository, tenants)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	started, err := service.Start(ctx, StartRequest{SessionID: testSessionID, MinigameID: "combat.duel", FounderID: testFounderID,
		CompanyStreamID: testStreamID, RunSeq: 1, EngineRef: "fixture.counter", EngineVersion: "1.0.0",
		ConstantsHash: testHash, ScalingInputs: map[string]int64{"era": 1, "trust_ppm": 500_000}, Seed: "8", Mode: ModeSolo})
	if err != nil {
		t.Fatal(err)
	}
	request := PlayRequest{FounderID: testFounderID, SessionID: testSessionID, ExpectedRevision: 1,
		Command: json.RawMessage(`{"add":2,"finish":true}`)}
	play := func() *CertifiedResolution {
		t.Helper()
		decision, playErr := service.Play(ctx, request)
		if playErr != nil || decision.Resolution == nil {
			t.Fatalf("terminal play resolution=%v err=%v", decision.Resolution, playErr)
		}
		return decision.Resolution
	}
	old := play()
	if _, err := db.ExecContext(ctx, `UPDATE minigame_sessions SET claimed_at=clock_timestamp()-interval '6 minutes' WHERE session_id=$1`, testSessionID); err != nil {
		t.Fatal(err)
	}
	current := play()
	if current.identity.claimToken == old.identity.claimToken {
		t.Fatal("lease replacement did not issue a new token")
	}
	if err := service.ReleaseResolutionClaim(ctx, old); !errors.Is(err, ErrClaimLost) {
		t.Fatalf("old worker release error=%v", err)
	}
	claimed, err := repository.Load(ctx, testFounderID, testSessionID)
	if err != nil || claimed.Status != StatusClaimed || claimed.ClaimToken != current.identity.claimToken || claimed.Revision != started.Revision || !bytes.Equal(claimed.State, started.State) {
		t.Fatalf("old worker changed newer claim=%+v err=%v", claimed, err)
	}
	if err := service.ReleaseResolutionClaim(ctx, current); err != nil {
		t.Fatal(err)
	}
	active, err := repository.Load(ctx, testFounderID, testSessionID)
	if err != nil || active.Status != StatusActive || active.ClaimToken != "" || active.ClaimedAt != nil || active.Revision != started.Revision || !bytes.Equal(active.State, started.State) {
		t.Fatalf("released snapshot=%+v err=%v", active, err)
	}
	if err := service.ReleaseResolutionClaim(ctx, current); !errors.Is(err, ErrClaimLost) {
		t.Fatalf("already released token error=%v", err)
	}
	terminal := play()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := resolveTx(ctx, tx, terminal.identity, terminal.command, terminal.state, terminal.bytes, json.RawMessage(`{"outcome":"applied"}`), 2, 2); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := service.ReleaseResolutionClaim(ctx, terminal); !errors.Is(err, ErrClaimLost) {
		t.Fatalf("committed resolution release error=%v", err)
	}
	resolved, err := repository.Load(ctx, testFounderID, testSessionID)
	if err != nil || resolved.Status != StatusResolved || resolved.Revision != started.Revision+1 || !bytes.Equal(resolved.State, terminal.state) || !jsonObjectEqual(resolved.ResolutionReceipt, []byte(`{"outcome":"applied"}`)) {
		t.Fatalf("cleanup changed committed resolution=%+v err=%v", resolved, err)
	}
}
