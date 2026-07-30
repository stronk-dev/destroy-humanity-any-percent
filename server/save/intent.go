package save

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
)

var ErrIdempotencyConflict = errors.New("idempotency key reused with a different request")

var (
	uuidV7Pattern       = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	mechanicalIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

type IntentOutcome string

const (
	IntentApplied  IntentOutcome = "applied"
	IntentRejected IntentOutcome = "rejected"
)

type EventKind string

const (
	EventGeneratorPurchased        EventKind = "generator_purchased"
	EventInvariantReported         EventKind = "invariant_reported"
	EventCompensation              EventKind = "compensation"
	EventGateCrossed               EventKind = "gate_crossed"
	EventRouteExecuted             EventKind = "route_executed"
	EventRouteHintPurchased        EventKind = "route_hint_purchased"
	EventRouteKnowledgeGranted     EventKind = "route_knowledge_granted"
	EventCompactSigned             EventKind = "compact_signed"
	EventCompactLeft               EventKind = "compact_left"
	EventCompactSampled            EventKind = "compact_sampled"
	EventCompactHealthBandChanged  EventKind = "compact_health_band_changed"
	EventCompactCascadeStarted     EventKind = "compact_cascade_started"
	EventCompactRecovered          EventKind = "compact_recovered"
	EventCompactRecruitmentOffered EventKind = "compact_recruitment_offered"
	EventExitOfferSpawned          EventKind = "exit_offer_spawned"
	EventExitOfferExpired          EventKind = "exit_offer_expired"
	EventExitOfferDeclined         EventKind = "exit_offer_declined"
	EventRunEnded                  EventKind = "run_ended"
	EventRunStarted                EventKind = "run_started"
	EventFounderAdvanced           EventKind = "founder_advanced"
)

type EventWrite struct {
	Kind          EventKind
	SchemaVersion int
	IntentID      string
	Payload       json.RawMessage
}

type EventRecord struct {
	EventID       string
	StreamID      string
	OwnerID       string
	Revision      int64
	Kind          EventKind
	IntentID      string
	ConstantsHash string
	OccurredAt    time.Time
	Payload       json.RawMessage
}

type IntentDecision struct {
	Outcome IntentOutcome
	Receipt json.RawMessage
	Events  []EventWrite
}

type IntentMutation func(state *State, revision Revision) (IntentDecision, error)

type IntentResult struct {
	Outcome IntentOutcome
	Receipt json.RawMessage
	Replay  bool
	Events  []EventRecord
}

type RevisionConflict struct {
	Expected int64
	Current  int64
}

func (e *RevisionConflict) Error() string {
	return fmt.Sprintf("%v: got %d, current %d", ErrConflict, e.Expected, e.Current)
}

func (e *RevisionConflict) Unwrap() error { return ErrConflict }

func (s *Store) ApplyIntent(
	ctx context.Context,
	streamID string,
	expectedRevision int64,
	intentID string,
	requestHash string,
	mutate IntentMutation,
) (IntentResult, error) {
	return s.applyIntent(ctx, streamID, expectedRevision, intentID, requestHash, nil, mutate)
}

func (s *Store) ApplyIntentLogged(
	ctx context.Context,
	streamID string,
	expectedRevision int64,
	intentID string,
	requestHash string,
	canonicalPayload []byte,
	mutate IntentMutation,
) (IntentResult, error) {
	if err := validateCanonicalPayload(canonicalPayload, requestHash); err != nil {
		return IntentResult{}, err
	}
	return s.applyIntent(ctx, streamID, expectedRevision, intentID, requestHash, canonicalPayload, mutate)
}

func (s *Store) applyIntent(
	ctx context.Context,
	streamID string,
	expectedRevision int64,
	intentID string,
	requestHash string,
	canonicalPayload []byte,
	mutate IntentMutation,
) (IntentResult, error) {
	if !uuidPattern.MatchString(streamID) || expectedRevision < 1 || !uuidV7Pattern.MatchString(intentID) ||
		!hashPattern.MatchString(requestHash) || mutate == nil {
		return IntentResult{}, ErrInvalidStream
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return IntentResult{}, err
	}
	defer tx.Rollback()

	var scope economy.Scope
	var ownerID string
	var archivedAt sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT scope, owner_id, archived_at FROM save_streams WHERE id=$1 FOR UPDATE`, streamID).Scan(&scope, &ownerID, &archivedAt); errors.Is(err, sql.ErrNoRows) {
		return IntentResult{}, ErrNotFound
	} else if err != nil {
		return IntentResult{}, err
	}
	if archivedAt.Valid {
		return IntentResult{}, ErrArchived
	}

	var recordedHash string
	var recordedOutcome IntentOutcome
	var recordedReceipt []byte
	err = tx.QueryRowContext(ctx, `
		SELECT request_hash,outcome,receipt FROM intent_records
		WHERE stream_id=$1 AND intent_id=$2`, streamID, intentID,
	).Scan(&recordedHash, &recordedOutcome, &recordedReceipt)
	if err == nil {
		if recordedHash != requestHash {
			return IntentResult{}, ErrIdempotencyConflict
		}
		recordedReceipt, err = normalizeJSON(recordedReceipt)
		if err != nil {
			return IntentResult{}, fmt.Errorf("%w: stored intent receipt: %v", ErrInvalidState, err)
		}
		events, err := s.eventsForIntent(ctx, tx, streamID, ownerID, intentID)
		if err != nil {
			return IntentResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return IntentResult{}, err
		}
		return IntentResult{Outcome: recordedOutcome, Receipt: cloneRaw(recordedReceipt), Replay: true, Events: events}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return IntentResult{}, err
	}

	var revision Revision
	var stateBytes []byte
	err = tx.QueryRowContext(ctx, `
		SELECT revision,version,constants_hash,created_at,state
		FROM save_revisions WHERE stream_id=$1 ORDER BY revision DESC LIMIT 1`, streamID,
	).Scan(&revision.Number, &revision.Version, &revision.ConstantsHash, &revision.CreatedAt, &stateBytes)
	if errors.Is(err, sql.ErrNoRows) {
		return IntentResult{}, ErrNotFound
	}
	if err != nil {
		return IntentResult{}, err
	}
	revision.StreamID = streamID
	revision.OwnerID = ownerID
	if revision.Number != expectedRevision {
		return IntentResult{}, &RevisionConflict{Expected: expectedRevision, Current: revision.Number}
	}
	catalog, ok := s.catalogs.Resolve(revision.ConstantsHash)
	if !ok {
		return IntentResult{}, fmt.Errorf("%w: unknown catalog %s", ErrInvalidState, revision.ConstantsHash)
	}
	state, err := RestoreState(stateBytes, revision.Version, catalog, scope, revision.CreatedAt)
	if err != nil {
		return IntentResult{}, err
	}
	var runLogSequence int64
	if scope == economy.ScopeCompany && len(canonicalPayload) != 0 {
		if err := requireRunEpochTx(ctx, tx, streamID, state.RunSeq, revision.ConstantsHash); err != nil {
			return IntentResult{}, err
		}
		runLogSequence, err = nextRunLogSequence(ctx, tx, streamID, state.RunSeq)
		if err != nil {
			return IntentResult{}, err
		}
		revision.RunLogSequence = runLogSequence
	}
	decision, err := mutate(state, revision)
	if err != nil {
		return IntentResult{}, err
	}
	if err := validateIntentDecision(decision, intentID); err != nil {
		return IntentResult{}, err
	}
	decision.Receipt, err = normalizeJSON(decision.Receipt)
	if err != nil {
		return IntentResult{}, fmt.Errorf("%w: normalize intent receipt: %v", ErrInvalidStream, err)
	}

	if decision.Outcome == IntentRejected {
		if runLogSequence != 0 {
			if err := insertRunLog(ctx, tx, streamID, state.RunSeq, runLogSequence, intentID, canonicalPayload, decision.Receipt, nil); err != nil {
				return IntentResult{}, err
			}
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO intent_records (stream_id,intent_id,request_hash,outcome,receipt)
			VALUES ($1,$2,$3,$4,$5)`, streamID, intentID, requestHash, decision.Outcome, decision.Receipt); err != nil {
			return IntentResult{}, err
		}
		if scope == economy.ScopeCompany {
			if err := insertReceiptOutbox(ctx, tx, ownerID, streamID, intentID, revision.Number, revision.ConstantsHash, decision.Receipt); err != nil {
				return IntentResult{}, err
			}
		}
		if err := tx.Commit(); err != nil {
			return IntentResult{}, err
		}
		return IntentResult{Outcome: decision.Outcome, Receipt: cloneRaw(decision.Receipt)}, nil
	}

	encodedState, err := s.validatedState(revision.ConstantsHash, scope, state)
	if err != nil {
		return IntentResult{}, err
	}
	newRevision := revision.Number + 1
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO save_revisions (stream_id,revision,version,state,constants_hash)
		VALUES ($1,$2,$3,$4,$5)`, streamID, newRevision, CurrentVersion, encodedState, revision.ConstantsHash); err != nil {
		return IntentResult{}, err
	}
	recordedEvents := make([]EventRecord, 0, len(decision.Events))
	for _, event := range decision.Events {
		var eventIntent any
		if event.IntentID != "" {
			eventIntent = event.IntentID
		}
		var record EventRecord
		var storedIntent sql.NullString
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO events (stream_id,revision,schema_version,kind,intent_id,constants_hash,payload)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
			RETURNING event_id,kind,intent_id,occurred_at,payload`, streamID, newRevision, event.SchemaVersion,
			event.Kind, eventIntent, revision.ConstantsHash, event.Payload).Scan(&record.EventID, &record.Kind, &storedIntent, &record.OccurredAt, &record.Payload); err != nil {
			return IntentResult{}, err
		}
		record.StreamID, record.OwnerID, record.Revision, record.ConstantsHash = streamID, ownerID, newRevision, revision.ConstantsHash
		if storedIntent.Valid {
			record.IntentID = storedIntent.String
		}
		recordedEvents = append(recordedEvents, record)
	}
	if runLogSequence != 0 {
		if err := insertRunLog(ctx, tx, streamID, state.RunSeq, runLogSequence, intentID, canonicalPayload, decision.Receipt, &newRevision); err != nil {
			return IntentResult{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO intent_records (stream_id,intent_id,request_hash,outcome,receipt)
		VALUES ($1,$2,$3,$4,$5)`, streamID, intentID, requestHash, decision.Outcome, decision.Receipt); err != nil {
		return IntentResult{}, err
	}
	if scope == economy.ScopeCompany {
		if err := insertReceiptOutbox(ctx, tx, ownerID, streamID, intentID, newRevision, revision.ConstantsHash, decision.Receipt); err != nil {
			return IntentResult{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM save_revisions WHERE stream_id=$1 AND revision <= $2`, streamID, newRevision-5); err != nil {
		return IntentResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return IntentResult{}, err
	}
	return IntentResult{Outcome: decision.Outcome, Receipt: cloneRaw(decision.Receipt), Events: recordedEvents}, nil
}

func (s *Store) eventsForIntent(ctx context.Context, tx *sql.Tx, streamID, ownerID, intentID string) ([]EventRecord, error) {
	rows, err := tx.QueryContext(ctx, `SELECT event_id,revision,kind,intent_id,constants_hash,occurred_at,payload FROM events WHERE stream_id=$1 AND intent_id=$2 ORDER BY event_id`, streamID, intentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []EventRecord
	for rows.Next() {
		var record EventRecord
		var storedIntent sql.NullString
		if err := rows.Scan(&record.EventID, &record.Revision, &record.Kind, &storedIntent, &record.ConstantsHash, &record.OccurredAt, &record.Payload); err != nil {
			return nil, err
		}
		record.StreamID, record.OwnerID = streamID, ownerID
		if storedIntent.Valid {
			record.IntentID = storedIntent.String
		}
		events = append(events, record)
	}
	return events, rows.Err()
}

func (s *Store) PruneIntentRecords(ctx context.Context, cutoff time.Time) (int64, error) {
	if cutoff.IsZero() {
		return 0, ErrInvalidStream
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM intent_records WHERE created_at < $1`, cutoff.UTC())
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func validateIntentDecision(decision IntentDecision, intentID string) error {
	if decision.Outcome != IntentApplied && decision.Outcome != IntentRejected {
		return ErrInvalidStream
	}
	var receipt map[string]json.RawMessage
	if err := decodeStrictJSON(decision.Receipt, &receipt); err != nil || receipt == nil {
		return fmt.Errorf("%w: invalid intent receipt", ErrInvalidStream)
	}
	if decision.Outcome == IntentRejected && len(decision.Events) != 0 {
		return fmt.Errorf("%w: rejected intent cannot emit events", ErrInvalidStream)
	}
	for _, event := range decision.Events {
		if event.SchemaVersion != 1 || event.IntentID != "" && !uuidV7Pattern.MatchString(event.IntentID) {
			return fmt.Errorf("%w: invalid event envelope", ErrInvalidStream)
		}
		if event.Kind != EventCompensation && event.IntentID != intentID {
			return fmt.Errorf("%w: event intent does not match mutation", ErrInvalidStream)
		}
		if err := validateEventPayload(event); err != nil {
			return err
		}
	}
	return nil
}

func validateEventPayload(event EventWrite) error {
	switch event.Kind {
	case EventGeneratorPurchased:
		var payload struct {
			GeneratorID    string `json:"generator_id"`
			Count          int64  `json:"count"`
			CostResourceID string `json:"cost_resource_id"`
			Cost           string `json:"cost"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !mechanicalIDPattern.MatchString(payload.GeneratorID) ||
			!mechanicalIDPattern.MatchString(payload.CostResourceID) || payload.Count <= 0 || payload.Count > decimal.MaxExactInteger {
			return fmt.Errorf("%w: invalid generator_purchased payload", ErrInvalidStream)
		}
		if value, err := decimal.ParseCanonical(payload.Cost); err != nil || value.Lt(decimal.Zero) {
			return fmt.Errorf("%w: invalid generator_purchased cost", ErrInvalidStream)
		}
	case EventInvariantReported:
		var payload struct {
			InvariantKind string `json:"invariant_kind"`
			Detail        string `json:"detail"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || payload.Detail == "" ||
			(payload.InvariantKind != "afford_fallback" && payload.InvariantKind != "residual_clamp" && payload.InvariantKind != "residual_abort") {
			return fmt.Errorf("%w: invalid invariant_reported payload", ErrInvalidStream)
		}
	case EventCompensation:
		var payload struct {
			CompensatesEventID string `json:"compensates_event_id"`
			ReasonKey          string `json:"reason_key"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidPattern.MatchString(payload.CompensatesEventID) || !mechanicalIDPattern.MatchString(payload.ReasonKey) {
			return fmt.Errorf("%w: invalid compensation payload", ErrInvalidStream)
		}
	case EventGateCrossed:
		var payload struct {
			GateID    string     `json:"gate_id"`
			RouteID   *string    `json:"route_id"`
			RunID     routeRunID `json:"run_id"`
			FounderID string     `json:"founder_id"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !mechanicalIDPattern.MatchString(payload.GateID) ||
			payload.RouteID != nil && !mechanicalIDPattern.MatchString(*payload.RouteID) || !validRouteRunID(payload.RunID) || !uuidPattern.MatchString(payload.FounderID) {
			return fmt.Errorf("%w: invalid gate_crossed payload", ErrInvalidStream)
		}
	case EventRouteExecuted:
		var payload struct {
			RouteID   string     `json:"route_id"`
			GateID    string     `json:"gate_id"`
			RunID     routeRunID `json:"run_id"`
			FounderID string     `json:"founder_id"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !mechanicalIDPattern.MatchString(payload.RouteID) ||
			!mechanicalIDPattern.MatchString(payload.GateID) || !validRouteRunID(payload.RunID) || !uuidPattern.MatchString(payload.FounderID) {
			return fmt.Errorf("%w: invalid route_executed payload", ErrInvalidStream)
		}
	case EventRouteHintPurchased:
		var payload struct {
			RouteID string `json:"route_id"`
			Cost    int64  `json:"cost"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !mechanicalIDPattern.MatchString(payload.RouteID) || payload.Cost <= 0 || payload.Cost > decimal.MaxExactInteger {
			return fmt.Errorf("%w: invalid route_hint_purchased payload", ErrInvalidStream)
		}
	case EventRouteKnowledgeGranted:
		var payload struct {
			RouteID string `json:"route_id"`
			Amount  int64  `json:"amount"`
			Source  string `json:"source"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !mechanicalIDPattern.MatchString(payload.RouteID) || payload.Amount <= 0 || payload.Amount > decimal.MaxExactInteger ||
			(payload.Source != "registry_first" && payload.Source != "founder_first" && payload.Source != "repeat" && payload.Source != "collapse_exit" && payload.Source != "region_draft") {
			return fmt.Errorf("%w: invalid route_knowledge_granted payload", ErrInvalidStream)
		}
	case EventCompactSigned, EventCompactLeft:
		var payload struct {
			FounderID   string     `json:"founder_id"`
			RunID       routeRunID `json:"run_id"`
			TithePPM    int64      `json:"tithe_ppm"`
			PriorMember bool       `json:"prior_member"`
			NewMember   bool       `json:"new_member"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidPattern.MatchString(payload.FounderID) || !validRouteRunID(payload.RunID) || payload.TithePPM < 0 || payload.TithePPM > 1_000_000 || event.Kind == EventCompactSigned && (payload.PriorMember || !payload.NewMember) || event.Kind == EventCompactLeft && (!payload.PriorMember || payload.NewMember) {
			return fmt.Errorf("%w: invalid compact membership payload", ErrInvalidStream)
		}
	case EventCompactSampled:
		var payload struct {
			FounderID     string     `json:"founder_id"`
			RunID         routeRunID `json:"run_id"`
			WeightPPM     int64      `json:"weight_ppm"`
			CompliancePPM int64      `json:"compliance_ppm"`
			Enclosure     string     `json:"enclosure"`
			Capacity      string     `json:"capacity"`
			SolidarityPPM int64      `json:"solidarity_ppm"`
			SampledMS     int64      `json:"sampled_ms"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidPattern.MatchString(payload.FounderID) || !validRouteRunID(payload.RunID) || payload.WeightPPM < 0 || payload.WeightPPM > 1_000_000 || payload.CompliancePPM < 0 || payload.CompliancePPM > 1_000_000 || payload.SolidarityPPM < 0 || payload.SolidarityPPM > 1_000_000 || payload.SampledMS <= 0 || payload.SampledMS > decimal.MaxExactInteger {
			return fmt.Errorf("%w: invalid compact_sampled payload", ErrInvalidStream)
		}
		if value, err := decimal.ParseCanonical(payload.Enclosure); err != nil || value.Lt(decimal.Zero) || value.Gt(decimal.One) {
			return fmt.Errorf("%w: invalid compact_sampled enclosure", ErrInvalidStream)
		}
		if value, err := decimal.ParseCanonical(payload.Capacity); err != nil || value.Lt(decimal.Zero) {
			return fmt.Errorf("%w: invalid compact_sampled capacity", ErrInvalidStream)
		}
	case EventCompactHealthBandChanged:
		var payload struct {
			ScopeKind string `json:"scope_kind"`
			ScopeID   string `json:"scope_id"`
			FromBand  string `json:"from_band"`
			ToBand    string `json:"to_band"`
			HealthPPM int64  `json:"health_ppm"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || payload.ScopeKind != "server" || !uuidPattern.MatchString(payload.ScopeID) || !validHealthBand(payload.FromBand) || !validHealthBand(payload.ToBand) || payload.FromBand == payload.ToBand || payload.HealthPPM < 0 || payload.HealthPPM > 1_000_000 {
			return fmt.Errorf("%w: invalid compact health band payload", ErrInvalidStream)
		}
	case EventCompactCascadeStarted, EventCompactRecovered:
		var payload struct {
			ScopeKind string `json:"scope_kind"`
			ScopeID   string `json:"scope_id"`
			HealthPPM int64  `json:"health_ppm"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || payload.ScopeKind != "server" || !uuidPattern.MatchString(payload.ScopeID) || payload.HealthPPM < 0 || payload.HealthPPM > 1_000_000 {
			return fmt.Errorf("%w: invalid compact transition payload", ErrInvalidStream)
		}
	case EventCompactRecruitmentOffered:
		var payload struct {
			FounderID string `json:"founder_id"`
			ReasonKey string `json:"reason_key"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidPattern.MatchString(payload.FounderID) || payload.ReasonKey != "compact.recruitment.mid_t3" {
			return fmt.Errorf("%w: invalid compact recruitment payload", ErrInvalidStream)
		}
	case EventExitOfferSpawned:
		var payload struct {
			OfferID       string             `json:"offer_id"`
			ExitType      string             `json:"exit_type"`
			ExpiresAtMS   int64              `json:"expires_at_ms"`
			PayoutPreview eventPrestigeTerms `json:"payout_preview"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidV7Pattern.MatchString(payload.OfferID) ||
			!validExitType(payload.ExitType) || payload.ExpiresAtMS <= 0 || payload.ExpiresAtMS > decimal.MaxExactInteger ||
			validatePrestigeTerms(payload.PayoutPreview) != nil {
			return fmt.Errorf("%w: invalid exit_offer_spawned payload", ErrInvalidStream)
		}
	case EventExitOfferExpired, EventExitOfferDeclined:
		var payload struct {
			OfferID string `json:"offer_id"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidV7Pattern.MatchString(payload.OfferID) {
			return fmt.Errorf("%w: invalid exit offer transition payload", ErrInvalidStream)
		}
	case EventRunEnded:
		var payload eventRunEnded
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidPattern.MatchString(payload.FounderID) ||
			!validRouteRunID(payload.RunID) || !validExitType(payload.ExitType) || payload.StartedAtMS <= 0 ||
			payload.EndedAtMS < payload.StartedAtMS || payload.EndedAtMS > decimal.MaxExactInteger ||
			payload.RTAMS != payload.EndedAtMS-payload.StartedAtMS || payload.AttendedMS < 0 || payload.AttendedMS > payload.RTAMS ||
			payload.TerminalSeq <= 0 || payload.TerminalSeq > decimal.MaxExactInteger || payload.Tier < 0 || payload.Tier > 9 ||
			validatePrestigeTerms(payload.Payout) != nil || !sortedMechanicalIDs(payload.LedgerFactKinds) || !sortedMechanicalIDs(payload.ExecutedRoutes) {
			return fmt.Errorf("%w: invalid run_ended payload", ErrInvalidStream)
		}
		if value, err := decimal.ParseCanonical(payload.LifetimeValue); err != nil || value.Lt(decimal.Zero) {
			return fmt.Errorf("%w: invalid run_ended lifetime value", ErrInvalidStream)
		}
	case EventRunStarted:
		var payload struct {
			FounderID   string        `json:"founder_id"`
			RunID       routeRunID    `json:"run_id"`
			StartedAtMS int64         `json:"started_at_ms"`
			Assisted    eventAssisted `json:"assisted"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidPattern.MatchString(payload.FounderID) ||
			!validRouteRunID(payload.RunID) || payload.StartedAtMS <= 0 || payload.StartedAtMS > decimal.MaxExactInteger {
			return fmt.Errorf("%w: invalid run_started payload", ErrInvalidStream)
		}
	case EventFounderAdvanced:
		var payload struct {
			FounderID       string     `json:"founder_id"`
			RunID           routeRunID `json:"run_id"`
			ExitType        string     `json:"exit_type"`
			ReputationDelta int64      `json:"reputation_delta"`
			RouteKnowledge  int64      `json:"route_knowledge"`
			OccurredAtMS    int64      `json:"occurred_at_ms"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidPattern.MatchString(payload.FounderID) ||
			!validRouteRunID(payload.RunID) || !validExitType(payload.ExitType) || payload.ReputationDelta < 0 ||
			payload.ReputationDelta > decimal.MaxExactInteger || payload.RouteKnowledge < 0 || payload.RouteKnowledge > decimal.MaxExactInteger ||
			payload.OccurredAtMS <= 0 || payload.OccurredAtMS > decimal.MaxExactInteger {
			return fmt.Errorf("%w: invalid founder_advanced payload", ErrInvalidStream)
		}
	default:
		return fmt.Errorf("%w: unknown event kind %q", ErrInvalidStream, event.Kind)
	}
	return nil
}

type eventPrestigeTerms struct {
	ReputationDelta    int64         `json:"reputation_delta"`
	NetworkSlotUnlocks []NetworkSlot `json:"network_slot_unlocks"`
	RouteKnowledge     int64         `json:"route_knowledge"`
	CloutReachNote     string        `json:"clout_reach_note"`
}

type eventAssisted struct {
	Commons bool `json:"commons"`
	Advisor bool `json:"advisor"`
}

type eventRunEnded struct {
	FounderID       string             `json:"founder_id"`
	RunID           routeRunID         `json:"run_id"`
	ExitType        string             `json:"exit_type"`
	StartedAtMS     int64              `json:"started_at_ms"`
	EndedAtMS       int64              `json:"ended_at_ms"`
	RTAMS           int64              `json:"rta_ms"`
	AttendedMS      int64              `json:"attended_ms"`
	TerminalSeq     int64              `json:"terminal_seq"`
	Payout          eventPrestigeTerms `json:"payout"`
	Tier            int64              `json:"tier"`
	LifetimeValue   string             `json:"lifetime_value"`
	LedgerFactKinds []string           `json:"ledger_fact_kinds"`
	ExecutedRoutes  []string           `json:"executed_routes"`
	Assisted        eventAssisted      `json:"assisted"`
}

func validatePrestigeTerms(terms eventPrestigeTerms) error {
	if terms.ReputationDelta < 0 || terms.ReputationDelta > decimal.MaxExactInteger || terms.RouteKnowledge < 0 ||
		terms.RouteKnowledge > decimal.MaxExactInteger || terms.CloutReachNote == "" {
		return ErrInvalidStream
	}
	last := ""
	for _, slot := range terms.NetworkSlotUnlocks {
		if !mechanicalIDPattern.MatchString(slot.Slot) || !mechanicalIDPattern.MatchString(slot.CarriedRef) || slot.Slot <= last {
			return ErrInvalidStream
		}
		last = slot.Slot
	}
	return nil
}

func sortedMechanicalIDs(values []string) bool {
	last := ""
	for _, value := range values {
		if !mechanicalIDPattern.MatchString(value) || value <= last {
			return false
		}
		last = value
	}
	return true
}

func validHealthBand(value string) bool {
	return value == "collapsed" || value == "strained" || value == "healthy"
}

type routeRunID struct {
	CompanyStreamID string `json:"company_stream_id"`
	RunSeq          int64  `json:"run_seq"`
}

func validRouteRunID(run routeRunID) bool {
	return uuidPattern.MatchString(run.CompanyStreamID) && run.RunSeq > 0 && run.RunSeq <= decimal.MaxExactInteger
}

func decodeStrictJSON(data []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	return ensureJSONEnd(decoder)
}

func cloneRaw(source []byte) json.RawMessage {
	return append(json.RawMessage(nil), source...)
}

func normalizeJSON(source []byte) (json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	if err := ensureJSONEnd(decoder); err != nil {
		return nil, err
	}
	return json.Marshal(value)
}
