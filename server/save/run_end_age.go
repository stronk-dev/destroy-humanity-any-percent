package save

import (
	"encoding/json"
	"fmt"
)

// run_ended v4 adds a required frozen Founder age. Legacy v2/v3 retain their
// existing grammar. Faction is the only nullable member of the new top-level arm.
func validRunEndedAgeEventFields(data []byte, branched bool) bool {
	keys := []string{"founder_id", "run_id", "exit_type", "started_at_ms", "ended_at_ms", "rta_ms", "attended_ms", "pre_timer", "terminal_seq", "payout", "tier", "lifetime_value", "ledger_fact_kinds", "executed_routes", "gates_crossed", "generators_purchased_total", "assisted", "faction", "founder_age_ms"}
	if branched {
		keys = append(keys, "branch", "starter_package")
	}
	return validRequiredEventFields(data, map[string]bool{"faction": true}, keys...)
}

// The value displayed from the event must be the Founder age this transaction
// actually commits, not a separately computed estimate or a later live read.
func validateExitFounderAge(decision ExitDecision, founder *State) error {
	if decision.Outcome != IntentApplied {
		return nil
	}
	for _, event := range decision.CompanyEndedEvents {
		if event.Kind != EventRunEnded || event.SchemaVersion != 4 {
			continue
		}
		var payload eventRunEnded
		if founder == nil || json.Unmarshal(event.Payload, &payload) != nil || payload.FounderAgeMS == nil || *payload.FounderAgeMS != founder.AgeMS {
			return fmt.Errorf("%w: run_ended Founder age does not match committed state", ErrInvalidStream)
		}
	}
	return nil
}
