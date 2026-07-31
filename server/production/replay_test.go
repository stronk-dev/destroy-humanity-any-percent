package production

import (
	"encoding/json"
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

func loadReplayTestBundle(t *testing.T, hash string, artifacts map[string][]byte) CatalogBundle {
	t.Helper()
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
	return CatalogBundle{ConstantsHash: hash, Economy: economyCatalog, Routes: routeCatalog,
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
		resolved.Accrual.GuildSettlementBatch == nil || resolved.Accrual.CommonsWeightPPM == nil || *resolved.Accrual.CommonsWeightPPM != weight {
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
	first, err := ApplyLogged(newState(), payload, catalogs, inputs, &fakeInvariantSink{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := ApplyLogged(newState(), payload, catalogs, inputs, &fakeInvariantSink{})
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
	if _, err := ApplyLogged(newState(), payload, catalogs, regressed, nil); err == nil {
		t.Fatal("clock-regressed replay input was accepted")
	}
}

func TestTerminalReplayInputsFreezeFounderCarry(t *testing.T) {
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	command := save.ReplayCommand{IntentID: "01985555-7200-7000-8000-000000000001", CompanyStreamID: "01985555-7200-4000-8000-000000000002", FounderID: "01985555-7200-4000-8000-000000000003", Revision: 9, RunSeq: 3, RunLogSeq: 12}
	founder := &save.State{ReputationLevel: 17, RouteKnowledgeBalance: 4, AgeMS: 50, Notoriety: 6, AdvisorMode: true,
		LedgerFactKinds: map[string]bool{"fact.z": true, "fact.a": true}, NetworkSlots: []save.NetworkSlot{{Slot: "slot.z", CarriedRef: "ref.z"}}}
	carry := founderCarry(founder)
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
