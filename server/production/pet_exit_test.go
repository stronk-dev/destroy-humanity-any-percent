package production

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"cloud-clicker/server/pet"
	"cloud-clicker/server/save"
)

// PA6/AC14: replay the same recorded economic Exit under two next-run pins.
// Only the required empty identity map may differ from the immutable source
// outputs. A newer bundle is available even in the no-activation control.
func TestPetAdoptionExitActivationUsesRecordedPin(t *testing.T) {
	source, _ := reputationInputShapeSource(t)
	var row reputationExitCase
	for _, candidate := range source.ExitCases {
		if candidate.Name == "plan-applies-with-in-plan-prerequisite" {
			row = candidate
		}
	}
	if row.Founder == nil || row.Founder.StateVersion != 22 || row.Company.ConstantsHash != row.Company.NextConstantsHash {
		t.Fatal("expected the pinned same-epoch v22 Exit source")
	}
	row.Company.Case.CanonicalPayload = []byte(canonicalFixtureJSON(t, row.Company.Case.CanonicalPayload))
	row.Founder.CanonicalPayload = []byte(canonicalFixtureJSON(t, row.Founder.CanonicalPayload))
	tree, species := reputationContentBundle(t), petSpeciesContentBundle(t)
	if row.Company.ConstantsHash != tree.ConstantsHash {
		t.Fatal("source artifact identity changed")
	}
	for _, activate := range []bool{false, true} {
		name := "old-pin-despite-new-bundle-available"
		if activate {
			name = "recorded-species-pin-activates-v23"
		}
		t.Run(name, func(t *testing.T) {
			current := tree
			current.Next = &species
			hash, version := tree.ConstantsHash, 22
			if activate {
				hash, version = species.ConstantsHash, 23
			}
			companyWire, err := parseReplayInputs(row.Company.Case.ReplayInputs)
			if err != nil {
				t.Fatal(err)
			}
			resolved := reputationShapeObject(t, companyWire.Resolved)
			resolved["next_constants_hash"] = reputationShapeJSON(t, hash)
			companyWire.Resolved = reputationShapeJSON(t, resolved)
			companyInputs := reputationShapeJSON(t, companyWire)
			company := replayFixtureStateFromEncoded(t, current, row.Company.Case.PreState)
			transition, err := ApplyLoggedExit(company, row.Company.Case.CanonicalPayload, current, companyInputs)
			if err != nil || transition.Decision.Outcome != save.IntentApplied {
				t.Fatalf("Company Exit outcome=%s err=%v", transition.Decision.Outcome, err)
			}
			if canonicalFixtureJSON(t, transition.Decision.Receipt) != row.Company.Case.ReceiptJSON ||
				canonicalFixtureJSON(t, mustEncodeState(t, transition.Decision.NewCompanyState)) != row.Company.Case.NewCompanyJSON {
				t.Fatal("species activation changed the recorded Company economics or receipt")
			}
			wantCarry := reputationShapeObject(t, []byte(row.Company.Case.FounderOutputJSON))
			wantState := reputationShapeObject(t, []byte(row.Founder.PostStateJSON))
			if activate {
				extensions := reputationShapeObject(t, wantCarry["founder_extensions"])
				extensions["pet_identities"] = json.RawMessage(`{}`)
				wantCarry["founder_extensions"] = reputationShapeJSON(t, extensions)
				wantState["pet_identities"] = json.RawMessage(`{}`)
			}
			var originalCarry replayExitResolved
			if err := json.Unmarshal(companyWire.Resolved, &originalCarry); err != nil {
				t.Fatal(err)
			}
			if canonicalFixtureValue(t, replayFounderOutput(transition.Founder, originalCarry.FounderCarry)) != canonicalFixtureValue(t, wantCarry) || save.VersionForState(transition.Founder) != version {
				t.Fatal("Company Exit did not preserve the expected carry and version")
			}
			founderWire, err := parseFounderReplayInputs(row.Founder.ReplayInputs)
			if err != nil {
				t.Fatal(err)
			}
			facts := reputationShapeObject(t, founderWire.Resolved)
			facts["result_constants_hash"] = reputationShapeJSON(t, hash)
			facts["result_founder_wire_version"] = reputationShapeJSON(t, version)
			founderWire.Resolved = reputationShapeJSON(t, facts)
			founder := reputationShapeRestore(t, *row.Founder, current)
			applied, err := ApplyFounderLogged(founder, row.Founder.CanonicalPayload, current, reputationShapeJSON(t, founderWire))
			if err != nil || applied.Outcome != save.IntentApplied || applied.ResultConstantsHash != hash || save.VersionForState(founder) != version ||
				canonicalFixtureJSON(t, mustEncodeState(t, founder)) != canonicalFixtureValue(t, wantState) {
				t.Fatalf("Founder Exit version=%d pin=%s err=%v", save.VersionForState(founder), applied.ResultConstantsHash, err)
			}
			if !activate {
				return
			}
			// The recorded species target is not replaceable by the old bundle.
			missing := current
			missing.Next = nil
			company = replayFixtureStateFromEncoded(t, current, row.Company.Case.PreState)
			before := mustEncodeState(t, company)
			if _, err := ApplyLoggedExit(company, row.Company.Case.CanonicalPayload, missing, companyInputs); err == nil || !bytes.Equal(before, mustEncodeState(t, company)) {
				t.Fatalf("missing target accepted or changed Company: %v", err)
			}
			facts["result_founder_wire_version"] = json.RawMessage(`22`)
			founderWire.Resolved = reputationShapeJSON(t, facts)
			founder = reputationShapeRestore(t, *row.Founder, current)
			before = mustEncodeState(t, founder)
			if _, err := ApplyFounderLogged(founder, row.Founder.CanonicalPayload, current, reputationShapeJSON(t, founderWire)); !errors.Is(err, ErrInvalidReplayInputs) || !bytes.Equal(before, mustEncodeState(t, founder)) {
				t.Fatalf("wrong activation floor accepted or changed Founder: %v", err)
			}
		})
	}
}

// PA3.5/AC6/AC8: reuse the already pinned nonempty Cosmetic Exit population.
// These are pure replay boundaries, not a claim of a live SQL adoption journey.
func TestPetIdentityAndCareSurviveBothExitPaths(t *testing.T) {
	raw, err := os.ReadFile(cosmeticCorpusPath)
	if err != nil {
		t.Fatal(err)
	}
	var source cosmeticCorpus
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	shop := cosmeticsContentBundle(t)
	count := 0
	for _, row := range source.ExitCases {
		if row.Name != "exit-wind-down-preserves-owned-equipped" && row.Name != "exit-accept-offer-preserves-owned-equipped" {
			continue
		}
		count++
		t.Run(row.Name, func(t *testing.T) {
			if row.Founder == nil || row.Company.ConstantsHash != shop.ConstantsHash || row.Company.NextConstantsHash != shop.ConstantsHash {
				t.Fatal("nonempty Exit source/pins changed")
			}
			row.Company.Case.CanonicalPayload = []byte(canonicalFixtureJSON(t, row.Company.Case.CanonicalPayload))
			row.Founder.CanonicalPayload = []byte(canonicalFixtureJSON(t, row.Founder.CanonicalPayload))
			before := reputationShapeObject(t, row.Founder.PreState)
			identities := reputationShapeObject(t, before["pet_identities"])
			if len(identities) != 1 || len(reputationShapeObject(t, before["pets"])) != 1 {
				t.Fatal("Exit must exercise nonempty identity and care maps")
			}
			company := replayFixtureStateFromEncoded(t, shop, row.Company.Case.PreState)
			transition, err := ApplyLoggedExit(company, row.Company.Case.CanonicalPayload, shop, row.Company.Case.ReplayInputs)
			if err != nil || transition.Decision.Outcome != save.IntentApplied {
				t.Fatalf("Company Exit: %v", err)
			}
			founder := reputationShapeRestore(t, *row.Founder, shop)
			applied, err := ApplyFounderLogged(founder, row.Founder.CanonicalPayload, shop, row.Founder.ReplayInputs)
			if err != nil || applied.Outcome != save.IntentApplied {
				t.Fatalf("Founder Exit: %v", err)
			}
			post := reputationShapeObject(t, mustEncodeState(t, founder))
			companyFounder := reputationShapeObject(t, reputationShapeJSON(t, founderCarry(transition.Founder).FounderExtensions))
			for _, field := range []string{"pet_identities", "pets"} {
				if canonicalFixtureJSON(t, before[field]) != canonicalFixtureJSON(t, post[field]) || canonicalFixtureJSON(t, before[field]) != canonicalFixtureJSON(t, companyFounder[field]) {
					t.Fatalf("%s changed across an Exit axis", field)
				}
			}
			wire, err := parseReplayInputs(row.Company.Case.ReplayInputs)
			if err != nil {
				t.Fatal(err)
			}
			resolved := reputationShapeObject(t, wire.Resolved)
			carry := reputationShapeObject(t, resolved["founder_carry"])
			extensions := reputationShapeObject(t, carry["founder_extensions"])
			delete(extensions, "pet_identities")
			carry["founder_extensions"] = reputationShapeJSON(t, extensions)
			resolved["founder_carry"] = reputationShapeJSON(t, carry)
			wire.Resolved = reputationShapeJSON(t, resolved)
			company = replayFixtureStateFromEncoded(t, shop, row.Company.Case.PreState)
			companyBefore := mustEncodeState(t, company)
			if _, err := ApplyLoggedExit(company, row.Company.Case.CanonicalPayload, shop, reputationShapeJSON(t, wire)); err == nil || !bytes.Equal(companyBefore, mustEncodeState(t, company)) {
				t.Fatalf("missing carry identities accepted or changed Company: %v", err)
			}
			// The existing test-only feature hook severs the actual Exit arm,
			// rather than testing the immutability helper alone.
			previous := founderTransitionTestArm
			founderTransitionTestArm = func(state *save.State) { state.PetIdentities = map[string]pet.Identity{} }
			defer func() { founderTransitionTestArm = previous }()
			founder = reputationShapeRestore(t, *row.Founder, shop)
			founderBefore := mustEncodeState(t, founder)
			if _, err := ApplyFounderLogged(founder, row.Founder.CanonicalPayload, shop, row.Founder.ReplayInputs); !errors.Is(err, ErrInvalidEngineState) || !bytes.Equal(founderBefore, mustEncodeState(t, founder)) {
				t.Fatalf("Exit dropped identities without rollback: %v", err)
			}
		})
	}
	if count != 2 {
		t.Fatalf("expected both nonempty Exit paths, got %d", count)
	}
}

// PA3.2 applies to the replay carry as well as the saved Founder: omitted
// and null coordinates are not an explicitly recorded numeric zero.
func TestPetExitIdentityFieldsAreRequired(t *testing.T) {
	raw, err := os.ReadFile(cosmeticCorpusPath)
	if err != nil {
		t.Fatal(err)
	}
	var source cosmeticCorpus
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	shop := cosmeticsContentBundle(t)
	count := 0
	for _, row := range source.ExitCases {
		if row.Name != "exit-wind-down-preserves-owned-equipped" && row.Name != "exit-accept-offer-preserves-owned-equipped" {
			continue
		}
		count++
		for _, field := range []string{"species_id", "temperament", "palette_id", "name_key", "adopted_at_ms", "adopted_at_attended_ms"} {
			for _, null := range []bool{false, true} {
				name := row.Name + "/" + field + "/omitted"
				if null {
					name = row.Name + "/" + field + "/null"
				}
				t.Run(name, func(t *testing.T) {
					wire := reputationShapeObject(t, row.Company.Case.ReplayInputs)
					resolved := reputationShapeObject(t, wire["resolved"])
					carry := reputationShapeObject(t, resolved["founder_carry"])
					extensions := reputationShapeObject(t, carry["founder_extensions"])
					identities := reputationShapeObject(t, extensions["pet_identities"])
					if len(identities) != 1 {
						t.Fatal("requires one recorded identity")
					}
					for id, data := range identities {
						identity := reputationShapeObject(t, data)
						if null {
							identity[field] = json.RawMessage(`null`)
						} else {
							delete(identity, field)
						}
						identities[id] = reputationShapeJSON(t, identity)
					}
					extensions["pet_identities"] = reputationShapeJSON(t, identities)
					carry["founder_extensions"] = reputationShapeJSON(t, extensions)
					resolved["founder_carry"] = reputationShapeJSON(t, carry)
					wire["resolved"] = reputationShapeJSON(t, resolved)
					company := replayFixtureStateFromEncoded(t, shop, row.Company.Case.PreState)
					before := mustEncodeState(t, company)
					transition, err := ApplyLoggedExit(company, []byte(canonicalFixtureJSON(t, row.Company.Case.CanonicalPayload)), shop, reputationShapeJSON(t, wire))
					if err == nil || !bytes.Equal(before, mustEncodeState(t, company)) {
						t.Fatalf("malformed identity accepted or changed Company: outcome=%s err=%v", transition.Decision.Outcome, err)
					}
				})
			}
		}
		t.Run(row.Name+"/explicit-zero-coordinates", func(t *testing.T) {
			wire, err := parseReplayInputs(row.Company.Case.ReplayInputs)
			if err != nil {
				t.Fatal(err)
			}
			var resolved replayExitResolved
			if err := json.Unmarshal(wire.Resolved, &resolved); err != nil {
				t.Fatal(err)
			}
			identities := resolved.FounderCarry.FounderExtensions.PetIdentities
			if identities == nil || len(*identities) != 1 {
				t.Fatal("requires one recorded identity")
			}
			for id, identity := range *identities {
				identity.AdoptedAtMS, identity.AdoptedAtAttendedMS = 0, 0
				(*identities)[id] = identity
			}
			wire.Resolved = reputationShapeJSON(t, resolved)
			company := replayFixtureStateFromEncoded(t, shop, row.Company.Case.PreState)
			transition, err := ApplyLoggedExit(company, []byte(canonicalFixtureJSON(t, row.Company.Case.CanonicalPayload)), shop, reputationShapeJSON(t, wire))
			if err != nil || transition.Decision.Outcome != save.IntentApplied ||
				canonicalFixtureValue(t, transition.Founder.PetIdentities) != canonicalFixtureValue(t, *identities) ||
				canonicalFixtureJSON(t, transition.Decision.Receipt) != row.Company.Case.ReceiptJSON {
				t.Fatalf("explicit zero coordinates rejected or changed: %v", err)
			}
		})
	}
	if count != 2 {
		t.Fatalf("expected both Exit paths, got %d", count)
	}
}
