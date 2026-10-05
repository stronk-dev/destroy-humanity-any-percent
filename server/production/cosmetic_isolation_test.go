package production

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/meters"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/save"
)

type cosmeticIsolationProbes struct {
	receipt      func(int, *save.IntentDecision)
	nextRunBonus func(int, []save.FrozenContribution)
}

// cosmeticIsolationRun is AC7's property: the existing 24-simulated-hour,
// 200-seed Company intent policy runs twice — plain, and with Founder cosmetic
// intents interleaved at seeded steps. Company bytes, frozen Founder
// contributions (the only Founder→production multiplier channel), and every
// non-cosmetics Founder byte must be identical. It returns the first
// divergence so the failing case can assert that a violating arm is caught.
func cosmeticIsolationRun(t *testing.T, seeds int64, probes ...cosmeticIsolationProbes) error {
	t.Helper()
	shop := cosmeticsContentBundle(t)
	catalog := shop.Economy
	service := &Service{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	baseState := reputationFounderState(t, shop, 24, now, 1)
	baseState.AgeMS = 5_000_000
	// A real declared, non-unit Founder contribution makes a disconnected
	// production consumer observable; Reputation level alone has unit factor.
	baseState.FiscalGeneratorLevels["generator.beige_tower"] = 1
	setup := &cosmeticRunner{t: t, catalogs: shop, bundle: "shop", state: baseState, revision: 1}
	adopted := setup.command("isolation-adopt", IntentAdoptPet, `"species_id":"pet_species.server_room_cat","name_key":"pet.name.server_room_cat.n07"`, 0, now)
	var adoption founderAdoptionReceipt
	if adopted.Outcome != save.IntentApplied || json.Unmarshal(adopted.Receipt, &adoption) != nil {
		t.Fatalf("isolation adoption: %s", adopted.Receipt)
	}
	founderBase := mustEncodeState(t, setup.state)
	baselineFrozen, err := FrozenFounderContributions(shop, setup.state)
	if err != nil {
		t.Fatal(err)
	}
	nonunit := false
	for _, row := range baselineFrozen {
		nonunit = nonunit || row.Factor != "1e0"
	}
	if !nonunit {
		t.Fatal("isolation population has no non-unit Founder bonus")
	}
	frozenContributions, err := ResolveFrozenContributions(catalog, baselineFrozen)
	if err != nil {
		t.Fatal(err)
	}
	applied := map[string]int{}
	for seed := int64(0); seed < seeds; seed++ {
		random := rand.New(rand.NewSource(seed))
		founder, err := save.RestoreState(founderBase, 24, shop.Economy, economy.ScopeFounder, time.Time{})
		if err != nil {
			t.Fatal(err)
		}
		plainFounder := mustEncodeState(t, founder)
		companies := [2]*save.State{}
		for arm := range companies {
			copyFounder, err := save.RestoreState(founderBase, 24, catalog, economy.ScopeFounder, time.Time{})
			if err != nil {
				t.Fatal(err)
			}
			prior := replayFixtureState(t, catalog, now)
			_, companyFloor := shop.versionFloors()
			prior.WireVersion, prior.MeterBands = companyFloor, nil
			meterState, err := meters.NewRunState(shop.Meters, copyFounder.Notoriety)
			if err != nil {
				t.Fatal(err)
			}
			prior.MeterValues, prior.MeterDecayRemainders, prior.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
			prior.AchievementsEarnedRun = map[string]bool{}
			if _, err := initializeActivePlayState(prior, shop.Opportunities, petFixtureFounderID); err != nil {
				t.Fatal(err)
			}
			companies[arm] = cosmeticIsolationNewCompany(t, shop, copyFounder, prior, now)
		}
		revisions := [2]int64{1, 1}
		founderRevision := setup.revision
		for step := 1; step <= 288; step++ {
			at := now.Add(time.Duration(step) * 5 * time.Minute)
			manual := step == 1 || random.Intn(2) == 0
			count := int64(random.Intn(80) + 1)
			decisions := [2]save.IntentDecision{}
			for arm := 0; arm < 2; arm++ {
				candidate, err := save.RestoreState(mustEncodeState(t, companies[arm]), save.VersionForState(companies[arm]), catalog, economy.ScopeCompany, time.Time{})
				if err != nil {
					t.Fatal(err)
				}
				var decision save.IntentDecision
				if manual {
					decision, err = service.performManualBatch(IntentRequest{IntentID: "018f6b7c-9abc-7def-8abc-999999999999", Kind: IntentPerformManualBatch,
						ExpectedRevision: revisions[arm], ActionID: "manual.click", Count: count, WindowMS: 300_000}, candidate, catalog, save.Revision{Number: revisions[arm]}, ModeOnline, at, frozenContributions, nil)
				} else {
					decision, err = service.buyGenerator(IntentRequest{IntentID: "018f6b7c-9abc-7def-8abc-999999999999", Kind: IntentBuyGenerator,
						ExpectedRevision: revisions[arm], GeneratorID: "generator.beige_tower", CountMode: "max"}, candidate, catalog, save.Revision{Number: revisions[arm]}, ModeOnline, at, frozenContributions, &invariantCollector{}, nil)
				}
				if err != nil {
					t.Fatalf("seed=%d step=%d arm=%d: %v", seed, step, arm, err)
				}
				for _, probe := range probes {
					if probe.receipt != nil {
						probe.receipt(arm, &decision)
					}
				}
				decisions[arm] = decision
				if decision.Outcome == save.IntentApplied {
					companies[arm], revisions[arm] = candidate, revisions[arm]+1
				}
			}
			if decisions[0].Outcome != decisions[1].Outcome || !bytes.Equal(decisions[0].Receipt, decisions[1].Receipt) {
				return fmt.Errorf("seed=%d step=%d: production receipts diverged", seed, step)
			}
			if !bytes.Equal(mustEncodeState(t, companies[0]), mustEncodeState(t, companies[1])) {
				return fmt.Errorf("seed=%d step=%d: Company bytes diverged", seed, step)
			}
			// The cosmetic arm only: a seeded Founder cosmetic intent (~1 in 8 steps).
			if random.Intn(8) != 0 {
				continue
			}
			kind, body, tier := IntentAcquireCosmetic, `"cosmetic_id":"horse_armor"`, int64(random.Intn(2))
			switch random.Intn(3) {
			case 1:
				kind, body = IntentEquipCosmetic, fmt.Sprintf(`"cosmetic_id":"horse_armor","pet_id":%q`, adoption.PetID)
			case 2:
				kind, body = IntentUnequipCosmetic, fmt.Sprintf(`"pet_id":%q`, adoption.PetID)
			}
			intentID := fmt.Sprintf("01986666-7f%02x-7000-8000-%012x", step%256, seed)
			request, parseErr := ParseIntent([]byte(fmt.Sprintf(`{"intent_id":%q,"kind":%q,"expected_revision":%d,%s}`, intentID, kind, founderRevision, body)))
			if parseErr != nil {
				t.Fatal(parseErr)
			}
			// Every Founder command runs the automatic Fiscal sweep; pinning the
			// server time to the Founder's period-open instant keeps the sweep a
			// no-op so only a cosmetic effect could diverge.
			command := save.FounderReplayCommand{IntentID: intentID, FounderStreamID: "01986666-5c00-4000-8000-000000000001",
				FounderID: petFixtureFounderID, Revision: founderRevision, FounderLogSeq: founderRevision, ServerTSMS: now.UnixMilli()}
			resolved := founderCosmeticResolved{Kind: kind}
			if kind == IntentAcquireCosmetic {
				resolved.ActiveCompany = &founderCosmeticActiveCompany{CompanyStreamID: "01986666-1900-7000-8000-000000000901", CompanyRevision: revisions[1], RunSeq: 2, Tier: tier}
			}
			inputs, marshalErr := save.MarshalFounderReplayInputs(command, resolved)
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}
			transition, applyErr := ApplyFounderLogged(founder, request.CanonicalPayload, shop, inputs)
			if applyErr != nil {
				return fmt.Errorf("seed=%d step=%d %s: %w", seed, step, kind, applyErr)
			}
			if transition.Outcome == save.IntentApplied {
				founderRevision++
				applied[kind]++
			}
			if err := sameExceptCosmetics(plainFounder, mustEncodeState(t, founder)); err != nil {
				return fmt.Errorf("seed=%d step=%d: %w", seed, step, err)
			}
		}
		if !bytes.Equal(mustEncodeState(t, companies[0]), mustEncodeState(t, companies[1])) {
			return fmt.Errorf("seed=%d: Company bytes diverged", seed)
		}
		if err := sameExceptCosmetics(plainFounder, mustEncodeState(t, founder)); err != nil {
			return fmt.Errorf("seed=%d: %w", seed, err)
		}
		before, err := save.RestoreState(founderBase, 24, shop.Economy, economy.ScopeFounder, time.Time{})
		if err != nil {
			t.Fatal(err)
		}
		plainFrozen, err := FrozenFounderContributions(shop, before)
		if err != nil {
			t.Fatal(err)
		}
		shopFrozen, err := FrozenFounderContributions(shop, founder)
		if err != nil {
			return fmt.Errorf("seed=%d: frozen contributions: %w", seed, err)
		}
		if fmt.Sprint(plainFrozen) != fmt.Sprint(shopFrozen) {
			return fmt.Errorf("seed=%d: frozen Founder contributions diverged", seed)
		}
		// Recompute only at the next-run boundary, never between commands of
		// the 24-hour run. Exercise the actual declared bonus consumer too.
		nextCompany := [2]*save.State{}
		nextReceipt := [2][]byte{}
		for arm, input := range []*save.State{before, founder} {
			at := now.Add(24 * time.Hour)
			copyFounder, err := save.RestoreState(mustEncodeState(t, input), 24, catalog, economy.ScopeFounder, time.Time{})
			if err != nil {
				t.Fatal(err)
			}
			candidate := cosmeticIsolationNewCompany(t, shop, copyFounder, companies[arm], at)
			rows, err := FrozenFounderContributions(shop, copyFounder)
			if err != nil {
				t.Fatal(err)
			}
			for _, probe := range probes {
				if probe.nextRunBonus != nil {
					probe.nextRunBonus(arm, rows)
				}
			}
			contributions, err := ResolveFrozenContributions(catalog, rows)
			if err != nil {
				t.Fatal(err)
			}
			// Identical test seed funds a real generator purchase, followed by
			// lazy accrual. A cosmetic bonus must affect neither cash nor receipt.
			setCash(t, candidate, "1e3")
			buy, err := service.buyGenerator(IntentRequest{IntentID: "01986666-7f00-7000-8000-000000000901", Kind: IntentBuyGenerator,
				ExpectedRevision: 1, GeneratorID: "generator.beige_tower", CountMode: "exact", Count: 1}, candidate, catalog,
				save.Revision{Number: 1}, ModeOnline, at, contributions, &invariantCollector{}, nil)
			if err != nil || buy.Outcome != save.IntentApplied {
				t.Fatalf("next-run generator seed: outcome=%s err=%v", buy.Outcome, err)
			}
			decision, err := service.performManualBatch(IntentRequest{IntentID: "01986666-7f00-7000-8000-000000000902", Kind: IntentPerformManualBatch,
				ExpectedRevision: 2, ActionID: "manual.click", Count: 1, WindowMS: 300_000}, candidate, catalog,
				save.Revision{Number: 2}, ModeOnline, at.Add(5*time.Minute), contributions, nil)
			if err != nil || decision.Outcome != save.IntentApplied {
				t.Fatalf("next-run consumer: outcome=%s err=%v", decision.Outcome, err)
			}
			nextCompany[arm], nextReceipt[arm] = candidate, decision.Receipt
		}
		if !bytes.Equal(mustEncodeState(t, nextCompany[0]), mustEncodeState(t, nextCompany[1])) {
			return fmt.Errorf("seed=%d: next-run Company bytes diverged", seed)
		}
		if !bytes.Equal(nextReceipt[0], nextReceipt[1]) {
			return fmt.Errorf("seed=%d: next-run production receipts diverged", seed)
		}
	}
	// Non-vacuity: every intent kind actually applied somewhere in the run.
	if applied[IntentAcquireCosmetic] == 0 || applied[IntentEquipCosmetic] == 0 || applied[IntentUnequipCosmetic] == 0 {
		t.Fatalf("isolation run applied too few cosmetic intents: %v", applied)
	}
	return nil
}

func cosmeticIsolationNewCompany(t *testing.T, bundle CatalogBundle, founder, prior *save.State, now time.Time) *save.State {
	t.Helper()
	company, err := prestigecore.NewRunState(bundle.Economy, prior, founder, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := settleAndActivateFoundations(bundle, bundle, founder, prior, company); err != nil {
		t.Fatal(err)
	}
	if _, err := initializeActivePlayState(company, bundle.Opportunities, petFixtureFounderID); err != nil {
		t.Fatal(err)
	}
	if err := bundle.ValidateFoundationState(company); err != nil {
		t.Fatal(err)
	}
	return company
}

func sameExceptCosmetics(left, right []byte) error {
	var a, b map[string]json.RawMessage
	if json.Unmarshal(left, &a) != nil || json.Unmarshal(right, &b) != nil || len(a) != len(b) {
		return fmt.Errorf("Founder shape diverged")
	}
	for key, value := range a {
		if key != "cosmetics" && !bytes.Equal(value, b[key]) {
			return fmt.Errorf("Founder %q diverged", key)
		}
	}
	return nil
}

// AC7: cosmetics are mechanically isolated across 200 seeded 24-hour policies.
func TestCosmeticsAreMechanicallyIsolated(t *testing.T) {
	if err := cosmeticIsolationRun(t, 200); err != nil {
		t.Fatal(err)
	}
}

// AC7 failing case: a test-only transition variant that feeds a multiplier
// input on every cosmetic transition must make the property fail.
func TestCosmeticIsolationCatchesAMultiplierLeak(t *testing.T) {
	founderTransitionTestArm = func(state *save.State) {
		if state.FiscalGeneratorLevels != nil {
			state.FiscalGeneratorLevels["generator.beige_tower"]++
		}
	}
	defer func() { founderTransitionTestArm = nil }()
	err := cosmeticIsolationRun(t, 3)
	if err == nil || !strings.Contains(err.Error(), "diverged") {
		t.Fatalf("a cosmetic arm that raises a Fiscal generator level was not caught as a divergence: %v", err)
	}
}

// A receipt-only leak must not be hidden by unchanged Company/Founder bytes.
func TestCosmeticIsolationCatchesAReceiptOnlyLeak(t *testing.T) {
	err := cosmeticIsolationRun(t, 1, cosmeticIsolationProbes{receipt: func(arm int, decision *save.IntentDecision) {
		if arm == 1 {
			var receipt map[string]json.RawMessage
			if err := json.Unmarshal(decision.Receipt, &receipt); err != nil {
				t.Fatal(err)
			}
			receipt["cosmetic_bonus"] = json.RawMessage(`"2e0"`)
			var err error
			decision.Receipt, err = json.Marshal(receipt)
			if err != nil {
				t.Fatal(err)
			}
		}
	}})
	if err == nil || !strings.Contains(err.Error(), "production receipts diverged") {
		t.Fatalf("receipt-only leak was not caught by its named oracle: %v", err)
	}
}

func TestCosmeticIsolationCatchesAConsumedBonusLeak(t *testing.T) {
	err := cosmeticIsolationRun(t, 1, cosmeticIsolationProbes{nextRunBonus: func(arm int, rows []save.FrozenContribution) {
		if arm == 1 {
			for index := range rows {
				if rows[index].SourceID == "reputation.founder_bonus" {
					rows[index].Factor = "2e0"
				}
			}
		}
	}})
	if err == nil || !strings.Contains(err.Error(), "next-run Company bytes diverged") {
		t.Fatalf("consumed bonus-only leak was not caught by its Company oracle: %v", err)
	}
}
