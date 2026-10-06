package production

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/garden"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/save"
)

// Literal permanent state: immature, mature and dormant plants, collected
// nonstarters, nondefault substrate and all clock/lockout metadata populated.
const retainedGardenJSON = `{"salt_hex":"0123456789abcdef","tick_anchor_wall_ms":1000000,"tick_seq":7,"substrate_id":"containerized","substrate_set_wall_ms":999999,"plots":[{"row":0,"col":0,"species_id":"strain_a","age_ticks":1,"matured_effect_ppm":null},{"row":1,"col":1,"species_id":"strain_b","age_ticks":4,"matured_effect_ppm":1234567},{"row":5,"col":5,"species_id":"strain_c","age_ticks":6,"matured_effect_ppm":7654321}],"seed_collection":["strain_a","strain_b","strain_c"]}`
const emptyGardenJSON = `{"salt_hex":null,"tick_anchor_wall_ms":null,"tick_seq":0,"substrate_id":"bare_metal","substrate_set_wall_ms":null,"plots":[],"seed_collection":["strain_a","strain_b"]}`

func retainedGarden(t *testing.T) *garden.State {
	t.Helper()
	state, err := garden.DecodeState([]byte(retainedGardenJSON))
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func assertGardenBytes(t *testing.T, state *garden.State, want string) {
	t.Helper()
	got, err := garden.EncodeState(state)
	if err != nil || string(got) != want {
		t.Fatalf("permanent Garden bytes changed: got=%s want=%s err=%v", got, want, err)
	}
}

func TestGardenActualNewFounderInitializer(t *testing.T) {
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	for _, pinned := range []bool{false, true} {
		t.Run(fmt.Sprint(pinned), func(t *testing.T) {
			bundle := cosmeticsContentBundle(t)
			wantFloor := 24
			if pinned {
				bundle, wantFloor = gardenContentBundle(t), 25
			}
			initializer := FounderInitializer{Catalogs: ReplayCatalogSet{bundle.ConstantsHash: bundle}}
			founder := foundationScopeState(t, bundle.Economy, economy.ScopeFounder)
			company := replayFixtureState(t, bundle.Economy, now)
			company.WireVersion = save.CurrentVersion
			company.RunStartedAt = now
			if _, err := initializer.InitializeNewFounder(bundle.ConstantsHash, "01986666-8d00-7000-8000-000000000001", now, founder, company); err != nil {
				t.Fatal(err)
			}
			_, companyFloor := bundle.versionFloors()
			if save.VersionForState(founder) != wantFloor || save.VersionForState(company) != companyFloor {
				t.Fatal("initializer changed the wrong scalar floor")
			}
			if pinned {
				assertGardenBytes(t, founder.ServerGarden, emptyGardenJSON)
			} else if founder.ServerGarden != nil {
				t.Fatal("Garden activated without its pin")
			}
			preexisting := foundationScopeState(t, bundle.Economy, economy.ScopeFounder)
			preexisting.ServerGarden = retainedGarden(t)
			freshCompany := replayFixtureState(t, bundle.Economy, now)
			freshCompany.WireVersion = save.CurrentVersion
			if _, err := initializer.InitializeNewFounder(bundle.ConstantsHash, "01986666-8d00-7000-8000-000000000001", now, preexisting, freshCompany); err == nil {
				t.Fatal("initializer accepted pre-existing Garden")
			}
		})
	}
}

// Unlike the helper-only activation proof, these are real locked Store exits.
func TestGardenRetainedExitIntegration(t *testing.T) {
	for _, kind := range []string{IntentWindDown, IntentAcceptExitOffer} {
		t.Run(kind, func(t *testing.T) {
			now := save.CanonicalServerTime(time.Now())
			fixture := newGardenHarvestFixtureWithGenesis(t, func(founder *save.State) {
				founder.ReputationLevel = 4
				founder.ExitHistory = []save.ExitRecord{{RunID: 1, ExitType: "collapse", OccurredAt: now.Add(-time.Hour)}}
				founder.FiscalGeneratorLevels = map[string]int64{gardenContentBundle(t).Garden.HostGeneratorID: 0}
				founder.ServerGarden = retainedGarden(t)
			}, func(company *save.State) {
				company.WireVersion, company.Tier, company.RunSeq = 18, 1, 2
				company.GatesCrossed["gate.t0_to_t1"] = true
				if _, err := initializeActivePlayState(company, gardenContentBundle(t).Opportunities, "01986666-7f10-7000-8000-000000000002"); err != nil {
					t.Fatal(err)
				}
				if kind == IntentAcceptExitOffer {
					seedCosmeticExitOffer(company, now)
				}
			})
			activity, err := minigame.NewRepository(fixture.db)
			if err != nil || WithMinigameActivity(activity)(fixture.service) != nil {
				t.Fatalf("activity fixture: %v", err)
			}
			before, err := fixture.store.LoadLatest(fixture.ctx, fixture.founderStreamID)
			if err != nil {
				t.Fatal(err)
			}
			assertGardenBytes(t, before.State.ServerGarden, retainedGardenJSON)
			body := fmt.Sprintf(`{"intent_id":"01986666-7f10-7000-8000-000000003001","kind":%q,"expected_revision":1,"expected_founder_revision":1`, kind)
			if kind == IntentAcceptExitOffer {
				body += `,"offer_id":"01986666-9e00-7000-8000-000000000099"`
			}
			request := []byte(body + "}")
			result, err := fixture.service.Handle(fixture.ctx, fixture.companyStreamID, ModeOnline, time.Now(), request)
			if err != nil || result.Replay || !strings.Contains(string(result.Receipt), `"outcome":"applied"`) {
				t.Fatalf("Exit did not apply: %s err=%v", result.Receipt, err)
			}
			after, err := fixture.store.LoadLatest(fixture.ctx, fixture.founderStreamID)
			if err != nil || after.Revision.Number != 2 {
				t.Fatalf("Founder revision: %+v err=%v", after.Revision, err)
			}
			assertGardenBytes(t, after.State.ServerGarden, retainedGardenJSON)
			next, err := fixture.store.LoadLatest(fixture.ctx, fixture.companyStreamID)
			if err != nil || next.State.RunSeq != 3 || next.Revision.Number != 3 {
				t.Fatalf("next run: %+v err=%v", next.Revision, err)
			}
			retry, err := fixture.service.Handle(fixture.ctx, fixture.companyStreamID, ModeOnline, time.Now(), request)
			if err != nil || !retry.Replay || !bytes.Equal(retry.Receipt, result.Receipt) {
				t.Fatalf("Exit retry: %s err=%v", retry.Receipt, err)
			}
			final, err := fixture.store.LoadLatest(fixture.ctx, fixture.founderStreamID)
			if err != nil || final.Revision.Number != after.Revision.Number || !bytes.Equal(mustEncodeState(t, after.State), mustEncodeState(t, final.State)) {
				t.Fatal("retry changed permanent state")
			}
			history, err := fixture.store.LoadFounderHistory(fixture.ctx, fixture.founderStreamID)
			if err != nil || len(history.Entries) != 1 || VerifyFounderHistory(history, ReplayCatalogSet{fixture.bundle.ConstantsHash: fixture.bundle}) != ReplayVerified {
				t.Fatalf("Founder history invalid: %v", err)
			}
			poisoned := history
			poisoned.Entries = append([]save.FounderHistoryEntry{}, history.Entries...)
			poisoned.Entries[0].Events = []save.EventWrite{}
			removed := 0
			for _, event := range history.Entries[0].Events {
				if event.Kind == save.EventFiscalPeriodHarvested {
					removed++
					continue
				}
				poisoned.Entries[0].Events = append(poisoned.Entries[0].Events, event)
			}
			if removed != 1 || after.State.FiscalCredit <= before.State.FiscalCredit || VerifyFounderHistory(poisoned, ReplayCatalogSet{fixture.bundle.ConstantsHash: fixture.bundle}) != ReplayStateDivergence {
				t.Fatal("due Fiscal prefix was missing or its owned-history removal survived")
			}
			genesis, version, entries := persistedPrestigeReplay(t, fixture.db, fixture.companyStreamID, fixture.founderStreamID, 2, &fixture.bundle)
			if len(entries) != 1 || !entries[0].Terminal || VerifyReplayRun(genesis, version, fixture.bundle, entries, fixture.bundle.ConstantsHash, false) != ReplayStateDivergence {
				t.Fatal("unscoped all-stream aggregate did not discriminate the due Founder prefix")
			}
			// RP-207 ownership: retain every Company event and every base Founder
			// Exit event. Only the separately verified Founder Fiscal prefix is excluded.
			rows, err := fixture.db.QueryContext(fixture.ctx, `SELECT kind,schema_version,intent_id,payload FROM events
WHERE intent_id=$3 AND (stream_id=$1 OR (stream_id=$2 AND kind<>'fiscal_period_harvested.v1'))
ORDER BY CASE WHEN stream_id=$1 THEN 1 ELSE 0 END,event_seq,event_id`, fixture.companyStreamID, fixture.founderStreamID, "01986666-7f10-7000-8000-000000003001")
			if err != nil {
				t.Fatal(err)
			}
			events := []save.EventWrite{}
			for rows.Next() {
				var event save.EventWrite
				if err := rows.Scan(&event.Kind, &event.SchemaVersion, &event.IntentID, &event.Payload); err != nil {
					rows.Close()
					t.Fatal(err)
				}
				events = append(events, event)
			}
			err = rows.Err()
			rows.Close()
			if err != nil {
				t.Fatal(err)
			}
			entries[0].EventsJSON = marshalReplayEvents(events)
			if verdict := VerifyReplayRun(genesis, version, fixture.bundle, entries, fixture.bundle.ConstantsHash, false); verdict != ReplayVerified {
				t.Fatalf("scoped Company terminal history: %s", verdict)
			}
		})
	}
}

// A loadable starter retune is an observed contract conflict, not permission
// to invent grants or restrict epochs. The ordinary numeric retune controls it.
func TestGardenStarterRetuneDiagnostic(t *testing.T) {
	current := gardenContentBundle(t)
	for _, promoteStarter := range []bool{false, true} {
		data := map[string]any{}
		if err := json.Unmarshal(current.Artifacts[garden.ArtifactName], &data); err != nil {
			t.Fatal(err)
		}
		rows := data["species"].([]any)
		rows[0].(map[string]any)["harvest_units"] = float64(6)
		if promoteStarter {
			rows[2].(map[string]any)["starter"] = true
		}
		encoded, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}
		next := current
		next.Artifacts = map[string][]byte{}
		for key, value := range current.Artifacts {
			next.Artifacts[key] = value
		}
		next.Artifacts[garden.ArtifactName] = encoded
		next.Garden, err = garden.LoadCatalog(encoded, gardenDeclarationsForTest(current))
		if err != nil {
			t.Fatal(err)
		}
		next.ConstantsHash, err = save.ConstantsHashArtifacts(next.Artifacts)
		if err != nil || !next.valid(next.ConstantsHash) || garden.ValidateTransition(current.Garden, next.Garden) != nil {
			t.Fatal("retune control not admitted")
		}
		state := garden.NewState(current.Garden)
		err = garden.ValidateAgainst(next.Garden, state)
		if !promoteStarter && err != nil {
			t.Fatalf("ordinary numeric retune refused: %v", err)
		}
		if promoteStarter && err == nil {
			t.Fatal("diagnostic changed: starter retune no longer conflicts with untouched carry")
		}
		t.Logf("starter_promotion=%t transition_admitted=true untouched_carry_validation=%v", promoteStarter, err)
		now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
		founder := reputationFounderState(t, current, 25, now, 4)
		company := gardenCompanyState(t, current, now)
		newCompany := gardenCompanyState(t, next, now.Add(time.Hour))
		err = settleAndActivateFoundations(current, next, founder, company, newCompany)
		if (err != nil) != promoteStarter {
			t.Fatalf("retune boundary diagnostic: starter=%t err=%v", promoteStarter, err)
		}
		assertGardenBytes(t, founder.ServerGarden, emptyGardenJSON)
		t.Logf("starter_promotion=%t actual_foundation_boundary=%v unchanged_garden=true", promoteStarter, err)
	}
}
