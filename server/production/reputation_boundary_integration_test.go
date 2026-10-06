package production

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"slices"
	"strings"
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
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

// RP-299: real Service/SQL old-pin activation and first-failure rollback.
// Earned6, run2 and stored offers are diagnostic genesis, not natural pacing.
func TestReputationExitBoundaryIntegration(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; Reputation boundary SQL NOT EXECUTED")
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
	live, tree := activeContentBundle(t), reputationContentBundle(t)
	good := []string{"reputation.unlock.p05", "reputation.starter.cash_small", "reputation.starter.generated_beige_tower"}
	profiles := []struct {
		name, category, detail string
		plan                   []string
	}{
		{"activate-plan", "", "", good},
		{"activate-absent", "", "", nil},
		{"activate-empty", "", "", []string{}},
		{"next-inactive", "not_eligible", "reputation_plan.tree_inactive", good[:1]},
		{"unknown-prefix", "unknown_id", "reputation_plan.unknown_id", []string{good[0], "reputation.missing"}},
		{"owned-prefix", "not_eligible", "reputation_plan.owned", []string{good[1], good[0]}},
		{"requires-prefix", "not_eligible", "reputation_plan.requires", []string{good[0], good[2]}},
		{"unaffordable-prefix", "unaffordable", "reputation_plan.reputation", append(slices.Clone(good), "reputation.unlock.p25")},
	}
	appliedCases, refusedCases, executed, fallbackCases := 0, 0, 0, 0
	for _, command := range []string{"wind_down", "acquihire", "acquisition"} {
		for _, profile := range profiles {
			t.Run(command+"-"+profile.name, func(t *testing.T) {
				executed++
				if _, err := db.ExecContext(ctx, `TRUNCATE accounts,save_streams,catalog_sets,epochs RESTART IDENTITY CASCADE`); err != nil {
					t.Fatal(err)
				}
				activation := strings.HasPrefix(profile.name, "activate-")
				current, next, version := tree, tree, 22
				if activation {
					current, version = live, 21
				}
				if profile.name == "next-inactive" {
					current, next, version = live, live, 21
				}
				seedProductionEpoch(t, db, current.ConstantsHash, current.Artifacts)
				resolver := integrationCatalogs{economy: map[string]*economy.Catalog{current.ConstantsHash: current.Economy, next.ConstantsHash: next.Economy},
					routes:   map[string]*routes.Catalog{current.ConstantsHash: current.Routes, next.ConstantsHash: next.Routes},
					prestige: map[string]*prestigecore.Policy{current.ConstantsHash: current.Prestige, next.ConstantsHash: next.Prestige},
					factions: map[string]*faction.Catalog{current.ConstantsHash: current.Faction, next.ConstantsHash: next.Faction}}
				set := ReplayCatalogSet{current.ConstantsHash: current, next.ConstantsHash: next}
				store, err := save.NewStore(db, reputationTaxonomyCatalogs{resolver, set}, nil)
				if err != nil {
					t.Fatal(err)
				}
				now := save.CanonicalServerTime(time.Now().UTC())
				const accountID = "01986666-e100-4000-8000-000000000001"
				const founderID = "01986666-e100-7000-8000-000000000001"
				if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')`, accountID); err != nil {
					t.Fatal(err)
				}
				if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)`, accountID, founderID); err != nil {
					t.Fatal(err)
				}
				founder := reputationFounderState(t, current, version, now, 6)
				founder.ExitHistory = []save.ExitRecord{{RunID: 1, ExitType: "collapse", OccurredAt: now.Add(-time.Hour)}}
				if profile.name == "owned-prefix" {
					founder.ReputationSpent, founder.ReputationUnlockPPM, founder.ReputationNodesOwned = 1, 50000, []string{good[0]}
				}
				company := replayFixtureState(t, current.Economy, now.Add(-time.Minute))
				company.WireVersion, company.Tier, company.RunSeq, company.MeterBands, company.LifetimeValue = 18, 3, 2, nil, decimal.Zero
				company.GatesCrossed["gate.t0_to_t1"] = true
				company.AchievementsEarnedRun = map[string]bool{}
				meterState, err := meters.NewRunState(current.Meters, 0)
				if err != nil {
					t.Fatal(err)
				}
				company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
				if _, err := initializeActivePlayState(company, current.Opportunities, founderID); err != nil {
					t.Fatal(err)
				}
				if command != "wind_down" {
					company.OfferState = &save.ExitOfferState{OfferID: "01986666-e101-7000-8000-000000000001", ExitType: command,
						SpawnedAt: now.Add(-time.Second), ExpiresAt: now.Add(time.Minute),
						TermsJSON: json.RawMessage(`{"market_modifier_ppm":1000000,"payout_preview":{"reputation_delta":0,"network_slot_unlocks":[],"route_knowledge":0,"clout_reach_note":"clout.reach.preserved"}}`)}
				}
				owner, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder}, current.ConstantsHash, founder, save.WriteContext{Cause: "reputation.boundary.integration"})
				if err != nil {
					t.Fatal(err)
				}
				player, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeCompany}, current.ConstantsHash, company, save.WriteContext{Cause: "reputation.boundary.integration"})
				if err != nil {
					t.Fatal(err)
				}
				tx, err := db.BeginTx(ctx, nil)
				if err != nil {
					t.Fatal(err)
				}
				defer tx.Rollback()
				if _, err := save.PinRunWithGenesisTx(ctx, tx, player.StreamID, founderID, 2, current.ConstantsHash, player.Version, mustEncodeState(t, company)); err != nil {
					t.Fatal(err)
				}
				frozen, err := FrozenFounderContributions(current, founder)
				if err != nil {
					t.Fatal(err)
				}
				if err := save.InsertRunFrozenContributionsTx(ctx, tx, player.StreamID, 2, frozen); err != nil {
					t.Fatal(err)
				}
				if err := tx.Commit(); err != nil {
					t.Fatal(err)
				}
				var oldPin string
				if err := db.QueryRowContext(ctx, `SELECT to_jsonb(row)::text FROM run_epochs row WHERE company_stream_id=$1 AND run_seq=2`, player.StreamID).Scan(&oldPin); err != nil {
					t.Fatal(err)
				}
				if activation {
					// Disposable fixture governance: close, then append a distinct
					// epoch. Never add a tree artifact as a same-epoch hotfix.
					if _, err := db.ExecContext(ctx, `UPDATE epochs SET ended_at=clock_timestamp() WHERE ended_at IS NULL`); err != nil {
						t.Fatal(err)
					}
					seedProductionEpoch(t, db, next.ConstantsHash, next.Artifacts)
				}
				repository, err := minigame.NewRepository(db)
				if err != nil {
					t.Fatal(err)
				}
				service, err := NewService(store, resolver, FrozenContributionProvider{DB: db}, nil, nil, WithProgressionRuntime(resolver),
					WithCurrentConstantsHash(next.ConstantsHash), WithReplayCatalogs(set), WithGuildSettlements(emptyGuildSettlements{}),
					WithMinigameActivity(repository), WithCompactPolicies(commons.CatalogSet{current.ConstantsHash: current.Commons.(commonsbinding.ReplayPolicy).Catalog, next.ConstantsHash: next.Commons.(commonsbinding.ReplayPolicy).Catalog}), WithCommonsWeightResolver(integrationWeight(1000000)))
				if err != nil {
					t.Fatal(err)
				}
				load := func(stream string) save.Loaded {
					value, err := store.LoadLatest(ctx, stream)
					if err != nil {
						t.Fatal(err)
					}
					return value
				}
				if activation {
					before := mustEncodeState(t, load(owner.StreamID).State)
					beforeCompany := load(player.StreamID)
					inactiveRequest := []byte(`{"intent_id":"01986666-e102-7000-8000-000000000001","kind":"purchase_reputation_node","expected_revision":1,"node_id":"reputation.unlock.p05"}`)
					result, err := service.Handle(ctx, player.StreamID, ModeOnline, now, inactiveRequest)
					if err != nil {
						t.Fatal(err)
					}
					category, detail := rejectionOf(t, result.Receipt)
					head := load(owner.StreamID)
					if category != "not_eligible" || detail != "reputation_tree_inactive" || head.Revision.Number != 1 || head.Revision.Version != 21 || head.Revision.ConstantsHash != current.ConstantsHash || !bytes.Equal(mustEncodeState(t, head.State), before) {
						t.Fatal("latest tree caused mid-run activation or changed inactive refusal")
					}
					companyHead := load(player.StreamID)
					if companyHead.Revision != beforeCompany.Revision || !bytes.Equal(mustEncodeState(t, companyHead.State), mustEncodeState(t, beforeCompany.State)) {
						t.Fatal("inactive mid-run purchase changed Company head")
					}
				}
				beforeOwner, beforePlayer := load(owner.StreamID), load(player.StreamID)
				beforeRows := reputationPlanDBSnapshot(t, ctx, db)
				body := map[string]any{"intent_id": "01986666-e103-7000-8000-000000000001", "kind": "wind_down", "expected_revision": 1, "expected_founder_revision": 1}
				if command != "wind_down" {
					body["kind"], body["offer_id"] = "accept_exit_offer", company.OfferState.OfferID
				}
				if profile.plan != nil {
					body["reputation_plan"] = profile.plan
				}
				request, err := json.Marshal(body)
				if err != nil {
					t.Fatal(err)
				}
				result, err := service.Handle(ctx, player.StreamID, ModeOnline, now, request)
				if err != nil || result.Replay {
					t.Fatalf("fresh boundary request err=%v replay=%v receipt=%s", err, result.Replay, result.Receipt)
				}
				afterOwner, afterPlayer := load(owner.StreamID), load(player.StreamID)
				if !activation {
					category, detail := rejectionOf(t, result.Receipt)
					if category != profile.category || detail != profile.detail {
						t.Fatalf("first failure=%s/%s want=%s/%s", category, detail, profile.category, profile.detail)
					}
					if afterOwner.Revision != beforeOwner.Revision || afterPlayer.Revision != beforePlayer.Revision || !bytes.Equal(mustEncodeState(t, afterOwner.State), mustEncodeState(t, beforeOwner.State)) || !bytes.Equal(mustEncodeState(t, afterPlayer.State), mustEncodeState(t, beforePlayer.State)) {
						t.Fatal("rejected prefix changed full heads or pending offer")
					}
					afterRows := reputationPlanDBSnapshot(t, ctx, db)
					for _, table := range []string{"save_streams", "save_revisions", "events", "run_epochs", "run_genesis", "run_frozen_contributions", "verification_queue"} {
						if beforeRows[table] != afterRows[table] {
							t.Fatalf("refused prefix changed %s", table)
						}
					}
					var logs, intents, outbox int
					if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM run_log),(SELECT count(*) FROM intent_records),(SELECT count(*) FROM transport_player_outbox)`).Scan(&logs, &intents, &outbox); err != nil || logs != 1 || intents != 1 || outbox < 1 {
						t.Fatalf("refusal not recorded: logs=%d intents=%d outbox=%d err=%v", logs, intents, outbox, err)
					}
				} else {
					var receipt struct {
						Outcome         string `json:"outcome"`
						NewRevision     int64  `json:"new_revision"`
						FounderRevision int64  `json:"founder_revision"`
					}
					if err := json.Unmarshal(result.Receipt, &receipt); err != nil || receipt.Outcome != "applied" {
						t.Fatalf("activation did not apply: %s/%v", result.Receipt, err)
					}
					if afterOwner.Revision.Number != 2 || afterPlayer.Revision.Number != 3 || afterOwner.Revision.Version != 22 || afterPlayer.Revision.Version != 18 || afterOwner.Revision.ConstantsHash != next.ConstantsHash || afterPlayer.Revision.ConstantsHash != next.ConstantsHash || afterPlayer.State.RunSeq != 3 || afterOwner.State.ReputationLevel != 6 {
						t.Fatal("cross-pin heads/versions/earned level differ")
					}
					if receipt.NewRevision != afterPlayer.Revision.Number || receipt.FounderRevision != afterOwner.Revision.Number {
						t.Fatal("applied receipt does not name committed heads")
					}
					planned := len(profile.plan) != 0
					wantSpent, wantUnlock, wantFactor, wantKind, wantPurchases := int64(0), int64(0), "1e0", "exit.v1", 0
					wantOwned := []string{}
					if planned {
						wantSpent, wantUnlock, wantFactor, wantKind, wantPurchases = 6, 50000, "1.003e0", "exit.v2", 3
						wantOwned = []string{good[1], good[2], good[0]}
					}
					if afterOwner.State.ReputationSpent != wantSpent || afterOwner.State.ReputationUnlockPPM != wantUnlock || !slices.Equal(afterOwner.State.ReputationNodesOwned, wantOwned) || afterPlayer.State.OfferState != nil {
						t.Fatal("activation plan state/accounting differs")
					}
					var factor, pinHash, startedFactor, arm string
					var pinEpoch, purchases int
					if err := db.QueryRowContext(ctx, `SELECT (SELECT factor FROM run_frozen_contributions WHERE company_stream_id=$1 AND run_seq=3 AND source_id='reputation.founder_bonus'),
						(SELECT constants_hash FROM run_epochs WHERE company_stream_id=$1 AND run_seq=3),(SELECT epoch_id FROM run_epochs WHERE company_stream_id=$1 AND run_seq=3),
						(SELECT payload->'reputation_tree'->>'bonus_factor' FROM events WHERE stream_id=$1 AND kind='run_started'),
						(SELECT replay_inputs->'resolved'->>'kind' FROM founder_log WHERE founder_stream_id=$2 ORDER BY seq DESC LIMIT 1),
						(SELECT count(*) FROM events WHERE stream_id=$2 AND kind='reputation_node_purchased.v1')`, player.StreamID, owner.StreamID).Scan(&factor, &pinHash, &pinEpoch, &startedFactor, &arm, &purchases); err != nil || factor != wantFactor || startedFactor != wantFactor || pinHash != next.ConstantsHash || pinEpoch != 2 || arm != wantKind || purchases != wantPurchases {
						t.Fatalf("cross-pin persisted summary differs: %s/%s pin=%s/%d arm=%s purchases=%d err=%v", factor, startedFactor, pinHash, pinEpoch, arm, purchases, err)
					}
					if planned {
						cash, ok := afterPlayer.State.Ledger.Balance("company.cash")
						if !ok || cash.String() != "1e3" || afterPlayer.State.GeneratorProvisioned["generator.beige_tower"] != 5 || afterPlayer.State.GeneratorCounts["generator.beige_tower"] != 0 || afterPlayer.State.GeneratorPurchasedTotal != 0 {
							t.Fatal("cross-pin starter assembly differs")
						}
					}
					genesis, oldVersion, entries := reputationCareerReplay(t, db, player.StreamID, owner.StreamID, 2, &next)
					if len(entries) != 1 || VerifyReplayRun(genesis, oldVersion, current, entries, current.ConstantsHash, false) != ReplayVerified {
						t.Fatal("completed old-pin run did not verify with next bundle")
					}
					if VerifyReplayRun(genesis, oldVersion, current, entries, next.ConstantsHash, false) != ReplayConstantsMismatch {
						t.Fatal("current hash substituted for completed run's old pin")
					}
					state, err := save.RestoreState(genesis, oldVersion, current.Economy, economy.ScopeCompany, time.Time{})
					if err != nil {
						t.Fatal(err)
					}
					linked := current
					linked.Next = &next
					terminal, err := ApplyLoggedExit(state, entries[0].CanonicalPayload, linked, entries[0].ReplayInputs)
					if err != nil || !bytes.Equal(mustEncodeState(t, terminal.Decision.NewCompanyState), mustEncodeState(t, afterPlayer.State)) {
						t.Fatalf("new Company replay differs from complete saved head: %v", err)
					}
					var started struct {
						ReputationTree reputationRunStarted `json:"reputation_tree"`
					}
					if len(terminal.Decision.CompanyStartedEvents) != 1 || terminal.Decision.CompanyStartedEvents[0].SchemaVersion != 2 {
						t.Fatal("missing v2 start summary")
					}
					if err := json.Unmarshal(terminal.Decision.CompanyStartedEvents[0].Payload, &started); err != nil {
						t.Fatal(err)
					}
					wantIDs := []string{}
					if planned {
						wantIDs = good[1:]
					}
					if started.ReputationTree.BonusFactor != wantFactor || !slices.Equal(started.ReputationTree.AppliedStarterNodeIDs, wantIDs) {
						t.Fatal("complete starter summary differs")
					}
				}
				var unchangedPin string
				if err := db.QueryRowContext(ctx, `SELECT to_jsonb(row)::text FROM run_epochs row WHERE company_stream_id=$1 AND run_seq=2`, player.StreamID).Scan(&unchangedPin); err != nil || unchangedPin != oldPin {
					t.Fatal("old run pin changed")
				}
				history, err := store.LoadFounderHistory(ctx, owner.StreamID)
				wantEntries := 1
				if activation {
					wantEntries = 2
				}
				if err != nil || len(history.Entries) != wantEntries || VerifyFounderHistory(history, set) != ReplayVerified {
					t.Fatalf("complete Founder history did not verify: entries=%d err=%v", len(history.Entries), err)
				}
				if activation {
					last := history.Entries[len(history.Entries)-1]
					wantEvents := 1
					if len(profile.plan) != 0 {
						wantEvents = 4
					}
					if len(last.Events) != wantEvents || last.Events[0].Kind != save.EventFounderAdvanced {
						t.Fatal("activated Founder events count/order differs")
					}
					if len(profile.plan) != 0 {
						for index, id := range good {
							var purchase struct {
								NodeID string `json:"node_id"`
								Cost   int64  `json:"cost"`
								Source string `json:"source"`
							}
							if err := json.Unmarshal(last.Events[index+1].Payload, &purchase); err != nil {
								t.Fatal(err)
							}
							if last.Events[index+1].Kind != save.EventReputationNodePurchased || last.Events[index+1].SchemaVersion != 1 || purchase.NodeID != id || purchase.Cost != int64(index+1) || purchase.Source != "exit_plan" {
								t.Fatal("activated ordered purchase payload differs")
							}
						}
					}
					copied := history
					copied.HeadConstants = current.ConstantsHash
					if VerifyFounderHistory(copied, set) != ReplayStateDivergence {
						t.Fatal("copied wrong Founder head pin did not diverge")
					}
				} else {
					genesis, oldVersion, entries := reputationCareerReplay(t, db, player.StreamID, owner.StreamID, 2, &next)
					if len(entries) != 1 {
						t.Fatal("refusal run-log population differs")
					}
					state, err := save.RestoreState(genesis, oldVersion, current.Economy, economy.ScopeCompany, time.Time{})
					if err != nil {
						t.Fatal(err)
					}
					linked := current
					linked.Next = &next
					replayed, err := ApplyLoggedExit(state, entries[0].CanonicalPayload, linked, entries[0].ReplayInputs)
					// Postgres jsonb and the codec reorder/compact nested raw offer
					// terms. Compare every JSON value with exact-number canonicalization,
					// not incidental key order or whitespace; persisted heads above
					// still require byte equality before/after the rejected transaction.
					wantState := canonicalFixtureJSON(t, mustEncodeState(t, beforePlayer.State))
					if err != nil || replayed.Decision.Outcome != save.IntentRejected || !canonicalJSONEqual(replayed.Decision.Receipt, entries[0].ReceiptJSON) || len(replayed.Decision.FounderEvents) != 0 || len(replayed.Decision.CompanyEndedEvents) != 0 || canonicalFixtureJSON(t, mustEncodeState(t, state)) != wantState {
						t.Fatalf("recorded refusal replay differs: %v", err)
					}
					copied, err := save.RestoreState(genesis, oldVersion, current.Economy, economy.ScopeCompany, time.Time{})
					if err != nil {
						t.Fatal(err)
					}
					copied.Tier++
					if canonicalFixtureJSON(t, mustEncodeState(t, copied)) == wantState {
						t.Fatal("full canonical state oracle missed a changed tier")
					}
					// A logged Exit attempt is not a completed run when rejected.
				}
				committed := reputationPlanDBSnapshot(t, ctx, db)
				retry, err := service.Handle(ctx, player.StreamID, ModeOnline, now.Add(time.Second), request)
				if err != nil || !retry.Replay || !bytes.Equal(retry.Receipt, result.Receipt) {
					t.Fatalf("boundary exact retry differs: %v", err)
				}
				for table, rows := range reputationPlanDBSnapshot(t, ctx, db) {
					if rows != committed[table] {
						t.Fatalf("retry rewrote %s", table)
					}
				}
				if profile.name == "next-inactive" {
					fallbackRequest := []byte(`{"intent_id":"01986666-e104-7000-8000-000000000001","kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`)
					fallback, err := service.Handle(ctx, player.StreamID, ModeOnline, now.Add(time.Second), fallbackRequest)
					if err != nil {
						t.Fatal(err)
					}
					var receipt struct {
						Outcome string `json:"outcome"`
					}
					if err := json.Unmarshal(fallback.Receipt, &receipt); err != nil || receipt.Outcome != "applied" || fallback.Replay {
						t.Fatalf("no-plan WindDown door closed: %s/%v", fallback.Receipt, err)
					}
					ownerHead, companyHead := load(owner.StreamID), load(player.StreamID)
					if ownerHead.Revision.Number != 2 || ownerHead.Revision.Version != 21 || ownerHead.Revision.ConstantsHash != current.ConstantsHash || ownerHead.State.ReputationLevel != 6 || ownerHead.State.ReputationSpent != 0 || len(ownerHead.State.ReputationNodesOwned) != 0 || companyHead.Revision.Number != 3 || companyHead.State.RunSeq != 3 || companyHead.Revision.ConstantsHash != current.ConstantsHash || companyHead.State.OfferState != nil {
						t.Fatal("absent-tree fallback changed progression/pin contract")
					}
					var purchases, reputationRows int
					if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM events WHERE kind='reputation_node_purchased.v1'),(SELECT count(*) FROM run_frozen_contributions WHERE company_stream_id=$1 AND run_seq=3 AND source_id='reputation.founder_bonus')`, player.StreamID).Scan(&purchases, &reputationRows); err != nil || purchases != 0 || reputationRows != 0 {
						t.Fatalf("fallback activated absent Reputation tree: purchases=%d rows=%d err=%v", purchases, reputationRows, err)
					}
					fallbackHistory, err := store.LoadFounderHistory(ctx, owner.StreamID)
					if err != nil || len(fallbackHistory.Entries) != 2 || VerifyFounderHistory(fallbackHistory, set) != ReplayVerified {
						t.Fatal("fallback Founder history does not verify")
					}
					genesis, oldVersion, entries := reputationCareerReplay(t, db, player.StreamID, owner.StreamID, 2, &next)
					if len(entries) != 2 || VerifyReplayRun(genesis, oldVersion, current, entries, current.ConstantsHash, false) != ReplayVerified {
						t.Fatal("refused then no-plan completed run does not verify")
					}
					committedFallback := reputationPlanDBSnapshot(t, ctx, db)
					fallbackRetry, err := service.Handle(ctx, player.StreamID, ModeOnline, now.Add(2*time.Second), fallbackRequest)
					if err != nil || !fallbackRetry.Replay || !bytes.Equal(fallbackRetry.Receipt, fallback.Receipt) {
						t.Fatal("fallback exact retry differs")
					}
					for table, rows := range reputationPlanDBSnapshot(t, ctx, db) {
						if rows != committedFallback[table] {
							t.Fatalf("fallback retry rewrote %s", table)
						}
					}
					fallbackCases++
				}
				if activation {
					appliedCases++
				} else {
					refusedCases++
				}
			})
		}
	}
	if executed != 24 || appliedCases != 9 || refusedCases != 15 || fallbackCases != 3 {
		t.Errorf("full boundary population executed=%d applied=%d refused=%d fallback=%d want24/9/15/3", executed, appliedCases, refusedCases, fallbackCases)
	}
}
