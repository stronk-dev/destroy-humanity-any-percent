package production

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/commonsbinding"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routeprojection"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

// Diagnostic run2 genesis only. The offer and all subsequent transitions are
// produced by Handle; this is not natural pacing or a default browser career.
func TestReputationOfferPlanPayoutIntegration(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; offer-plan SQL NOT EXECUTED")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	bundle := reputationContentBundle(t)
	for _, tc := range []struct {
		kind, factor string
		level        int64
		modifier     int64
	}{
		{"acquihire", "1.009e0", 18, 900000},
		{"acquisition", "1.01e0", 20, 1000000},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			if modifier, ok := bundle.Prestige.Modifier(tc.kind); !ok || modifier != tc.modifier {
				t.Fatal("pinned modifier changed; independent payout population must be reconciled")
			}
			if _, err := db.ExecContext(ctx, `TRUNCATE accounts,save_streams,catalog_sets,epochs RESTART IDENTITY CASCADE`); err != nil {
				t.Fatal(err)
			}
			seedProductionEpoch(t, db, bundle.ConstantsHash, bundle.Artifacts)
			resolver := integrationCatalogs{economy: map[string]*economy.Catalog{bundle.ConstantsHash: bundle.Economy},
				routes: map[string]*routes.Catalog{bundle.ConstantsHash: bundle.Routes}, prestige: map[string]*prestigecore.Policy{bundle.ConstantsHash: bundle.Prestige},
				factions: map[string]*faction.Catalog{bundle.ConstantsHash: bundle.Faction}}
			set := ReplayCatalogSet{bundle.ConstantsHash: bundle}
			store, err := save.NewStore(db, reputationTaxonomyCatalogs{resolver, set}, nil)
			if err != nil {
				t.Fatal(err)
			}
			repository, err := minigame.NewRepository(db)
			if err != nil {
				t.Fatal(err)
			}
			projector, err := routeprojection.New(db, resolver)
			if err != nil {
				t.Fatal(err)
			}
			service, err := NewService(store, resolver, FrozenContributionProvider{DB: db}, nil, nil,
				WithRouteCatalogs(resolver), WithRouteProjector(projector), WithProgressionRuntime(resolver),
				WithCurrentConstantsHash(bundle.ConstantsHash), WithReplayCatalogs(set), WithGuildSettlements(emptyGuildSettlements{}),
				WithMinigameActivity(repository), WithCompactPolicies(commons.CatalogSet{bundle.ConstantsHash: bundle.Commons.(commonsbinding.ReplayPolicy).Catalog}),
				WithCommonsWeightResolver(integrationWeight(1000000)))
			if err != nil {
				t.Fatal(err)
			}
			// Same diagnostic search ceiling as offerFixtureFounder; both kinds
			// must run. No statistical inference or silently excluded population.
			founderID := ""
			for candidate := int64(1); candidate <= 1000000; candidate++ {
				id := fmt.Sprintf("01986666-c100-7000-8000-%012d", candidate)
				draw, kind, _ := prestigecore.OfferDraws(id, 2, 3, 0)
				if draw < bundle.Prestige.SpawnGatePPM[3] && kind == tc.kind {
					founderID = id
					break
				}
			}
			if founderID == "" {
				t.Fatal("diagnostic offer-kind fixture search exhausted")
			}
			now := save.CanonicalServerTime(time.Now().UTC())
			const accountID = "01986666-c100-4000-8000-000000000001"
			if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')`, accountID); err != nil {
				t.Fatal(err)
			}
			if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)`, accountID, founderID); err != nil {
				t.Fatal(err)
			}
			founder := reputationFounderState(t, bundle, 22, now, 0)
			founder.ExitHistory = []save.ExitRecord{{RunID: 1, ExitType: "collapse", OccurredAt: now.Add(-time.Hour)}}
			company := replayFixtureState(t, bundle.Economy, now)
			company.WireVersion, company.Tier, company.RunSeq, company.MeterBands = 18, 2, 2, nil
			setCash(t, company, "1e9")
			// Cube root of8000 is20; expected modifier payouts18/20 are fixed
			// independently, not taken from ComputeTerms or the stored preview.
			company.LifetimeValue = bundle.Prestige.ThresholdValue().Mul(decimal.New(8, 3))
			meterState, err := meters.NewRunState(bundle.Meters, 0)
			if err != nil {
				t.Fatal(err)
			}
			company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
			company.AchievementsEarnedRun = map[string]bool{}
			if _, err := initializeActivePlayState(company, bundle.Opportunities, founderID); err != nil {
				t.Fatal(err)
			}
			ownerRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder}, bundle.ConstantsHash, founder, save.WriteContext{Cause: "reputation.offer.integration"})
			if err != nil {
				t.Fatal(err)
			}
			companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeCompany}, bundle.ConstantsHash, company, save.WriteContext{Cause: "reputation.offer.integration"})
			if err != nil {
				t.Fatal(err)
			}
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			if _, err := save.PinRunWithGenesisTx(ctx, tx, companyRevision.StreamID, founderID, 2, bundle.ConstantsHash, companyRevision.Version, mustEncodeState(t, company)); err != nil {
				t.Fatal(err)
			}
			frozen, err := FrozenFounderContributions(bundle, founder)
			if err != nil {
				t.Fatal(err)
			}
			if err := save.InsertRunFrozenContributionsTx(ctx, tx, companyRevision.StreamID, 2, frozen); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
			load := func() (save.Loaded, save.Loaded) {
				t.Helper()
				owner, err := store.LoadLatest(ctx, ownerRevision.StreamID)
				if err != nil {
					t.Fatal(err)
				}
				current, err := store.LoadLatest(ctx, companyRevision.StreamID)
				if err != nil {
					t.Fatal(err)
				}
				return owner, current
			}
			cross := []byte(`{"intent_id":"01986666-c101-7000-8000-000000000001","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t2_to_t3","route_id":null}`)
			crossed, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, cross)
			reputationOfferRequireReceipt(t, crossed, err, "applied", "", "")
			owner, offered := load()
			if offered.State.OfferState == nil || offered.State.OfferState.ExitType != tc.kind || offered.State.Tier != 3 || !offered.State.GatesCrossed["gate.t2_to_t3"] || owner.State.ReputationLevel != 0 || owner.State.ReputationSpent != 0 {
				t.Fatal("real gate did not produce the required offer with zero Reputation")
			}
			promise, err := prestigecore.DecodeStoredOfferTerms(offered.State.OfferState.TermsJSON)
			if err != nil || promise.PayoutPreview.ReputationDelta != tc.level || promise.MarketModifierPPM != 1000000 {
				t.Fatalf("independent payout preview differs: %+v err=%v", promise, err)
			}
			offerID := offered.State.OfferState.OfferID
			ownerBefore, companyBefore := mustEncodeState(t, owner.State), mustEncodeState(t, offered.State)
			request := func(id string, plan []string) []byte {
				t.Helper()
				wire, err := json.Marshal(map[string]any{"intent_id": id, "kind": IntentAcceptExitOffer, "expected_revision": offered.Revision.Number,
					"expected_founder_revision": owner.Revision.Number, "offer_id": offerID, "reputation_plan": plan})
				if err != nil {
					t.Fatal(err)
				}
				return wire
			}
			plan := []string{"reputation.unlock.p05", "reputation.starter.cash_small", "reputation.starter.generated_beige_tower"}
			badPlan := append(append([]string{}, plan...), "reputation.unlock.p25", "reputation.starter.cash_large") // Total23 >18/20, last entry unaffordable.
			badWire := request("01986666-c101-7000-8000-000000000002", badPlan)
			before := reputationPlanDBSnapshot(t, ctx, db)
			rejected, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, badWire)
			reputationOfferRequireReceipt(t, rejected, err, "rejected", "unaffordable", "reputation_plan.reputation")
			ownerAfter, companyAfter := load()
			if ownerAfter.Revision.Number != owner.Revision.Number || companyAfter.Revision.Number != offered.Revision.Number || !bytes.Equal(mustEncodeState(t, ownerAfter.State), ownerBefore) || !bytes.Equal(mustEncodeState(t, companyAfter.State), companyBefore) {
				t.Fatal("unaffordable last entry changed stream head or pending offer")
			}
			after := reputationPlanDBSnapshot(t, ctx, db)
			for _, table := range []string{"save_streams", "save_revisions", "events", "run_epochs", "run_genesis", "run_frozen_contributions", "verification_queue"} {
				if after[table] != before[table] {
					t.Fatalf("rejected offer plan changed game/evidence rows in %s", table)
				}
			}
			reputationOfferRequireRetry(t, ctx, db, service, companyRevision.StreamID, now, badWire, rejected)
			validID := "01986666-c101-7000-8000-000000000003"
			validWire := request(validID, plan)
			applied, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, validWire)
			reputationOfferRequireReceipt(t, applied, err, "applied", "", "")
			ownerAfter, companyAfter = load()
			var receiptRevisions struct {
				NewRevision     int64 `json:"new_revision"`
				FounderRevision int64 `json:"founder_revision"`
				AppliedCount    int64 `json:"applied_count"`
			}
			if json.Unmarshal(applied.Receipt, &receiptRevisions) != nil || receiptRevisions.NewRevision != companyAfter.Revision.Number || receiptRevisions.FounderRevision != ownerAfter.Revision.Number || receiptRevisions.AppliedCount != 1 {
				t.Fatal("applied receipt revisions/count do not match committed heads")
			}
			if ownerAfter.Revision.Number != owner.Revision.Number+1 || companyAfter.Revision.Number != offered.Revision.Number+2 || ownerAfter.State.ReputationLevel != tc.level || ownerAfter.State.ReputationSpent != 6 || ownerAfter.State.ReputationUnlockPPM != 50000 || !slices.Equal(ownerAfter.State.ReputationNodesOwned, []string{"reputation.starter.cash_small", "reputation.starter.generated_beige_tower", "reputation.unlock.p05"}) || len(ownerAfter.State.ExitHistory) != 2 || ownerAfter.State.ExitHistory[1].ExitType != tc.kind || companyAfter.State.RunSeq != 3 || companyAfter.State.OfferState != nil || companyAfter.State.GeneratorProvisioned["generator.beige_tower"] != 5 || companyAfter.State.GeneratorCounts["generator.beige_tower"] != 0 {
				t.Fatalf("payout-funded applied plan accounting/state differs: revisions=%d/%d want=%d/%d level=%d want=%d spent=%d unlock=%d owned=%v exits=%+v run=%d offer=%+v generated=%d purchased=%d", ownerAfter.Revision.Number, companyAfter.Revision.Number, owner.Revision.Number+1, offered.Revision.Number+2, ownerAfter.State.ReputationLevel, tc.level, ownerAfter.State.ReputationSpent, ownerAfter.State.ReputationUnlockPPM, ownerAfter.State.ReputationNodesOwned, ownerAfter.State.ExitHistory, companyAfter.State.RunSeq, companyAfter.State.OfferState, companyAfter.State.GeneratorProvisioned["generator.beige_tower"], companyAfter.State.GeneratorCounts["generator.beige_tower"])
			}
			if cash, ok := companyAfter.State.Ledger.Balance("company.cash"); !ok || cash.String() != "1e3" {
				t.Fatal("plan cash starter differs")
			}
			var factor string
			if err := db.QueryRowContext(ctx, `SELECT factor FROM run_frozen_contributions WHERE company_stream_id=$1 AND run_seq=3 AND source_id='reputation.founder_bonus'`, companyRevision.StreamID).Scan(&factor); err != nil || factor != tc.factor {
				t.Fatalf("post-plan frozen factor=%s want=%s err=%v", factor, tc.factor, err)
			}
			rows, err := db.QueryContext(ctx, `SELECT kind,payload FROM events WHERE stream_id=$1 AND intent_id=$2 ORDER BY event_seq,event_id`, ownerRevision.StreamID, validID)
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			var kinds []save.EventKind
			purchases := 0
			for rows.Next() {
				var kind save.EventKind
				var payload []byte
				if err := rows.Scan(&kind, &payload); err != nil {
					t.Fatal(err)
				}
				kinds = append(kinds, kind)
				if kind == save.EventReputationNodePurchased {
					var event struct {
						NodeID, Source             string
						Cost, Level, Before, After int64
					}
					object := reputationShapeObject(t, payload)
					if json.Unmarshal(object["node_id"], &event.NodeID) != nil || json.Unmarshal(object["source"], &event.Source) != nil || json.Unmarshal(object["cost"], &event.Cost) != nil || json.Unmarshal(object["reputation_level"], &event.Level) != nil || json.Unmarshal(object["reputation_spent_before"], &event.Before) != nil || json.Unmarshal(object["reputation_spent_after"], &event.After) != nil || purchases >= 3 || event.NodeID != plan[purchases] || event.Source != "exit_plan" || event.Cost != int64(purchases+1) || event.Level != tc.level || event.Before != []int64{0, 1, 3}[purchases] || event.After != []int64{1, 3, 6}[purchases] {
						t.Fatalf("ordered purchase event differs: %+v", event)
					}
					purchases++
				}
			}
			if err := rows.Err(); err != nil {
				t.Fatal(err)
			}
			rows.Close()
			if purchases != 3 || !slices.Equal(kinds, []save.EventKind{save.EventFounderAdvanced, save.EventReputationNodePurchased, save.EventReputationNodePurchased, save.EventReputationNodePurchased}) {
				t.Fatalf("Founder event population/order=%v", kinds)
			}
			var resolution, endedSeq, resolvedSeq string
			if err := db.QueryRowContext(ctx, `SELECT
				(SELECT payload->>'resolution' FROM events WHERE intent_id=$1 AND kind='exit_offer_resolved' AND payload->>'offer_id'=$2),
				(SELECT event_seq::text FROM events WHERE intent_id=$1 AND kind='exit_offer_resolved'),
				(SELECT event_seq::text FROM events WHERE intent_id=$1 AND kind='run_ended')`, validID, offerID).Scan(&resolution, &resolvedSeq, &endedSeq); err != nil || resolution != "accepted" {
				t.Fatalf("offer identity/resolution differs: %q err=%v", resolution, err)
			}
			var inOrder bool
			if err := db.QueryRowContext(ctx, `SELECT $1::bigint<$2::bigint`, resolvedSeq, endedSeq).Scan(&inOrder); err != nil || !inOrder {
				t.Fatal("offer resolution must precede run_ended")
			}
			var started []byte
			var schema int
			if err := db.QueryRowContext(ctx, `SELECT schema_version,payload->'reputation_tree' FROM events WHERE intent_id=$1 AND kind='run_started'`, validID).Scan(&schema, &started); err != nil {
				t.Fatal(err)
			}
			var summary reputationRunStarted
			if json.Unmarshal(started, &summary) != nil || schema != 2 || summary.BonusFactor != tc.factor || !slices.Equal(summary.AppliedStarterNodeIDs, plan[1:]) {
				t.Fatalf("run_started v2 summary differs: %s", started)
			}
			genesis, version, entries := reputationCareerReplay(t, db, companyRevision.StreamID, ownerRevision.StreamID, 2, &bundle)
			if len(entries) != 3 || VerifyReplayRun(genesis, version, bundle, entries, bundle.ConstantsHash, false) != ReplayVerified {
				t.Fatal("gate/rejected plan/applied offer Company run did not verify")
			}
			history, err := store.LoadFounderHistory(ctx, ownerRevision.StreamID)
			if err != nil || len(history.Entries) != 2 || VerifyFounderHistory(history, set) != ReplayVerified {
				t.Fatalf("rejected/applied Founder history differs: entries=%d err=%v", len(history.Entries), err)
			}
			immutableBefore := reputationPlanDBSnapshot(t, ctx, db)
			forgedRun := append([]ReplayLogEntry{}, entries...)
			object := reputationShapeObject(t, forgedRun[2].CanonicalPayload)
			object["reputation_plan"] = reputationShapeJSON(t, []string{plan[0], plan[2], plan[1]})
			forgedRun[2].CanonicalPayload = reputationShapeJSON(t, object)
			if VerifyReplayRun(genesis, version, bundle, forgedRun, bundle.ConstantsHash, false) != ReplayStateDivergence {
				t.Fatal("changed copied plan order was not refused")
			}
			forgedFounder := cloneReputationHistory(t, history)
			object = reputationShapeObject(t, forgedFounder.Entries[1].ReplayInputs)
			resolved := reputationShapeObject(t, object["resolved"])
			var frozenPurchases []reputationPlanPurchase
			if json.Unmarshal(resolved["reputation_purchases"], &frozenPurchases) != nil || len(frozenPurchases) != 3 || frozenPurchases[0].ResolvedCost != 1 {
				t.Fatal("copied cost mutation lacks its exact persisted target")
			}
			frozenPurchases[0].ResolvedCost = 2
			resolved["reputation_purchases"] = reputationShapeJSON(t, frozenPurchases)
			object["resolved"] = reputationShapeJSON(t, resolved)
			forgedFounder.Entries[1].ReplayInputs = reputationShapeJSON(t, object)
			if VerifyFounderHistory(forgedFounder, set) != ReplayStateDivergence {
				t.Fatal("changed copied plan cost was not refused")
			}
			immutableAfter := reputationPlanDBSnapshot(t, ctx, db)
			for table, values := range immutableBefore {
				if immutableAfter[table] != values {
					t.Fatalf("copied-evidence negative rewrote SQL %s", table)
				}
			}
			reputationOfferRequireRetry(t, ctx, db, service, companyRevision.StreamID, now, validWire, applied)
			before = reputationPlanDBSnapshot(t, ctx, db)
			conflictingWire := request(validID, plan[:2])
			conflict, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, conflictingWire)
			reputationOfferRequireReceipt(t, conflict, err, "rejected", "idempotency_conflict", validID)
			after = reputationPlanDBSnapshot(t, ctx, db)
			for table, values := range before {
				if after[table] != values {
					t.Fatalf("changed-plan conflict rewrote %s", table)
				}
			}
			t.Logf("%s: real offer from gate, zero initial Reputation → payout%d → plan6; both replay consumers and retry/conflict/copy negatives discriminate", tc.kind, tc.level)
		})
	}
}

func reputationOfferRequireReceipt(t *testing.T, result HandleResult, err error, outcome, category, detail string) {
	t.Helper()
	var receipt struct {
		Outcome   string `json:"outcome"`
		Rejection *struct {
			Category string `json:"category"`
			Detail   string `json:"detail"`
		} `json:"rejection"`
	}
	if err != nil || result.Replay || json.Unmarshal(result.Receipt, &receipt) != nil || receipt.Outcome != outcome ||
		(outcome == "applied" && receipt.Rejection != nil) ||
		(outcome == "rejected" && (receipt.Rejection == nil || receipt.Rejection.Category != category || receipt.Rejection.Detail != detail)) {
		t.Fatalf("receipt=%s replay=%v err=%v want=%s/%s/%s", result.Receipt, result.Replay, err, outcome, category, detail)
	}
}

func reputationOfferRequireRetry(t *testing.T, ctx context.Context, db *sql.DB, service *Service, stream string, now time.Time, wire []byte, previous HandleResult) {
	t.Helper()
	before := reputationPlanDBSnapshot(t, ctx, db)
	retry, err := service.Handle(ctx, stream, ModeOnline, now, wire)
	if err != nil || !retry.Replay || !bytes.Equal(retry.Receipt, previous.Receipt) {
		t.Fatalf("identical retry changed receipt: %s replay=%v err=%v", retry.Receipt, retry.Replay, err)
	}
	after := reputationPlanDBSnapshot(t, ctx, db)
	for table, values := range before {
		if after[table] != values {
			t.Fatalf("identical retry rewrote %s", table)
		}
	}
}
