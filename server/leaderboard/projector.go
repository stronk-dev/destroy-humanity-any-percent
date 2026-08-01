package leaderboard

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"time"

	"cloud-clicker/server/routes"
)

type QueueProjector struct{}

func NewQueueProjector() *QueueProjector { return &QueueProjector{} }

type terminalRun struct {
	FounderID                string          `json:"founder_id"`
	RunID                    terminalRunID   `json:"run_id"`
	ExitType                 string          `json:"exit_type"`
	StartedAtMS              int64           `json:"started_at_ms"`
	EndedAtMS                int64           `json:"ended_at_ms"`
	RTAMS                    int64           `json:"rta_ms"`
	AttendedMS               int64           `json:"attended_ms"`
	PreTimer                 bool            `json:"pre_timer"`
	TerminalSeq              int64           `json:"terminal_seq"`
	Payout                   json.RawMessage `json:"payout"`
	Tier                     int64           `json:"tier"`
	LifetimeValue            string          `json:"lifetime_value"`
	LedgerFactKinds          []string        `json:"ledger_fact_kinds"`
	ExecutedRoutes           []string        `json:"executed_routes"`
	GatesCrossed             *[]string       `json:"gates_crossed"`
	GeneratorsPurchasedTotal *int64          `json:"generators_purchased_total"`
	Assisted                 struct {
		Commons bool `json:"commons"`
		Advisor bool `json:"advisor"`
	} `json:"assisted"`
	Faction *string `json:"faction"`
}

type terminalRunID struct {
	CompanyStreamID string `json:"company_stream_id"`
	RunSeq          int64  `json:"run_seq"`
}

// ProjectVerifiedRun implements replayverify.Projector. The caller owns tx,
// and the projection claim plus every matching category row therefore commit
// atomically with the verification queue's token-checked mark.
func (projector *QueueProjector) ProjectVerifiedRun(ctx context.Context, tx *sql.Tx, streamID string, runSeq int64) error {
	if projector == nil || tx == nil || !uuidPattern.MatchString(streamID) || runSeq < 1 {
		return ErrInvalidEpoch
	}
	terminal, eventID, occurredAt, err := loadTerminalRun(ctx, tx, streamID, runSeq)
	if err != nil {
		return err
	}
	var epochID int64
	var constantsHash string
	if err := tx.QueryRowContext(ctx, `SELECT epoch_id,constants_hash FROM run_epochs WHERE company_stream_id=$1 AND run_seq=$2`, streamID, runSeq).Scan(&epochID, &constantsHash); err != nil {
		return err
	}
	var imported, drifted bool
	if err := tx.QueryRowContext(ctx, `SELECT imported FROM account_founders WHERE founder_id=$1`, terminal.FounderID).Scan(&imported); err != nil {
		return err
	}
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM run_version_drift WHERE company_stream_id=$1 AND run_seq=$2)`, streamID, runSeq).Scan(&drifted); err != nil {
		return err
	}
	if imported || drifted {
		claimed, err := claimProjection(ctx, tx, eventID)
		if err != nil || claimed {
			return err
		}
		return verifyProjectedCategories(ctx, tx, eventID, nil)
	}
	catalog, err := loadPinnedCategoryCatalog(ctx, tx, constantsHash)
	if err != nil {
		return err
	}
	faction, glitched, err := scanRunVariables(ctx, tx, streamID, runSeq)
	if err != nil {
		return err
	}
	if !sameOptionalString(faction, terminal.Faction) || glitched != (len(terminal.ExecutedRoutes) != 0) {
		return fmt.Errorf("%w: terminal variables disagree with event history", ErrInvalidEpoch)
	}
	matching, err := catalog.Matching(TerminalFacts{GatesCrossed: *terminal.GatesCrossed,
		Facts: terminal.LedgerFactKinds, GeneratorsPurchasedTotal: *terminal.GeneratorsPurchasedTotal})
	if err != nil {
		return err
	}
	variables := Variables{Commons: terminal.Assisted.Commons, Advisor: terminal.Assisted.Advisor, Glitched: glitched, Faction: faction}
	encodedVariables, _ := json.Marshal(variables)
	runID := streamID + ":" + strconv.FormatInt(runSeq, 10)
	expected := make([]string, 0, len(matching))
	for _, category := range matching {
		if !terminal.PreTimer || category.Timer == TimerNone {
			expected = append(expected, category.ID)
		}
	}
	claimed, err := claimProjection(ctx, tx, eventID)
	if err != nil {
		return err
	}
	if !claimed {
		return verifyProjectedCategories(ctx, tx, eventID, expected)
	}
	for _, category := range matching {
		if terminal.PreTimer && category.Timer != TimerNone {
			continue
		}
		run := VerifiedRun{EventID: eventID, RunID: runID, FounderID: terminal.FounderID, CategoryID: category.ID,
			Variables: variables, EpochID: epochID, MandateLevel: 0, VerifiedAt: occurredAt}
		switch category.Timer {
		case TimerRTA:
			key := terminal.RTAMS
			run.KeyMS = &key
		case TimerAttended:
			key := terminal.AttendedMS
			run.KeyMS = &key
		case TimerNone:
			key, err := magnitudeKeyFromCanonical(terminal.LifetimeValue)
			if err != nil {
				return err
			}
			run.KeyMagnitude = &key
		default:
			return ErrInvalidEpoch
		}
		if _, err := insertBoardRowTx(ctx, tx, run, encodedVariables); err != nil {
			return err
		}
	}
	return nil
}

func loadPinnedCategoryCatalog(ctx context.Context, tx *sql.Tx, constantsHash string) (*CategoryCatalog, error) {
	rows, err := tx.QueryContext(ctx, `SELECT artifact_name,bytes FROM catalog_artifacts
		WHERE constants_hash=$1 AND artifact_name IN ('categories','routes') ORDER BY artifact_name`, constantsHash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	artifacts := map[string][]byte{}
	for rows.Next() {
		var name string
		var data []byte
		if err := rows.Scan(&name, &data); err != nil {
			return nil, err
		}
		artifacts[name] = bytes.Clone(data)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(artifacts) != 2 {
		return nil, fmt.Errorf("%w: pinned category artifacts", ErrInvalidEpoch)
	}
	routeCatalog, err := routes.LoadCatalog(artifacts["routes"])
	if err != nil {
		return nil, fmt.Errorf("%w: pinned routes catalog", ErrInvalidEpoch)
	}
	gates := routeCatalog.Gates()
	gateIDs := make([]string, len(gates))
	for index, gate := range gates {
		gateIDs[index] = gate.ID
	}
	sort.Strings(gateIDs)
	return LoadCategoryCatalog(artifacts["categories"], gateIDs)
}

func verifyProjectedCategories(ctx context.Context, tx *sql.Tx, eventID string, expected []string) error {
	rows, err := tx.QueryContext(ctx, `SELECT category_id FROM verified_runs WHERE event_id=$1 ORDER BY category_id`, eventID)
	if err != nil {
		return err
	}
	defer rows.Close()
	actual := []string{}
	for rows.Next() {
		var categoryID string
		if err := rows.Scan(&categoryID); err != nil {
			return err
		}
		actual = append(actual, categoryID)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if !sameStrings(actual, expected) {
		return fmt.Errorf("%w: incomplete prior projection", ErrInvalidEpoch)
	}
	return nil
}

func loadTerminalRun(ctx context.Context, tx *sql.Tx, streamID string, runSeq int64) (terminalRun, string, time.Time, error) {
	rows, err := tx.QueryContext(ctx, `SELECT event_id,schema_version,occurred_at,payload::text
		FROM events WHERE stream_id=$1 AND kind='run_ended'
		  AND payload->'run_id'->>'company_stream_id'=$1::text
		  AND payload->'run_id'->>'run_seq'=$2::bigint::text
		ORDER BY event_seq`, streamID, runSeq)
	if err != nil {
		return terminalRun{}, "", time.Time{}, err
	}
	defer rows.Close()
	var terminal terminalRun
	var eventID, payload string
	var occurredAt time.Time
	var schema int
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return terminalRun{}, "", time.Time{}, err
		}
		return terminalRun{}, "", time.Time{}, fmt.Errorf("%w: missing run_ended", ErrInvalidEpoch)
	}
	if err := rows.Scan(&eventID, &schema, &occurredAt, &payload); err != nil {
		return terminalRun{}, "", time.Time{}, err
	}
	if rows.Next() {
		return terminalRun{}, "", time.Time{}, fmt.Errorf("%w: duplicate run_ended", ErrInvalidEpoch)
	}
	if err := rows.Err(); err != nil {
		return terminalRun{}, "", time.Time{}, err
	}
	if schema != 2 || decodeProjectorStrict([]byte(payload), &terminal) != nil || terminal.RunID.CompanyStreamID != streamID || terminal.RunID.RunSeq != runSeq ||
		terminal.FounderID == "" || terminal.TerminalSeq < 1 || terminal.RTAMS < 0 || terminal.AttendedMS < 0 || terminal.AttendedMS > terminal.RTAMS ||
		!sortedUniqueMechanical(terminal.LedgerFactKinds) || !sortedUniqueMechanical(terminal.ExecutedRoutes) || terminal.GatesCrossed == nil ||
		!sortedUniqueMechanical(*terminal.GatesCrossed) || terminal.GeneratorsPurchasedTotal == nil || *terminal.GeneratorsPurchasedTotal < 0 {
		return terminalRun{}, "", time.Time{}, fmt.Errorf("%w: invalid run_ended", ErrInvalidEpoch)
	}
	var maxSeq sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT max(seq) FROM run_log WHERE company_stream_id=$1 AND run_seq=$2`, streamID, runSeq).Scan(&maxSeq); err != nil {
		return terminalRun{}, "", time.Time{}, err
	}
	if !maxSeq.Valid || maxSeq.Int64 != terminal.TerminalSeq {
		return terminalRun{}, "", time.Time{}, fmt.Errorf("%w: terminal sequence", ErrInvalidEpoch)
	}
	return terminal, eventID, occurredAt.UTC(), nil
}

func scanRunVariables(ctx context.Context, tx *sql.Tx, streamID string, runSeq int64) (*string, bool, error) {
	rows, err := tx.QueryContext(ctx, `SELECT kind,payload::text FROM events
		WHERE stream_id=$1 AND kind IN ('route_executed','incorporated')
		  AND payload->'run_id'->>'company_stream_id'=$1::text
		  AND payload->'run_id'->>'run_seq'=$2::bigint::text
		ORDER BY event_seq`, streamID, runSeq)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	var faction *string
	glitched := false
	for rows.Next() {
		var kind, payload string
		if err := rows.Scan(&kind, &payload); err != nil {
			return nil, false, err
		}
		if kind == "route_executed" {
			glitched = true
			continue
		}
		var incorporated struct {
			FactionID string `json:"faction_id"`
		}
		if json.Unmarshal([]byte(payload), &incorporated) != nil || !mechanicalPattern.MatchString(incorporated.FactionID) || faction != nil && *faction != incorporated.FactionID {
			return nil, false, fmt.Errorf("%w: faction event", ErrInvalidEpoch)
		}
		value := incorporated.FactionID
		faction = &value
	}
	return faction, glitched, rows.Err()
}

func claimProjection(ctx context.Context, tx *sql.Tx, eventID string) (bool, error) {
	result, err := tx.ExecContext(ctx, `INSERT INTO verification_projection_events(event_id) VALUES($1) ON CONFLICT DO NOTHING`, eventID)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected == 1, err
}

func sameOptionalString(left, right *string) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

func decodeProjectorStrict(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return ErrInvalidEpoch
	}
	return nil
}
