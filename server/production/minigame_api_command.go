package production

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"cloud-clicker/server/minigame"
	"cloud-clicker/server/save"
)

var (
	minigameAPIOpaqueIDPattern = regexp.MustCompile(`^[A-Za-z0-9-]{1,64}$`)
	minigameAPIHashPattern     = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

type PlayMinigameAPIRequest struct {
	FounderID        string
	SessionID        string
	CommandID        string
	ExpectedRevision int64
	Command          json.RawMessage
}

// PlayMinigameAPICommand is the typed, not-yet-mounted MA-C13 coordinator.
// It checks the durable command receipt before tenant execution, commits a
// nonterminal response with the snapshot update, and auto-resolves a terminal
// result through the cross-stream transaction in the same request.
func (s *Service) PlayMinigameAPICommand(ctx context.Context, platform *minigame.Service,
	request PlayMinigameAPIRequest, now time.Time, fault save.ExitFaultInjector,
) (save.IntentResult, error) {
	if s == nil || platform == nil || !minigameAPIOpaqueIDPattern.MatchString(request.CommandID) ||
		request.ExpectedRevision < 1 {
		return save.IntentResult{}, ErrInvalidIntent
	}
	command, err := normalizeReplayJSON(request.Command)
	if err != nil || len(command) == 0 || command[0] != '{' {
		return save.IntentResult{}, ErrInvalidIntent
	}
	identity, err := normalizeReplayJSON(mustJSON(map[string]any{
		"command":           json.RawMessage(command),
		"command_id":        request.CommandID,
		"expected_revision": request.ExpectedRevision,
		"session_id":        request.SessionID,
	}))
	if err != nil {
		return save.IntentResult{}, ErrInvalidIntent
	}
	requestHash := soulRequestHash(identity)
	decision, err := platform.PlayWithReceipt(ctx, minigame.PlayRequest{
		FounderID: request.FounderID, SessionID: request.SessionID,
		ExpectedRevision: request.ExpectedRevision, Command: command,
	}, request.CommandID, requestHash, func(session minigame.Session) (json.RawMessage, error) {
		return MinigameAPISessionReceipt(session, nil)
	})
	if err != nil {
		return save.IntentResult{}, err
	}
	if decision.Replay {
		return save.IntentResult{Outcome: save.IntentApplied, Receipt: decision.Receipt, Replay: true}, nil
	}
	if decision.Resolution == nil {
		return save.IntentResult{Outcome: save.IntentApplied, Receipt: decision.Receipt}, nil
	}
	resolved, err := s.ResolveMinigameAPICommand(ctx, platform, decision.Resolution,
		request.CommandID, requestHash, now, fault)
	if err != nil {
		return save.IntentResult{}, err
	}
	return save.IntentResult{Outcome: save.IntentApplied, Receipt: resolved.Receipt, Replay: resolved.Replay}, nil
}

func MinigameAPISessionReceipt(session minigame.Session, resolutionReceipt json.RawMessage) (json.RawMessage, error) {
	if len(session.State) == 0 || session.Revision < 1 {
		return nil, fmt.Errorf("%w: invalid minigame API session", ErrInvalidIntent)
	}
	value := map[string]any{
		"constants_hash": session.ConstantsHash,
		"engine_ref":     session.EngineRef,
		"engine_version": session.EngineVersion,
		"minigame_id":    session.MinigameID,
		"mode":           session.Mode,
		"revision":       session.Revision,
		"session_id":     session.SessionID,
		"snapshot":       json.RawMessage(session.State),
		"status":         session.Status,
	}
	if len(resolutionReceipt) != 0 {
		value["resolution_receipt"] = json.RawMessage(resolutionReceipt)
	}
	return normalizeReplayJSON(mustJSON(value))
}
