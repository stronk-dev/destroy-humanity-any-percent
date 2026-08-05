package production

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/leaderboard"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

type prestigeNoopProjector struct{}

func (prestigeNoopProjector) Project(context.Context, []save.EventRecord) error        { return nil }
func (prestigeNoopProjector) RepairFounder(context.Context, string, *save.State) error { return nil }

func TestPrestigeWindDownAndScriptedExitIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
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
	if _, err := db.ExecContext(ctx, `TRUNCATE epochs,catalog_sets,save_streams RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	bundle, err := epochseed.Load(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	replayBundle := loadReplayTestBundle(t, bundle.Hash, bundle.Artifacts)
	economyBytes := bundle.Artifacts["economy"]
	routeBytes := bundle.Artifacts["routes"]
	prestigeBytes := bundle.Artifacts["prestige"]
	catalog, err := economy.LoadCatalog(economyBytes)
	if err != nil {
		t.Fatal(err)
	}
	routeCatalog, err := routes.LoadCatalog(routeBytes)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := prestigecore.LoadPolicy(prestigeBytes)
	if err != nil {
		t.Fatal(err)
	}
	commonsCatalog, err := commons.LoadCatalog(bundle.Artifacts["commons"])
	if err != nil {
		t.Fatal(err)
	}
	factionCatalog, err := faction.LoadCatalog(bundle.Artifacts["factions"], faction.CompactTitheBand{
		MinimumPPM: commonsCatalog.MinimumTithePPM,
		DefaultPPM: commonsCatalog.DefaultTithePPM,
		MaximumPPM: commonsCatalog.MaximumTithePPM,
	})
	if err != nil {
		t.Fatal(err)
	}
	hash := bundle.Hash
	seedProductionEpoch(t, db, hash, bundle.Artifacts)
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{hash: catalog}, routes: map[string]*routes.Catalog{hash: routeCatalog}, prestige: map[string]*prestigecore.Policy{hash: policy}, factions: map[string]*faction.Catalog{hash: factionCatalog}}
	commonsCatalogs := commons.CatalogSet{hash: commonsCatalog}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	projectionFailure := &failNextProjection{}
	service, err := NewService(store, resolver, nil, nil, nil, WithRouteCatalogs(resolver), WithRouteProjector(prestigeNoopProjector{}), WithCompactPolicies(commonsCatalogs), WithProgressionRuntime(resolver), WithCurrentConstantsHash(hash), WithCommonsWeightResolver(integrationWeight(1_000_000)), WithReplayCatalogs(ReplayCatalogSet{hash: replayBundle}), WithGuildSettlements(emptyGuildSettlements{}), WithEventProjector(projectionFailure))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 29, 16, 0, 0, 0, time.UTC)

	t.Run("elective wind down", func(t *testing.T) {
		owner := "01985555-1000-7000-8000-000000000001"
		founderRevision, companyRevision := createPrestigeStreams(t, ctx, store, catalog, hash, owner, now, now.Add(-10*time.Minute), now, "0", decimal.New(8, 12), 1)
		if _, err := store.PinRunToCurrentEpoch(ctx, companyRevision.StreamID, owner, 1, hash); err != nil {
			t.Fatal(err)
		}
		companyBefore, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil {
			t.Fatal(err)
		}
		routePayload, _ := json.Marshal(map[string]any{"route_id": "route.nonprofit_wrapper_zip", "gate_id": "gate.t1_to_t2", "run_id": map[string]any{"company_stream_id": companyRevision.StreamID, "run_seq": 1}, "founder_id": owner})
		if _, err := db.ExecContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,1,1,'route_executed',$2,$3,$4,$5)`, companyRevision.StreamID, "01985555-1000-7000-8000-000000000099", hash, now.Add(-time.Minute), routePayload); err != nil {
			t.Fatal(err)
		}
		request := []byte(`{"intent_id":"01985555-1001-7000-8000-000000000001","kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`)
		projectionFailure.fail = true
		result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, request)
		if !errors.Is(err, ErrInvalidEngineState) {
			t.Fatalf("post-commit Exit projection error=%v", err)
		}
		result, err = service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, request)
		if err != nil || !result.Replay {
			t.Fatalf("result=%+v err=%v", result, err)
		}
		var receipt struct {
			Outcome         string `json:"outcome"`
			NewRevision     int64  `json:"new_revision"`
			FounderRevision int64  `json:"founder_revision"`
		}
		if err := json.Unmarshal(result.Receipt, &receipt); err != nil || receipt.Outcome != "applied" || receipt.NewRevision != 3 || receipt.FounderRevision != 2 {
			t.Fatalf("receipt=%s err=%v", result.Receipt, err)
		}
		founder, err := store.LoadLatest(ctx, founderRevision.StreamID)
		if err != nil || founder.State.ReputationLevel != 2 || founder.State.RouteKnowledgeBalance != 25 || len(founder.State.ExitHistory) != 1 || founder.State.ExitHistory[0].ExitType != "scripted_first" {
			t.Fatalf("founder=%+v err=%v", founder.State, err)
		}
		company, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil || company.Revision.Number != 3 || company.State.RunSeq != 2 || company.State.Tier != 0 || !company.State.LifetimeValue.Eq(decimal.Zero) {
			t.Fatalf("company=%+v err=%v", company, err)
		}
		var endedPayload []byte
		if err := db.QueryRowContext(ctx, `SELECT payload FROM events WHERE stream_id=$1 AND kind='run_ended'`, companyRevision.StreamID).Scan(&endedPayload); err != nil {
			t.Fatal(err)
		}
		var ended struct {
			ExecutedRoutes []string `json:"executed_routes"`
			TerminalSeq    int64    `json:"terminal_seq"`
		}
		if err := json.Unmarshal(endedPayload, &ended); err != nil || len(ended.ExecutedRoutes) != 1 || ended.ExecutedRoutes[0] != "route.nonprofit_wrapper_zip" || ended.TerminalSeq != 1 {
			t.Fatalf("run_ended=%s err=%v", endedPayload, err)
		}
		var sequence int64
		var loggedPayload, replayInputs []byte
		if err := db.QueryRowContext(ctx, `SELECT seq,canonical_payload,replay_inputs FROM run_log WHERE company_stream_id=$1 AND run_seq=1`, companyRevision.StreamID).Scan(&sequence, &loggedPayload, &replayInputs); err != nil || sequence != ended.TerminalSeq {
			t.Fatalf("run log sequence=%d terminal=%d payload=%s err=%v", sequence, ended.TerminalSeq, loggedPayload, err)
		}
		parsedInputs, err := parseReplayInputs(replayInputs)
		if err != nil {
			t.Fatal(err)
		}
		var terminal replayExitResolved
		if err := decodeReplayStrict(parsedInputs.Resolved, &terminal); err != nil || terminal.Kind != "exit" || terminal.SelectedExitType != "scripted_first" || terminal.NextConstantsHash != hash || len(terminal.ExecutedRouteIDs) != 1 {
			t.Fatalf("terminal replay inputs=%s err=%v", replayInputs, err)
		}
		var founderLogSeq, sourceRunSeq, sourceLogSeq int64
		var sourceStream, founderInputHash, founderKind string
		var founderAudit []byte
		if err := db.QueryRowContext(ctx, `SELECT seq,source_company_stream_id,source_run_seq,
			source_run_log_seq,constants_hash,replay_inputs->'resolved'->>'kind',receipt
			FROM founder_log WHERE founder_stream_id=$1 AND intent_id=$2`, founderRevision.StreamID,
			parsedInputs.Command.IntentID).Scan(&founderLogSeq, &sourceStream, &sourceRunSeq, &sourceLogSeq,
			&founderInputHash, &founderKind, &founderAudit); err != nil || founderLogSeq != 1 ||
			sourceStream != companyRevision.StreamID || sourceRunSeq != 1 || sourceLogSeq != sequence ||
			founderInputHash != hash || founderKind != founderExitResolvedKind {
			t.Fatalf("Founder Exit log seq=%d source=%s/%d/%d hash=%q kind=%q audit=%s err=%v",
				founderLogSeq, sourceStream, sourceRunSeq, sourceLogSeq, founderInputHash, founderKind,
				founderAudit, err)
		}
		var audit founderExitAuditReceipt
		if err := json.Unmarshal(founderAudit, &audit); err != nil || audit.Outcome != "applied" ||
			audit.FounderRevision == nil || *audit.FounderRevision != 2 ||
			audit.ResultConstantsHash != hash {
			t.Fatalf("Founder Exit audit=%+v raw=%s err=%v", audit, founderAudit, err)
		}
		founderGenesis, err := store.LoadFounderGenesis(ctx, founderRevision.StreamID)
		if err != nil || founderGenesis.Revision != 1 || founderGenesis.ConstantsHash != hash {
			t.Fatalf("Founder Exit genesis=%+v err=%v", founderGenesis, err)
		}
		terminalBundle := replayBundle
		terminalBundle.Next = &terminalBundle
		replayed, err := ApplyLoggedExit(companyBefore.State, loggedPayload, terminalBundle, replayInputs)
		if err != nil || string(replayed.Decision.Receipt) != string(result.Receipt) {
			t.Fatalf("terminal replay receipt=%s live=%s err=%v", replayed.Decision.Receipt, result.Receipt, err)
		}
		expectedEvents := append([]save.EventWrite(nil), replayed.Decision.FounderEvents...)
		expectedEvents = append(expectedEvents, replayed.Decision.CompanyEndedEvents...)
		expectedEvents = append(expectedEvents, replayed.Decision.CompanyStartedEvents...)
		rows, err := db.QueryContext(ctx, `SELECT kind,payload FROM events WHERE intent_id=$1 ORDER BY event_seq`, parsedInputs.Command.IntentID)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		index := 0
		for rows.Next() {
			var kind save.EventKind
			var payload []byte
			if err := rows.Scan(&kind, &payload); err != nil || index >= len(expectedEvents) || expectedEvents[index].Kind != kind || !jsonSemanticallyEqual(expectedEvents[index].Payload, payload) {
				t.Fatalf("terminal event[%d] replay=%v live=%s/%s err=%v", index, expectedEvents, kind, payload, err)
			}
			index++
		}
		if err := rows.Err(); err != nil || index != len(expectedEvents) {
			t.Fatalf("terminal replay events=%d live=%d err=%v", len(expectedEvents), index, err)
		}
		replay, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, request)
		if err != nil || !replay.Replay || string(replay.Receipt) != string(result.Receipt) {
			t.Fatalf("replay=%+v err=%v", replay, err)
		}
	})

	t.Run("incorporated run exits and may incorporate again", func(t *testing.T) {
		owner := "01985555-1500-7000-8000-000000000001"
		founderRevision, companyRevision := createPrestigeStreams(t, ctx, store, catalog, hash, owner, now, now.Add(-10*time.Minute), now, "0", decimal.New(8, 12), 2)
		if _, err := store.PinRunToCurrentEpoch(ctx, companyRevision.StreamID, owner, 1, hash); err != nil {
			t.Fatal(err)
		}
		incorporate := []byte(`{"intent_id":"01985555-1501-7000-8000-000000000001","kind":"incorporate","expected_revision":1,"faction_id":"open_source"}`)
		if result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, incorporate); err != nil || result.Replay {
			t.Fatalf("incorporate=%+v err=%v", result, err)
		}
		exit := []byte(`{"intent_id":"01985555-1502-7000-8000-000000000001","kind":"wind_down","expected_revision":2,"expected_founder_revision":1}`)
		if result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now.Add(time.Minute), exit); err != nil || result.Replay {
			t.Fatalf("exit=%+v err=%v", result, err)
		}
		var endedFaction string
		if err := db.QueryRowContext(ctx, `SELECT payload->>'faction' FROM events WHERE stream_id=$1 AND kind='run_ended'`, companyRevision.StreamID).Scan(&endedFaction); err != nil || endedFaction != "open_source" {
			t.Fatalf("ended faction=%v err=%v", endedFaction, err)
		}
		company, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil || company.Revision.Number != 4 || company.State.RunSeq != 2 || company.State.FactionID != "" {
			t.Fatalf("new run company=%+v err=%v", company, err)
		}
		advanced, err := store.ApplyIntent(ctx, companyRevision.StreamID, company.Revision.Number, "01985555-1503-7000-8000-000000000001", "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", func(state *save.State, revision save.Revision) (save.IntentDecision, error) {
			state.Tier = 2
			return save.IntentDecision{Outcome: save.IntentApplied, Receipt: json.RawMessage(`{"outcome":"applied","new_revision":5}`)}, nil
		})
		if err != nil || advanced.Outcome != save.IntentApplied {
			t.Fatalf("test progression=%+v err=%v", advanced, err)
		}
		reincorporate := []byte(`{"intent_id":"01985555-1504-7000-8000-000000000001","kind":"incorporate","expected_revision":5,"faction_id":"enterprise"}`)
		if result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now.Add(2*time.Minute), reincorporate); err != nil || result.Replay {
			t.Fatalf("reincorporate=%+v err=%v", result, err)
		}
		company, err = store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil || company.State.FactionID != "enterprise" {
			t.Fatalf("reincorporated company=%+v err=%v", company, err)
		}
		founder, err := store.LoadLatest(ctx, founderRevision.StreamID)
		if err != nil || len(founder.State.ExitHistory) != 1 {
			t.Fatalf("founder=%+v err=%v", founder, err)
		}
	})

	t.Run("scripted first threshold", func(t *testing.T) {
		owner := "01985555-2000-7000-8000-000000000002"
		founderRevision, companyRevision := createPrestigeStreams(t, ctx, store, catalog, hash, owner, now, now.Add(-16*time.Minute), now.Add(-time.Second), "1e10", decimal.New(1, 12), 2)
		if _, err := store.PinRunToCurrentEpoch(ctx, companyRevision.StreamID, owner, 1, hash); err != nil {
			t.Fatal(err)
		}
		request := []byte(`{"intent_id":"01985555-2001-7000-8000-000000000002","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t2_to_t3","route_id":null}`)
		result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, request)
		if err != nil || result.Replay {
			t.Fatalf("result=%+v err=%v receipt=%s", result, err, result.Receipt)
		}
		founder, err := store.LoadLatest(ctx, founderRevision.StreamID)
		if err != nil || len(founder.State.ExitHistory) != 1 || founder.State.ExitHistory[0].ExitType != "scripted_first" || founder.State.ReputationLevel != 1 {
			t.Fatalf("founder=%+v err=%v", founder.State, err)
		}
		var kinds []string
		rows, err := db.QueryContext(ctx, `SELECT kind FROM events WHERE stream_id=$1 ORDER BY revision,kind`, companyRevision.StreamID)
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			var kind string
			_ = rows.Scan(&kind)
			kinds = append(kinds, kind)
		}
		rows.Close()
		if len(kinds) != 3 || kinds[0] != "gate_crossed" || kinds[1] != "run_ended" || kinds[2] != "run_started" {
			t.Fatalf("company event kinds=%v", kinds)
		}
	})

	t.Run("offer preview remains a promise", func(t *testing.T) {
		owner := "01985555-3000-7000-8000-000000000003"
		founderRevision, companyRevision := createPrestigeStreams(t, ctx, store, catalog, hash, owner, now, now, now, "1e25", decimal.New(8, 12), 7)
		founderLoaded, err := store.LoadLatest(ctx, founderRevision.StreamID)
		if err != nil {
			t.Fatal(err)
		}
		founderLoaded.State.ExitHistory = append(founderLoaded.State.ExitHistory, save.ExitRecord{RunID: 1, ExitType: "collapse", OccurredAt: now.Add(-time.Hour), ReputationDelta: 1})
		founderRevision, err = store.Write(ctx, founderRevision.StreamID, founderRevision.Number, hash, founderLoaded.State, save.WriteContext{Cause: "prestige.test.prior_exit"})
		if err != nil {
			t.Fatal(err)
		}
		companyLoaded, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil {
			t.Fatal(err)
		}
		companyLoaded.State.RunSeq = 2
		companyRevision, err = store.Write(ctx, companyRevision.StreamID, companyRevision.Number, hash, companyLoaded.State, save.WriteContext{Cause: "prestige.test.second_run"})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := store.PinRunToCurrentEpoch(ctx, companyRevision.StreamID, owner, 2, hash); err != nil {
			t.Fatal(err)
		}
		cross := []byte(`{"intent_id":"01985555-3001-7000-8000-000000000003","kind":"cross_gate","expected_revision":2,"gate_id":"gate.t7_to_t8","route_id":null}`)
		if result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, cross); err != nil || result.Replay {
			t.Fatalf("cross=%+v err=%v", result, err)
		}
		offered, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil || offered.State.OfferState == nil {
			t.Fatalf("offered=%+v err=%v", offered.State, err)
		}
		stored, err := prestigecore.DecodeStoredOfferTerms(offered.State.OfferState.TermsJSON)
		if err != nil {
			t.Fatal(err)
		}
		offerID := offered.State.OfferState.OfferID
		offered.State.LifetimeValue = decimal.New(27, 12)
		written, err := store.Write(ctx, companyRevision.StreamID, 3, hash, offered.State, save.WriteContext{Cause: "prestige.test.progress"})
		if err != nil || written.Number != 4 {
			t.Fatalf("write=%+v err=%v", written, err)
		}
		acceptBody, _ := json.Marshal(map[string]any{"intent_id": "01985555-3002-7000-8000-000000000003", "kind": "accept_exit_offer", "expected_revision": 4, "expected_founder_revision": founderRevision.Number, "offer_id": offerID})
		if result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, acceptBody); err != nil || result.Replay {
			t.Fatalf("accept=%+v err=%v", result, err)
		}
		founder, err := store.LoadLatest(ctx, founderRevision.StreamID)
		if err != nil || founder.State.ReputationLevel < stored.PayoutPreview.ReputationDelta {
			t.Fatalf("founder reputation=%d preview=%d err=%v", founder.State.ReputationLevel, stored.PayoutPreview.ReputationDelta, err)
		}
	})

	t.Run("offers wait for scripted first exit", func(t *testing.T) {
		owner := "01985555-3500-7000-8000-000000000003"
		_, companyRevision := createPrestigeStreams(t, ctx, store, catalog, hash, owner, now, now, now, "1e25", decimal.New(8, 12), 7)
		if _, err := store.PinRunToCurrentEpoch(ctx, companyRevision.StreamID, owner, 1, hash); err != nil {
			t.Fatal(err)
		}
		cross := []byte(`{"intent_id":"01985555-3501-7000-8000-000000000003","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t7_to_t8","route_id":null}`)
		if result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, cross); err != nil || result.Replay {
			t.Fatalf("cross=%+v err=%v", result, err)
		}
		company, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil || company.State.OfferState != nil {
			t.Fatalf("company=%+v err=%v", company.State, err)
		}
		var offers int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM events WHERE stream_id=$1 AND kind='exit_offer_spawned'`, companyRevision.StreamID).Scan(&offers); err != nil || offers != 0 {
			t.Fatalf("offers=%d err=%v", offers, err)
		}
	})

	ownerAcrossEpoch := "01985555-4000-7000-8000-000000000004"
	founderAcrossEpoch, companyAcrossEpoch := createPrestigeStreams(t, ctx, store, catalog, hash, ownerAcrossEpoch, now, now.Add(-10*time.Minute), now, "0", decimal.New(8, 12), 1)
	if _, err := store.PinRunToCurrentEpoch(ctx, companyAcrossEpoch.StreamID, ownerAcrossEpoch, 1, hash); err != nil {
		t.Fatal(err)
	}
	changedEconomyBytes := append(append([]byte(nil), economyBytes...), '\n')
	changedArtifacts := make(map[string][]byte, len(bundle.Artifacts))
	for name, data := range bundle.Artifacts {
		changedArtifacts[name] = append([]byte(nil), data...)
	}
	changedArtifacts["economy"] = changedEconomyBytes
	currentHash, err := save.ConstantsHashArtifacts(changedArtifacts)
	if err != nil || currentHash == hash {
		t.Fatalf("changed hash=%s old=%s err=%v", currentHash, hash, err)
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "changelog"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "changelog", "epoch-2.md"), []byte("# changed balance bytes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	epochRepository, err := leaderboard.NewRepository(db, root)
	if err != nil {
		t.Fatal(err)
	}
	artifacts := make([]leaderboard.Artifact, 0, len(changedArtifacts))
	for _, declaration := range bundle.Seed.Artifacts {
		artifacts = append(artifacts, leaderboard.Artifact{Name: declaration.Name, Bytes: changedArtifacts[declaration.Name]})
	}
	var firstEpochStarted time.Time
	if err := db.QueryRowContext(ctx, `SELECT started_at FROM epochs WHERE ended_at IS NULL`).Scan(&firstEpochStarted); err != nil {
		t.Fatal(err)
	}
	secondEpoch, err := epochRepository.MintEpoch(ctx, "Phase 0.1", firstEpochStarted.Add(time.Hour), "changelog/epoch-2.md", artifacts)
	if err != nil || secondEpoch.ID != 2 || secondEpoch.Hashes[0] != currentHash {
		t.Fatalf("second epoch=%+v err=%v", secondEpoch, err)
	}
	resolver.economy[currentHash], resolver.routes[currentHash], resolver.prestige[currentHash], resolver.factions[currentHash] = catalog, routeCatalog, policy, factionCatalog
	commonsCatalogs[currentHash] = commonsCatalog
	currentReplayBundle := loadReplayTestBundle(t, currentHash, changedArtifacts)
	currentService, err := NewService(store, resolver, nil, nil, nil, WithRouteCatalogs(resolver), WithRouteProjector(prestigeNoopProjector{}), WithCompactPolicies(commonsCatalogs), WithProgressionRuntime(resolver), WithCurrentConstantsHash(currentHash), WithCommonsWeightResolver(integrationWeight(1_000_000)), WithReplayCatalogs(ReplayCatalogSet{hash: replayBundle, currentHash: currentReplayBundle}), WithGuildSettlements(emptyGuildSettlements{}))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("epoch mint moves only the new run to changed bytes", func(t *testing.T) {
		request := []byte(`{"intent_id":"01985555-4001-7000-8000-000000000004","kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`)
		if result, err := currentService.Handle(ctx, companyAcrossEpoch.StreamID, ModeOnline, now.Add(time.Hour), request); err != nil || result.Replay {
			t.Fatalf("epoch-crossing exit=%+v err=%v", result, err)
		}
		company, err := store.LoadLatest(ctx, companyAcrossEpoch.StreamID)
		if err != nil || company.Revision.ConstantsHash != currentHash || company.State.RunSeq != 2 {
			t.Fatalf("company revision=%+v state=%+v err=%v", company.Revision, company.State, err)
		}
		founder, err := store.LoadLatest(ctx, founderAcrossEpoch.StreamID)
		if err != nil || founder.Revision.ConstantsHash != currentHash {
			t.Fatalf("founder revision=%+v err=%v", founder.Revision, err)
		}
		var oldEpoch, newEpoch int64
		var oldHash, newHash string
		if err := db.QueryRowContext(ctx, `SELECT epoch_id,constants_hash FROM run_epochs WHERE company_stream_id=$1 AND run_seq=1`, companyAcrossEpoch.StreamID).Scan(&oldEpoch, &oldHash); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRowContext(ctx, `SELECT epoch_id,constants_hash FROM run_epochs WHERE company_stream_id=$1 AND run_seq=2`, companyAcrossEpoch.StreamID).Scan(&newEpoch, &newHash); err != nil || oldEpoch != 1 || oldHash != hash || newEpoch != 2 || newHash != currentHash {
			t.Fatalf("pins old=(%d,%s) new=(%d,%s) err=%v", oldEpoch, oldHash, newEpoch, newHash, err)
		}
		rows, err := db.QueryContext(ctx, `SELECT kind,constants_hash FROM events WHERE stream_id=$1 AND kind IN ('run_ended','run_started') ORDER BY revision`, companyAcrossEpoch.StreamID)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var pairs [][2]string
		for rows.Next() {
			var pair [2]string
			if err := rows.Scan(&pair[0], &pair[1]); err != nil {
				t.Fatal(err)
			}
			pairs = append(pairs, pair)
		}
		if len(pairs) != 2 || pairs[0] != [2]string{"run_ended", hash} || pairs[1] != [2]string{"run_started", currentHash} {
			t.Fatalf("transition events=%v", pairs)
		}
	})

	t.Run("company v6 migrates to an exitable pre-timer run", func(t *testing.T) {
		owner := "01985555-5000-7000-8000-000000000005"
		founderRevision, companyRevision := createPrestigeStreams(t, ctx, store, catalog, currentHash, owner, now, now, now, "1e10", decimal.Zero, 0)
		loaded, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := save.EncodeState(loaded.State)
		if err != nil {
			t.Fatal(err)
		}
		var legacy map[string]any
		if err := json.Unmarshal(encoded, &legacy); err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"tier", "lifetime_value", "offer_state", "run_started_at_ms", "run_pre_timer", "offline_spans", "collapsed_offline_ms", "reputation_level", "reputation_unlock_ppm", "network_slots", "clout_lifetime", "soul", "age_ms", "notoriety", "advisor_mode", "exit_history", "faction_id", "incorporated_at_ms", "stock_units", "stock_progress_ms", "consumed_stock_units", "guild_tithe_carry_ppm", "guild_boundary_seq", "guild_consumed_window_units", "guild_boundary_guild_id", "generators_purchased_total", "upgrades_owned", "generators_provisioned", "provision_remainders_ppm", "stock_rate_remainder_ppm"} {
			delete(legacy, key)
		}
		legacyBytes, _ := json.Marshal(legacy)
		if _, err := db.ExecContext(ctx, `UPDATE save_revisions SET version=6,state=$2 WHERE stream_id=$1 AND revision=1`, companyRevision.StreamID, legacyBytes); err != nil {
			t.Fatal(err)
		}
		if _, err := store.PinRunToCurrentEpoch(ctx, companyRevision.StreamID, owner, 1, currentHash); err != nil {
			t.Fatal(err)
		}
		request := []byte(`{"intent_id":"01985555-5001-7000-8000-000000000005","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t2_to_t3","route_id":null}`)
		result, err := currentService.Handle(ctx, companyRevision.StreamID, ModeOnline, now.Add(16*time.Minute), request)
		if err != nil || result.Replay {
			t.Fatalf("v6 cross=%+v err=%v", result, err)
		}
		exit := []byte(`{"intent_id":"01985555-5002-7000-8000-000000000005","kind":"wind_down","expected_revision":2,"expected_founder_revision":1}`)
		result, err = currentService.Handle(ctx, companyRevision.StreamID, ModeOnline, now.Add(16*time.Minute), exit)
		if err != nil || result.Replay {
			t.Fatalf("v6 exit=%+v err=%v", result, err)
		}
		company, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil || company.State.RunSeq != 2 || company.State.RunPreTimer {
			t.Fatalf("new company=%+v err=%v", company.State, err)
		}
		founder, err := store.LoadLatest(ctx, founderRevision.StreamID)
		if err != nil || len(founder.State.ExitHistory) != 1 || founder.State.ExitHistory[0].ExitType != "scripted_first" {
			t.Fatalf("founder=%+v err=%v", founder.State, err)
		}
		var preTimer bool
		if err := db.QueryRowContext(ctx, `SELECT (payload->>'pre_timer')::boolean FROM events WHERE stream_id=$1 AND kind='run_ended'`, companyRevision.StreamID).Scan(&preTimer); err != nil || !preTimer {
			t.Fatalf("pre-timer event=%v err=%v", preTimer, err)
		}
	})
}

func createPrestigeStreams(t *testing.T, ctx context.Context, store *save.Store, catalog *economy.Catalog, hash, owner string, now, started, evaluated time.Time, cash string, lifetime decimal.Decimal, tier int64) (save.Revision, save.Revision) {
	t.Helper()
	founderLedger, err := economy.NewLedger(catalog, economy.ScopeFounder)
	if err != nil {
		t.Fatal(err)
	}
	founder := &save.State{Ledger: founderLedger, GeneratorCounts: map[string]int64{}, EvaluatedThrough: evaluated, ManualTokenRefilledAt: evaluated,
		GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{},
		CompactSamples: []save.CompactSample{}, LifetimeValue: decimal.Zero, OfflineSpans: []save.OfflineSpan{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}}
	companyLedger, err := economy.RestoreLedger(catalog, economy.ScopeCompany, map[string]string{"company.cash": cash})
	if err != nil {
		t.Fatal(err)
	}
	company := &save.State{Ledger: companyLedger, GeneratorCounts: map[string]int64{"generator.beige_tower": 0}, EvaluatedThrough: evaluated,
		ManualTokenMilli: catalog.ManualPolicy().BucketCapMilli, ManualTokenRefilledAt: evaluated,
		GatesCrossed: map[string]bool{}, RunSeq: 1, DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{},
		CompactSamples: []save.CompactSample{}, Tier: tier, LifetimeValue: lifetime, RunStartedAt: started, OfflineSpans: []save.OfflineSpan{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}}
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: owner, Scope: economy.ScopeFounder}, hash, founder, save.WriteContext{Cause: "prestige.integration"})
	if err != nil {
		t.Fatal(err)
	}
	companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: owner, Scope: economy.ScopeCompany}, hash, company, save.WriteContext{Cause: "prestige.integration"})
	if err != nil {
		t.Fatal(err)
	}
	return founderRevision, companyRevision
}
