package production

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/garden"
	"cloud-clicker/server/save"
)

// Server Garden SG5 (Founder scope) plus the SG3 advance pre-step (SG-P3).

func isGardenIntent(kind string) bool {
	return kind == IntentGardenPlant || kind == IntentGardenUproot || kind == IntentGardenSetSubstrate || kind == IntentGardenHarvest
}

// isGardenAdvanceTrigger is SG3's trigger set (OD-20): every garden command
// and spend_fiscal_credit, which can raise the host level or buy the unlock.
// AC6 proves the set sufficient; any future transition that mutates a garden
// input must join it.
func isGardenAdvanceTrigger(kind string) bool {
	return isGardenIntent(kind) || kind == gardenHarvestCreditedKind || kind == IntentSpendFiscalCredit
}

// founderGardenResolved is the `garden_command.v1` resolved arm: the server
// wall timestamp, the hidden salt when this advance draws it (OD-12), and the
// advance summary replay recomputes and must reproduce.
type founderGardenResolved struct {
	Kind          string         `json:"kind"`
	ServerMS      int64          `json:"server_ms"`
	GardenSaltHex string         `json:"garden_salt_hex,omitempty"`
	Advance       garden.Advance `json:"advance"`
}

type gardenEvent struct {
	Kind    save.EventKind `json:"kind"`
	Payload any            `json:"payload"`
}

type founderGardenReceipt struct {
	IntentID        string      `json:"intent_id"`
	Outcome         string      `json:"outcome"`
	FounderRevision int64       `json:"founder_revision"`
	Kind            string      `json:"kind"`
	Event           gardenEvent `json:"event"`
}

// gardenInputs resolves SG3's non-state inputs from replay-owned Founder state.
func gardenInputs(catalog *garden.Catalog, state *save.State) (hostLevel int64, unlocked bool) {
	return state.FiscalGeneratorLevels[catalog.HostGeneratorID], state.FiscalUnlocks[catalog.UnlockID]
}

func gardenActive(catalogs CatalogBundle, state *save.State) bool {
	return catalogs.Garden != nil && save.VersionForState(state) >= 25 && state.ServerGarden != nil
}

// gardenSaltDrawer is the live salt source; tests pin it for determinism.
var gardenSaltDrawer = drawGardenSalt

// liveGardenResolved builds the `garden_command.v1` resolved arm the live path
// freezes: the advance summary is computed on a clone after the same Fiscal
// sweep ApplyFounderLogged performs, which recomputes and must agree.
func liveGardenResolved(bundle CatalogBundle, state *save.State, kind string, serverMS int64) (founderGardenResolved, error) {
	salt, err := liveGardenSalt(bundle, state)
	if err != nil {
		return founderGardenResolved{}, err
	}
	live := founderGardenResolved{Kind: kind, ServerMS: serverMS, GardenSaltHex: salt,
		Advance: garden.Advance{Matured: []garden.PlotRef{}, Spawned: []garden.Spawn{}}}
	if !gardenActive(bundle, state) {
		return live, nil
	}
	probe, err := cloneFounderReplayState(state, bundle.Economy)
	if err != nil {
		return founderGardenResolved{}, err
	}
	if bundle.Fiscal != nil {
		fiscalState := fiscalStateFromSave(probe)
		if _, err := bundle.Fiscal.Sweep(&fiscalState, serverMS); err != nil {
			return founderGardenResolved{}, err
		}
		fiscalStateToSave(probe, fiscalState)
	}
	advance, err := preAdvanceGarden(probe, bundle, serverMS, salt)
	if err != nil {
		return founderGardenResolved{}, err
	}
	live.Advance = *advance
	return live, nil
}

// drawGardenSalt is the live server draw of the hidden per-Founder salt. It is
// frozen into resolved inputs, so replay never draws.
func drawGardenSalt() (string, error) {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes[:]), nil
}

// liveGardenSalt returns a fresh salt exactly when this command's advance
// would initialize it, else "".
func liveGardenSalt(catalogs CatalogBundle, state *save.State) (string, error) {
	if !gardenActive(catalogs, state) {
		return "", nil
	}
	_, unlocked := gardenInputs(catalogs.Garden, state)
	if !garden.NeedsSalt(state.ServerGarden, unlocked) {
		return "", nil
	}
	return gardenSaltDrawer()
}

// resolvedGardenSalt reads the optional salt field every trigger arm carries.
func resolvedGardenSalt(resolved json.RawMessage) string {
	var field struct {
		GardenSaltHex string `json:"garden_salt_hex"`
	}
	_ = json.Unmarshal(resolved, &field)
	return field.GardenSaltHex
}

// preAdvanceGarden is the SG-P3 pre-step: after the Fiscal sweep and before
// the command body. A rejected command rolls it back with everything else.
func preAdvanceGarden(state *save.State, catalogs CatalogBundle, serverMS int64, salt string) (*garden.Advance, error) {
	if !gardenActive(catalogs, state) {
		if salt != "" {
			return nil, fmt.Errorf("%w: garden salt without an active garden", ErrInvalidReplayInputs)
		}
		return nil, nil
	}
	hostLevel, unlocked := gardenInputs(catalogs.Garden, state)
	advance, err := catalogs.Garden.Advance(state.ServerGarden, garden.AdvanceInput{ServerMS: serverMS, HostLevel: hostLevel, Unlocked: unlocked, Salt: salt})
	if err != nil {
		return nil, fmt.Errorf("%w: garden advance: %v", ErrInvalidReplayInputs, err)
	}
	return &advance, nil
}

// decorateGardenAdvance adds `garden_advance` to the applied receipt and
// prepends `garden_advanced.v1` when SG8 emits it.
func decorateGardenAdvance(result *FounderLoggedTransition, intentID string, advance *garden.Advance) error {
	var receipt map[string]json.RawMessage
	if err := json.Unmarshal(result.Receipt, &receipt); err != nil || receipt == nil {
		return fmt.Errorf("%w: garden receipt", ErrInvalidReplayInputs)
	}
	encoded, err := json.Marshal(advance)
	if err != nil {
		return err
	}
	receipt["garden_advance"] = encoded
	if result.Receipt, err = json.Marshal(receipt); err != nil {
		return err
	}
	if advance.Visible() {
		result.Events = append([]save.EventWrite{{Kind: save.EventGardenAdvanced, SchemaVersion: 1, IntentID: intentID, Payload: encoded}}, result.Events...)
	}
	return nil
}

// checkGardenTransition is SG2/AC11's carry law: only the advance trigger set
// may change server_garden; every other Founder transition, Exit included,
// carries it byte-identically (activation from absent excepted).
func checkGardenTransition(before, after *garden.State, trigger bool) error {
	if before == nil {
		if after != nil && (after.SaltHex != nil || after.TickAnchorWallMS != nil || after.TickSeq != 0 || len(after.Plots) != 0) {
			return fmt.Errorf("%w: server_garden appeared outside activation", ErrInvalidEngineState)
		}
		return nil
	}
	if !trigger && !before.Equal(after) {
		return fmt.Errorf("%w: a non-garden Founder transition changed server_garden", ErrInvalidEngineState)
	}
	return nil
}

func (s *Service) handleFounderGarden(ctx context.Context, companyStreamID string, now time.Time, request IntentRequest) (HandleResult, error) {
	if s.replayCatalogs == nil {
		return HandleResult{}, fmt.Errorf("%w: Founder replay runtime unavailable", ErrInvalidIntent)
	}
	if request.Kind == IntentGardenHarvest && request.InvalidDetail == "" {
		return s.handleGardenHarvest(ctx, companyStreamID, request, now)
	}
	founder, err := s.store.LoadSiblingLatest(ctx, companyStreamID, economy.ScopeFounder)
	if err != nil {
		return HandleResult{}, err
	}
	result, err := s.store.ApplyFounderLogged(ctx, founder.Revision.StreamID, request.ExpectedRevision,
		request.IntentID, request.RequestHash, request.CanonicalPayload,
		func(state *save.State, revision save.Revision, command save.FounderReplayCommand) (save.IntentDecision, json.RawMessage, error) {
			bundle, ok := s.replayCatalogs.ResolveReplayCatalogs(revision.ConstantsHash)
			if !ok {
				return save.IntentDecision{}, nil, fmt.Errorf("%w: replay catalog bundle unavailable", ErrInvalidIntent)
			}
			var resolved any = founderInvalidResolved{Kind: "invalid", Detail: request.InvalidDetail}
			if request.InvalidDetail == "" {
				live, liveErr := liveGardenResolved(bundle, state, request.Kind, command.ServerTSMS)
				if liveErr != nil {
					return save.IntentDecision{}, nil, liveErr
				}
				resolved = live
			}
			replayInputs, marshalErr := marshalFounderReplayInputs(command, resolved)
			if marshalErr != nil {
				return save.IntentDecision{}, nil, marshalErr
			}
			transition, transitionErr := ApplyFounderLogged(state, request.CanonicalPayload, bundle, replayInputs)
			return save.IntentDecision{Outcome: transition.Outcome, Receipt: transition.Receipt,
				Events: transition.Events}, replayInputs, transitionErr
		})
	if err != nil {
		var conflict *save.RevisionConflict
		switch {
		case errors.As(err, &conflict):
			return HandleResult{Receipt: marshalRejection(request.IntentID, conflict.Current, "revision_conflict", "expected_revision")}, nil
		case errors.Is(err, save.ErrIdempotencyConflict):
			current := request.ExpectedRevision
			if loaded, loadErr := s.store.LoadLatest(ctx, founder.Revision.StreamID); loadErr == nil {
				current = loaded.Revision.Number
			}
			return HandleResult{Receipt: marshalRejection(request.IntentID, current, "idempotency_conflict", request.IntentID)}, nil
		default:
			return HandleResult{}, err
		}
	}
	if err := s.projectCommittedEvents(ctx, result.Events); err != nil {
		return HandleResult{}, err
	}
	return HandleResult{Receipt: result.Receipt, Replay: result.Replay}, nil
}

// applyFounderGardenResolved is SG5's replay arm for plant, uproot, and
// set_substrate, shared by live and history. The advance already ran as the
// pre-step; advance is its result (nil when the garden is inactive).
func applyFounderGardenResolved(state *save.State, request IntentRequest, revision save.Revision, catalogs CatalogBundle,
	serverMS int64, resolvedJSON json.RawMessage, advance *garden.Advance) (FounderLoggedTransition, error) {
	if !isGardenIntent(request.Kind) || request.Kind == IntentGardenHarvest || request.ExpectedRevision != revision.Number || request.InvalidDetail != "" {
		return FounderLoggedTransition{}, fmt.Errorf("%w: garden Founder command", ErrInvalidReplayInputs)
	}
	var resolved founderGardenResolved
	if err := decodeReplayStrict(resolvedJSON, &resolved); err != nil || resolved.Kind != request.Kind || resolved.ServerMS != serverMS {
		return FounderLoggedTransition{}, fmt.Errorf("%w: garden Founder inputs", ErrInvalidReplayInputs)
	}
	recomputed := garden.Advance{Matured: []garden.PlotRef{}, Spawned: []garden.Spawn{}}
	if advance != nil {
		recomputed = *advance
	}
	if !advanceEqual(recomputed, resolved.Advance) {
		return FounderLoggedTransition{}, fmt.Errorf("%w: garden advance diverged from resolved inputs", ErrInvalidReplayInputs)
	}
	reject := func(category, detail string) (FounderLoggedTransition, error) {
		decision, err := rejectedDecision(request, revision.Number, category, detail)
		return founderDecisionTransition(state, decision, catalogs.ConstantsHash, err)
	}
	if !gardenActive(catalogs, state) {
		return reject("not_eligible", garden.DetailInactive)
	}
	hostLevel, unlocked := gardenInputs(catalogs.Garden, state)
	humanLocked := false
	if catalogs.Garden.SoulGate == garden.SoulGateHumanHobby {
		if catalogs.Soul == nil {
			return FounderLoggedTransition{}, fmt.Errorf("%w: human_hobby garden without Soul", ErrInvalidReplayInputs)
		}
		locked, err := catalogs.Soul.HumanContentLocked(state.Soul)
		if err != nil {
			return FounderLoggedTransition{}, err
		}
		humanLocked = locked
	}
	var event gardenEvent
	var err error
	if err = catalogs.Garden.Gate(unlocked, humanLocked); err == nil {
		switch request.Kind {
		case IntentGardenPlant:
			var planted garden.Planted
			planted, err = catalogs.Garden.Plant(state.ServerGarden, hostLevel, request.GardenRow, request.GardenCol, request.GardenSpeciesID)
			event = gardenEvent{Kind: save.EventGardenPlanted, Payload: planted}
		case IntentGardenUproot:
			var uprooted garden.Uprooted
			uprooted, err = catalogs.Garden.Uproot(state.ServerGarden, request.GardenRow, request.GardenCol)
			event = gardenEvent{Kind: save.EventGardenUprooted, Payload: uprooted}
		case IntentGardenSetSubstrate:
			var set garden.SubstrateSet
			set, err = catalogs.Garden.SetSubstrate(state.ServerGarden, serverMS, request.GardenSubstrateID)
			event = gardenEvent{Kind: save.EventGardenSubstrateSet, Payload: set}
		}
	}
	if err != nil {
		rejection, ok := garden.AsRejection(err)
		if !ok {
			return FounderLoggedTransition{}, err
		}
		return reject(rejection.Category, rejection.Detail)
	}
	receipt, err := json.Marshal(founderGardenReceipt{IntentID: request.IntentID, Outcome: string(save.IntentApplied),
		FounderRevision: revision.Number + 1, Kind: request.Kind, Event: event})
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	return FounderLoggedTransition{State: state, Outcome: save.IntentApplied, Receipt: receipt,
		Events:              []save.EventWrite{{Kind: event.Kind, SchemaVersion: 1, IntentID: request.IntentID, Payload: payload}},
		ResultConstantsHash: catalogs.ConstantsHash}, nil
}

func advanceEqual(left, right garden.Advance) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftJSON) == string(rightJSON)
}
