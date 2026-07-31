package production

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

func TestReplayInputsAreClosedCanonicalInputs(t *testing.T) {
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	command := save.ReplayCommand{IntentID: "01985555-7100-7000-8000-000000000001", CompanyStreamID: "01985555-7100-4000-8000-000000000002", FounderID: "01985555-7100-4000-8000-000000000003", Revision: 7, RunSeq: 2, RunLogSeq: 4}
	weight := int64(812_345)
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOffline, Now: now, IntentKind: IntentBuyGenerator,
		CommonsWeightPPM: &weight, RouteContextVersion: 3, Contributions: []multiplier.Contribution{
			{Slot: multiplier.SlotPrestige, SourceID: "source.z", Target: "all", Factor: decimal.New(13, -1)},
			{Slot: multiplier.SlotFaction, SourceID: "source.a", Target: "generator.one", Factor: decimal.One},
		}})
	if err != nil {
		t.Fatal(err)
	}
	wire, err := parseReplayInputs(inputs)
	if err != nil || wire.Command != command || wire.EvaluatedAtMS != now.UnixMilli() || wire.EvaluationMode != ModeOffline {
		t.Fatalf("wire=%+v err=%v", wire, err)
	}
	var resolved replayAccrualResolved
	if err := decodeReplayStrict(wire.Resolved, &resolved); err != nil || len(resolved.Accrual.Contributions) != 2 ||
		resolved.Accrual.Contributions[0].SourceID != "source.a" || resolved.Accrual.Contributions[1].SourceID != "source.z" ||
		resolved.Accrual.GuildSettlementBatch == nil || resolved.Accrual.CommonsWeightPPM == nil || *resolved.Accrual.CommonsWeightPPM != weight {
		t.Fatalf("resolved=%+v err=%v", resolved, err)
	}

	var object map[string]any
	if err := json.Unmarshal(inputs, &object); err != nil {
		t.Fatal(err)
	}
	object["future_field"] = true
	tampered, _ := json.Marshal(object)
	if _, err := parseReplayInputs(tampered); err == nil {
		t.Fatal("unknown replay-input field accepted")
	}
	if strings.Contains(string(inputs), `"factor":1.3`) {
		t.Fatal("factor was encoded as a binary JSON number instead of canonical Decimal string")
	}
}

func TestTerminalReplayInputsFreezeFounderCarry(t *testing.T) {
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	command := save.ReplayCommand{IntentID: "01985555-7200-7000-8000-000000000001", CompanyStreamID: "01985555-7200-4000-8000-000000000002", FounderID: "01985555-7200-4000-8000-000000000003", Revision: 9, RunSeq: 3, RunLogSeq: 12}
	founder := &save.State{ReputationLevel: 17, RouteKnowledgeBalance: 4, AgeMS: 50, Notoriety: 6, AdvisorMode: true,
		LedgerFactKinds: map[string]bool{"fact.z": true, "fact.a": true}, NetworkSlots: []save.NetworkSlot{{Slot: "slot.z", CarriedRef: "ref.z"}}}
	carry := founderCarry(founder)
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: IntentWindDown,
		FounderCarry: &carry, Terminal: true, ExecutedRouteIDs: []string{"route.z", "route.a"}, SelectedExitType: "collapse",
		SelectedTerms: json.RawMessage(`{}`), NextConstantsHash: "sha256:" + strings.Repeat("a", 64)})
	if err != nil {
		t.Fatal(err)
	}
	founder.ReputationLevel = 99
	wire, _ := parseReplayInputs(inputs)
	var terminal replayExitResolved
	if err := decodeReplayStrict(wire.Resolved, &terminal); err != nil || terminal.FounderCarry.ReputationLevel != 17 ||
		terminal.ExecutedRouteIDs[0] != "route.a" || terminal.ExecutedRouteIDs[1] != "route.z" {
		t.Fatalf("terminal=%+v err=%v", terminal, err)
	}
}
