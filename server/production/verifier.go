package production

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

type ReplayVerdict string

const (
	ReplayVerified          ReplayVerdict = "verified"
	ReplayLogGap            ReplayVerdict = "log_gap"
	ReplayStateDivergence   ReplayVerdict = "state_divergence"
	ReplayConstantsMismatch ReplayVerdict = "constants_mismatch"
	ReplayClockViolation    ReplayVerdict = "clock_violation"
	ReplayEngineMismatch    ReplayVerdict = "engine_mismatch"
)

type ReplayLogEntry struct {
	Sequence         int64
	CanonicalPayload []byte
	ReplayInputs     []byte
	ReceiptJSON      []byte
	EventsJSON       []byte
	Terminal         bool
	// NextCatalog belongs to this log entry. A run may contain rejected Exit
	// attempts on either side of an epoch change, so a run-wide mutable Next
	// slot cannot reproduce the catalog resolved for each attempt.
	NextCatalog *CatalogBundle
}

// VerifyReplayRun replays one immutable Company run through the same owned
// transition boundary used by live commands. Receipt and ordered event bytes
// are canonical JSON captured from persistence, never an alternate oracle.
func VerifyReplayRun(genesis []byte, version int, catalogs CatalogBundle, entries []ReplayLogEntry, constantsHash string, engineMismatch bool) ReplayVerdict {
	if engineMismatch {
		return ReplayEngineMismatch
	}
	if constantsHash != catalogs.ConstantsHash || !catalogs.valid(constantsHash) {
		return ReplayConstantsMismatch
	}
	state, err := save.RestoreState(genesis, version, catalogs.Economy, economy.ScopeCompany, time.Time{})
	if err != nil {
		return ReplayConstantsMismatch
	}
	terminal := false
	for index, entry := range entries {
		if entry.Sequence != int64(index+1) || terminal {
			return ReplayLogGap
		}
		wire, wireErr := parseReplayInputs(entry.ReplayInputs)
		if wireErr != nil {
			return ReplayStateDivergence
		}
		if wire.Command.RunLogSeq != entry.Sequence {
			return ReplayLogGap
		}
		var discriminator struct {
			Kind string `json:"kind"`
		}
		if json.Unmarshal(wire.Resolved, &discriminator) != nil {
			return ReplayStateDivergence
		}
		if discriminator.Kind == "exit" {
			entryCatalogs := catalogs
			entryCatalogs.Next = entry.NextCatalog
			transition, transitionErr := ApplyLoggedExit(state, entry.CanonicalPayload, entryCatalogs, entry.ReplayInputs)
			if transitionErr != nil {
				return replayErrorVerdict(transitionErr)
			}
			events := append([]save.EventWrite(nil), transition.Decision.FounderEvents...)
			events = append(events, transition.Decision.CompanyEndedEvents...)
			events = append(events, transition.Decision.CompanyStartedEvents...)
			if !canonicalJSONEqual(transition.Decision.Receipt, entry.ReceiptJSON) || !canonicalJSONEqual(marshalReplayEvents(events), entry.EventsJSON) {
				return ReplayStateDivergence
			}
			state = transition.Company
			terminal = transition.Decision.Outcome == save.IntentApplied
			continue
		}
		transition, transitionErr := ApplyLogged(state, entry.CanonicalPayload, catalogs, entry.ReplayInputs)
		if transitionErr != nil {
			return replayErrorVerdict(transitionErr)
		}
		if !canonicalJSONEqual(transition.Receipt, entry.ReceiptJSON) || !canonicalJSONEqual(marshalReplayEvents(transition.Events), entry.EventsJSON) {
			return ReplayStateDivergence
		}
		state = transition.State
	}
	if !terminal {
		return ReplayLogGap
	}
	return ReplayVerified
}

func replayErrorVerdict(err error) ReplayVerdict {
	if errors.Is(err, ErrReplayClockViolation) {
		return ReplayClockViolation
	}
	return ReplayStateDivergence
}

func marshalReplayEvents(events []save.EventWrite) []byte {
	values := make([]map[string]any, len(events))
	for index, event := range events {
		var payload any
		decoder := json.NewDecoder(bytes.NewReader(event.Payload))
		decoder.UseNumber()
		if decoder.Decode(&payload) != nil {
			return nil
		}
		values[index] = map[string]any{"kind": event.Kind, "schema_version": event.SchemaVersion, "intent_id": event.IntentID, "payload": payload}
	}
	encoded, _ := json.Marshal(values)
	return encoded
}

func canonicalJSONEqual(left, right []byte) bool {
	canonical := func(source []byte) ([]byte, bool) {
		decoder := json.NewDecoder(bytes.NewReader(source))
		decoder.UseNumber()
		var value any
		if decoder.Decode(&value) != nil {
			return nil, false
		}
		if decoder.Decode(&struct{}{}) != io.EOF {
			return nil, false
		}
		encoded, err := json.Marshal(value)
		return encoded, err == nil
	}
	leftCanonical, leftOK := canonical(left)
	rightCanonical, rightOK := canonical(right)
	return leftOK && rightOK && bytes.Equal(leftCanonical, rightCanonical)
}
