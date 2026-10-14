package production

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

type externalReplayControl struct {
	Name              string             `json:"name"`
	Omitted           axisSequenceAction `json:"omitted"`
	DuplicatePreState string             `json:"duplicate_pre_state_json"`
	Rejected          axisSequenceAction `json:"rejected"`
}

type externalReplayReport struct {
	Version    int                     `json:"version"`
	Acceptance string                  `json:"acceptance_status"`
	Sources    map[string]string       `json:"source_sha256"`
	Bundle     reputationCorpusBundle  `json:"bundle"`
	Actions    []axisSequenceAction    `json:"actions"`
	Controls   []externalReplayControl `json:"controls"`
}

func TestAxisExternalReplayCompatibilityResearch(t *testing.T) {
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	report := externalReplayReport{Version: 1,
		Acceptance: "NOT_PROVEN: bounded external/Guild replay compatibility; production AC6 remains red",
		Bundle:     reputationCorpusBundle{ConstantsHash: bundle.ConstantsHash, Artifacts: stringArtifacts(bundle.Artifacts)},
		Sources: sourceResearchHashes(t, []string{
			"server/production/axis_origin_projection_test.go", "server/production/axis_external_replay_research_test.go",
			"client/test/axis-external-replay.test.ts", "server/production/replay.go", "client/src/replay.ts",
			"server/save/state.go", "server/production/axis_timing_test.go", "server/production/axis_stack_test.go",
		}), Controls: []externalReplayControl{}}
	report.Actions = append(originExternalReplayActions(t), originGuildReplayActions(t)...)
	if len(report.Actions) != 12 {
		t.Fatal("incomplete primary capture")
	}
	index := 0
	for _, profile := range []string{"commons", "fiscal", "commons_fiscal_guild", "single", "ordered", "nonzero_base"} {
		for _, mode := range []string{"online", "offline"} {
			if report.Actions[index].Name != profile+"/"+mode || report.Actions[index].Outcome != string(save.IntentApplied) {
				t.Fatal("primary capture identity/order/outcome differs from declared population")
			}
			index++
		}
	}
	for index, action := range report.Actions {
		if index < 6 {
			continue
		}
		t.Run("controls/"+action.Name, func(t *testing.T) {
			var wire replayInputsWire
			var resolved replayAccrualResolved
			if err := json.Unmarshal(action.Inputs, &wire); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(wire.Resolved, &resolved); err != nil {
				t.Fatal(err)
			}
			originalAccrual := resolved.Accrual
			resolved.Accrual.GuildSettlementBatch = replayGuildSettlementBatch{Settlements: []replayGuildSettlement{}}
			var err error
			wire.Resolved, err = json.Marshal(resolved)
			if err != nil {
				t.Fatal(err)
			}
			omittedInputs, err := json.Marshal(wire)
			if err != nil {
				t.Fatal(err)
			}
			state := activationRestore(t, bundle, []byte(action.PreStateJSON), 19, start)
			omitted, err := ApplyLogged(state, action.Payload, bundle, omittedInputs)
			ordinary := activationRestore(t, bundle, []byte(action.PostStateJSON), 19, start)
			if err != nil || omitted.Outcome != save.IntentApplied ||
				canonicalFixtureJSON(t, mustEncodeState(t, state)) == action.PostStateJSON ||
				state.Ledger.Snapshot()["company.cash"] != ordinary.Ledger.Snapshot()["company.cash"] {
				t.Fatalf("omission control failed to expose noncash settlement state: %v", err)
			}
			request, err := parseLoggedIntent(action.Payload, wire.Command.IntentID)
			if err != nil {
				t.Fatal(err)
			}
			control := externalReplayControl{Name: action.Name,
				Omitted:           originReplayAction(t, action.Name+"/omitted", []byte(action.PreStateJSON), request, omittedInputs, state, omitted),
				DuplicatePreState: action.PostStateJSON}
			before := mustEncodeState(t, ordinary)
			if _, err := ApplyLogged(ordinary, action.Payload, bundle, action.Inputs); err == nil || !bytes.Equal(before, mustEncodeState(t, ordinary)) {
				t.Fatal("duplicate clearing did not refuse unchanged")
			}
			state = activationRestore(t, bundle, []byte(action.PreStateJSON), 19, start)
			request, err = ParseIntent([]byte(`{"intent_id":"01986666-0901-7000-8000-000000000902","kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":1}}`))
			if err != nil {
				t.Fatal(err)
			}
			inputs := originRequestInputs(t, bundle, state, start.Add(2001*time.Millisecond), wire.EvaluationMode, originalAccrual, request)
			rejected, err := ApplyLogged(state, request.CanonicalPayload, bundle, inputs)
			if err != nil || rejected.Outcome != save.IntentRejected || len(rejected.Events) != 0 ||
				canonicalFixtureJSON(t, mustEncodeState(t, state)) != action.PreStateJSON ||
				!bytes.Equal(rejected.Receipt, marshalRejection(request.IntentID, 1, "unaffordable", "generator.beige_tower")) {
				t.Fatalf("late purchase control did not retain exact rollback: %v", err)
			}
			control.Rejected = originReplayAction(t, action.Name+"/rejected", []byte(action.PreStateJSON), request, inputs, state, rejected)
			report.Controls = append(report.Controls, control)
		})
	}
	if len(report.Controls) != 6 || t.Failed() {
		t.Fatal("incomplete settlement-control population")
	}
	sourceResearchRegressionArtifact(t, "../../testdata/axis-stack/external-replay-research-v1.json", "UPDATE_EXTERNAL_REPLAY_RESEARCH", &report, &report.Sources)
}
