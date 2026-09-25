package production

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"cloud-clicker/server/cosmetic"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

// Cosmetic Shop v1 §4 (Founder scope). None of these reads or writes the
// ledger, and no receipt carries an amount, price, currency, or payment field.
const (
	IntentAcquireCosmetic = "acquire_cosmetic"
	IntentEquipCosmetic   = "equip_cosmetic"
	IntentUnequipCosmetic = "unequip_cosmetic"
)

// cosmeticStreamIDPattern is any lowercase UUID: stream IDs are not all v7.
var cosmeticStreamIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func isCosmeticIntent(kind string) bool {
	return kind == IntentAcquireCosmetic || kind == IntentEquipCosmetic || kind == IntentUnequipCosmetic
}

// founderCosmeticActiveCompany is §4.2 step 5's frozen active-Company
// context: the server selects the active Company sibling and freezes its
// authoritative tier, as care_action freezes attendance. The client never
// supplies a tier.
type founderCosmeticActiveCompany struct {
	CompanyStreamID string `json:"company_stream_id"`
	CompanyRevision int64  `json:"company_revision"`
	RunSeq          int64  `json:"run_seq"`
	Tier            int64  `json:"tier"`
}

type founderCosmeticResolved struct {
	Kind          string                        `json:"kind"`
	ActiveCompany *founderCosmeticActiveCompany `json:"active_company,omitempty"`
}

type cosmeticEvent struct {
	Kind    save.EventKind `json:"kind"`
	Payload map[string]any `json:"payload"`
}

type founderCosmeticReceipt struct {
	IntentID        string          `json:"intent_id"`
	Outcome         string          `json:"outcome"`
	FounderRevision int64           `json:"founder_revision"`
	Kind            string          `json:"kind"`
	Cosmetics       *cosmetic.State `json:"cosmetics"`
	Event           cosmeticEvent   `json:"event"`
}

func (s *Service) resolveCosmeticActiveCompany(ctx context.Context, companyStreamID string, founderOwnerID string) (*founderCosmeticActiveCompany, error) {
	company, err := s.store.LoadLatest(ctx, companyStreamID)
	if err != nil {
		return nil, err
	}
	if company.ArchivedAt != nil || company.Key.OwnerKind != save.OwnerFounder || company.Key.OwnerID != founderOwnerID ||
		company.Key.Scope != economy.ScopeCompany || company.State == nil {
		return nil, fmt.Errorf("%w: active Company context unavailable", ErrInvalidIntent)
	}
	return &founderCosmeticActiveCompany{CompanyStreamID: companyStreamID, CompanyRevision: company.Revision.Number,
		RunSeq: company.State.RunSeq, Tier: company.State.Tier}, nil
}

func (s *Service) handleFounderCosmetic(ctx context.Context, companyStreamID string, request IntentRequest) (HandleResult, error) {
	if s.replayCatalogs == nil {
		return HandleResult{}, fmt.Errorf("%w: Founder replay runtime unavailable", ErrInvalidIntent)
	}
	founder, err := s.store.LoadSiblingLatest(ctx, companyStreamID, economy.ScopeFounder)
	if err != nil {
		return HandleResult{}, err
	}
	var active *founderCosmeticActiveCompany
	if request.Kind == IntentAcquireCosmetic && request.InvalidDetail == "" {
		if active, err = s.resolveCosmeticActiveCompany(ctx, companyStreamID, founder.Key.OwnerID); err != nil {
			return HandleResult{}, err
		}
	}
	// §4.5: cosmetic intents are never blocked as exclusive_activity, so the
	// unguarded Founder boundary is used deliberately.
	result, err := s.store.ApplyFounderLogged(ctx, founder.Revision.StreamID, request.ExpectedRevision,
		request.IntentID, request.RequestHash, request.CanonicalPayload,
		func(state *save.State, revision save.Revision, command save.FounderReplayCommand) (save.IntentDecision, json.RawMessage, error) {
			bundle, ok := s.replayCatalogs.ResolveReplayCatalogs(revision.ConstantsHash)
			if !ok {
				return save.IntentDecision{}, nil, fmt.Errorf("%w: replay catalog bundle unavailable", ErrInvalidIntent)
			}
			var resolved any = founderInvalidResolved{Kind: "invalid", Detail: request.InvalidDetail}
			if request.InvalidDetail == "" {
				resolved = founderCosmeticResolved{Kind: request.Kind, ActiveCompany: active}
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

// applyFounderCosmeticResolved is §4's replay arm, shared by live and history:
// the §4.1 prefix, the per-intent validation order, and the single effect.
func applyFounderCosmeticResolved(state *save.State, request IntentRequest, revision save.Revision, catalogs CatalogBundle,
	resolvedJSON json.RawMessage) (FounderLoggedTransition, error) {
	if !isCosmeticIntent(request.Kind) || request.ExpectedRevision != revision.Number || request.InvalidDetail != "" {
		return FounderLoggedTransition{}, fmt.Errorf("%w: cosmetic Founder command", ErrInvalidReplayInputs)
	}
	var resolved founderCosmeticResolved
	if err := decodeReplayStrict(resolvedJSON, &resolved); err != nil || resolved.Kind != request.Kind ||
		(request.Kind == IntentAcquireCosmetic) != (resolved.ActiveCompany != nil) {
		return FounderLoggedTransition{}, fmt.Errorf("%w: cosmetic Founder inputs", ErrInvalidReplayInputs)
	}
	if active := resolved.ActiveCompany; active != nil && (!cosmeticStreamIDPattern.MatchString(active.CompanyStreamID) ||
		active.CompanyRevision < 1 || active.RunSeq < 0 || active.RunSeq > decimal.MaxExactInteger || active.Tier < 0 || active.Tier > 8) {
		return FounderLoggedTransition{}, fmt.Errorf("%w: cosmetic active Company context", ErrInvalidReplayInputs)
	}
	reject := func(category, detail string) (FounderLoggedTransition, error) {
		decision, err := rejectedDecision(request, revision.Number, category, detail)
		return founderDecisionTransition(state, decision, catalogs.ConstantsHash, err)
	}
	// §4.1 prefix.
	if catalogs.Cosmetics == nil || save.VersionForState(state) < 24 || state.Cosmetics == nil {
		return reject("not_eligible", "inactive")
	}
	var item cosmetic.Item
	if request.Kind != IntentUnequipCosmetic {
		var ok bool
		if item, ok = catalogs.Cosmetics.Item(request.CosmeticID); !ok {
			return reject("unknown_id", "cosmetic_id")
		}
	}
	next := state.Cosmetics.Clone()
	var event cosmeticEvent
	switch request.Kind {
	case IntentAcquireCosmetic:
		if next.Owns(item.CosmeticID) {
			return reject("not_eligible", "owned")
		}
		if resolved.ActiveCompany.Tier < item.Unlock.Tier {
			return reject("not_eligible", "locked")
		}
		next.Owned = insertSorted(next.Owned, item.CosmeticID)
		event = cosmeticEvent{Kind: save.EventCosmeticAcquired, Payload: map[string]any{"cosmetic_id": item.CosmeticID, "order_number": int64(len(next.Owned))}}
	case IntentEquipCosmetic:
		if _, ok := state.Pets[request.CosmeticPetID]; !ok {
			return reject("unknown_id", "pet_id")
		}
		if !next.Owns(item.CosmeticID) {
			return reject("not_eligible", "not_owned")
		}
		previous, wearing := next.Equipped[request.CosmeticPetID]
		if wearing && previous == item.CosmeticID {
			return reject("not_eligible", "already_equipped")
		}
		var replaced any
		if wearing {
			replaced = previous
		}
		next.Equipped[request.CosmeticPetID] = item.CosmeticID
		event = cosmeticEvent{Kind: save.EventCosmeticEquipped, Payload: map[string]any{"cosmetic_id": item.CosmeticID, "pet_id": request.CosmeticPetID, "replaced_cosmetic_id": replaced}}
	case IntentUnequipCosmetic:
		if _, ok := state.Pets[request.CosmeticPetID]; !ok {
			return reject("unknown_id", "pet_id")
		}
		worn, wearing := next.Equipped[request.CosmeticPetID]
		if !wearing {
			return reject("not_eligible", "nothing_equipped")
		}
		delete(next.Equipped, request.CosmeticPetID)
		event = cosmeticEvent{Kind: save.EventCosmeticUnequipped, Payload: map[string]any{"cosmetic_id": worn, "pet_id": request.CosmeticPetID}}
	}
	state.Cosmetics = next
	receipt, err := json.Marshal(founderCosmeticReceipt{IntentID: request.IntentID, Outcome: string(save.IntentApplied),
		FounderRevision: revision.Number + 1, Kind: request.Kind, Cosmetics: next.Clone(), Event: event})
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

// checkCosmeticsTransition is §4.5/AC8: only the three cosmetic intents may
// change `cosmetics`; every other Founder transition, Exit included, carries
// it byte-identically (activation from absent to empty excepted).
func checkCosmeticsTransition(before, after *cosmetic.State, cosmeticIntent bool) error {
	if before == nil {
		if after != nil && !after.Equal(cosmetic.NewState()) {
			return fmt.Errorf("%w: cosmetics appeared outside activation", ErrInvalidEngineState)
		}
		return nil
	}
	if !cosmeticIntent && !before.Equal(after) {
		return fmt.Errorf("%w: a non-cosmetic Founder transition changed cosmetics", ErrInvalidEngineState)
	}
	return nil
}

func insertSorted(values []string, value string) []string {
	result := make([]string, 0, len(values)+1)
	inserted := false
	for _, existing := range values {
		if !inserted && value < existing {
			result = append(result, value)
			inserted = true
		}
		result = append(result, existing)
	}
	if !inserted {
		result = append(result, value)
	}
	return result
}
