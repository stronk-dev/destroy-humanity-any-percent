package production

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/save"
)

var (
	ErrFounderAttendanceContext = errors.New("invalid Founder attendance context")
	ErrFounderAttendanceStale   = errors.New("stale Founder attendance sample")
	ErrFounderAttendanceBounds  = errors.New("Founder attendance exceeds exact bounds")
)

var (
	founderAttendanceStreamPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	founderAttendanceHashPattern   = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

// FounderAttendanceSample is the complete A2 tuple frozen before a
// Founder-scoped transition. It is a total attended-time sample, never a
// delta, and is safe to persist verbatim in Founder replay inputs.
type FounderAttendanceSample struct {
	CompanyStreamID             string `json:"company_stream_id"`
	RunSeq                      int64  `json:"run_seq"`
	CompanyRevision             int64  `json:"company_revision"`
	CompanyConstantsHash        string `json:"company_constants_hash"`
	CompletedAttendedMS         int64  `json:"completed_attended_ms"`
	CurrentRunPartialAttendedMS int64  `json:"current_run_partial_attended_ms"`
	EffectiveFounderAttendedMS  int64  `json:"effective_founder_attended_ms"`
}

// CompletedFounderAttendedMS names the existing Founder age_ms authority.
// It deliberately does not introduce another persisted cursor.
func CompletedFounderAttendedMS(founder *save.State) (int64, error) {
	if founder == nil || founder.Ledger == nil || founder.Ledger.Scope() != economy.ScopeFounder ||
		founder.AgeMS < 0 || founder.AgeMS > decimal.MaxExactInteger {
		return 0, ErrFounderAttendanceBounds
	}
	return founder.AgeMS, nil
}

func EffectiveFounderAttendedMS(completed, partial int64) (int64, error) {
	if completed < 0 || partial < 0 || completed > decimal.MaxExactInteger ||
		partial > decimal.MaxExactInteger-completed {
		return 0, ErrFounderAttendanceBounds
	}
	return completed + partial, nil
}

// ResolveFounderAttendanceSample is the projection-free resolver used by the
// live service and deterministic tests. The Company clone is classified for
// an unresolved offline gap before attended time is measured; the clone is
// never persisted.
func ResolveFounderAttendanceSample(founder *save.State, founderRevision save.Revision, company *save.State,
	companyRevision save.Revision, bundle CatalogBundle, now time.Time) (FounderAttendanceSample, error) {
	if founderRevision.OwnerID == "" || founderRevision.OwnerID != companyRevision.OwnerID ||
		companyRevision.StreamID == "" || companyRevision.Number < 1 || companyRevision.ConstantsHash == "" ||
		company == nil || company.Ledger == nil || company.Ledger.Scope() != economy.ScopeCompany ||
		company.RunSeq < 1 || company.RunStartedAt.IsZero() || !bundle.valid(companyRevision.ConstantsHash) {
		return FounderAttendanceSample{}, ErrFounderAttendanceContext
	}
	completed, err := CompletedFounderAttendedMS(founder)
	if err != nil {
		return FounderAttendanceSample{}, err
	}
	effectiveNow := save.CanonicalServerTime(now)
	if effectiveNow.Before(company.EvaluatedThrough) {
		return FounderAttendanceSample{}, ErrFounderAttendanceContext
	}
	clone, err := cloneReplayState(company, bundle.Economy)
	if err != nil {
		return FounderAttendanceSample{}, fmt.Errorf("%w: clone Company state: %v", ErrFounderAttendanceContext, err)
	}
	if effectiveNow.After(clone.EvaluatedThrough) {
		if err := prestigecore.RecordOfflineSpan(clone, clone.EvaluatedThrough, effectiveNow, bundle.Prestige.CatchupCeilingMS); err != nil {
			return FounderAttendanceSample{}, fmt.Errorf("%w: classify offline gap: %v", ErrFounderAttendanceContext, err)
		}
	}
	partial, err := prestigecore.AttendedMS(clone, effectiveNow)
	if err != nil {
		return FounderAttendanceSample{}, fmt.Errorf("%w: attended partial: %v", ErrFounderAttendanceContext, err)
	}
	effective, err := EffectiveFounderAttendedMS(completed, partial)
	if err != nil {
		return FounderAttendanceSample{}, err
	}
	return FounderAttendanceSample{
		CompanyStreamID: companyRevision.StreamID, RunSeq: company.RunSeq,
		CompanyRevision: companyRevision.Number, CompanyConstantsHash: companyRevision.ConstantsHash,
		CompletedAttendedMS: completed, CurrentRunPartialAttendedMS: partial,
		EffectiveFounderAttendedMS: effective,
	}, nil
}

// ResolveFounderAttendance reads Founder first and Company second. If Exit
// commits between those reads, the later Founder lock rejects the old
// completed base through ValidateFounderAttendanceSample; this order cannot
// combine a post-Exit age_ms with a pre-Exit Company partial.
func (s *Service) ResolveFounderAttendance(ctx context.Context, founderStreamID, companyStreamID string, now time.Time) (FounderAttendanceSample, error) {
	if s == nil || s.store == nil || s.replayCatalogs == nil || founderStreamID == "" || companyStreamID == "" {
		return FounderAttendanceSample{}, ErrFounderAttendanceContext
	}
	founder, err := s.store.LoadSiblingLatest(ctx, companyStreamID, economy.ScopeFounder)
	if err != nil {
		return FounderAttendanceSample{}, fmt.Errorf("%w: load Founder: %v", ErrFounderAttendanceContext, err)
	}
	if founder.Revision.StreamID != founderStreamID || founder.ArchivedAt != nil {
		return FounderAttendanceSample{}, ErrFounderAttendanceContext
	}
	company, err := s.store.LoadLatest(ctx, companyStreamID)
	if err != nil {
		return FounderAttendanceSample{}, fmt.Errorf("%w: load Company: %v", ErrFounderAttendanceContext, err)
	}
	if company.ArchivedAt != nil || company.Key.OwnerKind != save.OwnerFounder ||
		founder.Key.OwnerID != company.Key.OwnerID || founder.Key.Scope != economy.ScopeFounder || company.Key.Scope != economy.ScopeCompany {
		return FounderAttendanceSample{}, ErrFounderAttendanceContext
	}
	founder.Revision.OwnerID = founder.Key.OwnerID
	company.Revision.OwnerID = company.Key.OwnerID
	bundle, ok := s.replayCatalogs.ResolveReplayCatalogs(company.Revision.ConstantsHash)
	if !ok {
		return FounderAttendanceSample{}, ErrFounderAttendanceContext
	}
	return ResolveFounderAttendanceSample(founder.State, founder.Revision, company.State, company.Revision, bundle, now)
}

// ValidateFounderAttendanceSample is called inside the Founder-locked
// transition. A concurrent Exit changes age_ms and/or the Founder revision, so
// the consumer must re-resolve instead of applying a stale partial.
func ValidateFounderAttendanceSample(founder *save.State, actualFounderRevision, expectedFounderRevision int64, sample FounderAttendanceSample) error {
	completed, err := CompletedFounderAttendedMS(founder)
	if err != nil {
		return err
	}
	if actualFounderRevision < 1 || actualFounderRevision > decimal.MaxExactInteger || expectedFounderRevision != actualFounderRevision ||
		!founderAttendanceStreamPattern.MatchString(sample.CompanyStreamID) || sample.RunSeq < 1 || sample.RunSeq > decimal.MaxExactInteger ||
		sample.CompanyRevision < 1 || sample.CompanyRevision > decimal.MaxExactInteger ||
		!founderAttendanceHashPattern.MatchString(sample.CompanyConstantsHash) || sample.CompletedAttendedMS != completed {
		return ErrFounderAttendanceStale
	}
	effective, err := EffectiveFounderAttendedMS(sample.CompletedAttendedMS, sample.CurrentRunPartialAttendedMS)
	if err != nil {
		return err
	}
	if effective != sample.EffectiveFounderAttendedMS {
		return ErrFounderAttendanceBounds
	}
	return nil
}
