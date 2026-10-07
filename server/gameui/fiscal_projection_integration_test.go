package gameui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

// GS1-A1: project a persisted Founder, harvest through the production service,
// and reload the actual receipt/state. No substituted catalog or frozen clock.
func TestFiscalProjectionMatchesPersistedHarvestIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set; Fiscal projection/harvest NOT EXECUTED")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE accounts,save_streams,catalog_sets,epochs RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	bundle, _, _, _ := reputationProjectionSource(t)
	if _, err := db.ExecContext(ctx, `INSERT INTO catalog_sets(constants_hash) VALUES($1)`, bundle.ConstantsHash); err != nil {
		t.Fatal(err)
	}
	for name, data := range bundle.Artifacts {
		if _, err := db.ExecContext(ctx, `INSERT INTO catalog_artifacts(constants_hash,artifact_name,bytes) VALUES($1,$2,$3)`, bundle.ConstantsHash, name, data); err != nil {
			t.Fatal(err)
		}
	}
	var epochID int64
	if err := db.QueryRowContext(ctx, `INSERT INTO epochs(name,started_at,changelog_ref) VALUES('GS1 diagnostic',now(),'changelog/epoch-1.md') RETURNING epoch_id`).Scan(&epochID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES($1,$2)`, epochID, bundle.ConstantsHash); err != nil {
		t.Fatal(err)
	}
	catalogs := reputationProjectionCatalogs{production.ReplayCatalogSet{bundle.ConstantsHash: bundle}}
	store, err := save.NewStore(db, catalogs, nil)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := minigame.NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	provider := production.FrozenContributionProvider{DB: db}
	service, err := production.NewService(store, catalogs, provider, nil, nil,
		production.WithProgressionRuntime(catalogs), production.WithCurrentConstantsHash(bundle.ConstantsHash),
		production.WithReplayCatalogs(catalogs), production.WithGuildSettlements(reputationProjectionSettlements{}))
	if err != nil {
		t.Fatal(err)
	}
	projector, err := New(store, catalogs, provider, WithMinigameActivity(repository))
	if err != nil {
		t.Fatal(err)
	}
	databaseNow := func(t *testing.T) time.Time {
		t.Helper()
		var ms int64
		if err := db.QueryRowContext(ctx, `SELECT (extract(epoch FROM clock_timestamp())*1000)::bigint`).Scan(&ms); err != nil {
			t.Fatal(err)
		}
		return time.UnixMilli(ms).UTC()
	}
	// Include state and history, not only revision numbers: a read must not
	// quietly sweep credit or emit commands/events without advancing the head.
	recordedRows := func(t *testing.T) map[string]string {
		t.Helper()
		rows := map[string]string{}
		for _, table := range []string{"save_streams", "save_revisions", "events", "founder_genesis", "founder_log", "intent_records", "transport_player_outbox", "verification_queue"} {
			query := fmt.Sprintf(`SELECT coalesce(jsonb_agg(to_jsonb(row) ORDER BY to_jsonb(row)::text),'[]'::jsonb)::text FROM %s row`, table)
			var value string
			if err := db.QueryRowContext(ctx, query).Scan(&value); err != nil {
				t.Fatal(err)
			}
			rows[table] = value
		}
		return rows
	}
	cap, yield, autoMS := bundle.Fiscal.Credit.Hardcap, bundle.Fiscal.Credit.CreditPerPeriod, bundle.Fiscal.Clock.AutoMS
	for index, row := range []struct {
		name      string
		credit    int64
		saturates bool
	}{
		{"normal-live-clock", 4, false},
		{"overdue-from-low-credit", 4, true},
		{"overdue-one-below-cap", cap - 1, true},
		{"overdue-at-cap", cap, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			_, company, founder, _ := reputationProjectionSource(t)
			now := databaseNow(t)
			founder.EvaluatedThrough, founder.ManualTokenRefilledAt = now, now
			founder.FiscalCredit, founder.FiscalPeriodSequence = row.credit, 17
			ageMS := 2*autoMS + autoMS/2
			if row.saturates {
				// The same credit outcome on either side of a real 300 ms
				// boundary; periods still have to match each actual timestamp.
				ageMS = (cap/yield + 2) * autoMS
			}
			founder.FiscalPeriodOpenedWallMS = now.UnixMilli() - ageMS
			company.EvaluatedThrough, company.ManualTokenRefilledAt, company.RunStartedAt = now, now, now
			accountID := fmt.Sprintf("01986666-f100-4000-8000-%012d", index+1)
			ownerID := fmt.Sprintf("01986666-f101-4000-8000-%012d", index+1)
			spawn, err := bundle.Opportunities.Spawn(ownerID, company.RunSeq, 0, 0)
			if err != nil {
				t.Fatal(err)
			}
			company.OpportunitySpawnSeq, company.NextOpportunityAttendedMS = 0, spawn.SpawnedAttendedMS
			if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')`, accountID); err != nil {
				t.Fatal(err)
			}
			if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)`, accountID, ownerID); err != nil {
				t.Fatal(err)
			}
			owner, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: ownerID, Scope: economy.ScopeFounder}, bundle.ConstantsHash, founder, save.WriteContext{Cause: "fiscal.projection.integration"})
			if err != nil {
				t.Fatal(err)
			}
			player, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: ownerID, Scope: economy.ScopeCompany}, bundle.ConstantsHash, company, save.WriteContext{Cause: "fiscal.projection.integration"})
			if err != nil {
				t.Fatal(err)
			}
			genesis, err := save.EncodeState(company)
			if err != nil {
				t.Fatal(err)
			}
			frozen, err := production.FrozenFounderContributions(bundle, founder)
			if err != nil {
				t.Fatal(err)
			}
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			if _, err := save.PinRunWithGenesisTx(ctx, tx, player.StreamID, ownerID, company.RunSeq, bundle.ConstantsHash, player.Version, genesis); err != nil {
				t.Fatal(err)
			}
			if err := save.InsertRunFrozenContributionsTx(ctx, tx, player.StreamID, company.RunSeq, frozen); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
			beforeRead := recordedRows(t)
			projectionTime := databaseNow(t)
			raw, err := projector.GameUISnapshot(ctx, player.StreamID, projectionTime)
			if err != nil {
				t.Fatal(err)
			}
			var view snapshot
			if err := json.Unmarshal(raw, &view); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(beforeRead, recordedRows(t)) {
				t.Fatal("Fiscal projection mutated persisted state/history")
			}
			arm := view.Features.Fiscal
			if arm == nil || arm.Credit != row.credit || view.FounderRevision != owner.Number || view.Revision != player.Number || view.ServerNowMS != projectionTime.UnixMilli() {
				t.Fatalf("wrong persisted Fiscal arm/coordinates: %s", raw)
			}
			// Independent arithmetic, not a second call to the implementation's Sweep.
			periods := (view.ServerNowMS - founder.FiscalPeriodOpenedWallMS) / autoMS
			credited := min(periods*yield, cap-row.credit)
			wantPreview := fiscalSweepPreview{Periods: periods, Credited: credited, CreditAfter: row.credit + credited, Saturated: periods*yield > cap-row.credit}
			if periods < 2 || arm.SweepPreview != wantPreview || arm.SweepPreview.Saturated != row.saturates {
				t.Fatalf("persisted sweep preview=%+v want=%+v (population saturates=%t)", arm.SweepPreview, wantPreview, row.saturates)
			}
			request, err := json.Marshal(map[string]any{"intent_id": fmt.Sprintf("01986666-f102-7000-8000-%012d", index+1), "kind": production.IntentHarvestFiscalPeriod, "expected_revision": view.FounderRevision})
			if err != nil {
				t.Fatal(err)
			}
			result, err := service.Handle(ctx, player.StreamID, production.ModeOnline, projectionTime, request)
			if err != nil {
				t.Fatal(err)
			}
			var receipt struct {
				Outcome         string `json:"outcome"`
				HarvestOutcome  string `json:"harvest_outcome"`
				FounderRevision int64  `json:"founder_revision"`
				CreditAfter     int64  `json:"fiscal_credit_after"`
				Sweep           *struct {
					Periods        int64 `json:"periods"`
					Credited       int64 `json:"credited"`
					CreditBefore   int64 `json:"credit_before"`
					CreditAfter    int64 `json:"credit_after"`
					OpenedBeforeMS int64 `json:"opened_before_ms"`
					OpenedAfterMS  int64 `json:"opened_after_ms"`
					SeqBefore      int64 `json:"seq_before"`
					SeqAfter       int64 `json:"seq_after"`
					Saturated      bool  `json:"saturated"`
				} `json:"fiscal_sweep"`
			}
			if err := json.Unmarshal(result.Receipt, &receipt); err != nil || result.Replay || receipt.Outcome != "applied" || receipt.HarvestOutcome != "consumed_by_auto" || receipt.Sweep == nil || receipt.FounderRevision != owner.Number+1 {
				t.Fatalf("real harvest receipt: %s, err=%v", result.Receipt, err)
			}
			history, err := store.LoadFounderHistory(ctx, owner.StreamID)
			if err != nil {
				t.Fatal(err)
			}
			if len(history.Entries) != 1 {
				t.Fatalf("harvest recorded %d commands, want one", len(history.Entries))
			}
			commandMS := history.Entries[0].ServerTSMS
			commandPeriods := (commandMS - founder.FiscalPeriodOpenedWallMS) / autoMS
			commandCredited := min(commandPeriods*yield, cap-row.credit)
			sweep := receipt.Sweep
			if commandMS < view.ServerNowMS || commandPeriods < periods || sweep.Periods != commandPeriods || sweep.CreditBefore != row.credit || sweep.Credited != commandCredited || sweep.CreditAfter != row.credit+commandCredited || sweep.Saturated != (commandPeriods*yield > cap-row.credit) || sweep.OpenedBeforeMS != founder.FiscalPeriodOpenedWallMS || sweep.OpenedAfterMS != founder.FiscalPeriodOpenedWallMS+commandPeriods*autoMS || sweep.SeqBefore != founder.FiscalPeriodSequence || sweep.SeqAfter != founder.FiscalPeriodSequence+commandPeriods {
				t.Fatalf("harvest disagrees with recorded DB time %d: %s", commandMS, result.Receipt)
			}
			// Normal credit can increase across a live clock boundary. Saturating
			// populations must agree exactly even when another period elapses.
			if row.saturates || commandPeriods == periods {
				if sweep.CreditAfter != arm.SweepPreview.CreditAfter || sweep.Credited != arm.SweepPreview.Credited || sweep.Saturated != arm.SweepPreview.Saturated {
					t.Fatalf("preview/next receipt mismatch: preview=%+v receipt=%s", arm.SweepPreview, result.Receipt)
				}
			}
			persisted, err := store.LoadLatest(ctx, owner.StreamID)
			if err != nil {
				t.Fatal(err)
			}
			companyAfter, err := store.LoadLatest(ctx, player.StreamID)
			if err != nil {
				t.Fatal(err)
			}
			companyBytes, err := save.EncodeState(companyAfter.State)
			if err != nil {
				t.Fatal(err)
			}
			if persisted.Revision.Number != receipt.FounderRevision || persisted.State.FiscalCredit != sweep.CreditAfter || receipt.CreditAfter != sweep.CreditAfter || persisted.State.FiscalPeriodOpenedWallMS != sweep.OpenedAfterMS || persisted.State.FiscalPeriodSequence != sweep.SeqAfter || companyAfter.Revision != player || !bytes.Equal(companyBytes, genesis) {
				t.Fatal("reloaded Founder/Company disagree with harvest receipt")
			}
			t.Logf("preview periods=%d credit=%d; recorded harvest periods=%d credit=%d; persisted Founder revision=%d", periods, arm.SweepPreview.CreditAfter, commandPeriods, sweep.CreditAfter, persisted.Revision.Number)
		})
	}
}
