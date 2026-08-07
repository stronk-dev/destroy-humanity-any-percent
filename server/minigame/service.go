package minigame

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
)

var ErrSessionRevision = errors.New("minigame session revision conflict")

type Service struct {
	repository      *Repository
	tenants         *TenantRegistry
	contentResolver TenantContentResolver
}

type StartRequest struct {
	SessionID       string
	MinigameID      string
	FounderID       string
	CompanyStreamID string
	RunSeq          int64
	EngineRef       string
	EngineVersion   string
	ConstantsHash   string
	ScalingInputs   map[string]int64
	Seed            string
	Mode            Mode
}

type PlayRequest struct {
	FounderID        string
	SessionID        string
	ExpectedRevision int64
	Command          json.RawMessage
}

type PlayDecision struct {
	Session    Session
	Resolution *CertifiedResolution
}

type CertifiedResolution struct {
	identity resolutionIdentity
	command  json.RawMessage
	state    json.RawMessage
	result   *Result
	bytes    json.RawMessage
}

type CertifiedResolutionView struct {
	SessionID       string
	MinigameID      string
	FounderID       string
	CompanyStreamID string
	RunSeq          int64
	EngineRef       string
	EngineVersion   string
	ConstantsHash   string
	Result          *Result
	ResultBytes     json.RawMessage
}

type PreparedResolution struct {
	Session Session
	Faucet  FaucetApplication
}

func (resolution *CertifiedResolution) Result() *Result {
	if resolution == nil {
		return nil
	}
	return cloneResult(resolution.result)
}

func (resolution *CertifiedResolution) View() (CertifiedResolutionView, error) {
	if resolution == nil || !validResolutionIdentity(resolution.identity) || !validJSONObject(resolution.bytes) {
		return CertifiedResolutionView{}, ErrInvalidSession
	}
	return CertifiedResolutionView{SessionID: resolution.identity.sessionID, MinigameID: resolution.identity.minigameID, FounderID: resolution.identity.founderID,
		CompanyStreamID: resolution.identity.companyStreamID, RunSeq: resolution.identity.runSeq,
		EngineRef: resolution.identity.engineRef, EngineVersion: resolution.identity.engineVersion,
		ConstantsHash: resolution.identity.constantsHash, Result: cloneResult(resolution.result),
		ResultBytes: bytes.Clone(resolution.bytes)}, nil
}

func NewService(repository *Repository, tenants *TenantRegistry, contentResolvers ...TenantContentResolver) (*Service, error) {
	if repository == nil || tenants == nil || len(contentResolvers) > 1 || len(contentResolvers) == 1 && contentResolvers[0] == nil {
		return nil, errors.New("minigame service requires repository and tenant registry")
	}
	service := &Service{repository: repository, tenants: tenants}
	if len(contentResolvers) == 1 {
		service.contentResolver = contentResolvers[0]
	}
	return service, nil
}

// Start consumes only server-resolved identity and scaling values. It asks the
// tenant for a deterministic genesis snapshot, then freezes both inputs in the
// authoritative session row.
func (service *Service) Start(ctx context.Context, request StartRequest) (Session, error) {
	if service == nil || request.EngineVersion == "" {
		return Session{}, ErrInvalidSession
	}
	descriptor, ok := service.tenants.Descriptor(request.EngineRef)
	if !ok || descriptor.EngineVersion != request.EngineVersion {
		return Session{}, ErrInvalidTenant
	}
	seed, err := strconv.ParseUint(request.Seed, 10, 64)
	if err != nil || request.Seed != "0" && request.Seed[0] == '0' {
		return Session{}, ErrInvalidSession
	}
	content := service.resolveContent(request.ConstantsHash, request.EngineRef, request.EngineVersion)
	genesis, err := service.tenants.Create(request.EngineRef, request.EngineVersion, CreateInput{
		Mode: request.Mode, Seed: seed, ScalingInputs: request.ScalingInputs, Content: content.Bytes,
		ContentHash: content.Hash, ContentSchemaVersion: content.SchemaVersion,
	})
	if err != nil {
		return Session{}, err
	}
	scaling, err := json.Marshal(request.ScalingInputs)
	if err != nil {
		return Session{}, ErrInvalidSession
	}
	return service.repository.create(ctx, CreateSession{
		SessionID: request.SessionID, MinigameID: request.MinigameID, FounderID: request.FounderID,
		CompanyStreamID: request.CompanyStreamID, RunSeq: request.RunSeq, EngineRef: request.EngineRef,
		EngineVersion: request.EngineVersion, ConstantsHash: request.ConstantsHash,
		ScalingInputs: scaling, Seed: request.Seed, Mode: request.Mode, Genesis: genesis,
	})
}

// Play claims the authoritative row before invoking the pure tenant. Rejected
// or stale commands release without advancing revision. A terminal result
// remains claim-owned so the caller can commit it with the Company payout via
// ResolveTx; clients never supply that result.
func (service *Service) Play(ctx context.Context, request PlayRequest) (decision PlayDecision, err error) {
	if service == nil || request.ExpectedRevision < 1 || !validJSONObject(request.Command) {
		return PlayDecision{}, ErrInvalidSession
	}
	claimed, err := service.repository.claim(ctx, request.FounderID, request.SessionID)
	if err != nil {
		return PlayDecision{}, err
	}
	release := true
	defer func() {
		if release {
			if releaseErr := service.repository.releaseClaim(context.WithoutCancel(ctx), request.FounderID, request.SessionID, claimed.ClaimToken); releaseErr != nil {
				err = errors.Join(err, releaseErr)
			}
		}
	}()
	if claimed.Revision != request.ExpectedRevision {
		return PlayDecision{}, ErrSessionRevision
	}
	scaling, ok := decodeScalingInputs(claimed.ScalingInputs)
	if !ok {
		return PlayDecision{}, ErrTenantDivergence
	}
	seed, seedErr := strconv.ParseUint(claimed.Seed, 10, 64)
	if seedErr != nil {
		return PlayDecision{}, ErrTenantDivergence
	}
	content := service.resolveContent(claimed.ConstantsHash, claimed.EngineRef, claimed.EngineVersion)
	output, err := service.tenants.Apply(claimed.EngineRef, claimed.EngineVersion, ApplyInput{
		Mode: claimed.Mode, Seed: seed, Revision: claimed.Revision, Snapshot: claimed.State,
		Command: request.Command, ScalingInputs: scaling, Content: content.Bytes,
		ContentHash: content.Hash, ContentSchemaVersion: content.SchemaVersion,
	})
	if err != nil {
		return PlayDecision{}, err
	}
	if output.Result != nil {
		resultBytes, marshalErr := json.Marshal(output.Result)
		if marshalErr != nil || !validJSONObject(resultBytes) {
			return PlayDecision{}, ErrTenantDivergence
		}
		claimed.State = output.Snapshot
		decision = PlayDecision{Session: claimed, Resolution: &CertifiedResolution{
			identity: resolutionIdentity{sessionID: claimed.SessionID, minigameID: claimed.MinigameID, founderID: claimed.FounderID,
				companyStreamID: claimed.CompanyStreamID, runSeq: claimed.RunSeq, engineRef: claimed.EngineRef,
				engineVersion: claimed.EngineVersion, constantsHash: claimed.ConstantsHash, claimToken: claimed.ClaimToken},
			command: bytes.Clone(request.Command), state: output.Snapshot,
			result: cloneResult(output.Result), bytes: resultBytes,
		}}
		release = false
		return decision, nil
	}
	updated, err := service.repository.completePlay(ctx, request.FounderID, request.SessionID,
		claimed.ClaimToken, request.Command, output.Snapshot)
	if err != nil {
		return PlayDecision{}, err
	}
	release = false
	return PlayDecision{Session: updated}, nil
}

// PrepareResolutionTx verifies the claimed session by replaying its immutable
// command log and applies its pinned faucet row. The caller owns the global
// Founder→Company lock order and must either finalize in the same transaction
// or roll the transaction back.
func (service *Service) PrepareResolutionTx(ctx context.Context, tx *sql.Tx, resolution *CertifiedResolution,
	definition Definition, founderAttendedMS int64,
) (PreparedResolution, error) {
	if service == nil || tx == nil || resolution == nil || !validResolutionIdentity(resolution.identity) ||
		!validJSONObject(resolution.command) || !validJSONObject(resolution.state) || !validJSONObject(resolution.bytes) ||
		service.tenants.validateCertifiedResult(resolution.identity.engineRef, resolution.identity.engineVersion, resolution.result) != nil {
		return PreparedResolution{}, ErrInvalidSession
	}
	canonicalResult, err := json.Marshal(resolution.result)
	if err != nil || !bytes.Equal(canonicalResult, resolution.bytes) {
		return PreparedResolution{}, ErrInvalidSession
	}
	if definition.MinigameID == "" || definition.EngineRef != resolution.identity.engineRef ||
		definition.EngineVersion != resolution.identity.engineVersion {
		return PreparedResolution{}, ErrInvalidSession
	}
	if err := service.validateReplayTx(ctx, tx, resolution); err != nil {
		return PreparedResolution{}, err
	}
	score, err := SelectPayoutScore(resolution.result, definition.Payout)
	if err != nil {
		return PreparedResolution{}, err
	}
	faucet, err := applyFaucetWindowTx(ctx, tx, resolution.identity.founderID, definition.MinigameID,
		founderAttendedMS, definition.Payout, score, definition.Fallback.RateReductionPPM)
	if err != nil {
		return PreparedResolution{}, err
	}
	session, _, err := loadClaimedReplay(ctx, tx, resolution.identity)
	if err != nil {
		return PreparedResolution{}, err
	}
	return PreparedResolution{Session: session, Faucet: faucet}, nil
}

func (service *Service) validateCertifiedTx(ctx context.Context, tx *sql.Tx, resolution *CertifiedResolution) error {
	if service == nil || tx == nil || resolution == nil || !validResolutionIdentity(resolution.identity) ||
		!validJSONObject(resolution.command) || !validJSONObject(resolution.state) || !validJSONObject(resolution.bytes) ||
		service.tenants.validateCertifiedResult(resolution.identity.engineRef, resolution.identity.engineVersion, resolution.result) != nil {
		return ErrInvalidSession
	}
	canonical, err := json.Marshal(resolution.result)
	if err != nil || !bytes.Equal(canonical, resolution.bytes) {
		return ErrInvalidSession
	}
	return service.validateReplayTx(ctx, tx, resolution)
}

// FinalizeResolutionTx is callable only after PrepareResolutionTx in the same
// transaction. The write-once receipt tuple makes the session terminal state
// a durable idempotency receipt rather than a process-local outcome.
func (service *Service) FinalizeResolutionTx(ctx context.Context, tx *sql.Tx, resolution *CertifiedResolution,
	receipt json.RawMessage, companyRevision, founderRevision int64,
) (Session, error) {
	if service == nil || tx == nil || resolution == nil {
		return Session{}, ErrInvalidSession
	}
	return resolveTx(ctx, tx, resolution.identity, resolution.command, resolution.state, resolution.bytes,
		receipt, companyRevision, founderRevision)
}

func (service *Service) validateReplayTx(ctx context.Context, tx *sql.Tx, resolution *CertifiedResolution) error {
	session, commands, err := loadClaimedReplay(ctx, tx, resolution.identity)
	if err != nil {
		return err
	}
	scaling, ok := decodeScalingInputs(session.ScalingInputs)
	if !ok {
		return ErrTenantDivergence
	}
	seed, err := strconv.ParseUint(session.Seed, 10, 64)
	if err != nil {
		return ErrTenantDivergence
	}
	content := service.resolveContent(session.ConstantsHash, session.EngineRef, session.EngineVersion)
	snapshot := bytes.Clone(session.Genesis)
	revision := int64(1)
	for _, command := range commands {
		output, applyErr := service.tenants.Apply(session.EngineRef, session.EngineVersion, ApplyInput{
			Mode: session.Mode, Seed: seed, Revision: revision, Snapshot: snapshot,
			Command: command.Command, ScalingInputs: scaling, Content: content.Bytes,
			ContentHash: content.Hash, ContentSchemaVersion: content.SchemaVersion,
		})
		if applyErr != nil || output.Result != nil {
			return ErrTenantDivergence
		}
		snapshot, revision = output.Snapshot, revision+1
	}
	if revision != session.Revision || !bytes.Equal(snapshot, session.State) {
		return ErrTenantDivergence
	}
	output, err := service.tenants.Apply(session.EngineRef, session.EngineVersion, ApplyInput{
		Mode: session.Mode, Seed: seed, Revision: revision, Snapshot: snapshot,
		Command: resolution.command, ScalingInputs: scaling, Content: content.Bytes,
		ContentHash: content.Hash, ContentSchemaVersion: content.SchemaVersion,
	})
	if err != nil || output.Result == nil || !bytes.Equal(output.Snapshot, resolution.state) ||
		service.tenants.validateCertifiedResult(session.EngineRef, session.EngineVersion, output.Result) != nil {
		return ErrTenantDivergence
	}
	encoded, err := json.Marshal(output.Result)
	if err != nil || !bytes.Equal(encoded, resolution.bytes) {
		return ErrTenantDivergence
	}
	return nil
}

func (service *Service) resolveContent(constantsHash, engineRef, engineVersion string) TenantContent {
	if service == nil || service.contentResolver == nil {
		return TenantContent{}
	}
	content, ok := service.contentResolver.ResolveTenantContent(constantsHash, engineRef, engineVersion)
	if !ok {
		return TenantContent{}
	}
	content.Bytes = bytes.Clone(content.Bytes)
	return content
}

func decodeScalingInputs(data []byte) (map[string]int64, bool) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var raw map[string]json.Number
	if err := decoder.Decode(&raw); err != nil || raw == nil {
		return nil, false
	}
	result := make(map[string]int64, len(raw))
	for key, value := range raw {
		parsed, err := strconv.ParseInt(value.String(), 10, 64)
		if err != nil || parsed < -9_007_199_254_740_991 || parsed > 9_007_199_254_740_991 {
			return nil, false
		}
		result[key] = parsed
	}
	return result, true
}
