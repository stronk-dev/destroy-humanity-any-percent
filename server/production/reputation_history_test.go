package production

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"cloud-clicker/server/save"
)

type reputationHistoryProfile struct {
	name    string
	history save.FounderHistory
}

func validateReputationHistorySource(source reputationCorpus, raw []byte) error {
	if err := validateReputationTaxonomySource(source, raw); err != nil {
		return err
	}
	chain := 0
	for _, row := range source.Cases {
		if strings.HasPrefix(row.Name, "chain-") {
			chain++
		}
	}
	if chain != 9 {
		return fmt.Errorf("history source requires nine chain rows, got %d", chain)
	}
	want := map[string]bool{"plan-applies-with-in-plan-prerequisite": false, "exit-activates-founder-v22": false, "activating-exit-applies-plan": false}
	paired := 0
	for _, row := range source.ExitCases {
		if row.Founder == nil {
			continue
		}
		seen, ok := want[row.Name]
		if !ok || seen {
			return fmt.Errorf("unexpected or duplicate paired Founder Exit %s", row.Name)
		}
		want[row.Name], paired = true, paired+1
	}
	if paired != 3 {
		return fmt.Errorf("history source requires three paired Founder Exits, got %d", paired)
	}
	return nil
}

// This wraps immutable source expectations, not new outputs from the replay
// being tested. A single-case history begins at that source row's revision;
// only its local log sequence is rebased. The nine-purchase chain is unchanged.
func reputationHistoryFromRows(t *testing.T, rows []reputationCorpusCase, hashes map[string]string, rebase bool) save.FounderHistory {
	t.Helper()
	var history save.FounderHistory
	for index, row := range rows {
		wire, err := parseFounderReplayInputs(row.ReplayInputs)
		if err != nil {
			t.Fatal(err)
		}
		if index == 0 {
			history = save.FounderHistory{FounderStreamID: wire.Command.FounderStreamID, FounderID: wire.Command.FounderID,
				Genesis: save.FounderGenesis{FounderStreamID: wire.Command.FounderStreamID, Revision: wire.Command.Revision,
					State: bytes.Clone(row.PreState), Version: row.StateVersion, ConstantsHash: hashes[row.Bundle]},
				Entries: []save.FounderHistoryEntry{}}
		} else if canonicalFixtureJSON(t, row.PreState) != string(history.HeadState) ||
			wire.Command.Revision != history.HeadRevision || hashes[row.Bundle] != history.HeadConstants {
			t.Fatal("chain source states/revisions/pins are not contiguous")
		}
		if rebase {
			wire.Command.FounderLogSeq = 1
		}
		if wire.Command.FounderLogSeq != int64(index+1) || wire.Command.FounderStreamID != history.FounderStreamID || wire.Command.FounderID != history.FounderID {
			t.Fatal("history coordinates are inconsistent")
		}
		var events []fixtureEvent
		if err := json.Unmarshal([]byte(row.EventsJSON), &events); err != nil {
			t.Fatal(err)
		}
		entry := save.FounderHistoryEntry{Sequence: int64(index + 1), IntentID: wire.Command.IntentID,
			CanonicalPayload: []byte(canonicalFixtureJSON(t, row.CanonicalPayload)), ReplayInputs: reputationShapeJSON(t, wire),
			Receipt: []byte(row.ReceiptJSON), ConstantsHash: hashes[row.Bundle], ServerTSMS: wire.Command.ServerTSMS,
			Events: []save.EventWrite{}}
		for _, event := range events {
			entry.Events = append(entry.Events, save.EventWrite{Kind: save.EventKind(event.Kind), SchemaVersion: event.SchemaVersion, IntentID: event.IntentID, Payload: bytes.Clone(event.Payload)})
		}
		history.HeadRevision, history.HeadVersion, history.HeadConstants = wire.Command.Revision, row.StateVersion, hashes[row.Bundle]
		if row.Outcome == string(save.IntentApplied) {
			history.HeadRevision++
			applied := history.HeadRevision
			entry.AppliedRevision = &applied
		} else if row.Outcome != string(save.IntentRejected) {
			t.Fatal("unexpected source outcome")
		}
		var discriminator struct {
			Kind string `json:"kind"`
		}
		if err := json.Unmarshal(wire.Resolved, &discriminator); err != nil {
			t.Fatal(err)
		}
		if discriminator.Kind == founderExitResolvedKind || discriminator.Kind == founderExitPlanResolvedKind {
			var facts founderExitResolvedWire
			if err := decodeReplayStrict(wire.Resolved, &facts); err != nil {
				t.Fatal(err)
			}
			entry.Source = &save.FounderLogSource{CompanyStreamID: facts.CompanyStreamID, RunSeq: facts.RunSeq, RunLogSeq: facts.RunLogSeq}
			history.HeadVersion, history.HeadConstants = facts.ResultFounderWireVersion, facts.ResultConstantsHash
		}
		history.Entries = append(history.Entries, entry)
		history.HeadState = []byte(row.PostStateJSON)
	}
	return history
}

func reputationHistoryProfiles(t *testing.T) ([]reputationHistoryProfile, ReplayCatalogSet) {
	t.Helper()
	source, raw := reputationInputShapeSource(t)
	if err := validateReputationHistorySource(source, raw); err != nil {
		t.Fatal(err)
	}
	byName := reputationShapeCatalogs(t, source)
	set, hashes := ReplayCatalogSet{}, map[string]string{}
	for name, bundle := range byName {
		set[bundle.ConstantsHash], hashes[name] = bundle, bundle.ConstantsHash
	}
	profiles, chain := []reputationHistoryProfile{}, []reputationCorpusCase{}
	for _, row := range source.Cases {
		profiles = append(profiles, reputationHistoryProfile{row.Name, reputationHistoryFromRows(t, []reputationCorpusCase{row}, hashes, true)})
		if strings.HasPrefix(row.Name, "chain-") {
			chain = append(chain, row)
		}
	}
	for _, row := range source.ExitCases {
		if row.Founder == nil {
			continue
		}
		// Every paired source here uses one of the two already byte-checked
		// corpus bundles; a future third bundle must be loaded explicitly.
		if _, ok := set[row.Company.ConstantsHash]; !ok {
			t.Fatal("missing paired current bundle")
		}
		if _, ok := set[row.Company.NextConstantsHash]; !ok {
			t.Fatal("missing paired next bundle")
		}
		hashes[row.Founder.Bundle] = row.Company.ConstantsHash
		profiles = append(profiles, reputationHistoryProfile{row.Name, reputationHistoryFromRows(t, []reputationCorpusCase{*row.Founder}, hashes, true)})
	}
	profiles = append(profiles, reputationHistoryProfile{"complete-nine-node-chain", reputationHistoryFromRows(t, chain, hashes, false)})
	if len(profiles) != 24 || len(chain) != 9 {
		t.Fatal("history population changed")
	}
	return profiles, set
}

func cloneReputationHistory(t *testing.T, history save.FounderHistory) save.FounderHistory {
	t.Helper()
	var clone save.FounderHistory
	if err := json.Unmarshal(reputationShapeJSON(t, history), &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}

func TestReputationHistoryVerifiesPinnedCorpus(t *testing.T) {
	profiles, set := reputationHistoryProfiles(t)
	for _, profile := range profiles {
		t.Run(profile.name, func(t *testing.T) {
			if verdict := VerifyFounderHistory(profile.history, set); verdict != ReplayVerified {
				t.Fatalf("honest history verdict=%s", verdict)
			}
		})
	}
}

func TestReputationHistoryRefusesCorruptedEvidence(t *testing.T) {
	profiles, set := reputationHistoryProfiles(t)
	mutations := []struct {
		name    string
		verdict ReplayVerdict
		change  func(*testing.T, *save.FounderHistory)
	}{
		{"head-mirror", ReplayStateDivergence, func(t *testing.T, history *save.FounderHistory) {
			head := reputationShapeObject(t, history.HeadState)
			if string(head["reputation_unlock_ppm"]) == "0" {
				head["reputation_unlock_ppm"] = json.RawMessage("50000")
			} else {
				head["reputation_unlock_ppm"] = json.RawMessage("0")
			}
			history.HeadState = reputationShapeJSON(t, head)
		}},
		{"head-pin", ReplayStateDivergence, func(_ *testing.T, history *save.FounderHistory) {
			history.HeadConstants = "sha256:" + strings.Repeat("0", 64)
		}},
		{"receipt-outcome", ReplayStateDivergence, func(t *testing.T, history *save.FounderHistory) {
			value := reputationShapeObject(t, history.Entries[0].Receipt)
			if string(value["outcome"]) == `"applied"` {
				value["outcome"] = json.RawMessage(`"rejected"`)
			} else {
				value["outcome"] = json.RawMessage(`"applied"`)
			}
			history.Entries[0].Receipt = reputationShapeJSON(t, value)
		}},
		{"extra-event", ReplayStateDivergence, func(_ *testing.T, history *save.FounderHistory) {
			entry := &history.Entries[0]
			entry.Events = append(entry.Events, save.EventWrite{Kind: "reputation_node_purchased.v1", SchemaVersion: 1, IntentID: entry.IntentID,
				Payload: json.RawMessage(`{"node_id":"reputation.unlock.p05","node_kind":"bonus_unlock","cost":1,"reputation_level":1,"reputation_spent_before":0,"reputation_spent_after":1,"unlock_ppm_after":50000,"source":"direct"}`)})
		}},
		{"log-sequence", ReplayLogGap, func(_ *testing.T, history *save.FounderHistory) {
			history.Entries[0].Sequence = 2
		}},
		{"source-presence", ReplayStateDivergence, func(_ *testing.T, history *save.FounderHistory) {
			entry := &history.Entries[0]
			if entry.Source == nil {
				entry.Source = &save.FounderLogSource{CompanyStreamID: "01986666-8e00-7000-8000-000000000001", RunSeq: 2, RunLogSeq: 1}
			} else {
				entry.Source = nil
			}
		}},
	}
	if len(mutations)*len(profiles) != 144 {
		t.Fatal("history negative population changed")
	}
	for _, profile := range profiles {
		for _, mutation := range mutations {
			t.Run(profile.name+"/"+mutation.name, func(t *testing.T) {
				history := cloneReputationHistory(t, profile.history)
				mutation.change(t, &history)
				if verdict := VerifyFounderHistory(history, set); verdict != mutation.verdict {
					t.Fatalf("corrupt history verdict=%s want=%s", verdict, mutation.verdict)
				}
			})
		}
	}
}

func TestReputationHistorySourceRefusesMissingPopulation(t *testing.T) {
	source, raw := reputationInputShapeSource(t)
	for _, mutation := range []string{"source-hash", "missing-case", "missing-chain", "missing-paired-exit"} {
		t.Run(mutation, func(t *testing.T) {
			copy := source
			copy.Cases = slices.Clone(source.Cases)
			copy.ExitCases = slices.Clone(source.ExitCases)
			data := bytes.Clone(raw)
			switch mutation {
			case "source-hash":
				data = append(data, '\n')
			case "missing-case":
				copy.Cases = copy.Cases[1:]
			case "missing-chain":
				index := slices.IndexFunc(copy.Cases, func(row reputationCorpusCase) bool { return strings.HasPrefix(row.Name, "chain-") })
				copy.Cases[index].Name = "not-a-chain-row"
			case "missing-paired-exit":
				index := slices.IndexFunc(copy.ExitCases, func(row reputationExitCase) bool { return row.Founder != nil })
				copy.ExitCases[index].Founder = nil
			}
			if err := validateReputationHistorySource(copy, data); err == nil {
				t.Fatal("corrupt history population admitted")
			}
		})
	}
}
