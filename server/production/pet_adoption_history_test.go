package production

import (
	"encoding/json"
	"os"
	"testing"

	"cloud-clicker/server/save"
)

// AC7 checks the final history verdict, not just whether a changed nonce
// changes a transition's output. Expected states/receipts/events come from
// the committed corpus; this test does not regenerate them from its subject.
func TestPetAdoptionHistoryRefusesCorruptedEvidence(t *testing.T) {
	raw, err := os.ReadFile(petAdoptionCorpusPath)
	if err != nil {
		t.Fatal(err)
	}
	var source petAdoptionCorpus
	if err := json.Unmarshal(raw, &source); err != nil || source.Version != 1 || len(source.Cases) != 10 {
		t.Fatalf("incomplete adoption history source: %v", err)
	}
	species, tree := petSpeciesContentBundle(t), reputationContentBundle(t)
	set := ReplayCatalogSet{species.ConstantsHash: species, tree.ConstantsHash: tree}
	hashes := map[string]string{"species": species.ConstantsHash, "tree": tree.ConstantsHash}
	for name, hash := range hashes {
		if source.Bundles[name].ConstantsHash != hash {
			t.Fatalf("pinned %s artifact identity differs from loaded fixture", name)
		}
	}
	var adoption *save.FounderHistory
	for _, row := range source.Cases {
		t.Run(row.Name, func(t *testing.T) {
			// Single-row histories retain their source revision. Only local log
			// sequence is rebased; this is not a SQL chronology claim.
			history := reputationHistoryFromRows(t, []reputationCorpusCase{row}, hashes, true)
			if verdict := VerifyFounderHistory(history, set); verdict != ReplayVerified {
				t.Fatalf("honest pinned history: %s", verdict)
			}
			if row.Name == "applies-starter-adoption" {
				adoption = &history
			}
		})
	}
	if adoption == nil {
		t.Fatal("applied adoption history is missing")
	}
	assertPetAdoptionHistoryIntegrity(t, *adoption, set)
}

// Reused with the actual Postgres-loaded history, so the same refusals also
// cover persisted evidence when the integration population executes.
func assertPetAdoptionHistoryIntegrity(t *testing.T, history save.FounderHistory, set ReplayCatalogSet) {
	t.Helper()
	if verdict := VerifyFounderHistory(history, set); verdict != ReplayVerified {
		t.Fatalf("honest adoption history: %s", verdict)
	}
	entryIndex, eventIndex := -1, -1
	for i, entry := range history.Entries {
		for j, event := range entry.Events {
			if event.Kind == save.EventPetAdopted {
				if entryIndex >= 0 || entry.AppliedRevision == nil {
					t.Fatal("expected exactly one applied adoption event")
				}
				entryIndex, eventIndex = i, j
			}
		}
	}
	if entryIndex < 0 {
		t.Fatal("adoption event is missing")
	}
	var receipt founderAdoptionReceipt
	if err := json.Unmarshal(history.Entries[entryIndex].Receipt, &receipt); err != nil || receipt.PetID == "" {
		t.Fatalf("adoption receipt: %v", err)
	}
	for _, change := range []string{"nonce", "missing-nonce", "receipt-pet", "missing-event", "event-name", "identity-name", "care-watermark", "applied-revision", "head-revision"} {
		t.Run("refuses-"+change, func(t *testing.T) {
			copy := cloneReputationHistory(t, history)
			entry := &copy.Entries[entryIndex]
			want := ReplayStateDivergence
			switch change {
			case "nonce", "missing-nonce":
				inputs := reputationShapeObject(t, entry.ReplayInputs)
				resolved := reputationShapeObject(t, inputs["resolved"])
				if change == "missing-nonce" {
					delete(resolved, "adoption_nonce")
				} else if string(resolved["adoption_nonce"]) == `"`+nonceB+`"` {
					resolved["adoption_nonce"] = reputationShapeJSON(t, nonceA)
				} else {
					resolved["adoption_nonce"] = reputationShapeJSON(t, nonceB)
				}
				inputs["resolved"] = reputationShapeJSON(t, resolved)
				entry.ReplayInputs = reputationShapeJSON(t, inputs)
			case "receipt-pet":
				value := reputationShapeObject(t, entry.Receipt)
				value["pet_id"] = reputationShapeJSON(t, "01986666-bbbb-7bbb-8bbb-bbbbbbbbbbbb")
				entry.Receipt = reputationShapeJSON(t, value)
			case "missing-event":
				entry.Events = append(entry.Events[:eventIndex], entry.Events[eventIndex+1:]...)
			case "event-name":
				value := reputationShapeObject(t, entry.Events[eventIndex].Payload)
				value["name_key"] = reputationShapeJSON(t, changedPetHistoryName(receipt.NameKey))
				entry.Events[eventIndex].Payload = reputationShapeJSON(t, value)
			case "identity-name", "care-watermark":
				head := reputationShapeObject(t, copy.HeadState)
				field, key := "pet_identities", "name_key"
				if change == "care-watermark" {
					field, key = "pets", "evaluated_through_attended_ms"
				}
				pets := reputationShapeObject(t, head[field])
				value := reputationShapeObject(t, pets[receipt.PetID])
				if change == "identity-name" {
					value[key] = reputationShapeJSON(t, changedPetHistoryName(receipt.NameKey))
				} else {
					var attended int64
					if err := json.Unmarshal(value[key], &attended); err != nil {
						t.Fatal(err)
					}
					value[key] = reputationShapeJSON(t, attended+1)
				}
				pets[receipt.PetID] = reputationShapeJSON(t, value)
				head[field] = reputationShapeJSON(t, pets)
				copy.HeadState = reputationShapeJSON(t, head)
			case "applied-revision":
				*entry.AppliedRevision++
				want = ReplayLogGap
			case "head-revision":
				copy.HeadRevision++
			}
			if verdict := VerifyFounderHistory(copy, set); verdict != want {
				t.Fatalf("corrupted %s verdict=%s want=%s", change, verdict, want)
			}
			if verdict := VerifyFounderHistory(history, set); verdict != ReplayVerified {
				t.Fatalf("negative control mutated original history: %s", verdict)
			}
		})
	}
	if verdict := VerifyFounderHistory(history, ReplayCatalogSet{}); verdict != ReplayConstantsMismatch {
		t.Fatalf("missing artifacts verdict=%s", verdict)
	}
	// Remove only pet_species, retaining its old label and every other artifact
	// and bundle. A deploy-current/other-epoch fallback must not repair this pin.
	stripped := set[history.Genesis.ConstantsHash]
	stripped.Artifacts = make(map[string][]byte, len(stripped.Artifacts))
	for name, data := range set[history.Genesis.ConstantsHash].Artifacts {
		if name != "pet_species" {
			stripped.Artifacts[name] = data
		}
	}
	if len(stripped.Artifacts)+1 != len(set[history.Genesis.ConstantsHash].Artifacts) {
		t.Fatal("species-only artifact removal did not occur")
	}
	available := ReplayCatalogSet{}
	for hash, bundle := range set {
		available[hash] = bundle
	}
	available[history.Genesis.ConstantsHash] = stripped
	if verdict := VerifyFounderHistory(history, available); verdict != ReplayConstantsMismatch {
		t.Fatalf("stripped pet_species artifact verdict=%s", verdict)
	}
}

func changedPetHistoryName(name string) string {
	if name == "pet.name.server_room_cat.n01" {
		return "pet.name.server_room_cat.n02"
	}
	return "pet.name.server_room_cat.n01"
}
