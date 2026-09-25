package gameui

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/minigameapi"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

func transitionPreviewFounder(t *testing.T, catalog *economy.Catalog, now time.Time) *save.State {
	t.Helper()
	ledger, err := economy.NewLedger(catalog, economy.ScopeFounder)
	if err != nil {
		t.Fatal(err)
	}
	return &save.State{Ledger: ledger, GeneratorCounts: map[string]int64{}, EvaluatedThrough: now,
		ManualTokenRefilledAt: now, GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{},
		LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{}, RegionTraits: map[string]bool{},
		HintsUnlocked: map[string]bool{}, CompactSamples: []save.CompactSample{}, LifetimeValue: decimal.Zero,
		OfflineSpans: []save.OfflineSpan{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}}
}

func TestPhaseATransitionPreviewUsesProductionTransitionWithoutMutatingInputs(t *testing.T) {
	catalog, routeCatalog := loadCandidateCatalogs(t)
	now := time.UnixMilli(1_800_000_000_000).UTC()
	company := candidateState(t, catalog)
	company.ManualTokenRefilledAt = now
	founder := transitionPreviewFounder(t, catalog, now)
	bundle := production.CatalogBundle{ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Economy: catalog, Routes: routeCatalog}
	revision := save.Revision{OwnerID: "01985555-1111-7111-8111-111111111111", Number: 7, ConstantsHash: bundle.ConstantsHash}

	if _, err := company.Ledger.Apply(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: decimal.FromString("-1e9")}}}); err != nil {
		t.Fatal(err)
	}
	companyBefore, err := save.EncodeState(company)
	if err != nil {
		t.Fatal(err)
	}
	founderBefore, err := save.EncodeState(founder)
	if err != nil {
		t.Fatal(err)
	}
	below, err := previewPhaseATransitions(bundle, company, founder, revision, now, nil, false)
	if err != nil || below.CrossGate == nil || below.CrossGate.GateID != phaseAStandardGateID || below.CrossGate.Eligible || below.WindDown {
		t.Fatalf("below preview=%+v err=%v", below, err)
	}
	companyAfter, _ := save.EncodeState(company)
	founderAfter, _ := save.EncodeState(founder)
	if !bytes.Equal(companyBefore, companyAfter) || !bytes.Equal(founderBefore, founderAfter) {
		t.Fatal("transition preview mutated persisted inputs")
	}

	if _, err := company.Ledger.Apply(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: decimal.FromString("1e5")}}}); err != nil {
		t.Fatal(err)
	}
	ready, err := previewPhaseATransitions(bundle, company, founder, revision, now, nil, false)
	if err != nil || ready.CrossGate == nil || !ready.CrossGate.Eligible || ready.WindDown {
		t.Fatalf("ready preview=%+v err=%v", ready, err)
	}
	clone, err := cloneCompanyState(company, catalog)
	if err != nil {
		t.Fatal(err)
	}
	decision, err := production.TransitionWithRoutes(production.IntentRequest{IntentID: "00000000-0000-7000-8000-000000000001",
		Kind: production.IntentCrossGate, ExpectedRevision: revision.Number, GateID: phaseAStandardGateID},
		clone, catalog, routeCatalog, revision, production.ModeOnline, now, nil, nil)
	if err != nil || (decision.Outcome == save.IntentApplied) != ready.CrossGate.Eligible {
		t.Fatalf("production outcome=%s preview=%+v err=%v", decision.Outcome, ready, err)
	}

	company.Tier = 1
	later, err := previewPhaseATransitions(bundle, company, founder, revision, now, nil, false)
	if err != nil || later.CrossGate != nil || !later.WindDown {
		t.Fatalf("later preview=%+v err=%v", later, err)
	}
	company.Tier = 0
	company.GatesCrossed[phaseAStandardGateID] = true
	crossed, err := previewPhaseATransitions(bundle, company, founder, revision, now, nil, false)
	if err != nil || crossed.CrossGate != nil || crossed.WindDown {
		t.Fatalf("crossed preview=%+v err=%v", crossed, err)
	}
}

func TestPhaseATransitionPreviewWithholdsWindDownWhileMinigameSessionActive(t *testing.T) {
	catalog, routeCatalog := loadCandidateCatalogs(t)
	now := time.UnixMilli(1_800_000_000_000).UTC()
	company := candidateState(t, catalog)
	company.ManualTokenRefilledAt = now
	company.Tier = 1
	founder := transitionPreviewFounder(t, catalog, now)
	bundle := production.CatalogBundle{ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Economy: catalog, Routes: routeCatalog}
	revision := save.Revision{OwnerID: "01985555-1111-7111-8111-111111111111", Number: 7, ConstantsHash: bundle.ConstantsHash}
	idle, err := previewPhaseATransitions(bundle, company, founder, revision, now, nil, false)
	if err != nil || !idle.WindDown {
		t.Fatalf("idle preview=%+v err=%v", idle, err)
	}
	// MA-C12: Exit rejects not_eligible/minigame_session_active; the preview agrees.
	busy, err := previewPhaseATransitions(bundle, company, founder, revision, now, nil, true)
	if err != nil || busy.WindDown {
		t.Fatalf("active-session preview offered wind_down: %+v err=%v", busy, err)
	}
}

type fixedMinigameActivity struct {
	active  bool
	founder string
}

func (fixed *fixedMinigameActivity) ActiveMinigame(_ context.Context, founderID string) (bool, error) {
	fixed.founder = founderID
	return fixed.active, nil
}

func TestProjectorMinigameActivityMirrorsExitGate(t *testing.T) {
	ctx := context.Background()
	const founderID = "01985555-1111-7111-8111-111111111111"
	plain := production.CatalogBundle{}
	pinned := production.CatalogBundle{MinigameAPI: &minigameapi.Catalog{}}
	bare := &Projector{}
	if active, err := bare.minigameActive(ctx, plain, founderID); err != nil || active {
		t.Fatalf("bundle without minigame_api: active=%v err=%v", active, err)
	}
	if _, err := bare.minigameActive(ctx, pinned, founderID); !errors.Is(err, ErrInvalidProjection) {
		t.Fatalf("pinned bundle without resolver must fail loud, err=%v", err)
	}
	resolver := &fixedMinigameActivity{active: true}
	projector := &Projector{}
	if err := WithMinigameActivity(resolver)(projector); err != nil {
		t.Fatal(err)
	}
	if active, err := projector.minigameActive(ctx, pinned, founderID); err != nil || !active || resolver.founder != founderID {
		t.Fatalf("resolver not consulted: active=%v founder=%q err=%v", active, resolver.founder, err)
	}
	if err := WithMinigameActivity(nil)(&Projector{}); !errors.Is(err, ErrInvalidProjection) {
		t.Fatalf("nil resolver accepted: %v", err)
	}
}
