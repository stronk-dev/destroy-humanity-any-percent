package gameserver

import (
	"context"
	"encoding/json"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/production"
)

type minigameAPIAdapter struct {
	accounts   *account.Repository
	production *production.Service
	platform   *minigame.Service
}

func (adapter minigameAPIAdapter) CreateMinigameSession(ctx context.Context, accountID, minigameID,
	sessionID, idempotencyKey string, now time.Time,
) (json.RawMessage, error) {
	state, err := adapter.accounts.ActiveCompanyState(ctx, accountID)
	if err != nil {
		return nil, err
	}
	result, err := adapter.production.StartMinigameAPISession(ctx, adapter.platform,
		production.StartMinigameAPIRequest{SessionID: sessionID, FounderID: state.FounderID,
			CompanyStreamID: state.StreamID, MinigameID: minigameID, IdempotencyKey: idempotencyKey}, now, nil)
	return result.Receipt, err
}

func (adapter minigameAPIAdapter) PlayMinigameCommand(ctx context.Context, accountID, sessionID, commandID string,
	expectedRevision int64, command json.RawMessage, now time.Time,
) (json.RawMessage, error) {
	founder, err := adapter.accounts.ActiveFounder(ctx, accountID)
	if err != nil {
		return nil, err
	}
	result, err := adapter.production.PlayMinigameAPICommand(ctx, adapter.platform,
		production.PlayMinigameAPIRequest{FounderID: founder.ID, SessionID: sessionID, CommandID: commandID,
			ExpectedRevision: expectedRevision, Command: command}, now, nil)
	return result.Receipt, err
}

func (adapter minigameAPIAdapter) CurrentMinigameSession(ctx context.Context, accountID string) (json.RawMessage, error) {
	founder, err := adapter.accounts.ActiveFounder(ctx, accountID)
	if err != nil {
		return nil, err
	}
	session, active, err := adapter.platform.Current(ctx, founder.ID)
	if err != nil {
		return nil, err
	}
	if !active {
		return json.RawMessage(`{"kind":"none"}`), nil
	}
	descriptor := map[string]any{
		"constants_hash": session.ConstantsHash,
		"engine_ref":     session.EngineRef,
		"engine_version": session.EngineVersion,
		"minigame_id":    session.MinigameID,
		"mode":           session.Mode,
		"revision":       session.Revision,
		"session_id":     session.SessionID,
		"status":         session.Status,
	}
	return json.Marshal(map[string]any{"kind": "active", "session": descriptor, "snapshot": session.State})
}

func (adapter minigameAPIAdapter) ResolveMinigameSession(ctx context.Context, accountID, sessionID string) (json.RawMessage, error) {
	founder, err := adapter.accounts.ActiveFounder(ctx, accountID)
	if err != nil {
		return nil, err
	}
	session, err := adapter.platform.Load(ctx, founder.ID, sessionID)
	if err != nil {
		return nil, err
	}
	if session.Status != minigame.StatusResolved || len(session.ResolutionReceipt) == 0 {
		return nil, minigame.ErrSessionBusy
	}
	return production.MinigameAPISessionReceipt(session, session.ResolutionReceipt)
}
