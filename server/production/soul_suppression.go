package production

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"cloud-clicker/server/guild"
	"cloud-clicker/server/multiplier"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/save"
)

const (
	soulRecoveryResolveKind = "resolve_soul_recovery"
	soulRecoveryCancelKind  = "cancel_soul_recovery"
)

type soulSuppression struct {
	FromEvaluatedMS      int64  `json:"from_evaluated_ms"`
	ToEvaluatedMS        int64  `json:"to_evaluated_ms"`
	FounderAttendedStart int64  `json:"founder_attended_start_ms"`
	FounderAttendedEnd   int64  `json:"founder_attended_end_ms"`
	SessionID            string `json:"session_id"`
}

type soulSuppressionResolved struct {
	Kind        string                      `json:"kind"`
	IntentKind  string                      `json:"intent_kind"`
	Suppression soulSuppression             `json:"suppression"`
	Accrual     replayAccrual               `json:"accrual"`
	ActivePlay  *activePlayScheduleEvidence `json:"active_play"`
}

type soulRecoveryPayload struct {
	Kind      string `json:"kind"`
	SessionID string `json:"session_id"`
}

type SuppressedTransition struct {
	State   *save.State
	Receipt json.RawMessage
	Events  []save.EventWrite
}

// ApplySuppressedLogged is the sole zero-output Company transition boundary.
// It consumes the same frozen accrual inputs as ordinary replay, runs the
// complete time-hook chain so every cursor/watermark advances, then restores
// every production-derived output authority. Meter and achievement hooks are
// deliberately absent: a recovery interval observes no production action.
func ApplySuppressedLogged(state *save.State, canonicalPayload []byte, catalogs CatalogBundle, replayInputs []byte) (SuppressedTransition, error) {
	if state == nil || !catalogs.valid(catalogs.ConstantsHash) {
		return SuppressedTransition{}, fmt.Errorf("%w: suppressed catalog bundle", ErrInvalidReplayInputs)
	}
	wire, err := parseReplayInputs(replayInputs)
	if err != nil {
		return SuppressedTransition{}, err
	}
	var payload soulRecoveryPayload
	if err := decodeReplayStrict(canonicalPayload, &payload); err != nil ||
		(payload.Kind != soulRecoveryResolveKind && payload.Kind != soulRecoveryCancelKind) ||
		payload.SessionID != wire.Command.IntentID {
		return SuppressedTransition{}, fmt.Errorf("%w: suppressed command", ErrInvalidReplayInputs)
	}
	var resolved soulSuppressionResolved
	if err := decodeReplayStrict(wire.Resolved, &resolved); err != nil || resolved.Kind != "soul_recovery_suppression" ||
		resolved.IntentKind != payload.Kind || resolved.Suppression.SessionID != payload.SessionID {
		return SuppressedTransition{}, fmt.Errorf("%w: suppressed resolved input", ErrInvalidReplayInputs)
	}
	suppression := resolved.Suppression
	if suppression.FromEvaluatedMS != state.EvaluatedThrough.UnixMilli() || suppression.ToEvaluatedMS < suppression.FromEvaluatedMS ||
		suppression.FounderAttendedStart < 0 || suppression.FounderAttendedEnd < suppression.FounderAttendedStart ||
		suppression.ToEvaluatedMS != wire.EvaluatedAtMS || wire.EvaluationMode != ModeOnline ||
		wire.Command.RunSeq != state.RunSeq {
		return SuppressedTransition{}, fmt.Errorf("%w: suppressed coordinates", ErrInvalidReplayInputs)
	}
	contributions, err := contributionsFromReplay(resolved.Accrual)
	if err != nil || resolved.Accrual.RouteContextVersion != catalogs.Routes.ContextVersion() ||
		len(resolved.Accrual.GuildSettlementBatch.Settlements) != 0 {
		return SuppressedTransition{}, fmt.Errorf("%w: suppressed accrual inputs", ErrInvalidReplayInputs)
	}
	if state.CompactMember != (resolved.Accrual.CommonsWeightPPM != nil) {
		return SuppressedTransition{}, fmt.Errorf("%w: suppressed commons weight", ErrInvalidReplayInputs)
	}
	if (state.WireVersion == 18) != (resolved.ActivePlay != nil) {
		return SuppressedTransition{}, fmt.Errorf("%w: suppressed active-play evidence", ErrInvalidReplayInputs)
	}
	before, err := cloneReplayState(state, catalogs.Economy)
	if err != nil {
		return SuppressedTransition{}, err
	}
	beforeOutputs, err := suppressionOutputSnapshot(before)
	if err != nil {
		return SuppressedTransition{}, err
	}
	effectiveNow := time.UnixMilli(suppression.ToEvaluatedMS).UTC()
	var scheduleEvents []save.EventWrite
	if resolved.ActivePlay != nil {
		scheduleEvents, err = applyActivePlaySchedule(state, catalogs.Opportunities, catalogs.Prestige,
			wire.Command.FounderID, effectiveNow, *resolved.ActivePlay)
		if err != nil {
			return SuppressedTransition{}, fmt.Errorf("%w: suppressed active-play schedule", ErrInvalidReplayInputs)
		}
		for index := range scheduleEvents {
			scheduleEvents[index].IntentID = wire.Command.IntentID
		}
	}
	if effectiveNow.After(state.EvaluatedThrough) {
		if err := prestigecore.RecordOfflineSpan(state, state.EvaluatedThrough, effectiveNow, catalogs.Prestige.CatchupCeilingMS); err != nil {
			return SuppressedTransition{}, err
		}
	}
	evaluation, err := Evaluate(state, catalogs.Economy, effectiveNow, ModeOnline, contributions)
	if err != nil {
		return SuppressedTransition{}, err
	}
	revision := save.Revision{StreamID: wire.Command.CompanyStreamID, OwnerID: wire.Command.FounderID,
		Number: wire.Command.Revision, ConstantsHash: catalogs.ConstantsHash, RunLogSequence: wire.Command.RunLogSeq}
	if _, err := runAccrualHook(closedReplayAccrualHook(catalogs, resolved.Accrual.CommonsWeightPPM),
		wire.Command.IntentID, state, catalogs.Economy, revision, evaluation, contributions); err != nil {
		return SuppressedTransition{}, err
	}
	// Preserve temporal state produced by Evaluate/the hooks, but restore every
	// output-bearing authority. The elapsed interval is therefore consumed and
	// can never be accrued later.
	state.Ledger = before.Ledger
	state.GeneratorProvisioned = before.GeneratorProvisioned
	state.ProvisionRemaindersPPM = before.ProvisionRemaindersPPM
	state.StockUnits = before.StockUnits
	state.StockProgressMS = before.StockProgressMS
	state.StockRateRemainderPPM = before.StockRateRemainderPPM
	state.ConsumedStockUnits = before.ConsumedStockUnits
	state.GuildTitheCarryPPM = before.GuildTitheCarryPPM
	state.GuildBoundaryGuildID = before.GuildBoundaryGuildID
	state.GuildBoundarySeq = before.GuildBoundarySeq
	state.GuildConsumedWindow = before.GuildConsumedWindow
	state.MeterValues = before.MeterValues
	state.MeterDecayRemainders = before.MeterDecayRemainders
	state.MeterInputRemainders = before.MeterInputRemainders
	state.AchievementsEarnedRun = before.AchievementsEarnedRun
	state.AchievementScoreRun = before.AchievementScoreRun
	state.LifetimeValue = before.LifetimeValue
	refillManualTokens(state, catalogs.Economy.ManualPolicy(), effectiveNow)
	afterOutputs, err := suppressionOutputSnapshot(state)
	if err != nil || !bytes.Equal(beforeOutputs, afterOutputs) {
		return SuppressedTransition{}, ErrInvalidEngineState
	}
	receipt, _ := json.Marshal(map[string]any{"intent_id": wire.Command.IntentID, "outcome": string(save.IntentApplied),
		"revision": wire.Command.Revision + 1, "session_id": suppression.SessionID,
		"from_evaluated_ms": suppression.FromEvaluatedMS, "to_evaluated_ms": suppression.ToEvaluatedMS,
		"suppressed_output": true})
	return SuppressedTransition{State: state, Receipt: receipt, Events: scheduleEvents}, nil
}

func buildSoulSuppressionInputs(command save.ReplayCommand, kind string, suppression soulSuppression,
	contributions []multiplier.Contribution, commonsWeight *int64, routeContextVersion int,
	activePlay *activePlayScheduleEvidence,
) (json.RawMessage, error) {
	accrual, err := makeReplayAccrual(contributions, commonsWeight, guild.SettlementBatch{}, routeContextVersion)
	if err != nil {
		return nil, err
	}
	resolved, err := json.Marshal(soulSuppressionResolved{Kind: "soul_recovery_suppression", IntentKind: kind,
		Suppression: suppression, Accrual: accrual, ActivePlay: activePlay})
	if err != nil {
		return nil, err
	}
	wire, err := json.Marshal(replayInputsWire{Version: save.ReplayInputsVersion, Command: command,
		EvaluatedAtMS: suppression.ToEvaluatedMS, EvaluationMode: ModeOnline, Resolved: resolved})
	if err != nil {
		return nil, err
	}
	if _, err := parseReplayInputs(wire); err != nil {
		return nil, err
	}
	return wire, nil
}

func isSoulRecoveryPayload(data []byte) bool {
	var payload soulRecoveryPayload
	return decodeReplayStrict(data, &payload) == nil &&
		(payload.Kind == soulRecoveryResolveKind || payload.Kind == soulRecoveryCancelKind)
}

func suppressionOutputSnapshot(state *save.State) ([]byte, error) {
	if state == nil || state.Ledger == nil {
		return nil, ErrInvalidEngineState
	}
	return json.Marshal(map[string]any{
		"ledger":                      state.Ledger.Snapshot(),
		"generator_provisioned":       state.GeneratorProvisioned,
		"provision_remainders_ppm":    state.ProvisionRemaindersPPM,
		"stock_units":                 state.StockUnits,
		"stock_progress_ms":           state.StockProgressMS,
		"stock_rate_remainder_ppm":    state.StockRateRemainderPPM,
		"consumed_stock_units":        state.ConsumedStockUnits,
		"guild_tithe_carry_ppm":       state.GuildTitheCarryPPM,
		"guild_boundary_guild_id":     state.GuildBoundaryGuildID,
		"guild_boundary_seq":          state.GuildBoundarySeq,
		"guild_consumed_window_units": state.GuildConsumedWindow,
		"meter_values":                state.MeterValues,
		"meter_decay_remainders":      state.MeterDecayRemainders,
		"meter_input_remainders":      state.MeterInputRemainders,
		"achievements_earned_run":     state.AchievementsEarnedRun,
		"achievement_score_run":       state.AchievementScoreRun,
		"lifetime_value":              state.LifetimeValue,
	})
}
