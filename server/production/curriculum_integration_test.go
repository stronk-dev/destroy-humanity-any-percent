package production

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/commonsbinding"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/pet"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routeprojection"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

func TestCurriculumAutomaticFailurePersistsAndReplaysIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE epochs,catalog_sets,founder_route_state,save_streams RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	current := activeContentBundle(t)
	if current.Curriculum == nil {
		t.Fatal("minted curriculum artifact is unavailable")
	}
	next := current
	seedProductionEpoch(t, db, current.ConstantsHash, current.Artifacts)
	resolver := integrationCatalogs{
		economy:  map[string]*economy.Catalog{current.ConstantsHash: current.Economy, next.ConstantsHash: next.Economy},
		routes:   map[string]*routes.Catalog{current.ConstantsHash: current.Routes, next.ConstantsHash: next.Routes},
		prestige: map[string]*prestigecore.Policy{current.ConstantsHash: current.Prestige, next.ConstantsHash: next.Prestige},
		factions: map[string]*faction.Catalog{current.ConstantsHash: current.Faction, next.ConstantsHash: next.Faction},
	}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 15, 18, 0, 0, 0, time.UTC)
	const founderID = "01986666-ad00-7000-8000-000000000001"
	founder := replayFounderFixtureState(t, current, now)
	founder.WireVersion = 21
	if err := activateMinigameState(founder, current.Minigames); err != nil {
		t.Fatal(err)
	}
	founder.Pets = map[string]pet.CareState{}
	founder.FiscalPeriodOpenedWallMS, founder.FiscalGeneratorLevels, founder.FiscalUnlocks = now.UnixMilli(), map[string]int64{}, map[string]bool{}
	for _, row := range current.Fiscal.GeneratorLevelRows() {
		founder.FiscalGeneratorLevels[row.GeneratorID] = 0
	}
	founder.Soul, founder.SoulExhaustedSourceIDs = 80, []string{}
	company := replayFixtureState(t, current.Economy, now.Add(-15*time.Minute))
	company.WireVersion, company.Tier = 18, 1
	company.GatesCrossed["gate.t0_to_t1"] = true
	company.MeterBands = nil
	meterState, meterErr := meters.NewRunState(current.Meters, 0)
	if meterErr != nil {
		t.Fatal(meterErr)
	}
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	company.AchievementsEarnedRun = map[string]bool{}
	if _, err := initializeActivePlayState(company, current.Opportunities, founderID); err != nil {
		t.Fatal(err)
	}
	advanceActivePlayFixtureAttendance(t, company, current.Opportunities, current.Prestige, founderID, now)
	company.ManualTokenRefilledAt = now
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder}, current.ConstantsHash, founder, save.WriteContext{Cause: "curriculum.integration"})
	if err != nil {
		t.Fatal(err)
	}
	companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeCompany}, current.ConstantsHash, company, save.WriteContext{Cause: "curriculum.integration"})
	if err != nil {
		t.Fatal(err)
	}
	pinTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	genesis, _ := save.EncodeState(company)
	_, err = save.PinRunWithGenesisTx(ctx, pinTx, companyRevision.StreamID, founderID, 1, current.ConstantsHash, companyRevision.Version, genesis)
	if err == nil {
		var frozen []save.FrozenContribution
		frozen, err = FrozenFiscalContributions(current.Fiscal, founder)
		if err == nil {
			err = save.InsertRunFrozenContributionsTx(ctx, pinTx, companyRevision.StreamID, 1, frozen)
		}
	}
	if err == nil {
		err = pinTx.Commit()
	} else {
		_ = pinTx.Rollback()
	}
	if err != nil {
		t.Fatal(err)
	}
	minigameRepository, err := minigame.NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	commonsCatalogs := commons.CatalogSet{current.ConstantsHash: current.Commons.(commonsbinding.ReplayPolicy).Catalog, next.ConstantsHash: next.Commons.(commonsbinding.ReplayPolicy).Catalog}
	routeProjector, err := routeprojection.New(db, resolver)
	if err != nil {
		t.Fatal(err)
	}
	hintProjection := &heldHintProjection{delegate: routeProjector, entered: make(chan struct{}), release: make(chan struct{})}
	unblockProjection := sync.OnceFunc(func() { close(hintProjection.release) })
	defer unblockProjection()
	service, err := NewService(store, resolver, FrozenContributionProvider{DB: db}, nil, nil,
		WithRouteCatalogs(resolver), WithRouteProjector(hintProjection),
		WithProgressionRuntime(resolver), WithCurrentConstantsHash(next.ConstantsHash),
		WithReplayCatalogs(ReplayCatalogSet{current.ConstantsHash: current, next.ConstantsHash: next}), WithGuildSettlements(emptyGuildSettlements{}),
		WithMinigameActivity(minigameRepository), WithCompactPolicies(commonsCatalogs), WithCommonsWeightResolver(integrationWeight(1_000_000)))
	if err != nil {
		t.Fatal(err)
	}
	request := []byte(`{"intent_id":"01986666-ad00-7000-8000-000000000002","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`)
	result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, request)
	if err != nil || !strings.Contains(string(result.Receipt), `"outcome":"applied"`) {
		t.Fatalf("receipt=%s err=%v", result.Receipt, err)
	}
	retry, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, request)
	if err != nil || !retry.Replay || string(retry.Receipt) != string(result.Receipt) {
		t.Fatalf("retry=%+v err=%v", retry, err)
	}
	loadedCompany, err := store.LoadLatest(ctx, companyRevision.StreamID)
	if err != nil || loadedCompany.State.RunSeq != 2 || loadedCompany.Revision.ConstantsHash != next.ConstantsHash || loadedCompany.State.GeneratorProvisioned["generator.beige_tower"] != 10 {
		t.Fatalf("company=%+v revision=%+v err=%v", loadedCompany.State, loadedCompany.Revision, err)
	}
	loadedFounder, err := store.LoadLatest(ctx, founderRevision.StreamID)
	if err != nil || loadedFounder.State.RouteKnowledgeBalance != 75 || len(loadedFounder.State.ExitHistory) != 1 {
		t.Fatalf("founder=%+v err=%v", loadedFounder.State, err)
	}
	var schema int
	var payload, replayInputs []byte
	if err := db.QueryRowContext(ctx, `SELECT schema_version,payload FROM events WHERE stream_id=$1 AND kind='run_ended'`, companyRevision.StreamID).Scan(&schema, &payload); err != nil {
		t.Fatal(err)
	}
	var ended struct {
		Branch string `json:"branch"`
	}
	if err := json.Unmarshal(payload, &ended); err != nil || schema != 3 || ended.Branch != "burnout" {
		t.Fatalf("run_ended schema=%d payload=%s err=%v", schema, payload, err)
	}
	if err := db.QueryRowContext(ctx, `SELECT replay_inputs FROM run_log WHERE company_stream_id=$1 AND run_seq=1`, companyRevision.StreamID).Scan(&replayInputs); err != nil {
		t.Fatal(err)
	}
	wire, err := parseReplayInputs(replayInputs)
	var terminal replayExitResolved
	if err == nil {
		err = decodeReplayStrict(wire.Resolved, &terminal)
	}
	if err != nil || terminal.SelectedBranch == nil || *terminal.SelectedBranch != "burnout" || terminal.IntentKind != IntentPerformManualBatch {
		t.Fatalf("terminal=%+v err=%v", terminal, err)
	}

	var routeIDs []string
	for _, gate := range current.Routes.Gates() {
		for _, route := range gate.Routes {
			routeIDs = append(routeIDs, route.RouteID)
		}
	}
	cost := current.Routes.KnowledgePolicy().HintCost
	if len(routeIDs) < 2 || loadedFounder.State.RouteKnowledgeBalance < cost || loadedFounder.State.RouteKnowledgeBalance >= 2*cost {
		t.Fatal("earned-credit fixture must afford exactly one of two distinct hints")
	}
	requestHint := func(id, route string, revision int64) []byte {
		return []byte(fmt.Sprintf(`{"intent_id":%q,"kind":"buy_route_hint","expected_revision":%d,"route_id":%q}`, id, revision, route))
	}
	purchase := requestHint("01986666-ad00-7000-8000-000000000003", routeIDs[0], loadedFounder.Revision.Number)
	wait, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	type completion struct {
		result HandleResult
		err    error
	}
	finished := make(chan completion, 1)
	go func() {
		result, err := service.Handle(wait, founderRevision.StreamID, ModeOnline, now, purchase)
		finished <- completion{result, err}
	}()
	select {
	case <-hintProjection.entered:
	case early := <-finished:
		t.Fatalf("affordable hint failed before projection: receipt=%s err=%v", early.result.Receipt, early.err)
	case <-wait.Done():
		t.Fatal("hint did not reach the post-commit projection barrier")
	}
	// Its authoritative debit has committed, while delivery remains held. A
	// second purchase must include that debit when repairing the projection.
	paid, err := store.LoadLatest(wait, founderRevision.StreamID)
	if err != nil || paid.State.RouteKnowledgeBalance != 75-cost || !paid.State.HintsUnlocked[routeIDs[0]] {
		t.Fatalf("earned hint did not persist: state=%+v err=%v", paid.State, err)
	}
	second := requestHint("01986666-ad00-7000-8000-000000000004", routeIDs[1], paid.Revision.Number)
	refused, err := service.Handle(wait, founderRevision.StreamID, ModeOnline, now, second)
	if err != nil || !strings.Contains(string(refused.Receipt), `"category":"insufficient_route_knowledge"`) {
		t.Fatalf("delayed debit allowed another hint: receipt=%s err=%v", refused.Receipt, err)
	}
	unblockProjection()
	var first completion
	select {
	case first = <-finished:
	case <-wait.Done():
		t.Fatal("released hint projection did not finish")
	}
	if first.err != nil || !strings.Contains(string(first.result.Receipt), `"outcome":"applied"`) {
		t.Fatalf("hint completion=%s err=%v", first.result.Receipt, first.err)
	}
	repeated, err := service.Handle(wait, founderRevision.StreamID, ModeOnline, now, purchase)
	if err != nil || !repeated.Replay || string(repeated.Receipt) != string(first.result.Receipt) {
		t.Fatal("hint retry changed its applied receipt")
	}
	after, err := store.LoadLatest(wait, founderRevision.StreamID)
	if err != nil || after.Revision.Number != paid.Revision.Number || after.State.RouteKnowledgeBalance != 75-cost || after.State.HintsUnlocked[routeIDs[1]] {
		t.Fatalf("rejected/retried hint changed state: state=%+v err=%v", after.State, err)
	}
	var projected, exitClaims, hintClaims int64
	if err := db.QueryRowContext(wait, `SELECT route_knowledge_balance,
		(SELECT count(*) FROM exit_knowledge_projection_events WHERE founder_id=$1),
		(SELECT count(*) FROM route_hint_projection_events WHERE founder_id=$1)
		FROM founder_route_state WHERE founder_id=$1`, founderID).Scan(&projected, &exitClaims, &hintClaims); err != nil || projected != 75-cost || exitClaims != 1 || hintClaims != 1 {
		t.Fatalf("projection=%d exit_claims=%d hint_claims=%d err=%v", projected, exitClaims, hintClaims, err)
	}
}

type heldHintProjection struct {
	delegate RouteProjector
	entered  chan struct{}
	release  chan struct{}
	held     atomic.Bool
}

func (p *heldHintProjection) Project(ctx context.Context, events []save.EventRecord) error {
	for _, event := range events {
		if event.Kind == save.EventRouteHintPurchased && p.held.CompareAndSwap(false, true) {
			close(p.entered)
			select {
			case <-p.release:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return p.delegate.Project(ctx, events)
}
func (p *heldHintProjection) RepairFounder(ctx context.Context, id string, state *save.State) error {
	return p.delegate.RepairFounder(ctx, id, state)
}
