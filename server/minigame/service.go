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
	repository *Repository
	tenants    *TenantRegistry
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

func (resolution *CertifiedResolution) Result() *Result {
	if resolution == nil {
		return nil
	}
	return cloneResult(resolution.result)
}

func NewService(repository *Repository, tenants *TenantRegistry) (*Service, error) {
	if repository == nil || tenants == nil {
		return nil, errors.New("minigame service requires repository and tenant registry")
	}
	return &Service{repository: repository, tenants: tenants}, nil
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
	genesis, err := service.tenants.Create(request.EngineRef, request.EngineVersion, CreateInput{
		Mode: request.Mode, Seed: seed, ScalingInputs: request.ScalingInputs,
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
	output, err := service.tenants.Apply(claimed.EngineRef, claimed.EngineVersion, ApplyInput{
		Mode: claimed.Mode, Revision: claimed.Revision, Snapshot: claimed.State,
		Command: request.Command, ScalingInputs: scaling,
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
			identity: resolutionIdentity{sessionID: claimed.SessionID, founderID: claimed.FounderID,
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

// ResolveTx locks the certified result's Company stream before its session,
// then writes the terminal snapshot/result through the unexported repository
// boundary. A caller cannot construct a non-empty certification outside this
// package or resolve it in a different Company's transaction.
func (service *Service) ResolveTx(ctx context.Context, tx *sql.Tx, resolution *CertifiedResolution) (Session, error) {
	if service == nil || tx == nil || resolution == nil || !validResolutionIdentity(resolution.identity) ||
		!validJSONObject(resolution.command) || !validJSONObject(resolution.state) || !validJSONObject(resolution.bytes) ||
		service.tenants.validateCertifiedResult(resolution.identity.engineRef, resolution.identity.engineVersion, resolution.result) != nil {
		return Session{}, ErrInvalidSession
	}
	canonicalResult, err := json.Marshal(resolution.result)
	if err != nil || !bytes.Equal(canonicalResult, resolution.bytes) {
		return Session{}, ErrInvalidSession
	}
	var ownsRun bool
	err = tx.QueryRowContext(ctx, lockResolutionCompanySQL, resolution.identity.companyStreamID,
		resolution.identity.founderID, resolution.identity.runSeq, resolution.identity.constantsHash).Scan(&ownsRun)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrInvalidSession
	}
	if err != nil {
		return Session{}, err
	}
	if err := service.validateReplayTx(ctx, tx, resolution); err != nil {
		return Session{}, err
	}
	return resolveTx(ctx, tx, resolution.identity, resolution.command, resolution.state, resolution.bytes)
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
	snapshot := bytes.Clone(session.Genesis)
	revision := int64(1)
	for _, command := range commands {
		output, applyErr := service.tenants.Apply(session.EngineRef, session.EngineVersion, ApplyInput{
			Mode: session.Mode, Revision: revision, Snapshot: snapshot,
			Command: command.Command, ScalingInputs: scaling,
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
		Mode: session.Mode, Revision: revision, Snapshot: snapshot,
		Command: resolution.command, ScalingInputs: scaling,
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

const lockResolutionCompanySQL = "SELECT true FROM save_streams s JOIN run_epochs r " +
	"ON r.company_stream_id=s.id AND r.run_seq=$3 WHERE s.id=$1 AND s.owner_kind='founder' " +
	"AND s.owner_id=$2 AND s.scope='company' AND s.archived_at IS NULL AND r.constants_hash=$4 FOR UPDATE OF s"
