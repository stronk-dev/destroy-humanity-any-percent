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
	"cloud-clicker/server/save"
)

// cosmeticIsolationRun is AC7's property: the existing 24-simulated-hour,
// 200-seed Company intent policy runs twice — plain, and with Founder cosmetic
// intents interleaved at seeded steps. Company bytes, frozen Founder
// contributions (the only Founder→production multiplier channel), and every
// non-cosmetics Founder byte must be identical. It returns the first
// divergence so the failing case can assert that a violating arm is caught.
func cosmeticIsolationRun(t *testing.T, seeds int64) error {
	t.Helper()
	catalog := phase0Catalog(t)
	shop := cosmeticsContentBundle(t)
	service := &Service{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	baseState := reputationFounderState(t, shop, 24, now, 1)
	baseState.AgeMS = 5_000_000
	setup := &cosmeticRunner{t: t, catalogs: shop, bundle: "shop", state: baseState, revision: 1}
	adopted := setup.command("isolation-adopt", IntentAdoptPet, `"species_id":"pet_species.server_room_cat","name_key":"pet.name.server_room_cat.n07"`, 0, now)
	var adoption founderAdoptionReceipt
	if adopted.Outcome != save.IntentApplied || json.Unmarshal(adopted.Receipt, &adoption) != nil {
		t.Fatalf("isolation adoption: %s", adopted.Receipt)
	}
	founderBase := mustEncodeState(t, setup.state)
	applied := map[string]int{}
	for seed := int64(0); seed < seeds; seed++ {
		random := rand.New(rand.NewSource(seed))
		companies := [2]*save.State{engineState(t, catalog, "0", 0), engineState(t, catalog, "0", 0)}
		revisions := [2]int64{1, 1}
		founder, err := save.RestoreState(founderBase, 24, shop.Economy, economy.ScopeFounder, time.Time{})
		if err != nil {
			t.Fatal(err)
		}
		plainFounder := mustEncodeState(t, founder)
		founderRevision := setup.revision
		for step := 1; step <= 288; step++ {
			at := engineCursor.Add(time.Duration(step) * 5 * time.Minute)
			manual := step == 1 || random.Intn(2) == 0
			count := int64(random.Intn(80) + 1)
			for arm := 0; arm < 2; arm++ {
				candidate := clonePolicyState(t, catalog, companies[arm])
				var decision save.IntentDecision
				if manual {
					decision, err = service.performManualBatch(IntentRequest{IntentID: "018f6b7c-9abc-7def-8abc-999999999999", Kind: IntentPerformManualBatch,
						ExpectedRevision: revisions[arm], ActionID: "manual.click", Count: count, WindowMS: 300_000}, candidate, catalog, save.Revision{Number: revisions[arm]}, ModeOnline, at, nil, nil)
				} else {
					decision, err = service.buyGenerator(IntentRequest{IntentID: "018f6b7c-9abc-7def-8abc-999999999999", Kind: IntentBuyGenerator,
						ExpectedRevision: revisions[arm], GeneratorID: "generator.beige_tower", CountMode: "max"}, candidate, catalog, save.Revision{Number: revisions[arm]}, ModeOnline, at, nil, &invariantCollector{}, nil)
				}
				if err != nil {
					t.Fatalf("seed=%d step=%d arm=%d: %v", seed, step, arm, err)
				}
				if decision.Outcome == save.IntentApplied {
					companies[arm], revisions[arm] = candidate, revisions[arm]+1
				}
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
	}
	// Non-vacuity: every intent kind actually applied somewhere in the run.
	if applied[IntentAcquireCosmetic] == 0 || applied[IntentEquipCosmetic] == 0 || applied[IntentUnequipCosmetic] == 0 {
		t.Fatalf("isolation run applied too few cosmetic intents: %v", applied)
	}
	return nil
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
