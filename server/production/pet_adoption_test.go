package production

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/save"
)

const petAdoptionCorpusPath = "../../testdata/replay/pet-adoption-v1.json"

type petAdoptionCorpus struct {
	Version int                               `json:"version"`
	Bundles map[string]reputationCorpusBundle `json:"bundles"`
	Cases   []reputationCorpusCase            `json:"cases"`
}

// petFounderState is an encodable Founder at the bundle's floor (v22 or v23)
// with attendance age ageMS and no pet.
func petFounderState(t *testing.T, catalogs CatalogBundle, version int, now time.Time, ageMS int64) *save.State {
	t.Helper()
	state := replayFounderFixtureState(t, catalogs, now)
	state.WireVersion = version
	state.MinigameRatings = map[string]save.MinigameRatingState{"pitch": {Elo: 1000, SeasonMember: "s1"}}
	state.MinigameOfflineQuality = map[string]save.MinigameOfflineQualityState{"pitch": {GradePPM: 200_000}}
	state.Pets = map[string]pet.CareState{}
	state.FiscalPeriodOpenedWallMS, state.FiscalGeneratorLevels, state.FiscalUnlocks = now.UnixMilli(), map[string]int64{}, map[string]bool{}
	for _, row := range catalogs.Fiscal.GeneratorLevelRows() {
		state.FiscalGeneratorLevels[row.GeneratorID] = 0
	}
	state.Soul, state.SoulExhaustedSourceIDs = catalogs.Soul.Policy.Initial, []string{}
	state.ReputationNodesOwned = []string{}
	state.AgeMS = ageMS
	if version >= 23 {
		state.PetIdentities = map[string]pet.Identity{}
	}
	if err := catalogs.ValidateFoundationState(state); err != nil {
		t.Fatalf("pet Founder fixture invalid: %v", err)
	}
	return state
}

type petAdoptionRunner struct {
	t        *testing.T
	catalogs CatalogBundle
	bundle   string
	state    *save.State
	revision int64
	cases    []reputationCorpusCase
}

const petFixtureFounderID = "01986666-5d00-7000-8000-000000000001"

func (runner *petAdoptionRunner) command(name, kind, body, nonce string, serverTS time.Time, partialMS int64) FounderLoggedTransition {
	t := runner.t
	t.Helper()
	intentID := fmt.Sprintf("01986666-6a%02d-7000-8000-000000000001", len(runner.cases)+1)
	request, err := ParseIntent([]byte(fmt.Sprintf(`{"intent_id":%q,"kind":%q,"expected_revision":%d,%s}`, intentID, kind, runner.revision, body)))
	if err != nil {
		t.Fatal(err)
	}
	pre := mustEncodeState(t, runner.state)
	command := save.FounderReplayCommand{IntentID: intentID, FounderStreamID: "01986666-5c00-4000-8000-000000000001",
		FounderID: petFixtureFounderID, Revision: runner.revision, FounderLogSeq: runner.revision, ServerTSMS: serverTS.UnixMilli()}
	attendance := FounderAttendanceSample{CompanyStreamID: "01986666-1900-7000-8000-000000000901", RunSeq: 1, CompanyRevision: 1,
		CompanyConstantsHash: runner.catalogs.ConstantsHash, CompletedAttendedMS: runner.state.AgeMS, CurrentRunPartialAttendedMS: partialMS,
		EffectiveFounderAttendedMS: runner.state.AgeMS + partialMS}
	var resolved any = founderInvalidResolved{Kind: "invalid", Detail: request.InvalidDetail}
	if request.InvalidDetail == "" {
		switch kind {
		case IntentAdoptPet:
			resolved = founderAdoptionResolved{Kind: IntentAdoptPet, Attendance: attendance, AdoptionNonce: nonce}
		case IntentCareAction:
			before := int64(0)
			if care, ok := runner.state.Pets[request.PetID]; ok {
				before = care.EvaluatedThroughAttendedMS
			}
			resolved = founderCareResolved{Kind: IntentCareAction, Attendance: attendance, PetAttendedBeforeMS: before}
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

const (
	nonceA = "0123456789abcdef0123456789abcdef"
	nonceB = "deadbeefcafebabe0011223344556677"
)

func buildPetAdoptionCorpus(t *testing.T) petAdoptionCorpus {
	t.Helper()
	species := petSpeciesContentBundle(t)
	tree := reputationContentBundle(t)
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	adopt := func(name string) string {
		return fmt.Sprintf(`"species_id":"pet_species.server_room_cat","name_key":%q`, name)
	}

	inactive := &petAdoptionRunner{t: t, catalogs: tree, bundle: "tree", state: petFounderState(t, tree, 22, now, 1_000), revision: 1}
	runner := &petAdoptionRunner{t: t, catalogs: species, bundle: "species", state: petFounderState(t, species, 23, now, 5_000_000), revision: 1}
	for _, row := range []struct {
		runner           *petAdoptionRunner
		name, body       string
		category, detail string
	}{
		{runner, "rejects-invalid-fields", adopt("pet.name.server_room_cat.n07") + `,"extra":1`, "invalid", "adopt_pet.fields"},
		{runner, "rejects-client-temperament", adopt("pet.name.server_room_cat.n07") + `,"temperament":"shy"`, "invalid", "adopt_pet.fields"},
		{runner, "rejects-invalid-species-id", `"species_id":"Not Mechanical","name_key":"pet.name.server_room_cat.n07"`, "invalid", "species_id"},
		{inactive, "rejects-inactive-adoption", adopt("pet.name.server_room_cat.n07"), "not_eligible", "adoption_inactive"},
		{runner, "rejects-unknown-species", `"species_id":"pet_species.robot_vacuum","name_key":"pet.name.server_room_cat.n07"`, "unknown_id", "unknown_species"},
		{runner, "rejects-unknown-name", adopt("pet.name.server_room_cat.n99"), "unknown_id", "unknown_name"},
	} {
		before := mustEncodeState(t, row.runner.state)
		transition := row.runner.command(row.name, IntentAdoptPet, row.body, nonceA, now, 400_000)
		category, detail := rejectionOf(t, transition.Receipt)
		if category != row.category || detail != row.detail || len(transition.Events) != 0 || !bytes.Equal(before, mustEncodeState(t, row.runner.state)) {
			t.Fatalf("%s: rejection %s/%s events=%d", row.name, category, detail, len(transition.Events))
		}
	}
	// Row 8: apply. The pet is watermarked at A = age + partial.
	applied := runner.command("applies-starter-adoption", IntentAdoptPet, adopt("pet.name.server_room_cat.n07"), nonceA, now, 400_000)
	if applied.Outcome != save.IntentApplied || len(applied.Events) != 1 || applied.Events[0].Kind != save.EventPetAdopted || len(runner.state.PetIdentities) != 1 {
		t.Fatalf("adoption not applied: %s", applied.Receipt)
	}
	var receipt founderAdoptionReceipt
	if err := json.Unmarshal(applied.Receipt, &receipt); err != nil {
		t.Fatal(err)
	}
	wantDraws, _ := pet.DrawAdoption(petFixtureFounderID, nonceA, now.UnixMilli(), mustSpeciesRow(t, species))
	identity := runner.state.PetIdentities[receipt.PetID]
	if receipt.PetID != wantDraws.PetID || identity.Temperament != wantDraws.Temperament || identity.PaletteID != wantDraws.PaletteID ||
		identity.AdoptedAtAttendedMS != 5_400_000 || identity.AdoptedAtMS != now.UnixMilli() || receipt.AdoptedAtAttendedMS != 5_400_000 {
		t.Fatalf("receipt %+v identity %+v want %+v", receipt, identity, wantDraws)
	}
	// AC5: the initial care record is canonical and watermarked at A.
	wantCare, _ := pet.InitialCareState(species.Pets, 5_400_000)
	if care := runner.state.Pets[receipt.PetID]; petJSON(t, care) != petJSON(t, wantCare) {
		t.Fatalf("initial care %s want %s", petJSON(t, care), petJSON(t, wantCare))
	}
	// Row 6 at cap, and rows 3 before 6: an unknown species at cap reports
	// unknown_species, not adoption_cap_reached (AC4).
	for _, row := range []struct{ name, body, category, detail string }{
		{"rejects-second-adoption-at-cap", adopt("pet.name.server_room_cat.n01"), "not_eligible", "adoption_cap_reached"},
		{"rejects-unknown-species-at-cap", `"species_id":"pet_species.robot_vacuum","name_key":"pet.name.server_room_cat.n01"`, "unknown_id", "unknown_species"},
	} {
		if category, detail := rejectionOf(t, runner.command(row.name, IntentAdoptPet, row.body, nonceB, now, 400_000).Receipt); category != row.category || detail != row.detail {
			t.Fatalf("%s: %s/%s", row.name, category, detail)
		}
	}
	// AC5: care_action on the adopted pet applies, one attended second later.
	care := runner.command("care-applies-on-adopted-pet", IntentCareAction, fmt.Sprintf(`"pet_id":%q,"action_id":%q`, receipt.PetID, species.Pets.Actions[0].ActionID),
		"", now.Add(time.Second), 401_000)
	if care.Outcome != save.IntentApplied {
		t.Fatalf("care on adopted pet: %s", care.Receipt)
	}
	cases := append(append([]reputationCorpusCase{}, inactive.cases...), runner.cases...)
	return petAdoptionCorpus{Version: 1, Cases: cases, Bundles: map[string]reputationCorpusBundle{
		"species": {ConstantsHash: species.ConstantsHash, Artifacts: stringArtifacts(species.Artifacts)},
		"tree":    {ConstantsHash: tree.ConstantsHash, Artifacts: stringArtifacts(tree.Artifacts)},
	}}
}

func mustSpeciesRow(t *testing.T, bundle CatalogBundle) pet.SpeciesRow {
	t.Helper()
	row, ok := bundle.PetSpecies.Row("pet_species.server_room_cat")
	if !ok {
		t.Fatal("fixture species row missing")
	}
	return row
}

// TestPetAdoptionCorpus is AC4/AC5 and the Go half of AC7: it pins the
// Go-authored corpus the TS replay must byte-match. PET_ADOPTION_UPDATE_FIXTURE=1
// regenerates it.
func TestPetAdoptionCorpus(t *testing.T) {
	encoded, err := json.MarshalIndent(buildPetAdoptionCorpus(t), "", " ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if os.Getenv("PET_ADOPTION_UPDATE_FIXTURE") == "1" {
		if err := os.WriteFile(petAdoptionCorpusPath, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	pinned, err := os.ReadFile(petAdoptionCorpusPath)
	if err != nil {
		t.Fatalf("pinned corpus: %v", err)
	}
	if !bytes.Equal(pinned, encoded) {
		t.Fatal("pet adoption corpus drifted from the Go transition; regenerate with PET_ADOPTION_UPDATE_FIXTURE=1 and review")
	}
}

// AC7: replay re-derives the draws from the frozen nonce; a tampered nonce
// cannot reproduce the recorded receipt.
func TestPetAdoptionReplayRejectsTamperedNonce(t *testing.T) {
	corpus := buildPetAdoptionCorpus(t)
	species := petSpeciesContentBundle(t)
	for _, testCase := range corpus.Cases {
		if testCase.Name != "applies-starter-adoption" {
			continue
		}
		state, err := save.RestoreState(testCase.PreState, testCase.StateVersion, species.Economy, economy.ScopeFounder, time.Time{})
		if err != nil {
			t.Fatal(err)
		}
		tampered := bytes.Replace(testCase.ReplayInputs, []byte(nonceA), []byte(nonceB), 1)
		transition, err := ApplyFounderLogged(state, testCase.CanonicalPayload, species, tampered)
		if err != nil || canonicalFixtureJSON(t, transition.Receipt) == testCase.ReceiptJSON {
			t.Fatalf("tampered nonce reproduced the recorded receipt (err=%v)", err)
		}
		return
	}
	t.Fatal("applied adoption case missing")
}

// AC6: identity immutability is enforced by the transition layer.
func TestPetIdentityTransitionIsImmutableExceptOneAdoption(t *testing.T) {
	a := pet.Identity{SpeciesID: "pet_species.server_room_cat", Temperament: "curious", PaletteID: "pet_palette.fur_03", NameKey: "pet.name.server_room_cat.n07"}
	changed := a
	changed.Temperament = "shy"
	one := map[string]pet.Identity{adoptedPetID: a}
	two := map[string]pet.Identity{adoptedPetID: a, "01986666-bbbb-7bbb-8bbb-bbbbbbbbbbbb": a}
	for name, check := range map[string]error{
		"temperament mutated":         checkPetIdentityTransition(one, map[string]pet.Identity{adoptedPetID: changed}, false),
		"identity dropped":            checkPetIdentityTransition(one, map[string]pet.Identity{}, false),
		"non-adoption adds":           checkPetIdentityTransition(one, two, false),
		"adoption adds nothing":       checkPetIdentityTransition(one, one, true),
		"appeared outside activation": checkPetIdentityTransition(nil, one, false),
	} {
		if !errors.Is(check, ErrInvalidEngineState) {
			t.Fatalf("%s accepted: %v", name, check)
		}
	}
	if checkPetIdentityTransition(one, one, false) != nil || checkPetIdentityTransition(one, two, true) != nil || checkPetIdentityTransition(nil, map[string]pet.Identity{}, false) != nil {
		t.Fatal("legal identity transitions rejected")
	}
}

func petJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// AC6 failing cases: a feature arm that mutates a temperament, or an arm that
// drops the map, fails the authoritative transaction and restores state.
func TestTransitionLayerRejectsIdentityMutatingArms(t *testing.T) {
	corpus := buildPetAdoptionCorpus(t)
	species := petSpeciesContentBundle(t)
	var careCase reputationCorpusCase
	for _, testCase := range corpus.Cases {
		if testCase.Name == "care-applies-on-adopted-pet" {
			careCase = testCase
		}
	}
	for name, arm := range map[string]func(*save.State){
		"mutates a temperament": func(state *save.State) {
			for id, identity := range state.PetIdentities {
				identity.Temperament = "chaotic"
				if identity.Temperament == state.PetIdentities[id].Temperament {
					identity.Temperament = "lazy"
				}
				state.PetIdentities[id] = identity
			}
		},
		"drops the map": func(state *save.State) { state.PetIdentities = map[string]pet.Identity{} },
	} {
		state, err := save.RestoreState(careCase.PreState, careCase.StateVersion, species.Economy, economy.ScopeFounder, time.Time{})
		if err != nil {
			t.Fatal(err)
		}
		before := mustEncodeState(t, state)
		founderTransitionTestArm = arm
		_, err = ApplyFounderLogged(state, careCase.CanonicalPayload, species, careCase.ReplayInputs)
		founderTransitionTestArm = nil
		if !errors.Is(err, ErrInvalidEngineState) || !bytes.Equal(before, mustEncodeState(t, state)) {
			t.Fatalf("%s: err=%v restored=%t", name, err, bytes.Equal(before, mustEncodeState(t, state)))
		}
	}
}

// AC15: adoption touches no Company state and no Founder input to Company
// economics. The frozen Founder contributions pinned at every run start (the
// sole Founder→Company economic channel) are byte-identical with and without
// an adopted pet.
func TestPetAdoptionIsEconomicallyIsolated(t *testing.T) {
	species := petSpeciesContentBundle(t)
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	runner := &petAdoptionRunner{t: t, catalogs: species, bundle: "species", state: petFounderState(t, species, 23, now, 5_000_000), revision: 1}
	before, err := FrozenFounderContributions(species, runner.state)
	if err != nil {
		t.Fatal(err)
	}
	if transition := runner.command("isolation-adoption", IntentAdoptPet, `"species_id":"pet_species.server_room_cat","name_key":"pet.name.server_room_cat.n05"`, nonceA, now, 0); transition.Outcome != save.IntentApplied {
		t.Fatalf("adoption rejected: %s", transition.Receipt)
	}
	after, err := FrozenFounderContributions(species, runner.state)
	if err != nil {
		t.Fatal(err)
	}
	if petJSON(t, before) != petJSON(t, after) {
		t.Fatalf("adoption changed frozen Founder contributions:\n%s\n%s", petJSON(t, before), petJSON(t, after))
	}
}
