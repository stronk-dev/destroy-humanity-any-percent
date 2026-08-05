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
	"cloud-clicker/server/pet"
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
	EventUpgradePurchased          EventKind = "upgrade_purchased"
	EventInvariantReported         EventKind = "invariant_reported"
	EventCompensation              EventKind = "compensation"
	EventGateCrossed               EventKind = "gate_crossed"
	EventRouteExecuted             EventKind = "route_executed"
	EventRouteHintPurchased        EventKind = "route_hint_purchased"
	EventRouteKnowledgeGranted     EventKind = "route_knowledge_granted"
	EventCompactSigned             EventKind = "compact_signed"
	EventCompactTitheRaised        EventKind = "compact_tithe_raised"
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
	EventIncorporated              EventKind = "incorporated"
	EventFactionStockSaturated     EventKind = "faction_stock_saturated"
	EventGuildTitheAccrued         EventKind = "guild_tithe_accrued"
	EventGuildActivityEvaluated    EventKind = "guild_activity_evaluated"
	EventMeterBandChanged          EventKind = "meter_band_changed.v1"
	EventAchievementEarned         EventKind = "achievement_earned.v1"
	EventPetCareApplied            EventKind = "pet_care_applied.v1"
	EventPetStatusChanged          EventKind = "pet_status_changed.v1"
	EventMinigameResolved          EventKind = "minigame_resolved.v1"
	EventMinigameRatingChanged     EventKind = "minigame_rating_changed.v1"
)

// AllEventKinds is the closed structural authority consumed by catalog
// validators. It grows only with the event registry and is never balance data.
var AllEventKinds = [...]EventKind{
	EventCompactCascadeStarted, EventCompactHealthBandChanged, EventCompactLeft,
	EventCompactRecovered, EventCompactRecruitmentOffered, EventCompactSampled,
	EventCompactSigned, EventCompactTitheRaised, EventCompensation,
	EventExitOfferDeclined, EventExitOfferExpired, EventExitOfferSpawned,
	EventFactionStockSaturated, EventFounderAdvanced, EventGateCrossed,
	EventGeneratorPurchased, EventGuildActivityEvaluated, EventGuildTitheAccrued,
	EventIncorporated, EventInvariantReported, EventRouteExecuted,
	EventRouteHintPurchased, EventRouteKnowledgeGranted, EventRunEnded,
	EventRunStarted, EventUpgradePurchased, EventMeterBandChanged,
	EventAchievementEarned,
	EventPetCareApplied, EventPetStatusChanged,
	EventMinigameResolved, EventMinigameRatingChanged,
}

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
	// ActionDebits is deterministic transition evidence used by in-process
	// post-action hooks. It is deliberately not a wire or persistence surface:
	// replay recomputes it from the same transition authority.
	ActionDebits map[string]string
}

type IntentMutation func(state *State, revision Revision) (IntentDecision, error)

// LoggedIntentMutation owns the immutable inputs consumed by a replayable
// Company transition. The Store supplies the authoritative command envelope;
// the mutation returns the exact closed replay-input object persisted beside
// the canonical player payload and computed receipt.
type LoggedIntentMutation func(state *State, revision Revision, command ReplayCommand) (IntentDecision, json.RawMessage, error)

// FounderLoggedIntentMutation is the Founder-scope mirror. It receives no
// Company/run coordinates and must return the feature-owned resolved-input
// object inside the persistence-owned Founder replay envelope.
type FounderLoggedIntentMutation func(state *State, revision Revision, command FounderReplayCommand) (IntentDecision, json.RawMessage, error)

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
	return s.applyIntent(ctx, streamID, expectedRevision, intentID, requestHash, nil, mutate, nil, nil)
}

func (s *Store) ApplyIntentLogged(
	ctx context.Context,
	streamID string,
	expectedRevision int64,
	intentID string,
	requestHash string,
	canonicalPayload []byte,
	mutate LoggedIntentMutation,
) (IntentResult, error) {
	if err := validateCanonicalPayload(canonicalPayload, requestHash); err != nil {
		return IntentResult{}, err
	}
	return s.applyIntent(ctx, streamID, expectedRevision, intentID, requestHash, canonicalPayload, nil, mutate, nil)
}

func (s *Store) ApplyFounderLogged(
	ctx context.Context,
	streamID string,
	expectedRevision int64,
	intentID string,
	requestHash string,
	canonicalPayload []byte,
	mutate FounderLoggedIntentMutation,
) (IntentResult, error) {
	if err := validateCanonicalPayload(canonicalPayload, requestHash); err != nil {
		return IntentResult{}, err
	}
	return s.applyIntent(ctx, streamID, expectedRevision, intentID, requestHash, canonicalPayload, nil, nil, mutate)
}

func (s *Store) applyIntent(
	ctx context.Context,
	streamID string,
	expectedRevision int64,
	intentID string,
	requestHash string,
	canonicalPayload []byte,
	mutate IntentMutation,
	loggedMutate LoggedIntentMutation,
	founderLoggedMutate FounderLoggedIntentMutation,
) (IntentResult, error) {
	mutationCount := 0
	for _, present := range []bool{mutate != nil, loggedMutate != nil, founderLoggedMutate != nil} {
		if present {
			mutationCount++
		}
	}
	if !uuidPattern.MatchString(streamID) || expectedRevision < 1 || !uuidV7Pattern.MatchString(intentID) ||
		!hashPattern.MatchString(requestHash) || mutationCount != 1 {
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
	if mutate != nil && scope == economy.ScopeFounder || loggedMutate != nil && scope != economy.ScopeCompany ||
		founderLoggedMutate != nil && scope != economy.ScopeFounder {
		return IntentResult{}, ErrInvalidStream
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
	var command ReplayCommand
	if scope == economy.ScopeCompany && len(canonicalPayload) != 0 {
		if err := requireRunEpochTx(ctx, tx, streamID, state.RunSeq, revision.ConstantsHash); err != nil {
			return IntentResult{}, err
		}
		runLogSequence, err = nextRunLogSequence(ctx, tx, streamID, state.RunSeq)
		if err != nil {
			return IntentResult{}, err
		}
		revision.RunLogSequence = runLogSequence
		command = ReplayCommand{
			IntentID: intentID, CompanyStreamID: streamID, FounderID: ownerID,
			Revision: revision.Number, RunSeq: state.RunSeq, RunLogSeq: runLogSequence,
		}
	}
	var founderLogSequence int64
	var founderCommand FounderReplayCommand
	if scope == economy.ScopeFounder && founderLoggedMutate != nil {
		founderLogSequence, err = nextFounderLogSequence(ctx, tx, streamID)
		if err != nil {
			return IntentResult{}, err
		}
		serverTSMS, timestampErr := founderServerTimestamp(ctx, tx)
		if timestampErr != nil {
			return IntentResult{}, timestampErr
		}
		founderCommand = FounderReplayCommand{IntentID: intentID, FounderStreamID: streamID,
			FounderID: ownerID, Revision: revision.Number, FounderLogSeq: founderLogSequence,
			ServerTSMS: serverTSMS}
		if founderLogSequence == 1 {
			if err := InsertFounderGenesisTx(ctx, tx, FounderGenesis{FounderStreamID: streamID,
				Revision: revision.Number, State: stateBytes, Version: revision.Version,
				ConstantsHash: revision.ConstantsHash}); err != nil {
				return IntentResult{}, err
			}
		}
	}
	var decision IntentDecision
	var replayInputs json.RawMessage
	if loggedMutate != nil {
		decision, replayInputs, err = loggedMutate(state, revision, command)
	} else if founderLoggedMutate != nil {
		decision, replayInputs, err = founderLoggedMutate(state, revision, founderCommand)
	} else {
		decision, err = mutate(state, revision)
	}
	if err != nil {
		return IntentResult{}, err
	}
	if runLogSequence != 0 {
		replayInputs, err = ValidateReplayInputs(replayInputs, command)
		if err != nil {
			return IntentResult{}, err
		}
	}
	if founderLogSequence != 0 {
		replayInputs, err = ValidateFounderReplayInputs(replayInputs, founderCommand)
		if err != nil {
			return IntentResult{}, err
		}
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
			if err := insertRunLog(ctx, tx, streamID, state.RunSeq, runLogSequence, intentID, canonicalPayload, replayInputs, decision.Receipt, nil); err != nil {
				return IntentResult{}, err
			}
		}
		if founderLogSequence != 0 {
			if err := insertFounderLog(ctx, tx, founderCommand, revision.ConstantsHash, canonicalPayload,
				replayInputs, decision.Receipt, nil, nil); err != nil {
				return IntentResult{}, err
			}
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO intent_records (stream_id,intent_id,request_hash,outcome,receipt)
			VALUES ($1,$2,$3,$4,$5)`, streamID, intentID, requestHash, decision.Outcome, decision.Receipt); err != nil {
			return IntentResult{}, err
		}
		if scope == economy.ScopeCompany || scope == economy.ScopeFounder {
			if err := insertReceiptOutbox(ctx, tx, ownerID, streamID, intentID, scope, revision.Number, revision.ConstantsHash, decision.Receipt); err != nil {
				return IntentResult{}, err
			}
		}
		if err := tx.Commit(); err != nil {
			return IntentResult{}, err
		}
		return IntentResult{Outcome: decision.Outcome, Receipt: cloneRaw(decision.Receipt)}, nil
	}
	if VersionForState(state) != migratedWriteVersion(revision.Version) {
		return IntentResult{}, fmt.Errorf("%w: ordinary intent cannot change save version", ErrInvalidState)
	}

	encodedState, err := s.validatedState(revision.ConstantsHash, scope, state)
	if err != nil {
		return IntentResult{}, err
	}
	newRevision := revision.Number + 1
	version := VersionForState(state)
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO save_revisions (stream_id,revision,version,state,constants_hash)
		VALUES ($1,$2,$3,$4,$5)`, streamID, newRevision, version, encodedState, revision.ConstantsHash); err != nil {
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
		if err := insertRunLog(ctx, tx, streamID, state.RunSeq, runLogSequence, intentID, canonicalPayload, replayInputs, decision.Receipt, &newRevision); err != nil {
			return IntentResult{}, err
		}
	}
	if founderLogSequence != 0 {
		if err := insertFounderLog(ctx, tx, founderCommand, revision.ConstantsHash, canonicalPayload,
			replayInputs, decision.Receipt, &newRevision, nil); err != nil {
			return IntentResult{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO intent_records (stream_id,intent_id,request_hash,outcome,receipt)
		VALUES ($1,$2,$3,$4,$5)`, streamID, intentID, requestHash, decision.Outcome, decision.Receipt); err != nil {
		return IntentResult{}, err
	}
	if scope == economy.ScopeCompany || scope == economy.ScopeFounder {
		if err := insertReceiptOutbox(ctx, tx, ownerID, streamID, intentID, scope, newRevision, revision.ConstantsHash, decision.Receipt); err != nil {
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
		if !validEventSchemaVersion(event) || event.IntentID != "" && !uuidV7Pattern.MatchString(event.IntentID) {
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

func validEventSchemaVersion(event EventWrite) bool {
	if event.Kind == EventRunEnded {
		return event.SchemaVersion == 2
	}
	return event.SchemaVersion == 1
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
	case EventUpgradePurchased:
		var payload struct {
			UpgradeID      string `json:"upgrade_id"`
			CostResourceID string `json:"cost_resource_id"`
			Cost           string `json:"cost"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !mechanicalIDPattern.MatchString(payload.UpgradeID) || !mechanicalIDPattern.MatchString(payload.CostResourceID) {
			return fmt.Errorf("%w: invalid upgrade_purchased payload", ErrInvalidStream)
		}
		if value, err := decimal.ParseCanonical(payload.Cost); err != nil || !value.Gt(decimal.Zero) {
			return fmt.Errorf("%w: invalid upgrade_purchased cost", ErrInvalidStream)
		}
	case EventMeterBandChanged:
		var payload struct {
			RunID       routeRunID `json:"run_id"`
			MeterID     string     `json:"meter_id"`
			FromBand    string     `json:"from_band"`
			ToBand      string     `json:"to_band"`
			Direction   string     `json:"direction"`
			ValueBefore int        `json:"value_before"`
			ValueAfter  int        `json:"value_after"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !validRouteRunID(payload.RunID) ||
			!mechanicalIDPattern.MatchString(payload.MeterID) || !mechanicalIDPattern.MatchString(payload.FromBand) ||
			!mechanicalIDPattern.MatchString(payload.ToBand) || payload.FromBand == payload.ToBand ||
			(payload.Direction != "up" && payload.Direction != "down") || payload.ValueBefore < 0 || payload.ValueBefore > 100 ||
			payload.ValueAfter < 0 || payload.ValueAfter > 100 || payload.Direction == "up" && payload.ValueAfter <= payload.ValueBefore ||
			payload.Direction == "down" && payload.ValueAfter >= payload.ValueBefore {
			return fmt.Errorf("%w: invalid meter_band_changed.v1 payload", ErrInvalidStream)
		}
	case EventAchievementEarned:
		var payload struct {
			RunID          routeRunID `json:"run_id"`
			AchievementID  string     `json:"achievement_id"`
			ConditionScope string     `json:"condition_scope"`
			ScoreGrant     int64      `json:"score_grant"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !validRouteRunID(payload.RunID) ||
			!mechanicalIDPattern.MatchString(payload.AchievementID) || (payload.ConditionScope != "run" && payload.ConditionScope != "career") ||
			payload.ScoreGrant < 1 || payload.ScoreGrant > decimal.MaxExactInteger {
			return fmt.Errorf("%w: invalid achievement_earned.v1 payload", ErrInvalidStream)
		}
	case EventPetCareApplied:
		var payload struct {
			PetID                  string         `json:"pet_id"`
			ActionID               string         `json:"action_id"`
			StatID                 pet.StatID     `json:"stat_id"`
			BeforePPM              int64          `json:"before_ppm"`
			AppliedPPM             int64          `json:"applied_ppm"`
			AfterPPM               int64          `json:"after_ppm"`
			TrustBeforePPM         int64          `json:"trust_before_ppm"`
			TrustAfterPPM          int64          `json:"trust_after_ppm"`
			Mood                   pet.Mood       `json:"mood"`
			StatusBand             pet.StatusBand `json:"status_band"`
			NextEligibleAttendedMS int64          `json:"next_eligible_attended_ms"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidPattern.MatchString(payload.PetID) ||
			!mechanicalIDPattern.MatchString(payload.ActionID) || !pet.ValidStatID(payload.StatID) ||
			payload.BeforePPM < 0 || payload.BeforePPM > 1_000_000 || payload.AppliedPPM < 1 ||
			payload.AppliedPPM > 1_000_000-payload.BeforePPM || payload.AfterPPM != payload.BeforePPM+payload.AppliedPPM ||
			payload.TrustBeforePPM < 0 || payload.TrustBeforePPM > 1_000_000 ||
			payload.TrustAfterPPM < payload.TrustBeforePPM || payload.TrustAfterPPM > 1_000_000 ||
			!pet.ValidMood(payload.Mood) || !pet.ValidStatusBand(payload.StatusBand) ||
			payload.NextEligibleAttendedMS < 0 || payload.NextEligibleAttendedMS > decimal.MaxExactInteger {
			return fmt.Errorf("%w: invalid pet_care_applied.v1 payload", ErrInvalidStream)
		}
	case EventPetStatusChanged:
		var payload struct {
			PetID          string         `json:"pet_id"`
			FromStatusBand pet.StatusBand `json:"from_status_band"`
			ToStatusBand   pet.StatusBand `json:"to_status_band"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidPattern.MatchString(payload.PetID) ||
			!pet.ValidStatusBand(payload.FromStatusBand) || !pet.ValidStatusBand(payload.ToStatusBand) ||
			payload.FromStatusBand == payload.ToStatusBand {
			return fmt.Errorf("%w: invalid pet_status_changed.v1 payload", ErrInvalidStream)
		}
	case EventMinigameResolved:
		var payload struct {
			SessionID                 string `json:"session_id"`
			MinigameID                string `json:"minigame_id"`
			CertifiedResultHash       string `json:"certified_result_hash"`
			CreditedResourceID        string `json:"credited_resource_id"`
			CreditedDelta             string `json:"credited_delta"`
			ConfiguredCapForfeitUnits int64  `json:"configured_cap_forfeit_units"`
			CapReasonKey              string `json:"cap_reason_key"`
			FounderRevision           int64  `json:"founder_revision"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidV7Pattern.MatchString(payload.SessionID) ||
			!mechanicalIDPattern.MatchString(payload.MinigameID) || !hashPattern.MatchString(payload.CertifiedResultHash) ||
			!mechanicalIDPattern.MatchString(payload.CreditedResourceID) || payload.ConfiguredCapForfeitUnits < 0 ||
			payload.ConfiguredCapForfeitUnits > decimal.MaxExactInteger || payload.FounderRevision < 1 ||
			payload.FounderRevision > decimal.MaxExactInteger || payload.ConfiguredCapForfeitUnits == 0 && payload.CapReasonKey != "" ||
			payload.ConfiguredCapForfeitUnits > 0 && !mechanicalIDPattern.MatchString(payload.CapReasonKey) {
			return fmt.Errorf("%w: invalid minigame_resolved.v1 payload", ErrInvalidStream)
		}
		if value, err := decimal.ParseCanonical(payload.CreditedDelta); err != nil || value.Lt(decimal.Zero) {
			return fmt.Errorf("%w: invalid minigame credited delta", ErrInvalidStream)
		}
	case EventMinigameRatingChanged:
		var payload struct {
			SessionID           string                      `json:"session_id"`
			MinigameID          string                      `json:"minigame_id"`
			CertifiedResultHash string                      `json:"certified_result_hash"`
			OldElo              int64                       `json:"old_elo"`
			NewElo              int64                       `json:"new_elo"`
			SeasonMember        string                      `json:"season_member"`
			OldQuality          MinigameOfflineQualityState `json:"old_quality"`
			NewQuality          MinigameOfflineQualityState `json:"new_quality"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidV7Pattern.MatchString(payload.SessionID) ||
			!mechanicalIDPattern.MatchString(payload.MinigameID) || !hashPattern.MatchString(payload.CertifiedResultHash) ||
			payload.OldElo < -decimal.MaxExactInteger || payload.OldElo > decimal.MaxExactInteger ||
			payload.NewElo < -decimal.MaxExactInteger || payload.NewElo > decimal.MaxExactInteger ||
			!mechanicalIDPattern.MatchString(payload.SeasonMember) || !validMinigameQualityEvent(payload.OldQuality) ||
			!validMinigameQualityEvent(payload.NewQuality) {
			return fmt.Errorf("%w: invalid minigame_rating_changed.v1 payload", ErrInvalidStream)
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
	case EventCompactTitheRaised:
		var payload struct {
			FounderID     string     `json:"founder_id"`
			RunID         routeRunID `json:"run_id"`
			PriorTithePPM int64      `json:"prior_tithe_ppm"`
			NewTithePPM   int64      `json:"new_tithe_ppm"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidPattern.MatchString(payload.FounderID) || !validRouteRunID(payload.RunID) ||
			payload.PriorTithePPM < 0 || payload.PriorTithePPM > 1_000_000 || payload.NewTithePPM < payload.PriorTithePPM || payload.NewTithePPM > 1_000_000 {
			return fmt.Errorf("%w: invalid compact tithe raise payload", ErrInvalidStream)
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
	case EventExitOfferExpired:
		var payload struct {
			OfferID string `json:"offer_id"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidV7Pattern.MatchString(payload.OfferID) {
			return fmt.Errorf("%w: invalid exit offer transition payload", ErrInvalidStream)
		}
	case EventExitOfferDeclined:
		var payload struct {
			OfferID string `json:"offer_id"`
			RunSeq  int64  `json:"run_seq"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidV7Pattern.MatchString(payload.OfferID) ||
			payload.RunSeq < 1 || payload.RunSeq > decimal.MaxExactInteger {
			return fmt.Errorf("%w: invalid exit offer decline payload", ErrInvalidStream)
		}
	case EventRunEnded:
		var payload eventRunEnded
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidPattern.MatchString(payload.FounderID) ||
			!validRouteRunID(payload.RunID) || !validExitType(payload.ExitType) || payload.StartedAtMS <= 0 ||
			payload.EndedAtMS < payload.StartedAtMS || payload.EndedAtMS > decimal.MaxExactInteger ||
			payload.RTAMS != payload.EndedAtMS-payload.StartedAtMS || payload.AttendedMS < 0 || payload.AttendedMS > payload.RTAMS ||
			payload.TerminalSeq <= 0 || payload.TerminalSeq > decimal.MaxExactInteger || payload.Tier < 0 || payload.Tier > 9 ||
			validatePrestigeTerms(payload.Payout) != nil || !sortedMechanicalIDs(payload.LedgerFactKinds) || !sortedMechanicalIDs(payload.ExecutedRoutes) ||
			payload.GatesCrossed == nil || !sortedMechanicalIDs(*payload.GatesCrossed) || payload.GeneratorsPurchasedTotal == nil ||
			*payload.GeneratorsPurchasedTotal < 0 || *payload.GeneratorsPurchasedTotal > decimal.MaxExactInteger ||
			payload.Faction != nil && !mechanicalIDPattern.MatchString(*payload.Faction) {
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
	case EventIncorporated:
		var payload struct {
			FounderID         string     `json:"founder_id"`
			RunID             routeRunID `json:"run_id"`
			FactionID         string     `json:"faction_id"`
			StockResource     string     `json:"stock_resource"`
			IncorporatedAtMS  int64      `json:"incorporated_at_ms"`
			CompactAutoSigned bool       `json:"compact_auto_signed"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidPattern.MatchString(payload.FounderID) ||
			!validRouteRunID(payload.RunID) || !mechanicalIDPattern.MatchString(payload.FactionID) ||
			!mechanicalIDPattern.MatchString(payload.StockResource) || payload.IncorporatedAtMS <= 0 || payload.IncorporatedAtMS > decimal.MaxExactInteger {
			return fmt.Errorf("%w: invalid incorporated payload", ErrInvalidStream)
		}
	case EventFactionStockSaturated:
		var payload struct {
			FactionID     string `json:"faction_id"`
			StockResource string `json:"stock_resource"`
			StockCap      int64  `json:"stock_cap"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !mechanicalIDPattern.MatchString(payload.FactionID) ||
			!mechanicalIDPattern.MatchString(payload.StockResource) || payload.StockCap <= 0 || payload.StockCap > decimal.MaxExactInteger {
			return fmt.Errorf("%w: invalid faction_stock_saturated payload", ErrInvalidStream)
		}
	case EventGuildTitheAccrued, EventGuildActivityEvaluated:
		var payload struct {
			FounderID        string     `json:"founder_id"`
			RunID            routeRunID `json:"run_id"`
			ProgressDeltaPPM int64      `json:"progress_delta_ppm"`
			XPDelta          int64      `json:"xp_delta"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidPattern.MatchString(payload.FounderID) ||
			!validRouteRunID(payload.RunID) || payload.ProgressDeltaPPM < 0 || payload.ProgressDeltaPPM > 1_000_000 ||
			payload.XPDelta < 0 || payload.XPDelta > decimal.MaxExactInteger ||
			(event.Kind == EventGuildTitheAccrued && payload.XPDelta == 0) ||
			(event.Kind == EventGuildActivityEvaluated && payload.XPDelta != 0) {
			return fmt.Errorf("%w: invalid guild_tithe_accrued payload", ErrInvalidStream)
		}
	default:
		return fmt.Errorf("%w: unknown event kind %q", ErrInvalidStream, event.Kind)
	}
	return nil
}

func validMinigameQualityEvent(value MinigameOfflineQualityState) bool {
	return value.GradePPM >= 0 && value.GradePPM <= 1_000_000 && value.LastFounderAttendedMS >= 0 &&
		value.LastFounderAttendedMS <= decimal.MaxExactInteger && value.DecayRemainderPPM >= 0 && value.DecayRemainderPPM < 1_000_000
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
	FounderID                string             `json:"founder_id"`
	RunID                    routeRunID         `json:"run_id"`
	ExitType                 string             `json:"exit_type"`
	StartedAtMS              int64              `json:"started_at_ms"`
	EndedAtMS                int64              `json:"ended_at_ms"`
	RTAMS                    int64              `json:"rta_ms"`
	AttendedMS               int64              `json:"attended_ms"`
	PreTimer                 bool               `json:"pre_timer"`
	TerminalSeq              int64              `json:"terminal_seq"`
	Payout                   eventPrestigeTerms `json:"payout"`
	Tier                     int64              `json:"tier"`
	LifetimeValue            string             `json:"lifetime_value"`
	LedgerFactKinds          []string           `json:"ledger_fact_kinds"`
	ExecutedRoutes           []string           `json:"executed_routes"`
	GatesCrossed             *[]string          `json:"gates_crossed"`
	GeneratorsPurchasedTotal *int64             `json:"generators_purchased_total"`
	Assisted                 eventAssisted      `json:"assisted"`
	Faction                  *string            `json:"faction"`
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
