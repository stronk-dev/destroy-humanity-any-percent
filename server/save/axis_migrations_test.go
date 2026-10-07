package save_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

type axisMigrationCase struct {
	Name            string                     `json:"name"`
	SourceCase      string                     `json:"source_case"`
	SourceState     string                     `json:"source_state"`
	Operation       string                     `json:"operation"`
	FromVersion     int                        `json:"from_version"`
	ExpectedVersion int                        `json:"expected_version"`
	InputPatch      map[string]json.RawMessage `json:"input_patch"`
	ErrorStage      string                     `json:"error_stage"`
}

type axisMigrationAction struct {
	PreState      json.RawMessage `json:"pre_state"`
	PostState     json.RawMessage `json:"post_state"`
	Payload       json.RawMessage `json:"canonical_payload"`
	Inputs        json.RawMessage `json:"replay_inputs"`
	ReceiptJSON   string          `json:"receipt_json"`
	EventsJSON    string          `json:"events_json"`
	PostStateJSON string          `json:"post_state_json"`
}

type axisMigrationRow struct {
	ID   string              `json:"id"`
	Old  axisMigrationAction `json:"old"`
	Exit struct {
		PreState          json.RawMessage `json:"pre_state"`
		Payload           json.RawMessage `json:"canonical_payload"`
		Inputs            json.RawMessage `json:"replay_inputs"`
		ReceiptJSON       string          `json:"receipt_json"`
		FounderJSON       string          `json:"founder_output_json"`
		FinalJSON         string          `json:"final_company_json"`
		NewJSON           string          `json:"new_company_json"`
		FounderEventsJSON string          `json:"founder_events_json"`
		EndedEventsJSON   string          `json:"company_ended_events_json"`
		StartedEventsJSON string          `json:"company_started_events_json"`
	} `json:"exit"`
	Next []axisMigrationAction `json:"next"`
}

func axisMigrationEqual(t *testing.T, got any, expected string) {
	t.Helper()
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(canonicalFounderMigrationJSON(t, encoded), []byte(expected)) {
		t.Fatal("complete migration output differs")
	}
}

func axisMigrationEncode(t *testing.T, state *save.State) json.RawMessage {
	t.Helper()
	raw, err := save.EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func axisMigrationRestore(t *testing.T, raw json.RawMessage, version int, bundle production.CatalogBundle) *save.State {
	t.Helper()
	state, err := save.RestoreState(raw, version, bundle.Economy, economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if err := bundle.ValidateFoundationState(state); err != nil {
		t.Fatal(err)
	}
	axisMigrationEqual(t, axisMigrationEncode(t, state), string(canonicalFounderMigrationJSON(t, raw)))
	if save.VersionForState(state) != version {
		t.Fatal("loading activated a new run")
	}
	return state
}

func axisMigrationEvents(events []save.EventWrite) any {
	result := make([]map[string]any, len(events))
	for i, event := range events {
		result[i] = map[string]any{"kind": string(event.Kind), "schema_version": event.SchemaVersion, "intent_id": event.IntentID, "payload": event.Payload}
	}
	return result
}

// Mechanical projection of the PUBLIC replayed Founder fields into the source's
// carry-output envelope. Metadata comes from actual replay inputs, not expected
// output. No new production export or transition is introduced for this reader.
func axisMigrationFounderOutput(t *testing.T, founder *save.State, inputs json.RawMessage) any {
	t.Helper()
	if save.VersionForState(founder) != 21 {
		t.Fatal("unexpected replayed Founder floor")
	}
	var input struct {
		Resolved struct {
			Carry map[string]json.RawMessage `json:"founder_carry"`
		} `json:"resolved"`
	}
	if err := json.Unmarshal(inputs, &input); err != nil {
		t.Fatal(err)
	}
	sortedSet := func(values map[string]bool) []string {
		keys := make([]string, 0, len(values))
		for key, present := range values {
			if present {
				keys = append(keys, key)
			}
		}
		sort.Strings(keys)
		return keys
	}
	slots := append([]save.NetworkSlot{}, founder.NetworkSlots...)
	sort.Slice(slots, func(i, j int) bool { return slots[i].Slot < slots[j].Slot })
	result := map[string]any{
		"exit_history_count": len(founder.ExitHistory), "reputation_level": founder.ReputationLevel,
		"route_knowledge_balance": founder.RouteKnowledgeBalance, "age_ms": founder.AgeMS,
		"notoriety": founder.Notoriety, "advisor_mode": founder.AdvisorMode, "network_slots": slots,
		"ledger_fact_kinds":            sortedSet(founder.LedgerFactKinds),
		"achievements_earned_lifetime": sortedSet(founder.AchievementsEarnedLifetime),
		"achievement_score_lifetime":   founder.AchievementScoreLifetime,
	}
	for _, key := range []string{"founder_revision", "founder_constants_hash"} {
		if input.Resolved.Carry[key] == nil {
			t.Fatal("missing source carry identity")
		}
		result[key] = input.Resolved.Carry[key]
	}
	result["founder_extensions"] = map[string]any{
		"minigame_ratings": founder.MinigameRatings, "minigame_offline_quality": founder.MinigameOfflineQuality,
		"pets": founder.Pets, "fiscal_credit": founder.FiscalCredit,
		"fiscal_period_opened_wall_ms": founder.FiscalPeriodOpenedWallMS, "fiscal_period_seq": founder.FiscalPeriodSequence,
		"fiscal_generator_levels": founder.FiscalGeneratorLevels, "fiscal_unlocks": sortedSet(founder.FiscalUnlocks),
		"soul": founder.Soul, "soul_exhausted_source_ids": append([]string{}, founder.SoulExhaustedSourceIDs...),
		"minigame_session_seq": founder.MinigameSessionSeq,
	}
	return result
}

func axisMigrationStep(t *testing.T, state *save.State, action axisMigrationAction, bundle production.CatalogBundle, version int) *save.State {
	t.Helper()
	axisMigrationEqual(t, axisMigrationEncode(t, state), string(canonicalFounderMigrationJSON(t, action.PreState)))
	result, err := production.ApplyLogged(state, canonicalFounderMigrationJSON(t, action.Payload), bundle, action.Inputs)
	if err != nil || result.Outcome != save.IntentApplied {
		t.Fatalf("actual migration action: %v %s", err, result.Outcome)
	}
	axisMigrationEqual(t, result.Receipt, action.ReceiptJSON)
	axisMigrationEqual(t, axisMigrationEvents(result.Events), action.EventsJSON)
	axisMigrationEqual(t, axisMigrationEncode(t, result.State), action.PostStateJSON)
	return axisMigrationRestore(t, axisMigrationEncode(t, result.State), version, bundle)
}

func TestCompanyAxisMigrationCorpus(t *testing.T) {
	data, err := os.ReadFile("../../testdata/save-migrations.json")
	if err != nil {
		t.Fatal(err)
	}
	var table struct {
		Version       int               `json:"corpus_version"`
		Legacy        []json.RawMessage `json:"cases"`
		FounderSource json.RawMessage   `json:"founder_source"`
		Founder       []json.RawMessage `json:"founder_cases"`
		Source        struct {
			Path string `json:"path"`
			SHA  string `json:"sha256"`
		} `json:"company_source"`
		Cases []axisMigrationCase `json:"company_cases"`
	}
	strictFounderMigrationJSON(t, data, &table)
	baselineBytes, err := os.ReadFile("../../testdata/save-migrations-baseline.json")
	if err != nil {
		t.Fatal(err)
	}
	var baseline struct {
		Version int      `json:"schema_version"`
		Count   int      `json:"minimum_case_count"`
		Names   []string `json:"required_case_names"`
	}
	strictFounderMigrationJSON(t, baselineBytes, &baseline)
	if table.Version != 10 || len(table.Legacy) != 11 || len(table.Founder) != 4 || len(table.Cases) != 5 || baseline.Version != 1 || baseline.Count != 20 || len(baseline.Names) != 20 {
		t.Fatal("incomplete Company migration population/ratchet")
	}
	names := []string{}
	for _, raw := range append(table.Legacy, table.Founder...) {
		var row struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(raw, &row); err != nil {
			t.Fatal(err)
		}
		names = append(names, row.Name)
	}
	expectedNames := []string{"company-v18-to-v19-new-run", "company-v19-derived-mismatch-rejected", "company-v19-field-before-version-rejected", "company-v19-attained-not-superset-of-earned-rejected", "pre-activation-run-replays-through-exit"}
	expectedStates := []string{"exit_pre", "post_purchase", "old_post", "post_purchase", "old_pre"}
	expectedOperations := []string{"new_run_activation", "decode", "decode", "decode", "replay_exit"}
	expectedStages := []string{"none", "pinned_derived", "structural", "pinned_superset", "none"}
	fromVersions, toVersions := []int{18, 19, 18, 19, 18}, []int{19, 19, 18, 19, 19}
	for i, row := range table.Cases {
		if row.Name != expectedNames[i] || row.SourceCase != "veteran/online/3114" || row.InputPatch == nil || row.SourceState != expectedStates[i] || row.Operation != expectedOperations[i] || row.ErrorStage != expectedStages[i] || row.FromVersion != fromVersions[i] || row.ExpectedVersion != toVersions[i] {
			t.Fatal("wrong Company migration name/source/order")
		}
		names = append(names, row.Name)
	}
	sort.Strings(names)
	sort.Strings(baseline.Names)
	if !reflect.DeepEqual(names, baseline.Names) {
		t.Fatal("complete migration names disagree with ratchet")
	}
	for i, name := range names {
		if name == "" || i > 0 && name == names[i-1] {
			t.Fatal("duplicate migration name")
		}
	}
	if table.Source.Path != "axis-stack/activation-research-v1.json" {
		t.Fatal("wrong Company migration source path")
	}
	sourceBytes, err := os.ReadFile("../../testdata/" + table.Source.Path)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(sourceBytes)
	if hex.EncodeToString(digest[:]) != table.Source.SHA {
		t.Fatal("Company migration source SHA mismatch")
	}
	var source struct {
		Current struct {
			Hash      string            `json:"constants_hash"`
			Artifacts map[string]string `json:"artifacts"`
		} `json:"current"`
		Next struct {
			Hash      string            `json:"constants_hash"`
			Artifacts map[string]string `json:"artifacts"`
		} `json:"next"`
		Rows []axisMigrationRow `json:"rows"`
	}
	if err := json.Unmarshal(sourceBytes, &source); err != nil {
		t.Fatal(err)
	}
	selected := []axisMigrationRow{}
	for _, row := range source.Rows {
		if row.ID == "veteran/online/3114" {
			selected = append(selected, row)
		}
	}
	if len(selected) != 1 || len(selected[0].Next) != 2 {
		t.Fatal("missing/ambiguous complete Company migration source")
	}
	row := selected[0]
	current := loadFounderMigrationBundle(t, source.Current.Hash, source.Current.Artifacts)
	next := loadFounderMigrationBundle(t, source.Next.Hash, source.Next.Artifacts)
	current.Next = &next
	for _, entry := range table.Cases {
		t.Run(entry.Name, func(t *testing.T) {
			var raw json.RawMessage
			bundle := current
			switch entry.SourceState {
			case "exit_pre":
				raw = row.Exit.PreState
			case "old_pre":
				raw = row.Old.PreState
			case "old_post":
				raw = row.Old.PostState
			case "post_purchase":
				raw = row.Next[1].PostState
				bundle = next
			default:
				t.Fatal("unknown migration state")
			}
			state := axisMigrationRestore(t, raw, entry.FromVersion, bundle)
			if entry.Operation == "decode" {
				axisMigrationNegative(t, entry, raw, bundle)
				return
			}
			if entry.FromVersion != 18 || entry.ExpectedVersion != 19 || entry.ErrorStage != "none" || len(entry.InputPatch) != 0 || state.AchievementsAttainedRun != nil || state.AttainmentScoreRun != 0 {
				t.Fatal("invalid activation boundary")
			}
			if entry.Operation == "replay_exit" && entry.SourceState == "old_pre" {
				state = axisMigrationStep(t, state, row.Old, current, 18)
			} else if entry.Operation != "new_run_activation" || entry.SourceState != "exit_pre" {
				t.Fatal("unknown migration operation")
			}
			axisMigrationEqual(t, axisMigrationEncode(t, state), string(canonicalFounderMigrationJSON(t, row.Exit.PreState)))
			exit, err := production.ApplyLoggedExit(state, canonicalFounderMigrationJSON(t, row.Exit.Payload), current, row.Exit.Inputs)
			if err != nil || exit.Decision.Outcome != save.IntentApplied {
				t.Fatalf("actual activation Exit: %v", err)
			}
			axisMigrationEqual(t, exit.Decision.Receipt, row.Exit.ReceiptJSON)
			axisMigrationEqual(t, axisMigrationFounderOutput(t, exit.Founder, row.Exit.Inputs), row.Exit.FounderJSON)
			axisMigrationEqual(t, axisMigrationEncode(t, exit.Company), row.Exit.FinalJSON)
			axisMigrationEqual(t, axisMigrationEncode(t, exit.Decision.NewCompanyState), row.Exit.NewJSON)
			axisMigrationEqual(t, axisMigrationEvents(exit.Decision.FounderEvents), row.Exit.FounderEventsJSON)
			axisMigrationEqual(t, axisMigrationEvents(exit.Decision.CompanyEndedEvents), row.Exit.EndedEventsJSON)
			axisMigrationEqual(t, axisMigrationEvents(exit.Decision.CompanyStartedEvents), row.Exit.StartedEventsJSON)
			terminal := axisMigrationRestore(t, axisMigrationEncode(t, exit.Company), 18, current)
			if terminal.AchievementsAttainedRun != nil || terminal.AttainmentScoreRun != 0 {
				t.Fatal("old terminal retrofitted attainment")
			}
			state = axisMigrationRestore(t, axisMigrationEncode(t, exit.Decision.NewCompanyState), 19, next)
			if state.RunSeq != 3 || len(state.AchievementsAttainedRun) != 0 || state.AttainmentScoreRun != 0 || state.AchievementScoreRun != 0 || len(state.AchievementsEarnedRun) != 0 {
				t.Fatal("new run did not reset")
			}
			if entry.Operation == "new_run_activation" {
				for _, action := range row.Next {
					state = axisMigrationStep(t, state, action, next, 19)
				}
				if state.AttainmentScoreRun != 2 || len(state.AchievementsAttainedRun) != 1 || !state.AchievementsAttainedRun["achievement.generators_purchased_1"] {
					t.Fatal("new purchase did not re-attain")
				}
			}
		})
	}
}

func axisMigrationNegative(t *testing.T, entry axisMigrationCase, raw json.RawMessage, bundle production.CatalogBundle) {
	t.Helper()
	var patch, corrected map[string]json.RawMessage
	if err := json.Unmarshal(raw, &patch); err != nil {
		t.Fatal(err)
	}
	allowed := map[string]string{
		"pinned_derived":  `{"attainment_score_run":3}`,
		"structural":      `{"achievements_attained_run":[],"attainment_score_run":0}`,
		"pinned_superset": `{"achievements_attained_run":[],"attainment_score_run":0,"achievements_earned_run":["achievement.generators_purchased_1"],"achievement_score_run":2}`,
	}
	want, ok := allowed[entry.ErrorStage]
	if !ok || entry.FromVersion != entry.ExpectedVersion || entry.ErrorStage == "structural" && (entry.FromVersion != 18 || entry.SourceState != "old_post") || entry.ErrorStage != "structural" && (entry.FromVersion != 19 || entry.SourceState != "post_purchase") {
		t.Fatal("unknown negative boundary")
	}
	axisMigrationEqual(t, entry.InputPatch, string(canonicalFounderMigrationJSON(t, []byte(want))))
	for key, value := range entry.InputPatch {
		patch[key] = value
	}
	encoded, err := json.Marshal(patch)
	if err != nil {
		t.Fatal(err)
	}
	before := bytes.Clone(encoded)
	state, err := save.RestoreState(encoded, entry.FromVersion, bundle.Economy, economy.ScopeCompany, time.Time{})
	if entry.ErrorStage == "structural" {
		if !errors.Is(err, save.ErrInvalidState) {
			t.Fatalf("early fields admitted: %v", err)
		}
	} else {
		if err != nil {
			t.Fatalf("pinned negative failed structural stage: %v", err)
		}
		beforeState := axisMigrationEncode(t, state)
		err = bundle.ValidateFoundationState(state)
		message := "attainment score does not derive"
		if entry.ErrorStage == "pinned_superset" {
			message = "is not attained"
		}
		if !errors.Is(err, production.ErrInvalidEngineState) || !strings.Contains(err.Error(), message) {
			t.Fatalf("pinned %s admitted/wrong stage: %v", entry.ErrorStage, err)
		}
		if !bytes.Equal(beforeState, axisMigrationEncode(t, state)) {
			t.Fatal("negative validation mutated state")
		}
	}
	if !bytes.Equal(before, encoded) {
		t.Fatal("negative restore mutated raw input")
	}
	if err := json.Unmarshal(encoded, &corrected); err != nil {
		t.Fatal(err)
	}
	switch entry.ErrorStage {
	case "structural":
		delete(corrected, "achievements_attained_run")
		delete(corrected, "attainment_score_run")
	case "pinned_derived":
		corrected["attainment_score_run"] = json.RawMessage(`2`)
	case "pinned_superset":
		corrected["achievements_earned_run"] = json.RawMessage(`[]`)
		corrected["achievement_score_run"] = json.RawMessage(`0`)
	}
	fixed, err := json.Marshal(corrected)
	if err != nil {
		t.Fatal(err)
	}
	axisMigrationRestore(t, fixed, entry.FromVersion, bundle)
}
