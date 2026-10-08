package save

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/pet"
)

const v23PetID = "01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa"

func TestFounderPetIdentitiesRequireExactCaseKeys(t *testing.T) {
	for _, state := range []*State{v23FounderState(t), v24FounderState(t)} {
		t.Run(fmt.Sprintf("v%d", state.WireVersion), func(t *testing.T) {
			encoded, err := EncodeState(state)
			if err != nil {
				t.Fatal(err)
			}
			decode := func(data []byte) (*State, error) {
				return RestoreState(data, state.WireVersion, stateCatalog(t), economy.ScopeFounder, time.Time{})
			}
			object := func(data []byte) map[string]json.RawMessage {
				t.Helper()
				var value map[string]json.RawMessage
				if err := json.Unmarshal(data, &value); err != nil {
					t.Fatal(err)
				}
				return value
			}
			marshal := func(value any) []byte {
				t.Helper()
				data, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				return data
			}
			refuse := func(name string, data []byte) {
				t.Helper()
				t.Run(name, func(t *testing.T) {
					if _, err := decode(data); !errors.Is(err, ErrInvalidState) {
						t.Fatalf("noncanonical pet identity keys accepted or wrong error: %v", err)
					}
				})
			}
			for _, key := range []string{"pet_identities", "species_id", "temperament", "palette_id", "name_key", "adopted_at_ms", "adopted_at_attended_ms"} {
				for _, alias := range []string{strings.ToUpper(key[:1]) + key[1:], strings.ToUpper(key)} {
					for _, retainCanonical := range []bool{false, true} {
						root := object(encoded)
						target := root
						if key != "pet_identities" {
							pets := object(root["pet_identities"])
							target = object(pets[v23PetID])
						}
						target[alias] = target[key]
						if !retainCanonical {
							delete(target, key)
						}
						if key != "pet_identities" {
							root["pet_identities"] = marshal(map[string]json.RawMessage{v23PetID: marshal(target)})
						}
						refuse(fmt.Sprintf("%s/retain=%t", alias, retainCanonical), marshal(root))
					}
				}
			}
			for _, data := range [][]byte{encoded, bytes.ReplaceAll(bytes.ReplaceAll(encoded,
				[]byte(`"pet_identities"`), []byte(`"pet_\u0069dentities"`)), []byte(`"adopted_at_ms"`), []byte(`"adopted_at_\u006ds"`))} {
				restored, err := decode(data)
				if err != nil {
					t.Fatalf("valid decoded names rejected: %v", err)
				}
				roundTrip, err := EncodeState(restored)
				if err != nil || !bytes.Equal(roundTrip, encoded) {
					t.Fatalf("valid save bytes changed: %v", err)
				}
			}
			root := object(encoded)
			pets := object(root["pet_identities"])
			identity := object(pets[v23PetID])
			identity["adopted_at_ms"], identity["adopted_at_attended_ms"] = json.RawMessage(`0`), json.RawMessage(`0`)
			root["pet_identities"] = marshal(map[string]json.RawMessage{v23PetID: marshal(identity)})
			if restored, err := decode(marshal(root)); err != nil || restored.PetIdentities[v23PetID].AdoptedAtMS != 0 || restored.PetIdentities[v23PetID].AdoptedAtAttendedMS != 0 {
				t.Fatalf("explicit zero coordinates rejected: %v", err)
			}
		})
	}
}

func v23CareState() pet.CareState {
	return pet.CareState{
		StatsPPM:                map[pet.StatID]int64{"affection": 700_000, "cleanliness": 700_000, "energy": 700_000, "hunger": 700_000},
		StatDecayRemaindersPPM:  map[pet.StatID]int64{"affection": 0, "cleanliness": 0, "energy": 0, "hunger": 0},
		CooldownUntilAttendedMS: map[string]int64{}, TrustPPM: 500_000, EvaluatedThroughAttendedMS: 5_400_000,
		BehaviorState: pet.BehaviorIdle, BehaviorEnteredAtAttendedMS: 5_400_000, BehaviorQueue: []pet.BehaviorQueueEntry{},
	}
}

func v23FounderState(t *testing.T) *State {
	t.Helper()
	catalog := stateCatalog(t)
	ledger, err := economy.NewLedger(catalog, economy.ScopeFounder)
	if err != nil {
		t.Fatal(err)
	}
	return &State{WireVersion: 23, Ledger: ledger, GeneratorCounts: map[string]int64{}, GeneratorProvisioned: map[string]int64{},
		ProvisionRemaindersPPM: map[string]int64{}, UpgradesOwned: map[string]bool{}, EvaluatedThrough: testCursor,
		ManualTokenRefilledAt: testCursor, GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{},
		LedgerFactKinds: map[string]bool{}, MeterValues: map[string]int{}, MeterDecayRemainders: map[string]int64{}, MeterInputRemainders: map[string]int64{},
		AchievementsEarnedRun: map[string]bool{}, AchievementsEarnedLifetime: map[string]bool{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{},
		CompactSamples: []CompactSample{}, OfflineSpans: []OfflineSpan{}, NetworkSlots: []NetworkSlot{}, ExitHistory: []ExitRecord{},
		MinigameRatings: map[string]MinigameRatingState{}, MinigameOfflineQuality: map[string]MinigameOfflineQualityState{},
		Pets:                     map[string]pet.CareState{v23PetID: v23CareState()},
		FiscalPeriodOpenedWallMS: 1_786_000_000_000, FiscalGeneratorLevels: map[string]int64{}, FiscalUnlocks: map[string]bool{},
		Soul: 73, SoulExhaustedSourceIDs: []string{}, MinigameSessionSeq: 2, ReputationLevel: 9, ReputationNodesOwned: []string{},
		PetIdentities: map[string]pet.Identity{v23PetID: {SpeciesID: "pet_species.server_room_cat", Temperament: "curious", PaletteID: "pet_palette.fur_03",
			NameKey: "pet.name.server_room_cat.n07", AdoptedAtMS: 1_790_000_000_000, AdoptedAtAttendedMS: 5_400_000}}}
}

// AC2: the Founder v23 codec round-trips identities and rejects every listed
// invariant violation.
func TestFounderV23PetIdentitiesRoundTripAndInvariants(t *testing.T) {
	catalog := stateCatalog(t)
	state := v23FounderState(t)
	encoded, err := EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreState(encoded, 23, catalog, economy.ScopeFounder, time.Time{})
	if err != nil || restored.PetIdentities[v23PetID] != state.PetIdentities[v23PetID] {
		t.Fatalf("v23 restore=%+v err=%v", restored.PetIdentities, err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	load := func(mutate func(map[string]json.RawMessage), version int, scope economy.Scope) error {
		candidate := map[string]json.RawMessage{}
		for name, value := range object {
			candidate[name] = value
		}
		mutate(candidate)
		data, _ := json.Marshal(candidate)
		_, err := RestoreState(data, version, catalog, scope, time.Time{})
		return err
	}
	identity := string(object["pet_identities"])
	rejections := map[string]error{
		"v23 without pet_identities": load(func(value map[string]json.RawMessage) { delete(value, "pet_identities") }, 23, economy.ScopeFounder),
		"Company v23":                load(func(map[string]json.RawMessage) {}, 23, economy.ScopeCompany),
		"pet_identities at v22":      load(func(map[string]json.RawMessage) {}, 22, economy.ScopeFounder),
		"identity without care":      load(func(value map[string]json.RawMessage) { value["pets"] = json.RawMessage(`{}`) }, 23, economy.ScopeFounder),
		"care without identity":      load(func(value map[string]json.RawMessage) { value["pet_identities"] = json.RawMessage(`{}`) }, 23, economy.ScopeFounder),
		"identity missing a key": load(func(value map[string]json.RawMessage) {
			value["pet_identities"] = json.RawMessage(`{"` + v23PetID + `":{"species_id":"pet_species.server_room_cat","temperament":"curious","palette_id":"pet_palette.fur_03","name_key":"pet.name.server_room_cat.n07","adopted_at_ms":1}}`)
		}, 23, economy.ScopeFounder),
		"identity with an extra key": load(func(value map[string]json.RawMessage) {
			value["pet_identities"] = json.RawMessage(identity[:len(identity)-2] + `,"extra":1}}`)
		}, 23, economy.ScopeFounder),
	}
	for name, err := range rejections {
		if !errors.Is(err, ErrInvalidState) {
			t.Fatalf("%s accepted: %v", name, err)
		}
	}
	for name, mutate := range map[string]func(*State){
		"non-v7 pet id": func(value *State) {
			id := "01986666-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
			value.Pets = map[string]pet.CareState{id: v23CareState()}
			value.PetIdentities = map[string]pet.Identity{id: value.PetIdentities[v23PetID]}
		},
		"unknown temperament": func(value *State) {
			identity := value.PetIdentities[v23PetID]
			identity.Temperament = "grumpy"
			value.PetIdentities = map[string]pet.Identity{v23PetID: identity}
		},
		"identities before v23": func(value *State) { value.WireVersion = 22 },
		"nil identities at v23": func(value *State) { value.PetIdentities = nil },
	} {
		candidate := v23FounderState(t)
		mutate(candidate)
		if _, err := EncodeState(candidate); !errors.Is(err, ErrInvalidState) {
			t.Fatalf("encode accepted %s: %v", name, err)
		}
	}
}
