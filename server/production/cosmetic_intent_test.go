package production

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/cosmetic"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

const cosmeticCorpusPath = "../../testdata/replay/cosmetic-v1.json"

type cosmeticCorpus struct {
	Version   int                               `json:"version"`
	Bundles   map[string]reputationCorpusBundle `json:"bundles"`
	Cases     []reputationCorpusCase            `json:"cases"`
	ExitCases []reputationExitCase              `json:"exit_cases"`
}

// twoItemCosmeticsBundle pins a second cosmetic so AC6's replace and
// not_owned rows are reachable (the v1 catalog has only Horse Armor).
func twoItemCosmeticsBundle(t *testing.T) CatalogBundle {
	t.Helper()
	bundle := cosmeticsContentBundle(t)
	artifacts := map[string][]byte{}
	for name, data := range bundle.Artifacts {
		artifacts[name] = data
	}
	artifacts["cosmetics"] = []byte(`{"schema_version":1,"items":[{"cosmetic_id":"horse_armor","slot":"pet","unlock":{"kind":"active_company_tier_at_least","tier":1}},{"cosmetic_id":"zebra_armor","slot":"pet","unlock":{"kind":"active_company_tier_at_least","tier":0}}]}`)
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Cosmetics, err = cosmetic.Load(artifacts["cosmetics"]); err != nil {
		t.Fatal(err)
	}
	bundle.ConstantsHash, bundle.Artifacts = hash, artifacts
	if !bundle.valid(hash) {
		t.Fatal("two-item cosmetics bundle is not valid")
	}
	return bundle
}

type cosmeticRunner struct {
	t        *testing.T
	catalogs CatalogBundle
	bundle   string
	state    *save.State
	revision int64
	cases    []reputationCorpusCase
}

// command appends one Founder-log case. tier is the frozen active-Company tier
// for acquire_cosmetic; adopt_pet uses a fixed attendance sample and nonce.
func (runner *cosmeticRunner) command(name, kind, body string, tier int64, now time.Time) FounderLoggedTransition {
	t := runner.t
	t.Helper()
	intentID := fmt.Sprintf("01986666-6c%02d-7000-8000-000000000001", len(runner.cases)+1)
	request, err := ParseIntent([]byte(fmt.Sprintf(`{"intent_id":%q,"kind":%q,"expected_revision":%d,%s}`, intentID, kind, runner.revision, body)))
	if err != nil {
		t.Fatal(err)
	}
	pre := mustEncodeState(t, runner.state)
	command := save.FounderReplayCommand{IntentID: intentID, FounderStreamID: "01986666-5c00-4000-8000-000000000001",
		FounderID: petFixtureFounderID, Revision: runner.revision, FounderLogSeq: runner.revision, ServerTSMS: now.UnixMilli()}
	var resolved any = founderInvalidResolved{Kind: "invalid", Detail: request.InvalidDetail}
	if request.InvalidDetail == "" {
		switch {
		case kind == IntentAdoptPet:
			attendance := FounderAttendanceSample{CompanyStreamID: "01986666-1900-7000-8000-000000000901", RunSeq: 1, CompanyRevision: 1,
				CompanyConstantsHash: runner.catalogs.ConstantsHash, CompletedAttendedMS: runner.state.AgeMS, CurrentRunPartialAttendedMS: 400_000,
				EffectiveFounderAttendedMS: runner.state.AgeMS + 400_000}
			resolved = founderAdoptionResolved{Kind: IntentAdoptPet, Attendance: attendance, AdoptionNonce: nonceA}
		case kind == IntentAcquireCosmetic:
			resolved = founderCosmeticResolved{Kind: kind, ActiveCompany: &founderCosmeticActiveCompany{
				CompanyStreamID: "01986666-1900-7000-8000-000000000901", CompanyRevision: 7, RunSeq: 2, Tier: tier}}
		default:
			resolved = founderCosmeticResolved{Kind: kind}
		}
	}
	inputs, err := save.MarshalFounderReplayInputs(command, resolved)
	if err != nil {
		t.Fatal(err)
	}
	transition, err := ApplyFounderLogged(runner.state, request.CanonicalPayload, runner.catalogs, inputs)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if transition.Outcome == save.IntentApplied {
		runner.revision++
	}
	post := mustEncodeState(t, runner.state)
	runner.cases = append(runner.cases, reputationCorpusCase{Name: name, Bundle: runner.bundle, StateVersion: save.VersionForState(runner.state),
		PreState: pre, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs, Outcome: string(transition.Outcome),
		ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, fixtureEvents(transition.Events)), PostStateJSON: canonicalFixtureJSON(t, post)})
	return transition
}

func (runner *cosmeticRunner) expectRejection(name, kind, body string, tier int64, now time.Time, category, detail string) {
	t := runner.t
	t.Helper()
	before := mustEncodeState(t, runner.state)
	transition := runner.command(name, kind, body, tier, now)
	gotCategory, gotDetail := rejectionOf(t, transition.Receipt)
	if gotCategory != category || gotDetail != detail || len(transition.Events) != 0 || !bytes.Equal(before, mustEncodeState(t, runner.state)) {
		t.Fatalf("%s: rejection %s/%s events=%d receipt=%s", name, gotCategory, gotDetail, len(transition.Events), transition.Receipt)
	}
}

func (runner *cosmeticRunner) expectApplied(name, kind, body string, tier int64, now time.Time, event save.EventKind) FounderLoggedTransition {
	t := runner.t
	t.Helper()
	transition := runner.command(name, kind, body, tier, now)
	if transition.Outcome != save.IntentApplied || len(transition.Events) != 1 || transition.Events[0].Kind != event {
		t.Fatalf("%s: not applied: %s", name, transition.Receipt)
	}
	return transition
}

func buildCosmeticCorpus(t *testing.T) cosmeticCorpus {
	t.Helper()
	species := petSpeciesContentBundle(t)
	shop := cosmeticsContentBundle(t)
	pair := twoItemCosmeticsBundle(t)
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	acquire := func(id string) string { return fmt.Sprintf(`"cosmetic_id":%q`, id) }
	equip := func(id, petID string) string { return fmt.Sprintf(`"cosmetic_id":%q,"pet_id":%q`, id, petID) }
	unequip := func(petID string) string { return fmt.Sprintf(`"pet_id":%q`, petID) }
	const ghostPet = "01986666-bbbb-7bbb-8bbb-bbbbbbbbbbbb"

	inactive := &cosmeticRunner{t: t, catalogs: species, bundle: "species", state: reputationFounderState(t, species, 23, now, 1), revision: 1}
	inactive.expectRejection("rejects-inactive-below-v24", IntentAcquireCosmetic, acquire("horse_armor"), 1, now, "not_eligible", "inactive")

	shopState := reputationFounderState(t, shop, 24, now, 1)
	shopState.AgeMS = 5_000_000
	runner := &cosmeticRunner{t: t, catalogs: shop, bundle: "shop", state: shopState, revision: 1}
	// N2: a payment-shaped extra field is terminal invalid, never read.
	runner.expectRejection("rejects-price-field", IntentAcquireCosmetic, acquire("horse_armor")+`,"price":0`, 1, now, "invalid", "acquire_cosmetic.fields")
	runner.expectRejection("rejects-client-tier", IntentAcquireCosmetic, acquire("horse_armor")+`,"tier":1`, 1, now, "invalid", "acquire_cosmetic.fields")
	runner.expectRejection("rejects-unknown-cosmetic", IntentAcquireCosmetic, acquire("zebra_armor"), 1, now, "unknown_id", "cosmetic_id")
	runner.expectRejection("rejects-locked-at-tier-0", IntentAcquireCosmetic, acquire("horse_armor"), 0, now, "not_eligible", "locked")
	acquirePre := mustEncodeState(t, runner.state)
	applied := runner.expectApplied("applies-acquire-at-tier-1", IntentAcquireCosmetic, acquire("horse_armor"), 1, now, save.EventCosmeticAcquired)
	if string(applied.Events[0].Payload) != `{"cosmetic_id":"horse_armor","order_number":1}` {
		t.Fatalf("acquire payload %s", applied.Events[0].Payload)
	}
	// I2/AC5: acquisition leaves every non-cosmetics byte, ledger included, identical.
	assertOnlyCosmeticsChanged(t, acquirePre, mustEncodeState(t, runner.state))
	for _, forbidden := range []string{"price", "amount", "currency", "payment", "cost"} {
		if bytes.Contains(applied.Receipt, []byte(forbidden)) {
			t.Fatalf("acquire receipt carries %q: %s", forbidden, applied.Receipt)
		}
	}
	runner.expectRejection("rejects-second-acquire", IntentAcquireCosmetic, acquire("horse_armor"), 1, now, "not_eligible", "owned")
	runner.expectRejection("rejects-equip-unknown-pet", IntentEquipCosmetic, equip("horse_armor", ghostPet), 0, now, "unknown_id", "pet_id")
	adopted := runner.command("adopts-a-wearer", IntentAdoptPet, `"species_id":"pet_species.server_room_cat","name_key":"pet.name.server_room_cat.n07"`, 0, now)
	var adoption founderAdoptionReceipt
	if adopted.Outcome != save.IntentApplied || json.Unmarshal(adopted.Receipt, &adoption) != nil {
		t.Fatalf("adoption for the wearer: %s", adopted.Receipt)
	}
	runner.expectRejection("rejects-unequip-nothing-equipped", IntentUnequipCosmetic, unequip(adoption.PetID), 0, now, "not_eligible", "nothing_equipped")
	equipped := runner.expectApplied("applies-equip", IntentEquipCosmetic, equip("horse_armor", adoption.PetID), 0, now, save.EventCosmeticEquipped)
	if string(equipped.Events[0].Payload) != `{"cosmetic_id":"horse_armor","pet_id":"`+adoption.PetID+`","replaced_cosmetic_id":null}` {
		t.Fatalf("equip payload %s", equipped.Events[0].Payload)
	}
	runner.expectRejection("rejects-already-equipped", IntentEquipCosmetic, equip("horse_armor", adoption.PetID), 0, now, "not_eligible", "already_equipped")
	runner.expectRejection("rejects-unequip-unknown-pet", IntentUnequipCosmetic, unequip(ghostPet), 0, now, "unknown_id", "pet_id")
	runner.expectApplied("applies-unequip", IntentUnequipCosmetic, unequip(adoption.PetID), 0, now, save.EventCosmeticUnequipped)

	// AC6 replace and not_owned, on the two-item fixture catalog.
	pairState := reputationFounderState(t, pair, 24, now, 1)
	pairState.AgeMS = 5_000_000
	pairRunner := &cosmeticRunner{t: t, catalogs: pair, bundle: "pair", state: pairState, revision: 1}
	pairAdopted := pairRunner.command("pair-adopts-a-wearer", IntentAdoptPet, `"species_id":"pet_species.server_room_cat","name_key":"pet.name.server_room_cat.n07"`, 0, now)
	var pairAdoption founderAdoptionReceipt
	if pairAdopted.Outcome != save.IntentApplied || json.Unmarshal(pairAdopted.Receipt, &pairAdoption) != nil {
		t.Fatalf("pair adoption: %s", pairAdopted.Receipt)
	}
	pairRunner.expectRejection("pair-rejects-equip-not-owned", IntentEquipCosmetic, equip("zebra_armor", pairAdoption.PetID), 0, now, "not_eligible", "not_owned")
	pairRunner.expectApplied("pair-acquires-tier-0-item", IntentAcquireCosmetic, acquire("zebra_armor"), 0, now, save.EventCosmeticAcquired)
	second := pairRunner.expectApplied("pair-acquires-second-item", IntentAcquireCosmetic, acquire("horse_armor"), 1, now, save.EventCosmeticAcquired)
	if string(second.Events[0].Payload) != `{"cosmetic_id":"horse_armor","order_number":2}` {
		t.Fatalf("order number payload %s", second.Events[0].Payload)
	}
	pairRunner.expectApplied("pair-equips-first", IntentEquipCosmetic, equip("zebra_armor", pairAdoption.PetID), 0, now, save.EventCosmeticEquipped)
	replaced := pairRunner.expectApplied("pair-replaces", IntentEquipCosmetic, equip("horse_armor", pairAdoption.PetID), 0, now, save.EventCosmeticEquipped)
	if string(replaced.Events[0].Payload) != `{"cosmetic_id":"horse_armor","pet_id":"`+pairAdoption.PetID+`","replaced_cosmetic_id":"zebra_armor"}` {
		t.Fatalf("replace payload %s", replaced.Events[0].Payload)
	}

	cases := append(append(append([]reputationCorpusCase{}, inactive.cases...), runner.cases...), pairRunner.cases...)
	exitCases := []reputationExitCase{makeReputationPlanExitCase(t, "exit-activates-founder-v24", species, shop, 23, 4, nil, now)}
	return cosmeticCorpus{Version: 1, Cases: cases, ExitCases: exitCases, Bundles: map[string]reputationCorpusBundle{
		"species": {ConstantsHash: species.ConstantsHash, Artifacts: stringArtifacts(species.Artifacts)},
		"shop":    {ConstantsHash: shop.ConstantsHash, Artifacts: stringArtifacts(shop.Artifacts)},
		"pair":    {ConstantsHash: pair.ConstantsHash, Artifacts: stringArtifacts(pair.Artifacts)},
	}}
}

// TestCosmeticCorpus is AC5/AC6: every §4 row in Go, pinned as the corpus the
// TS replay must byte-match. COSMETIC_UPDATE_FIXTURE=1 regenerates it.
func TestCosmeticCorpus(t *testing.T) {
	encoded, err := json.MarshalIndent(buildCosmeticCorpus(t), "", " ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if os.Getenv("COSMETIC_UPDATE_FIXTURE") == "1" {
		if err := os.WriteFile(cosmeticCorpusPath, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	pinned, err := os.ReadFile(cosmeticCorpusPath)
	if err != nil {
		t.Fatalf("pinned corpus: %v", err)
	}
	if !bytes.Equal(pinned, encoded) {
		t.Fatal("cosmetic corpus drifted from the Go transition; regenerate with COSMETIC_UPDATE_FIXTURE=1 and review")
	}
}

// AC8/§4.5: only cosmetic intents may change cosmetics; a non-cosmetic arm
// that touches it fails the transaction and restores state.
func TestTransitionLayerRejectsCosmeticsMutatingArms(t *testing.T) {
	a := &cosmetic.State{Owned: []string{"horse_armor"}, Equipped: map[string]string{}}
	b := &cosmetic.State{Owned: []string{}, Equipped: map[string]string{}}
	if checkCosmeticsTransition(a, b, false) == nil || checkCosmeticsTransition(nil, a, false) == nil {
		t.Fatal("illegal cosmetics transitions accepted")
	}
	if checkCosmeticsTransition(a, b, true) != nil || checkCosmeticsTransition(nil, cosmetic.NewState(), false) != nil || checkCosmeticsTransition(a, a.Clone(), false) != nil {
		t.Fatal("legal cosmetics transitions rejected")
	}
	corpus := buildCosmeticCorpus(t)
	shop := cosmeticsContentBundle(t)
	var adoptCase reputationCorpusCase
	for _, testCase := range corpus.Cases {
		if testCase.Name == "adopts-a-wearer" {
			adoptCase = testCase
		}
	}
	state, err := save.RestoreState(adoptCase.PreState, adoptCase.StateVersion, shop.Economy, economy.ScopeFounder, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	before := mustEncodeState(t, state)
	founderTransitionTestArm = func(value *save.State) { value.Cosmetics.Owned = []string{} }
	_, err = ApplyFounderLogged(state, adoptCase.CanonicalPayload, shop, adoptCase.ReplayInputs)
	founderTransitionTestArm = nil
	if !errors.Is(err, ErrInvalidEngineState) || !bytes.Equal(before, mustEncodeState(t, state)) {
		t.Fatalf("adopt_pet arm that cleared cosmetics: err=%v restored=%t", err, bytes.Equal(before, mustEncodeState(t, state)))
	}
}

// §2/§6 fixture pets: one owned cosmetic may be worn by several pets (OD-15).
func TestOneCosmeticManyWearers(t *testing.T) {
	pets := map[string]struct{}{adoptedPetID: {}, "01986666-bbbb-7bbb-8bbb-bbbbbbbbbbbb": {}}
	state := &cosmetic.State{Owned: []string{"horse_armor"}, Equipped: map[string]string{adoptedPetID: "horse_armor", "01986666-bbbb-7bbb-8bbb-bbbbbbbbbbbb": "horse_armor"}}
	if err := cosmetic.ValidateShape(state, pets); err != nil {
		t.Fatalf("two wearers of one owned item rejected: %v", err)
	}
}

func assertOnlyCosmeticsChanged(t *testing.T, before, after []byte) {
	t.Helper()
	var left, right map[string]json.RawMessage
	if json.Unmarshal(before, &left) != nil || json.Unmarshal(after, &right) != nil || len(left) != len(right) {
		t.Fatal("state shape changed")
	}
	for key, value := range left {
		if key != "cosmetics" && !bytes.Equal(value, right[key]) {
			t.Fatalf("cosmetic intent changed %q: %s -> %s", key, value, right[key])
		}
	}
	if bytes.Equal(left["cosmetics"], right["cosmetics"]) {
		t.Fatal("applied cosmetic intent did not change cosmetics")
	}
}
