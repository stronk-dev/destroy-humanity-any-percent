package minigame

import (
	"context"
	"sort"
	"sync"
	"testing"
)

func TestFaucetWindowIntegrationCarriesAndCapsOnAttendedDay(t *testing.T) {
	db := minigameIntegrationDB(t)
	seedMinigameRun(t, db)
	ctx := context.Background()
	policy := PayoutPolicy{
		CreditedResourceID: "resource.compute", SendsPerDay: 2, PerSendCap: 4,
		ConversionPPM: 500000, PayoutScoreFactID: "score.total", CapReasonKey: "minigame.payout.cap",
	}
	apply := func(attended, score int64) FaucetApplication {
		t.Helper()
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		result, err := applyFaucetWindowTx(ctx, tx, testFounderID, "fixture.counter", attended, policy, score, 0)
		if err != nil {
			_ = tx.Rollback()
			t.Fatal(err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
		return result
	}
	first := apply(1, 11)
	if first.QuotaBefore != 0 || first.QuotaAfter != 1 || first.ConvertedUnits != 5 || first.CreditedUnits != 4 ||
		first.ForfeitedUnits != 1 || first.RemainderAfterPPM != 500000 || first.ConfiguredCapReasonKey != policy.CapReasonKey {
		t.Fatalf("first=%+v", first)
	}
	second := apply(2, 1)
	if second.QuotaBefore != 1 || second.QuotaAfter != 2 || second.ConvertedUnits != 1 || second.CreditedUnits != 1 ||
		second.ForfeitedUnits != 0 || second.RemainderBeforePPM != 500000 || second.RemainderAfterPPM != 0 || second.ConfiguredCapReasonKey != "" {
		t.Fatalf("second=%+v", second)
	}
	third := apply(3, 10)
	if third.QuotaBefore != 2 || third.QuotaAfter != 2 || third.ConvertedUnits != 5 || third.CreditedUnits != 0 ||
		third.ForfeitedUnits != 5 || third.ConfiguredCapReasonKey != policy.CapReasonKey {
		t.Fatalf("third=%+v", third)
	}
	nextDay := apply(attendedDayMS, 2)
	if nextDay.AttendedDay != 1 || nextDay.QuotaBefore != 0 || nextDay.QuotaAfter != 1 ||
		nextDay.RemainderBeforePPM != 0 || nextDay.CreditedUnits != 1 {
		t.Fatalf("next day=%+v", nextDay)
	}
}

func TestFaucetWindowIntegrationRollsBackAtomically(t *testing.T) {
	db := minigameIntegrationDB(t)
	seedMinigameRun(t, db)
	ctx := context.Background()
	policy := PayoutPolicy{CreditedResourceID: "resource.compute", SendsPerDay: 2, PerSendCap: 4,
		ConversionPPM: 500000, PayoutScoreFactID: "score.total", CapReasonKey: "minigame.payout.cap"}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := applyFaucetWindowTx(ctx, tx, testFounderID, "fixture.counter", 0, policy, 3, 0); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	var rows int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM minigame_faucet_window WHERE founder_id=$1`, testFounderID).Scan(&rows); err != nil || rows != 0 {
		t.Fatalf("rollback rows=%d err=%v", rows, err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO minigame_faucet_window(founder_id,minigame_id,attended_day,quota_used,conversion_remainder_ppm) VALUES($1,'fixture.counter',0,0,1000000)`, testFounderID); err == nil {
		t.Fatal("database accepted an invalid carried remainder")
	}
}

func TestFaucetWindowIntegrationSerializesConcurrentSessions(t *testing.T) {
	db := minigameIntegrationDB(t)
	seedMinigameRun(t, db)
	ctx := context.Background()
	policy := PayoutPolicy{CreditedResourceID: "resource.compute", SendsPerDay: 2, PerSendCap: 10,
		ConversionPPM: 333333, PayoutScoreFactID: "score.total", CapReasonKey: "minigame.payout.cap"}

	start := make(chan struct{})
	results := make(chan FaucetApplication, 2)
	errs := make(chan error, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	for _, score := range []int64{1, 2} {
		go func(score int64) {
			ready.Done()
			<-start
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				errs <- err
				return
			}
			result, err := applyFaucetWindowTx(ctx, tx, testFounderID, "fixture.counter", 0, policy, score, 0)
			if err == nil {
				err = tx.Commit()
			} else {
				_ = tx.Rollback()
			}
			if err != nil {
				errs <- err
				return
			}
			results <- result
		}(score)
	}
	ready.Wait()
	close(start)

	applications := make([]FaucetApplication, 0, 2)
	for range 2 {
		select {
		case err := <-errs:
			t.Fatal(err)
		case result := <-results:
			applications = append(applications, result)
		}
	}
	sort.Slice(applications, func(i, j int) bool { return applications[i].QuotaBefore < applications[j].QuotaBefore })
	if len(applications) != 2 || applications[0].QuotaBefore != 0 || applications[0].QuotaAfter != 1 ||
		applications[1].QuotaBefore != 1 || applications[1].QuotaAfter != 2 {
		t.Fatalf("applications=%+v", applications)
	}

	var quota, remainder int64
	if err := db.QueryRowContext(ctx, `SELECT quota_used,conversion_remainder_ppm
		FROM minigame_faucet_window WHERE founder_id=$1 AND minigame_id='fixture.counter' AND attended_day=0`,
		testFounderID).Scan(&quota, &remainder); err != nil {
		t.Fatal(err)
	}
	combined, err := ConvertPayout(3, 0, policy.ConversionPPM, 0)
	if err != nil || quota != 2 || remainder != combined.ConversionRemainderPPM {
		t.Fatalf("quota=%d remainder=%d combined=%+v err=%v", quota, remainder, combined, err)
	}
}
