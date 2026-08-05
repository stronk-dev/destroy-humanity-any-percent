package production

import (
	"encoding/json"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

// VerifyFounderHistory reproduces one immutable Founder stream from genesis to
// its repeatable-read head. Missing artifacts are deterministic evidence gaps;
// it never falls back to deploy-current catalogs.
func VerifyFounderHistory(history save.FounderHistory, catalogs ReplayCatalogResolver) ReplayVerdict {
	if catalogs == nil || history.FounderStreamID == "" || history.FounderID == "" ||
		history.Genesis.FounderStreamID != history.FounderStreamID || len(history.Entries) == 0 {
		return ReplayLogGap
	}
	bundle, ok := catalogs.ResolveReplayCatalogs(history.Genesis.ConstantsHash)
	if !ok || !bundle.valid(history.Genesis.ConstantsHash) {
		return ReplayConstantsMismatch
	}
	state, err := save.RestoreState(history.Genesis.State, history.Genesis.Version, bundle.Economy,
		economy.ScopeFounder, time.Time{})
	if err != nil {
		return ReplayStateDivergence
	}
	currentRevision, currentHash := history.Genesis.Revision, history.Genesis.ConstantsHash
	for index := range history.Entries {
		entry := history.Entries[index]
		if entry.Sequence != int64(index+1) || entry.ConstantsHash != currentHash {
			return ReplayLogGap
		}
		bundle, ok = catalogs.ResolveReplayCatalogs(currentHash)
		if !ok || !bundle.valid(currentHash) {
			return ReplayConstantsMismatch
		}
		wire, err := parseFounderReplayInputs(entry.ReplayInputs)
		if err != nil || wire.Command.IntentID != entry.IntentID ||
			wire.Command.FounderStreamID != history.FounderStreamID || wire.Command.FounderID != history.FounderID ||
			wire.Command.Revision != currentRevision || wire.Command.FounderLogSeq != entry.Sequence ||
			wire.Command.ServerTSMS != entry.ServerTSMS {
			return ReplayLogGap
		}
		var kind struct {
			Kind string `json:"kind"`
		}
		if json.Unmarshal(wire.Resolved, &kind) != nil {
			return ReplayStateDivergence
		}
		linked := kind.Kind == founderExitResolvedKind || kind.Kind == minigameResolutionKind
		if linked != (entry.Source != nil) {
			return ReplayStateDivergence
		}
		if entry.Source != nil && kind.Kind == founderExitResolvedKind {
			var facts founderExitResolvedWire
			if decodeReplayStrict(wire.Resolved, &facts) != nil || facts.CompanyStreamID != entry.Source.CompanyStreamID ||
				facts.RunSeq != entry.Source.RunSeq || facts.RunLogSeq != entry.Source.RunLogSeq {
				return ReplayStateDivergence
			}
			if facts.ResultConstantsHash != currentHash {
				next, nextOK := catalogs.ResolveReplayCatalogs(facts.ResultConstantsHash)
				if !nextOK || !next.valid(facts.ResultConstantsHash) {
					return ReplayConstantsMismatch
				}
				bundle.Next = &next
			}
		}
		transition, err := ApplyFounderLogged(state, entry.CanonicalPayload, bundle, entry.ReplayInputs)
		if err != nil || !jsonSemanticallyEqual(transition.Receipt, entry.Receipt) ||
			!founderEventsEqual(transition.Events, entry.Events) {
			return ReplayStateDivergence
		}
		if transition.Outcome == save.IntentApplied {
			if entry.AppliedRevision == nil || *entry.AppliedRevision != currentRevision+1 {
				return ReplayLogGap
			}
			currentRevision++
		} else if transition.Outcome == save.IntentRejected {
			if entry.AppliedRevision != nil {
				return ReplayLogGap
			}
		} else {
			return ReplayStateDivergence
		}
		currentHash = transition.ResultConstantsHash
	}
	if currentRevision != history.HeadRevision || currentHash != history.HeadConstants ||
		save.VersionForState(state) != history.HeadVersion {
		return ReplayStateDivergence
	}
	encoded, err := save.EncodeState(state)
	if err != nil || !jsonSemanticallyEqual(encoded, history.HeadState) {
		return ReplayStateDivergence
	}
	return ReplayVerified
}

func founderEventsEqual(left, right []save.EventWrite) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].Kind != right[index].Kind || left[index].SchemaVersion != right[index].SchemaVersion ||
			left[index].IntentID != right[index].IntentID || !jsonSemanticallyEqual(left[index].Payload, right[index].Payload) {
			return false
		}
	}
	return true
}
