package production

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/save"
	"cloud-clicker/server/soul"
)

// IntentAdoptPet is Pet Adoption v1 PA4 (Founder scope).
const IntentAdoptPet = "adopt_pet"

var adoptionNoncePattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

// founderAdoptionResolved is PA4.7's frozen resolved-input arm. Replay
// re-derives the pet ID, temperament and palette from the nonce; it never
// trusts a stored derived value.
type founderAdoptionResolved struct {
	Kind          string                  `json:"kind"`
	Attendance    FounderAttendanceSample `json:"attendance"`
	AdoptionNonce string                  `json:"adoption_nonce"`
}

type founderAdoptionReceipt struct {
	IntentID            string         `json:"intent_id"`
	Outcome             string         `json:"outcome"`
	FounderRevision     int64          `json:"founder_revision"`
	PetID               string         `json:"pet_id"`
	SpeciesID           string         `json:"species_id"`
	Temperament         string         `json:"temperament"`
	PaletteID           string         `json:"palette_id"`
	NameKey             string         `json:"name_key"`
	AdoptedAtAttendedMS int64          `json:"adopted_at_attended_ms"`
	StatusBand          pet.StatusBand `json:"status_band"`
	EligibleActionIDs   []string       `json:"eligible_action_ids"`
}

// WithAdoptionNonceSource replaces the server CSPRNG nonce source; tests use
// it to count draws (AC10) and pin vectors.
func WithAdoptionNonceSource(source func() (string, error)) ServiceOption {
	return func(service *Service) error {
		if source == nil {
			return ErrInvalidIntent
		}
		service.adoptionNonce = source
		return nil
	}
}

func cryptoAdoptionNonce() (string, error) {
	var nonce [pet.AdoptionNonceBytes]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(nonce[:]), nil
}

func (s *Service) handleFounderAdoption(ctx context.Context, companyStreamID string, now time.Time, request IntentRequest) (HandleResult, error) {
	if s.replayCatalogs == nil {
		return HandleResult{}, fmt.Errorf("%w: Founder replay runtime unavailable", ErrInvalidIntent)
	}
	founder, err := s.store.LoadSiblingLatest(ctx, companyStreamID, economy.ScopeFounder)
	if err != nil {
		return HandleResult{}, err
	}
	attendance, err := s.ResolveFounderAttendance(ctx, founder.Revision.StreamID, companyStreamID, now)
	if err != nil {
		return HandleResult{}, err
	}
	nonceSource := s.adoptionNonce
	if nonceSource == nil {
		nonceSource = cryptoAdoptionNonce
	}
	mutation := func(state *save.State, revision save.Revision, command save.FounderReplayCommand) (save.IntentDecision, json.RawMessage, error) {
		bundle, ok := s.replayCatalogs.ResolveReplayCatalogs(revision.ConstantsHash)
		if !ok {
			return save.IntentDecision{}, nil, fmt.Errorf("%w: replay catalog bundle unavailable", ErrInvalidIntent)
		}
		var resolved any = founderInvalidResolved{Kind: "invalid", Detail: request.InvalidDetail}
		if request.InvalidDetail == "" {
			if attendance.CompanyConstantsHash != revision.ConstantsHash {
				return save.IntentDecision{}, nil, fmt.Errorf("%w: adoption clock context unavailable", ErrInvalidIntent)
			}
			// The callback runs only on first execution; an idempotent retry
			// returns the recorded receipt, so exactly one nonce is drawn.
			nonce, nonceErr := nonceSource()
			if nonceErr != nil || !adoptionNoncePattern.MatchString(nonce) {
				return save.IntentDecision{}, nil, fmt.Errorf("%w: adoption nonce unavailable", ErrInvalidIntent)
			}
			resolved = founderAdoptionResolved{Kind: IntentAdoptPet, Attendance: attendance, AdoptionNonce: nonce}
		}
		replayInputs, marshalErr := marshalFounderReplayInputs(command, resolved)
		if marshalErr != nil {
			return save.IntentDecision{}, nil, marshalErr
		}
		transition, transitionErr := ApplyFounderLogged(state, request.CanonicalPayload, bundle, replayInputs)
		return save.IntentDecision{Outcome: transition.Outcome, Receipt: transition.Receipt,
			Events: transition.Events}, replayInputs, transitionErr
	}
	var result save.IntentResult
	if s.soulRecoveries != nil {
		result, err = s.store.ApplyFounderLoggedGuarded(ctx, founder.Revision.StreamID, request.ExpectedRevision,
			request.IntentID, request.RequestHash, request.CanonicalPayload, s.soulRecoveries.RequireInactiveTx, mutation)
	} else {
		result, err = s.store.ApplyFounderLogged(ctx, founder.Revision.StreamID, request.ExpectedRevision,
			request.IntentID, request.RequestHash, request.CanonicalPayload, mutation)
	}
	if err != nil {
		if errors.Is(err, soul.ErrRecoveryActive) {
			return HandleResult{Receipt: marshalRejection(request.IntentID, request.ExpectedRevision, "not_eligible", "exclusive_activity")}, nil
		}
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

// applyFounderAdoptionResolved is PA4's replay arm, shared by live and
// history: the PA4.2 validation order, the draws, and the single-transaction
// state effect.
func applyFounderAdoptionResolved(state *save.State, request IntentRequest, revision save.Revision, catalogs CatalogBundle,
	serverTSMS int64, resolvedJSON json.RawMessage) (FounderLoggedTransition, error) {
	if request.Kind != IntentAdoptPet || request.ExpectedRevision != revision.Number || request.InvalidDetail != "" {
		return FounderLoggedTransition{}, fmt.Errorf("%w: adoption Founder command", ErrInvalidReplayInputs)
	}
	var resolved founderAdoptionResolved
	if err := decodeReplayStrict(resolvedJSON, &resolved); err != nil || resolved.Kind != IntentAdoptPet ||
		!adoptionNoncePattern.MatchString(resolved.AdoptionNonce) || resolved.Attendance.CompanyConstantsHash != catalogs.ConstantsHash ||
		ValidateFounderAttendanceSample(state, revision.Number, request.ExpectedRevision, resolved.Attendance) != nil {
		return FounderLoggedTransition{}, fmt.Errorf("%w: adoption Founder inputs", ErrInvalidReplayInputs)
	}
	reject := func(category, detail string) (FounderLoggedTransition, error) {
		decision, err := rejectedDecision(request, revision.Number, category, detail)
		return founderDecisionTransition(state, decision, catalogs.ConstantsHash, err)
	}
	// PA4.2 rows 2–6; the first failure wins.
	if catalogs.PetSpecies == nil || catalogs.Pets == nil || save.VersionForState(state) < 23 || state.PetIdentities == nil {
		return reject("not_eligible", "adoption_inactive")
	}
	row, ok := catalogs.PetSpecies.Row(request.PetSpeciesID)
	if !ok {
		return reject("unknown_id", "unknown_species")
	}
	if !row.HasName(request.PetNameKey) {
		return reject("unknown_id", "unknown_name")
	}
	if row.Availability != pet.AvailabilityStarter {
		return reject("not_eligible", "species_locked")
	}
	if int64(len(state.PetIdentities)) >= catalogs.PetSpecies.MaxPetsPerFounder {
		return reject("not_eligible", "adoption_cap_reached")
	}
	draws, err := pet.DrawAdoption(revision.OwnerID, resolved.AdoptionNonce, serverTSMS, row)
	if err != nil {
		return FounderLoggedTransition{}, fmt.Errorf("%w: adoption draw", ErrInvalidReplayInputs)
	}
	if _, exists := state.PetIdentities[draws.PetID]; exists {
		return FounderLoggedTransition{}, fmt.Errorf("%w: adopted pet ID collides", ErrInvalidEngineState)
	}
	if _, exists := state.Pets[draws.PetID]; exists {
		return FounderLoggedTransition{}, fmt.Errorf("%w: adopted pet ID collides with care", ErrInvalidEngineState)
	}
	attended := resolved.Attendance.EffectiveFounderAttendedMS
	care, err := pet.InitialCareState(catalogs.Pets, attended)
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	band, _, err := pet.CareStatus(care, catalogs.Pets)
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	identity := pet.Identity{SpeciesID: row.SpeciesID, Temperament: draws.Temperament, PaletteID: draws.PaletteID,
		NameKey: request.PetNameKey, AdoptedAtMS: serverTSMS, AdoptedAtAttendedMS: attended}
	if state.Pets == nil {
		state.Pets = map[string]pet.CareState{}
	}
	state.Pets[draws.PetID] = care
	state.PetIdentities[draws.PetID] = identity
	eligible := pet.EligibleCareActions(care, catalogs.Pets, attended)
	if eligible == nil {
		eligible = []string{}
	}
	receipt, err := json.Marshal(founderAdoptionReceipt{IntentID: request.IntentID, Outcome: string(save.IntentApplied),
		FounderRevision: revision.Number + 1, PetID: draws.PetID, SpeciesID: row.SpeciesID, Temperament: draws.Temperament,
		PaletteID: draws.PaletteID, NameKey: request.PetNameKey, AdoptedAtAttendedMS: attended, StatusBand: band, EligibleActionIDs: eligible})
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	payload, err := json.Marshal(map[string]any{"pet_id": draws.PetID, "species_id": row.SpeciesID, "temperament": draws.Temperament,
		"palette_id": draws.PaletteID, "name_key": request.PetNameKey})
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	return FounderLoggedTransition{State: state, Outcome: save.IntentApplied, Receipt: receipt,
		Events:              []save.EventWrite{{Kind: save.EventPetAdopted, SchemaVersion: 1, IntentID: request.IntentID, Payload: payload}},
		ResultConstantsHash: catalogs.ConstantsHash}, nil
}

func checkPetIdentityTransition(before, after map[string]pet.Identity, adoption bool) error {
	if before == nil {
		if len(after) != 0 {
			return fmt.Errorf("%w: pet identities appeared outside activation", ErrInvalidEngineState)
		}
		return nil
	}
	added := 0
	for id, identity := range after {
		prior, existed := before[id]
		if !existed {
			added++
			continue
		}
		if prior != identity {
			return fmt.Errorf("%w: pet identity %q changed", ErrInvalidEngineState, id)
		}
	}
	for id := range before {
		if _, kept := after[id]; !kept {
			return fmt.Errorf("%w: pet identity %q disappeared", ErrInvalidEngineState, id)
		}
	}
	if adoption && added != 1 || !adoption && added != 0 {
		return fmt.Errorf("%w: %d pet identities added", ErrInvalidEngineState, added)
	}
	return nil
}
