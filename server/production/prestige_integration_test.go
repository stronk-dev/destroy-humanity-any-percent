package production

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
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
	if _, err := db.ExecContext(ctx, `TRUNCATE save_streams CASCADE`); err != nil {
		t.Fatal(err)
	}
	economyBytes, _ := os.ReadFile("../../balance/catalogs/phase0.json")
	routeBytes, _ := os.ReadFile("../../balance/routes/phase0.json")
	prestigeBytes, _ := os.ReadFile("../../balance/prestige/phase0.json")
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
	hash, err := save.ConstantsHashArtifacts(map[string][]byte{"economy": economyBytes, "routes": routeBytes, "prestige": prestigeBytes})
	if err != nil {
		t.Fatal(err)
	}
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{hash: catalog}, routes: map[string]*routes.Catalog{hash: routeCatalog}, prestige: map[string]*prestigecore.Policy{hash: policy}}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService(store, resolver, nil, nil, nil, WithRouteCatalogs(resolver), WithRouteProjector(prestigeNoopProjector{}), WithPrestigeRuntime(resolver, 5_000))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 29, 16, 0, 0, 0, time.UTC)

	t.Run("elective wind down", func(t *testing.T) {
		owner := "01985555-1000-7000-8000-000000000001"
		founderRevision, companyRevision := createPrestigeStreams(t, ctx, store, catalog, hash, owner, now, now.Add(-10*time.Minute), now, "0", decimal.New(8, 12), 1)
		routePayload, _ := json.Marshal(map[string]any{"route_id": "route.nonprofit_wrapper_zip", "gate_id": "gate.t1_to_t2", "run_id": map[string]any{"company_stream_id": companyRevision.StreamID, "run_seq": 1}, "founder_id": owner})
		if _, err := db.ExecContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,1,1,'route_executed',$2,$3,$4,$5)`, companyRevision.StreamID, "01985555-1000-7000-8000-000000000099", hash, now.Add(-time.Minute), routePayload); err != nil {
			t.Fatal(err)
		}
		request := []byte(`{"intent_id":"01985555-1001-7000-8000-000000000001","kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`)
		result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, request)
		if err != nil || result.Replay {
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
		if err != nil || founder.State.ReputationLevel != 1 || founder.State.RouteKnowledgeBalance != 25 || len(founder.State.ExitHistory) != 1 || founder.State.ExitHistory[0].ExitType != "collapse" {
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
		}
		if err := json.Unmarshal(endedPayload, &ended); err != nil || len(ended.ExecutedRoutes) != 1 || ended.ExecutedRoutes[0] != "route.nonprofit_wrapper_zip" {
			t.Fatalf("run_ended=%s err=%v", endedPayload, err)
		}
		replay, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, request)
		if err != nil || !replay.Replay || string(replay.Receipt) != string(result.Receipt) {
			t.Fatalf("replay=%+v err=%v", replay, err)
		}
	})

	t.Run("scripted first threshold", func(t *testing.T) {
		owner := "01985555-2000-7000-8000-000000000002"
		founderRevision, companyRevision := createPrestigeStreams(t, ctx, store, catalog, hash, owner, now, now.Add(-16*time.Minute), now.Add(-time.Second), "1e10", decimal.New(1, 12), 2)
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
		cross := []byte(`{"intent_id":"01985555-3001-7000-8000-000000000003","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t7_to_t8","route_id":null}`)
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
		written, err := store.Write(ctx, companyRevision.StreamID, 2, hash, offered.State, save.WriteContext{Cause: "prestige.test.progress"})
		if err != nil || written.Number != 3 {
			t.Fatalf("write=%+v err=%v", written, err)
		}
		acceptBody, _ := json.Marshal(map[string]any{"intent_id": "01985555-3002-7000-8000-000000000003", "kind": "accept_exit_offer", "expected_revision": 3, "expected_founder_revision": 1, "offer_id": offerID})
		if result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, acceptBody); err != nil || result.Replay {
			t.Fatalf("accept=%+v err=%v", result, err)
		}
		founder, err := store.LoadLatest(ctx, founderRevision.StreamID)
		if err != nil || founder.State.ReputationLevel < stored.PayoutPreview.ReputationDelta {
			t.Fatalf("founder reputation=%d preview=%d err=%v", founder.State.ReputationLevel, stored.PayoutPreview.ReputationDelta, err)
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
