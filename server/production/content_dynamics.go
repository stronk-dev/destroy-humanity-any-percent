package production

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/pitch"
	"cloud-clicker/server/save"
)

// ContentDynamicsActivePlayInput pins the natural first draw that a candidate
// scenario intends to measure. The simulation rejects rather than adapting if
// balance bytes change that draw or its selected target.
type ContentDynamicsActivePlayInput struct {
	FounderID         string
	RunSeq            int64
	EffectRowID       string
	TargetGeneratorID string
}

type ContentDynamicsActivePlayResult struct {
	SpawnedAttendedMS int64
	DurationMS        int64
	ClaimedOutput     decimal.Decimal
	ControlOutput     decimal.Decimal
	BonusOutput       decimal.Decimal
	Transitions       int64
}

type ContentDynamicsFiscalResult struct {
	Periods          int64
	SequenceBefore   int64
	SequenceAfter    int64
	CreditBefore     int64
	CreditAfter      int64
	Credited         int64
	Saturated        bool
	HardcapReasonKey string
	Transitions      int64
}

type ContentDynamicsPitchResult struct {
	FinalRound    int64
	PayoutScore   int64
	ConvertedCash int64
	Transitions   int64
}

type ContentDynamicsPermitResult struct {
	FiscalCredit   int64
	TimeToTwelveMS int64
	TimeToCapMS    int64
	CapReasonKey   string
	Transitions    int64
}

// SimulateContentDynamicsActivePlay is a pure harness boundary over the live
// schedule, claim, contribution, and accrual machinery. It owns no alternate
// active-play math and never manufactures a pending opportunity.
func SimulateContentDynamicsActivePlay(bundle CatalogBundle, input ContentDynamicsActivePlayInput) (ContentDynamicsActivePlayResult, error) {
	if bundle.Economy == nil || bundle.Opportunities == nil || bundle.Prestige == nil || input.FounderID == "" || input.RunSeq < 1 || input.EffectRowID == "" || input.TargetGeneratorID == "" {
		return ContentDynamicsActivePlayResult{}, ErrInvalidEngineState
	}
	state, err := newContentDynamicsCompany(bundle.Economy, 18, input.RunSeq)
	if err != nil {
		return ContentDynamicsActivePlayResult{}, err
	}
	if _, ok := bundle.Economy.GeneratorClass(input.TargetGeneratorID); !ok {
		return ContentDynamicsActivePlayResult{}, ErrInvalidEngineState
	}
	state.GeneratorCounts[input.TargetGeneratorID] = 1
	spawn, err := bundle.Opportunities.Spawn(input.FounderID, input.RunSeq, 0, 0)
	if err != nil || spawn.EffectRowID != input.EffectRowID {
		return ContentDynamicsActivePlayResult{}, fmt.Errorf("%w: active-play pinned first draw changed", ErrInvalidEngineState)
	}
	effect, ok := bundle.Opportunities.Effect(spawn.EffectRowID)
	if !ok || effect.DurationMS <= 0 || effect.Kind == "click_frenzy" || effect.Kind == "lucky_payout" ||
		effect.Kind == "building_special" && spawn.SelectedGenerator != input.TargetGeneratorID {
		return ContentDynamicsActivePlayResult{}, fmt.Errorf("%w: active-play draw has no declared production window", ErrInvalidEngineState)
	}
	if _, err := initializeActivePlayState(state, bundle.Opportunities, input.FounderID); err != nil {
		return ContentDynamicsActivePlayResult{}, err
	}
	schedule := activePlayScheduleEvidence{AttendedNowMS: spawn.SpawnedAttendedMS, BeforeSequence: 0,
		BeforeNextOpportunityMS: spawn.SpawnedAttendedMS, ExpiredBuffs: []activePlayExpiredBuff{}}
	if err := advanceActivePlaySchedule(state, bundle.Opportunities, input.FounderID, spawn.SpawnedAttendedMS, &schedule, true); err != nil ||
		state.PendingOpportunity == nil || schedule.Spawned == nil || schedule.Spawned.OpportunityID != spawn.OpportunityID {
		return ContentDynamicsActivePlayResult{}, fmt.Errorf("%w: active-play natural spawn failed", ErrInvalidEngineState)
	}
	started := contentDynamicsEpoch.Add(time.Duration(spawn.SpawnedAttendedMS) * time.Millisecond)
	state.EvaluatedThrough = started
	state.ManualTokenRefilledAt = started
	control, err := cloneReplayState(state, bundle.Economy)
	if err != nil {
		return ContentDynamicsActivePlayResult{}, err
	}
	claimEvidence := activePlayClaimEvidence{}
	request := IntentRequest{IntentID: "018f6b7c-9abc-7def-8abc-0123456789ab", Kind: IntentClaimOpportunity,
		ExpectedRevision: 1, OpportunityID: spawn.OpportunityID}
	decision, err := applyClaimOpportunity(request, state, bundle, save.Revision{Number: 1, OwnerID: input.FounderID}, ModeOnline, started,
		nil, nil, spawn.SpawnedAttendedMS, nil, &claimEvidence, true)
	if err != nil || decision.Outcome != save.IntentApplied || claimEvidence.EffectRowID != input.EffectRowID {
		return ContentDynamicsActivePlayResult{}, fmt.Errorf("%w: active-play claim failed: outcome=%s error=%v", ErrInvalidEngineState, decision.Outcome, err)
	}
	if effect.Kind == "production_frenzy" {
		found := false
		for _, target := range effect.Targets {
			found = found || target == "generator_production"
		}
		if !found {
			return ContentDynamicsActivePlayResult{}, ErrInvalidEngineState
		}
	}
	claimedOutput, err := contentDynamicsAccrueWindow(bundle, state, input.TargetGeneratorID, spawn.SpawnedAttendedMS, effect.DurationMS, false)
	if err != nil {
		return ContentDynamicsActivePlayResult{}, err
	}
	controlOutput, err := contentDynamicsAccrueWindow(bundle, control, input.TargetGeneratorID, spawn.SpawnedAttendedMS, effect.DurationMS, false)
	if err != nil {
		return ContentDynamicsActivePlayResult{}, err
	}
	partitioned, err := contentDynamicsAccrueWindow(bundle, state, input.TargetGeneratorID, spawn.SpawnedAttendedMS, effect.DurationMS, true)
	if err != nil || !partitioned.Eq(claimedOutput) {
		return ContentDynamicsActivePlayResult{}, fmt.Errorf("%w: active-play window is not partition invariant", ErrInvalidEngineState)
	}
	bonus := claimedOutput.Sub(controlOutput).Quantize(decimal.CanonicalSignificantDigits)
	if !bonus.IsStateValue() || !bonus.Gt(decimal.Zero) {
		return ContentDynamicsActivePlayResult{}, fmt.Errorf("%w: active-play control pair is non-discriminating", ErrInvalidEngineState)
	}
	return ContentDynamicsActivePlayResult{SpawnedAttendedMS: spawn.SpawnedAttendedMS, DurationMS: effect.DurationMS,
		ClaimedOutput: claimedOutput, ControlOutput: controlOutput, BonusOutput: bonus, Transitions: 5}, nil
}

func contentDynamicsAccrueWindow(bundle CatalogBundle, source *save.State, resourceGeneratorID string, attendedStart, durationMS int64, partitioned bool) (decimal.Decimal, error) {
	state, err := cloneReplayState(source, bundle.Economy)
	if err != nil {
		return decimal.NaN, err
	}
	generator, ok := bundle.Economy.GeneratorClass(resourceGeneratorID)
	if !ok || generator.Production == nil {
		return decimal.NaN, ErrInvalidEngineState
	}
	before, ok := state.Ledger.Balance(generator.Production.ResourceID)
	if !ok {
		return decimal.NaN, ErrInvalidEngineState
	}
	steps := []int64{durationMS}
	if partitioned {
		steps = []int64{durationMS / 2, durationMS - durationMS/2}
	}
	elapsed := int64(0)
	for _, step := range steps {
		active, activeErr := activePlayContributions(state, bundle.Opportunities, attendedStart+elapsed)
		if activeErr != nil {
			return decimal.NaN, activeErr
		}
		contributions, contributionErr := assembleContributions(state, bundle.Economy, active)
		if contributionErr != nil {
			return decimal.NaN, contributionErr
		}
		elapsed += step
		if _, evaluationErr := Evaluate(state, bundle.Economy, source.EvaluatedThrough.Add(time.Duration(elapsed)*time.Millisecond), ModeOnline, contributions); evaluationErr != nil {
			return decimal.NaN, evaluationErr
		}
	}
	after, ok := state.Ledger.Balance(generator.Production.ResourceID)
	if !ok {
		return decimal.NaN, ErrInvalidEngineState
	}
	return after.Sub(before).Quantize(decimal.CanonicalSignificantDigits), nil
}

// SimulateContentDynamicsFiscal exercises the production-owned lazy sweep at
// an exact complete-period boundary.
func SimulateContentDynamicsFiscal(bundle CatalogBundle, periods int64) (ContentDynamicsFiscalResult, error) {
	if bundle.Fiscal == nil || bundle.Soul == nil || bundle.MinigameAPI == nil || periods != 1 && periods != 4 {
		return ContentDynamicsFiscalResult{}, ErrInvalidEngineState
	}
	opened := contentDynamicsEpoch.UnixMilli()
	state, err := newContentDynamicsFounder(bundle, opened)
	if err != nil {
		return ContentDynamicsFiscalResult{}, err
	}
	now := opened + periods*bundle.Fiscal.Clock.AutoMS
	command := save.FounderReplayCommand{IntentID: "018f6b7c-9abc-7def-8abc-100000000001",
		FounderStreamID: "018f6b7c-9abc-7def-8abc-100000000002", FounderID: "018f6b7c-9abc-7def-8abc-100000000003",
		Revision: 1, FounderLogSeq: 1, ServerTSMS: now}
	resolved := founderFiscalHarvestResolved{Kind: IntentHarvestFiscalPeriod, NowWallMS: now,
		PeriodOpenedWallMSBefore: opened, PeriodsSwept: periods, SeqBefore: 0, Outcome: "consumed_by_auto"}
	inputs, err := save.MarshalFounderReplayInputs(command, resolved)
	if err != nil {
		return ContentDynamicsFiscalResult{}, err
	}
	transition, err := ApplyFounderLogged(state, []byte(`{"expected_revision":1,"kind":"harvest_fiscal_period"}`), bundle, inputs)
	if err != nil || transition.Outcome != save.IntentApplied || len(transition.Events) != 1 || transition.Events[0].Kind != save.EventFiscalPeriodHarvested {
		return ContentDynamicsFiscalResult{}, fmt.Errorf("%w: fiscal sweep reconciliation: outcome=%s events=%d error=%v", ErrInvalidEngineState, transition.Outcome, len(transition.Events), err)
	}
	var event struct {
		Source string `json:"source"`
		founderFiscalSweepWire
	}
	if json.Unmarshal(transition.Events[0].Payload, &event) != nil || event.Source != "automatic" || event.Periods != periods ||
		event.SeqAfter != state.FiscalPeriodSequence || event.CreditAfter != state.FiscalCredit || event.OpenedAfterMS != now ||
		event.HardcapReasonKey != bundle.Fiscal.Credit.HardcapReasonKey {
		return ContentDynamicsFiscalResult{}, fmt.Errorf("%w: fiscal event/state mismatch", ErrInvalidEngineState)
	}
	return ContentDynamicsFiscalResult{Periods: event.Periods, SequenceBefore: event.SeqBefore,
		SequenceAfter: event.SeqAfter, CreditBefore: event.CreditBefore, CreditAfter: event.CreditAfter,
		Credited: event.Credited, Saturated: event.Saturated, HardcapReasonKey: event.HardcapReasonKey, Transitions: 1}, nil
}

// SimulateContentDynamicsPitch executes the ruled deterministic tenant policy
// and the same certified score-selection/conversion kernels used by platform
// resolution. It never writes a session or faucet row.
func SimulateContentDynamicsPitch(bundle CatalogBundle, seed uint64) (ContentDynamicsPitchResult, error) {
	if bundle.Pitch == nil || bundle.Minigames == nil || len(bundle.Artifacts["pitch"]) == 0 {
		return ContentDynamicsPitchResult{}, ErrInvalidEngineState
	}
	definition, ok := bundle.Minigames.Definition("pitch")
	if !ok || definition.EngineRef != pitch.EngineRef || definition.EngineVersion != pitch.EngineVersion {
		return ContentDynamicsPitchResult{}, ErrInvalidEngineState
	}
	content := bundle.Artifacts["pitch"]
	contentHash := pitch.ContentHash(content)
	tenant := pitch.NewTenant()
	scaling := map[string]int64{pitch.ScalingDestination: 1}
	snapshotBytes, err := tenant.Create(minigame.CreateInput{Mode: minigame.ModeSolo, Seed: seed, ScalingInputs: scaling,
		Content: content, ContentHash: contentHash, ContentSchemaVersion: pitch.SchemaVersion})
	if err != nil {
		return ContentDynamicsPitchResult{}, err
	}
	transitions := int64(0)
	for transitions < 64 {
		var snapshot pitch.Snapshot
		if err := json.Unmarshal(snapshotBytes, &snapshot); err != nil {
			return ContentDynamicsPitchResult{}, err
		}
		var command []byte
		switch snapshot.Phase {
		case "playing":
			cards := append([]string(nil), snapshot.Hand...)
			sort.Strings(cards)
			if len(cards) > 4 {
				cards = cards[:4]
			}
			command, _ = json.Marshal(struct {
				Kind    string   `json:"kind"`
				CardIDs []string `json:"card_ids"`
			}{Kind: "play_hand", CardIDs: cards})
		case "shop":
			eligible := append([]pitch.ShopOffer(nil), snapshot.ShopOffers...)
			sort.Slice(eligible, func(i, j int) bool {
				if eligible[i].Price != eligible[j].Price {
					return eligible[i].Price < eligible[j].Price
				}
				return eligible[i].OfferID < eligible[j].OfferID
			})
			if len(snapshot.SlottedHacks) < 4 && len(eligible) > 0 && eligible[0].Price <= snapshot.RunCurrency {
				command, _ = json.Marshal(struct {
					Kind    string `json:"kind"`
					OfferID string `json:"offer_id"`
				}{Kind: "buy_hack", OfferID: eligible[0].OfferID})
			} else {
				command = []byte(`{"kind":"end_shop"}`)
			}
		case "terminal":
			return ContentDynamicsPitchResult{}, ErrInvalidEngineState
		default:
			return ContentDynamicsPitchResult{}, ErrInvalidEngineState
		}
		output, applyErr := tenant.Apply(minigame.ApplyInput{Mode: minigame.ModeSolo, Seed: seed, Revision: snapshot.Revision,
			Snapshot: snapshotBytes, Command: command, ScalingInputs: scaling, Content: content,
			ContentHash: contentHash, ContentSchemaVersion: pitch.SchemaVersion})
		if applyErr != nil {
			return ContentDynamicsPitchResult{}, applyErr
		}
		transitions++
		snapshotBytes = output.Snapshot
		if output.Result == nil {
			continue
		}
		if err := tenant.ValidateResult(output.Result); err != nil {
			return ContentDynamicsPitchResult{}, err
		}
		score, err := minigame.SelectPayoutScore(output.Result, definition.Payout)
		if err != nil {
			return ContentDynamicsPitchResult{}, err
		}
		converted, err := minigame.ConvertPayout(score, definition.Fallback.RateReductionPPM, definition.Payout.ConversionPPM, 0)
		if err != nil {
			return ContentDynamicsPitchResult{}, err
		}
		credited := converted.ConvertedUnits
		if credited > definition.Payout.PerSendCap {
			credited = definition.Payout.PerSendCap
		}
		finalRound := int64(0)
		for _, fact := range output.Result.ScoreFacts {
			if fact.Kind == "pitch.final_round" {
				finalRound = fact.Value
			}
		}
		if finalRound < 1 {
			return ContentDynamicsPitchResult{}, ErrInvalidEngineState
		}
		return ContentDynamicsPitchResult{FinalRound: finalRound, PayoutScore: score, ConvertedCash: credited, Transitions: transitions}, nil
	}
	return ContentDynamicsPitchResult{}, fmt.Errorf("%w: pitch command ceiling", ErrInvalidEngineState)
}

// SimulateContentDynamicsPermits finds the first millisecond at which the
// canonical production transition reaches each permit coordinate. Search is
// bounded and calls the production evaluator for every candidate; no harness
// formula mirrors the engine.
func SimulateContentDynamicsPermits(bundle CatalogBundle, fiscalCredit, horizonMS int64) (ContentDynamicsPermitResult, error) {
	if bundle.Economy == nil || bundle.Fiscal == nil || fiscalCredit != 0 && fiscalCredit != bundle.Fiscal.Hoard.CapCredits || horizonMS <= 0 {
		return ContentDynamicsPermitResult{}, ErrInvalidEngineState
	}
	state, err := newContentDynamicsCompany(bundle.Economy, 18, 1)
	if err != nil {
		return ContentDynamicsPermitResult{}, err
	}
	if _, ok := bundle.Economy.GeneratorClass("generator.legal_dept"); !ok {
		return ContentDynamicsPermitResult{}, ErrInvalidEngineState
	}
	state.GeneratorCounts["generator.legal_dept"] = 1
	founder := &save.State{WireVersion: 19, FiscalCredit: fiscalCredit, FiscalPeriodOpenedWallMS: 0,
		FiscalGeneratorLevels: map[string]int64{}, FiscalUnlocks: map[string]bool{}}
	for _, row := range bundle.Fiscal.GeneratorLevelRows() {
		founder.FiscalGeneratorLevels[row.GeneratorID] = 0
	}
	frozen, err := FrozenFiscalContributions(bundle.Fiscal, founder)
	if err != nil {
		return ContentDynamicsPermitResult{}, err
	}
	external, err := ResolveFrozenContributions(bundle.Economy, frozen)
	if err != nil {
		return ContentDynamicsPermitResult{}, err
	}
	transitions := int64(0)
	find := func(target decimal.Decimal) (int64, error) {
		low, high := int64(0), horizonMS
		for low < high {
			mid := low + (high-low)/2
			candidate, cloneErr := cloneReplayState(state, bundle.Economy)
			if cloneErr != nil {
				return 0, cloneErr
			}
			contributions, contributionErr := assembleContributions(candidate, bundle.Economy, external)
			if contributionErr != nil {
				return 0, contributionErr
			}
			if _, evaluateErr := Evaluate(candidate, bundle.Economy, contentDynamicsEpoch.Add(time.Duration(mid)*time.Millisecond), ModeOnline, contributions); evaluateErr != nil {
				return 0, evaluateErr
			}
			transitions++
			balance, ok := candidate.Ledger.Balance("company.permits")
			if !ok {
				return 0, ErrInvalidEngineState
			}
			if balance.Gte(target) {
				high = mid
			} else {
				low = mid + 1
			}
		}
		candidate, cloneErr := cloneReplayState(state, bundle.Economy)
		if cloneErr != nil {
			return 0, cloneErr
		}
		contributions, contributionErr := assembleContributions(candidate, bundle.Economy, external)
		if contributionErr != nil {
			return 0, contributionErr
		}
		if _, evaluateErr := Evaluate(candidate, bundle.Economy, contentDynamicsEpoch.Add(time.Duration(low)*time.Millisecond), ModeOnline, contributions); evaluateErr != nil {
			return 0, evaluateErr
		}
		transitions++
		balance, ok := candidate.Ledger.Balance("company.permits")
		if !ok || !balance.Gte(target) {
			return 0, fmt.Errorf("%w: permits target outside horizon", ErrInvalidEngineState)
		}
		return low, nil
	}
	twelve, err := find(decimal.FromFloat64(12))
	if err != nil {
		return ContentDynamicsPermitResult{}, err
	}
	capTime, err := find(decimal.FromFloat64(24))
	if err != nil {
		return ContentDynamicsPermitResult{}, err
	}
	resource, ok := bundle.Economy.Resource("company.permits")
	if !ok || resource.Hardcap == nil || !resource.Hardcap.Amount.Eq(decimal.FromFloat64(24)) || resource.Hardcap.ReasonKey == "" {
		return ContentDynamicsPermitResult{}, ErrInvalidEngineState
	}
	return ContentDynamicsPermitResult{FiscalCredit: fiscalCredit, TimeToTwelveMS: twelve, TimeToCapMS: capTime,
		CapReasonKey: resource.Hardcap.ReasonKey, Transitions: transitions}, nil
}

var contentDynamicsEpoch = time.UnixMilli(1_800_000_000_000).UTC()

func newContentDynamicsCompany(catalog *economy.Catalog, version int, runSeq int64) (*save.State, error) {
	ledger, err := economy.NewLedger(catalog, economy.ScopeCompany)
	if err != nil {
		return nil, err
	}
	counts, provisioned, remainders := map[string]int64{}, map[string]int64{}, map[string]int64{}
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		counts[generator.ID], provisioned[generator.ID] = 0, 0
		if generator.Provision != nil {
			remainders[generator.Provision.GeneratorID] = 0
		}
	}
	return &save.State{WireVersion: version, Ledger: ledger, GeneratorCounts: counts, GeneratorProvisioned: provisioned,
		ProvisionRemaindersPPM: remainders, UpgradesOwned: map[string]bool{}, EvaluatedThrough: contentDynamicsEpoch,
		ManualTokenRefilledAt: contentDynamicsEpoch, RunStartedAt: contentDynamicsEpoch, RunSeq: runSeq, GatesCrossed: map[string]bool{},
		DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{},
		MeterValues: map[string]int{}, MeterDecayRemainders: map[string]int64{}, MeterInputRemainders: map[string]int64{},
		AchievementsEarnedRun: map[string]bool{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{},
		ActiveBuffs: []save.ActiveBuff{}, OfflineSpans: []save.OfflineSpan{}}, nil
}

func newContentDynamicsFounder(bundle CatalogBundle, openedWallMS int64) (*save.State, error) {
	ledger, err := economy.NewLedger(bundle.Economy, economy.ScopeFounder)
	if err != nil {
		return nil, err
	}
	counts, provisioned, remainders := map[string]int64{}, map[string]int64{}, map[string]int64{}
	for _, generator := range bundle.Economy.GeneratorClassesForScope(economy.ScopeFounder) {
		counts[generator.ID], provisioned[generator.ID] = 0, 0
		if generator.Provision != nil {
			remainders[generator.Provision.GeneratorID] = 0
		}
	}
	state := &save.State{WireVersion: 16, Ledger: ledger, GeneratorCounts: counts, GeneratorProvisioned: provisioned,
		ProvisionRemaindersPPM: remainders, UpgradesOwned: map[string]bool{}, EvaluatedThrough: contentDynamicsEpoch,
		ManualTokenRefilledAt: contentDynamicsEpoch, GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{},
		LedgerFactKinds: map[string]bool{}, MeterValues: map[string]int{}, MeterDecayRemainders: map[string]int64{},
		MeterInputRemainders: map[string]int64{}, AchievementsEarnedRun: map[string]bool{}, AchievementsEarnedLifetime: map[string]bool{},
		RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{}, CompactSamples: []save.CompactSample{},
		LifetimeValue: decimal.Zero, OfflineSpans: []save.OfflineSpan{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}}
	band, ok := bundle.Soul.BandFor(bundle.Soul.Policy.Initial)
	if !ok {
		return nil, ErrInvalidEngineState
	}
	if err := activateFounderFeatureState(state, bundle, 21, openedWallMS,
		&nextSoulWire{SoulInitial: bundle.Soul.Policy.Initial, BandMember: string(band.Member)}); err != nil {
		return nil, err
	}
	state.WireVersion = 21
	if err := bundle.ValidateFoundationState(state); err != nil {
		return nil, err
	}
	return state, nil
}
