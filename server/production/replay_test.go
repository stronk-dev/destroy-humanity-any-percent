package production

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/commonsbinding"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/multiplier"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

type emptyGuildSettlements struct{}

func (emptyGuildSettlements) PendingSettlements(context.Context, string, string, int64) (guild.SettlementBatch, error) {
	return guild.SettlementBatch{}, nil
}

type fixedGuildSettlements struct{ batch guild.SettlementBatch }

func (resolver fixedGuildSettlements) PendingSettlements(context.Context, string, string, int64) (guild.SettlementBatch, error) {
	return resolver.batch, nil
}

type settlementObservingContributions struct{ called bool }

func (provider *settlementObservingContributions) Contributions(_ context.Context, state *save.State, _ *economy.Catalog, _ save.Revision) ([]multiplier.Contribution, error) {
	provider.called = true
	if state.StockUnits != 7 || state.ConsumedStockUnits != 7 || state.GuildConsumedWindow != 7 || state.GuildBoundarySeq != 1 {
		return nil, fmt.Errorf("contribution saw pre-settlement state stock=%d consumed=%d window=%d seq=%d", state.StockUnits, state.ConsumedStockUnits, state.GuildConsumedWindow, state.GuildBoundarySeq)
	}
	return []multiplier.Contribution{}, nil
}

func TestLiveReplayAccrualFreezesGuildSettlementBeforeContributions(t *testing.T) {
	guildID := "01985555-7100-7000-8000-000000000001"
	provider := &settlementObservingContributions{}
	service := &Service{contributions: provider, guildSettlements: fixedGuildSettlements{batch: guild.SettlementBatch{GuildID: guildID, BaseSeq: 0,
		Settlements: []guild.Settlement{{GuildID: guildID, BoundarySeq: 1, DebitUnits: 3, CreditUnits: 7}}}}}
	state := &save.State{StockUnits: 10}
	request := IntentRequest{Kind: IntentPerformManualBatch}
	contributions, batch, err := service.resolveReplayAccrual(context.Background(), state, save.Revision{OwnerID: "01985555-7100-4000-8000-000000000003"}, &economy.Catalog{}, 100, request)
	if err != nil || !provider.called || len(contributions) != 0 || batch.GuildID != guildID || len(batch.Settlements) != 1 {
		t.Fatalf("contributions=%v batch=%+v called=%t err=%v", contributions, batch, provider.called, err)
	}
	if state.StockUnits != 10 || state.ConsumedStockUnits != 0 || state.GuildConsumedWindow != 0 || state.GuildBoundaryGuildID != "" || state.GuildBoundarySeq != 0 {
		t.Fatalf("live prebuild leaked settlement mutation: %+v", state)
	}
}

func loadReplayTestBundle(t *testing.T, hash string, artifacts map[string][]byte) CatalogBundle {
	t.Helper()
	computed, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil || computed != hash {
		t.Fatalf("replay artifact identity computed=%s labeled=%s err=%v", computed, hash, err)
	}
	economyCatalog, err := economy.LoadCatalog(artifacts["economy"])
	if err != nil {
		t.Fatal(err)
	}
	routeCatalog, err := routes.LoadCatalog(artifacts["routes"])
	if err != nil {
		t.Fatal(err)
	}
	commonsCatalog, err := commons.LoadCatalog(artifacts["commons"])
	if err != nil {
		t.Fatal(err)
	}
	policy, err := prestigecore.LoadPolicy(artifacts["prestige"])
	if err != nil {
		t.Fatal(err)
	}
	factionCatalog, err := faction.LoadCatalog(artifacts["factions"], faction.CompactTitheBand{MinimumPPM: commonsCatalog.MinimumTithePPM, DefaultPPM: commonsCatalog.DefaultTithePPM, MaximumPPM: commonsCatalog.MaximumTithePPM})
	if err != nil {
		t.Fatal(err)
	}
	guildCatalog, err := guild.LoadCatalog(artifacts["guilds"])
	if err != nil {
		t.Fatal(err)
	}
	frozen := make(map[string][]byte, len(artifacts))
	for name, data := range artifacts {
		frozen[name] = append([]byte(nil), data...)
	}
	return CatalogBundle{ConstantsHash: hash, Artifacts: frozen, Economy: economyCatalog, Routes: routeCatalog,
		Commons: commonsbinding.ReplayPolicy{Catalog: commonsCatalog}, Prestige: policy, Faction: factionCatalog, Guild: guildCatalog}
}

func TestReplayInputsAreClosedCanonicalInputs(t *testing.T) {
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	command := save.ReplayCommand{IntentID: "01985555-7100-7000-8000-000000000001", CompanyStreamID: "01985555-7100-4000-8000-000000000002", FounderID: "01985555-7100-4000-8000-000000000003", Revision: 7, RunSeq: 2, RunLogSeq: 4}
	weight := int64(812_345)
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOffline, Now: now, IntentKind: IntentBuyGenerator,
		CommonsWeightPPM: &weight, RouteContextVersion: 3, Contributions: []multiplier.Contribution{
			{Slot: multiplier.SlotPrestige, SourceID: "source.z", Target: "all", Factor: decimal.New(13, -1)},
			{Slot: multiplier.SlotFaction, SourceID: "source.a", Target: "generator.one", Factor: decimal.One},
		}})
	if err != nil {
		t.Fatal(err)
	}
	wire, err := parseReplayInputs(inputs)
	if err != nil || wire.Command != command || wire.EvaluatedAtMS != now.UnixMilli() || wire.EvaluationMode != ModeOffline {
		t.Fatalf("wire=%+v err=%v", wire, err)
	}
	var resolved replayAccrualResolved
	if err := decodeReplayStrict(wire.Resolved, &resolved); err != nil || len(resolved.Accrual.Contributions) != 2 ||
		resolved.Accrual.Contributions[0].SourceID != "source.a" || resolved.Accrual.Contributions[1].SourceID != "source.z" ||
		resolved.Accrual.GuildSettlementBatch.Settlements == nil || resolved.Accrual.CommonsWeightPPM == nil || *resolved.Accrual.CommonsWeightPPM != weight {
		t.Fatalf("resolved=%+v err=%v", resolved, err)
	}

	var object map[string]any
	if err := json.Unmarshal(inputs, &object); err != nil {
		t.Fatal(err)
	}
	object["future_field"] = true
	tampered, _ := json.Marshal(object)
	if _, err := parseReplayInputs(tampered); err == nil {
		t.Fatal("unknown replay-input field accepted")
	}
	if strings.Contains(string(inputs), `"factor":1.3`) {
		t.Fatal("factor was encoded as a binary JSON number instead of canonical Decimal string")
	}
}

func TestReplaySettlementBatchIsRepresentableAndOrdered(t *testing.T) {
	valid := replayAccrual{Contributions: []replayContribution{}, CommonsWeightPPM: nil, RouteContextVersion: 1,
		GuildSettlementBatch: replayGuildSettlementBatch{GuildID: "01985555-7100-7000-8000-000000000001", BaseSeq: 6, Settlements: []replayGuildSettlement{
			{BoundarySeq: 7, DebitUnits: 0, CreditUnits: 0},
			{BoundarySeq: 8, DebitUnits: 21, CreditUnits: 42},
		}}}
	if _, err := contributionsFromReplay(valid); err != nil {
		t.Fatalf("well-formed non-empty settlement batch rejected: %v", err)
	}
	unsorted := valid
	unsorted.GuildSettlementBatch.Settlements = append([]replayGuildSettlement(nil), valid.GuildSettlementBatch.Settlements...)
	unsorted.GuildSettlementBatch.Settlements[0], unsorted.GuildSettlementBatch.Settlements[1] = unsorted.GuildSettlementBatch.Settlements[1], unsorted.GuildSettlementBatch.Settlements[0]
	if _, err := contributionsFromReplay(unsorted); !errors.Is(err, ErrInvalidReplayInputs) {
		t.Fatalf("out-of-order settlement batch error=%v", err)
	}
	invalid := valid
	invalid.GuildSettlementBatch = replayGuildSettlementBatch{GuildID: "not-a-guild", Settlements: []replayGuildSettlement{{BoundarySeq: 1, DebitUnits: -1}}}
	if _, err := contributionsFromReplay(invalid); !errors.Is(err, ErrInvalidReplayInputs) {
		t.Fatalf("invalid settlement batch error=%v", err)
	}
}

func TestFounderScopeRouteHintHasNoReplayUnionArm(t *testing.T) {
	command := save.ReplayCommand{IntentID: "01985555-7110-7000-8000-000000000001", CompanyStreamID: "01985555-7110-4000-8000-000000000002", FounderID: "01985555-7110-4000-8000-000000000003", Revision: 1, RunSeq: 1, RunLogSeq: 1}
	if _, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: time.Now(), IntentKind: IntentBuyRouteHint}); !errors.Is(err, ErrInvalidReplayInputs) {
		t.Fatalf("founder-scope route hint replay error=%v", err)
	}
}

func TestApplyLoggedDerivesFactionStockResourceInsideBoundary(t *testing.T) {
	bundleBytes, err := epochseed.Load("../..")
	if err != nil {
		t.Fatal(err)
	}
	catalogs := loadReplayTestBundle(t, bundleBytes.Hash, bundleBytes.Artifacts)
	member, ok := catalogs.Faction.Faction("open_source")
	if !ok {
		t.Fatal("open_source fixture missing")
	}
	request, err := ParseIntent([]byte(`{"intent_id":"01985555-7120-7000-8000-000000000001","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":0,"window_ms":10}`))
	if err != nil || request.InvalidDetail == "" {
		t.Fatalf("invalid fixture request=%+v err=%v", request, err)
	}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01985555-7120-4000-8000-000000000002", FounderID: "01985555-7120-4000-8000-000000000003", Revision: 1, RunSeq: 1, RunLogSeq: 1}
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind,
		Contributions: []multiplier.Contribution{}, RouteContextVersion: catalogs.Routes.ContextVersion()})
	if err != nil {
		t.Fatal(err)
	}
	state := replayFixtureState(t, catalogs.Economy, now)
	state.FactionID = "open_source"
	state.IncorporatedAt = now.Add(-time.Minute)
	result, err := ApplyLogged(state, request.CanonicalPayload, catalogs, inputs)
	if err != nil || result.Outcome != save.IntentRejected || state.FactionStockResource != member.Produces {
		t.Fatalf("outcome=%s stock=%q want=%q err=%v", result.Outcome, state.FactionStockResource, member.Produces, err)
	}
}

func TestApplyLoggedRejectsMismatchedFounderCatalogCarry(t *testing.T) {
	bundleBytes, err := epochseed.Load("../..")
	if err != nil {
		t.Fatal(err)
	}
	catalogs := loadReplayTestBundle(t, bundleBytes.Hash, bundleBytes.Artifacts)
	request, err := ParseIntent([]byte(`{"intent_id":"01985555-7130-7000-8000-000000000001","kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`))
	if err != nil {
		t.Fatal(err)
	}
	carry := founderCarry(&save.State{LedgerFactKinds: map[string]bool{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}})
	carry.FounderRevision = 1
	carry.FounderConstantsHash = "sha256:" + strings.Repeat("a", 64)
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01985555-7130-4000-8000-000000000002", FounderID: "01985555-7130-4000-8000-000000000003", Revision: 1, RunSeq: 1, RunLogSeq: 1}
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind,
		Contributions: []multiplier.Contribution{}, RouteContextVersion: catalogs.Routes.ContextVersion(), FounderCarry: &carry,
		Terminal: true, ExecutedRouteIDs: []string{}, SelectedExitType: "collapse", SelectedTerms: json.RawMessage(`{}`), NextConstantsHash: catalogs.ConstantsHash})
	if err != nil {
		t.Fatal(err)
	}
	company := &save.State{RunSeq: 1, EvaluatedThrough: now}
	if _, err := ApplyLoggedExit(company, request.CanonicalPayload, catalogs, inputs); !errors.Is(err, ErrInvalidReplayInputs) {
		t.Fatalf("mixed-catalog founder carry error=%v", err)
	}
}

func TestApplyLoggedReplaysByteIdenticalTransition(t *testing.T) {
	bundleBytes, err := epochseed.Load("../..")
	if err != nil {
		t.Fatal(err)
	}
	catalogs := loadReplayTestBundle(t, bundleBytes.Hash, bundleBytes.Artifacts)
	cursor := time.Date(2026, 7, 31, 8, 0, 0, 0, time.UTC)
	newState := func() *save.State {
		ledger, ledgerErr := economy.RestoreLedger(catalogs.Economy, economy.ScopeCompany, map[string]string{"company.cash": "1e2"})
		if ledgerErr != nil {
			t.Fatal(ledgerErr)
		}
		return &save.State{Ledger: ledger, GeneratorCounts: map[string]int64{"generator.beige_tower": 0},
			EvaluatedThrough: cursor, ManualTokenMilli: 50_000, ManualTokenRefilledAt: cursor,
			GatesCrossed: map[string]bool{}, RunSeq: 1, DoctrinesByTransition: map[string]string{},
			LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{}, RegionTraits: map[string]bool{},
			HintsUnlocked: map[string]bool{}, CompactSamples: []save.CompactSample{}, LifetimeValue: decimal.Zero,
			RunStartedAt: cursor, OfflineSpans: []save.OfflineSpan{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}}
	}
	request, err := ParseIntent([]byte(`{"intent_id":"01985555-7300-7000-8000-000000000001","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":3,"window_ms":10}`))
	if err != nil {
		t.Fatal(err)
	}
	payload := request.CanonicalPayload
	command := save.ReplayCommand{IntentID: "01985555-7300-7000-8000-000000000001", CompanyStreamID: "01985555-7300-4000-8000-000000000002", FounderID: "01985555-7300-4000-8000-000000000003", Revision: 1, RunSeq: 1, RunLogSeq: 1}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: cursor.Add(time.Second),
		IntentKind: IntentPerformManualBatch, Contributions: []multiplier.Contribution{}, RouteContextVersion: catalogs.Routes.ContextVersion()})
	if err != nil {
		t.Fatal(err)
	}
	first, err := ApplyLogged(newState(), payload, catalogs, inputs)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ApplyLogged(newState(), payload, catalogs, inputs)
	if err != nil {
		t.Fatal(err)
	}
	firstState, _ := json.Marshal(wireSnapshot(first.State))
	secondState, _ := json.Marshal(wireSnapshot(second.State))
	firstEvents, _ := json.Marshal(first.Events)
	secondEvents, _ := json.Marshal(second.Events)
	if first.Outcome != save.IntentApplied || string(first.Receipt) != string(second.Receipt) || string(firstState) != string(secondState) || string(firstEvents) != string(secondEvents) {
		t.Fatalf("first=%s state=%s events=%s second=%s state=%s events=%s", first.Receipt, firstState, firstEvents, second.Receipt, secondState, secondEvents)
	}

	var root map[string]any
	if err := json.Unmarshal(inputs, &root); err != nil {
		t.Fatal(err)
	}
	root["evaluated_at_ms"] = cursor.Add(-time.Second).UnixMilli()
	regressed, _ := json.Marshal(root)
	if _, err := ApplyLogged(newState(), payload, catalogs, regressed); err == nil {
		t.Fatal("clock-regressed replay input was accepted")
	}
}

func TestTerminalReplayInputsFreezeFounderCarry(t *testing.T) {
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	command := save.ReplayCommand{IntentID: "01985555-7200-7000-8000-000000000001", CompanyStreamID: "01985555-7200-4000-8000-000000000002", FounderID: "01985555-7200-4000-8000-000000000003", Revision: 9, RunSeq: 3, RunLogSeq: 12}
	founder := &save.State{ReputationLevel: 17, RouteKnowledgeBalance: 4, AgeMS: 50, Notoriety: 6, AdvisorMode: true,
		LedgerFactKinds: map[string]bool{"fact.z": true, "fact.a": true}, NetworkSlots: []save.NetworkSlot{{Slot: "slot.z", CarriedRef: "ref.z"}}}
	carry := founderCarry(founder)
	carry.FounderRevision = 1
	carry.FounderConstantsHash = "sha256:" + strings.Repeat("a", 64)
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: IntentWindDown,
		FounderCarry: &carry, Terminal: true, ExecutedRouteIDs: []string{"route.z", "route.a"}, SelectedExitType: "collapse",
		SelectedTerms: json.RawMessage(`{}`), NextConstantsHash: "sha256:" + strings.Repeat("a", 64)})
	if err != nil {
		t.Fatal(err)
	}
	founder.ReputationLevel = 99
	wire, _ := parseReplayInputs(inputs)
	var terminal replayExitResolved
	if err := decodeReplayStrict(wire.Resolved, &terminal); err != nil || terminal.FounderCarry.ReputationLevel != 17 ||
		terminal.ExecutedRouteIDs[0] != "route.a" || terminal.ExecutedRouteIDs[1] != "route.z" {
		t.Fatalf("terminal=%+v err=%v", terminal, err)
	}
}
