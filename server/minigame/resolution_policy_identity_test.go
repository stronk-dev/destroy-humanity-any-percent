package minigame

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestMinigameResolutionPolicyIdentityIntegration(t *testing.T) {
	for _, mode := range []Mode{ModeSolo, ModeAsyncSnapshot} {
		for _, existingWindow := range []bool{false, true} {
			name := string(mode) + "/empty_window"
			if existingWindow {
				name = string(mode) + "/existing_window"
			}
			t.Run(name, func(t *testing.T) {
				db := minigameIntegrationDB(t)
				seedMinigameRun(t, db)
				ctx := context.Background()
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
				data, err := os.ReadFile("../../testdata/minigame/catalog-v2.json")
				if err != nil {
					t.Fatal(err)
				}
				catalog, err := LoadCatalog(data)
				if err != nil {
					t.Fatal(err)
				}
				definition, ok := catalog.Definition("fixture.counter")
				if !ok {
					t.Fatal("missing fixture policy")
				}
				_, err = service.Start(ctx, StartRequest{SessionID: testSessionID, MinigameID: definition.MinigameID,
					FounderID: testFounderID, CompanyStreamID: testStreamID, RunSeq: 1,
					EngineRef: definition.EngineRef, EngineVersion: definition.EngineVersion, ConstantsHash: testHash,
					ScalingInputs: map[string]int64{"era": 1, "trust_ppm": 500_000}, Seed: "8", Mode: mode})
				if err != nil {
					t.Fatal(err)
				}
				terminal, err := service.Play(ctx, PlayRequest{FounderID: testFounderID, SessionID: testSessionID,
					ExpectedRevision: 1, Command: json.RawMessage(`{"add":7,"finish":true}`)})
				if err != nil || terminal.Resolution == nil {
					t.Fatalf("terminal certification missing: %v", err)
				}
				claimed, err := repository.Load(ctx, testFounderID, testSessionID)
				if err != nil {
					t.Fatal(err)
				}
				wrong := definition
				wrong.MinigameID = "fixture.other"
				if existingWindow {
					if _, err := db.ExecContext(ctx, `INSERT INTO minigame_faucet_window
						(founder_id,minigame_id,attended_day,quota_used,conversion_remainder_ppm)
						VALUES($1,$2,0,4,250000)`, testFounderID, wrong.MinigameID); err != nil {
						t.Fatal(err)
					}
				}
				tx, err := db.BeginTx(ctx, nil)
				if err != nil {
					t.Fatal(err)
				}
				defer tx.Rollback()
				windows := func() []byte {
					t.Helper()
					var value []byte
					if err := tx.QueryRowContext(ctx, `SELECT COALESCE(jsonb_agg(to_jsonb(w)
						ORDER BY minigame_id,attended_day),'[]'::jsonb)
						FROM minigame_faucet_window w WHERE founder_id=$1`, testFounderID).Scan(&value); err != nil {
						t.Fatal(err)
					}
					return value
				}
				before := windows()
				_, wrongErr := service.PrepareResolutionTx(ctx, tx, terminal.Resolution, wrong, 0)
				after := windows()
				if !errors.Is(wrongErr, ErrInvalidSession) || !bytes.Equal(before, after) {
					t.Errorf("foreign policy must fail before quota mutation: err=%v before=%s after=%s", wrongErr, before, after)
				}
				if err := tx.Rollback(); err != nil {
					t.Fatal(err)
				}
				afterRefusal, err := repository.Load(ctx, testFounderID, testSessionID)
				if err != nil || !reflect.DeepEqual(afterRefusal, claimed) {
					t.Fatalf("policy refusal changed certified claim: %v", err)
				}

				// The exact same certification remains usable with its own policy.
				// Commit the quota and token-owned terminal state together; this is
				// the platform boundary, not a Company/Founder payout coordinator.
				correctTx, err := db.BeginTx(ctx, nil)
				if err != nil {
					t.Fatal(err)
				}
				defer correctTx.Rollback()
				prepared, err := service.PrepareResolutionTx(ctx, correctTx, terminal.Resolution, definition, 0)
				if err != nil || prepared.Faucet.QuotaBefore != 0 || prepared.Faucet.QuotaAfter != 1 ||
					prepared.Faucet.CreditedUnits != 1 || prepared.Faucet.RemainderAfterPPM != 0 {
					t.Fatalf("own policy/certification failed: application=%+v err=%v", prepared.Faucet, err)
				}
				if _, err := service.FinalizeResolutionTx(ctx, correctTx, terminal.Resolution,
					json.RawMessage(`{"outcome":"applied"}`), 2, 2); err != nil {
					t.Fatal(err)
				}
				if err := correctTx.Commit(); err != nil {
					t.Fatal(err)
				}
				resolved, err := repository.Load(ctx, testFounderID, testSessionID)
				if err != nil || resolved.Status != StatusResolved || resolved.Revision != 2 ||
					!bytes.Equal(resolved.State, terminal.Session.State) || resolved.ClaimToken != "" {
					t.Fatalf("own policy did not finalize the certified state: session=%+v err=%v", resolved, err)
				}
				var quota, remainder int64
				if err := db.QueryRowContext(ctx, `SELECT quota_used,conversion_remainder_ppm
					FROM minigame_faucet_window WHERE founder_id=$1 AND minigame_id=$2 AND attended_day=0`,
					testFounderID, definition.MinigameID).Scan(&quota, &remainder); err != nil || quota != 1 || remainder != 0 {
					t.Fatalf("own policy quota=%d remainder=%d err=%v", quota, remainder, err)
				}
			})
		}
	}
}
