package prestige

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"sort"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/runidentity"
	"cloud-clicker/server/save"
)

type CatalogSet map[string]*Policy

func (set CatalogSet) ResolvePrestige(constantsHash string) (*Policy, bool) {
	policy, ok := set[constantsHash]
	return policy, ok
}

type Terms struct {
	ReputationDelta    int64              `json:"reputation_delta"`
	NetworkSlotUnlocks []save.NetworkSlot `json:"network_slot_unlocks"`
	RouteKnowledge     int64              `json:"route_knowledge"`
	CloutReachNote     string             `json:"clout_reach_note"`
}

type StoredOfferTerms struct {
	PayoutPreview     Terms `json:"payout_preview"`
	MarketModifierPPM int64 `json:"market_modifier_ppm"`
}

func ComputeTerms(company, founder *save.State, policy *Policy, exitType string) (Terms, error) {
	if company == nil || founder == nil || policy == nil {
		return Terms{}, ErrInvalidArithmetic
	}
	modifier, ok := policy.Modifier(exitType)
	if !ok {
		return Terms{}, ErrInvalidArithmetic
	}
	delta, err := ReputationDelta(company.LifetimeValue, policy.ThresholdValue(), founder.ReputationLevel, modifier)
	if err != nil {
		return Terms{}, err
	}
	knowledge := int64(0)
	if exitType == "collapse" || exitType == "scripted_first" {
		knowledge = policy.CollapseRouteKnowledge
	}
	return Terms{ReputationDelta: delta, NetworkSlotUnlocks: []save.NetworkSlot{}, RouteKnowledge: knowledge, CloutReachNote: "clout.reach.preserved"}, nil
}

func PromiseTerms(preview, current Terms) Terms {
	result := current
	if preview.ReputationDelta > result.ReputationDelta {
		result.ReputationDelta = preview.ReputationDelta
	}
	if preview.RouteKnowledge > result.RouteKnowledge {
		result.RouteKnowledge = preview.RouteKnowledge
	}
	bySlot := make(map[string]save.NetworkSlot, len(preview.NetworkSlotUnlocks)+len(current.NetworkSlotUnlocks))
	for _, slot := range current.NetworkSlotUnlocks {
		bySlot[slot.Slot] = slot
	}
	for _, slot := range preview.NetworkSlotUnlocks {
		bySlot[slot.Slot] = slot
	}
	keys := make([]string, 0, len(bySlot))
	for key := range bySlot {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result.NetworkSlotUnlocks = make([]save.NetworkSlot, 0, len(keys))
	for _, key := range keys {
		result.NetworkSlotUnlocks = append(result.NetworkSlotUnlocks, bySlot[key])
	}
	return result
}

func DecodeTerms(data []byte) (Terms, error) {
	var terms Terms
	if err := decodeStrict(data, &terms); err != nil {
		return Terms{}, ErrInvalidArithmetic
	}
	if terms.ReputationDelta < 0 || terms.ReputationDelta > decimal.MaxExactInteger || terms.RouteKnowledge < 0 || terms.RouteKnowledge > decimal.MaxExactInteger || terms.CloutReachNote == "" {
		return Terms{}, ErrInvalidArithmetic
	}
	last := ""
	for _, slot := range terms.NetworkSlotUnlocks {
		if slot.Slot <= last || !mechanicalIDPattern.MatchString(slot.Slot) || !mechanicalIDPattern.MatchString(slot.CarriedRef) {
			return Terms{}, ErrInvalidArithmetic
		}
		last = slot.Slot
	}
	return terms, nil
}

func DecodeStoredOfferTerms(data []byte) (StoredOfferTerms, error) {
	var stored StoredOfferTerms
	if err := decodeStrict(data, &stored); err != nil {
		return StoredOfferTerms{}, ErrInvalidArithmetic
	}
	if stored.MarketModifierPPM < 0 || stored.MarketModifierPPM > 2_000_000 {
		return StoredOfferTerms{}, ErrInvalidArithmetic
	}
	preview, err := json.Marshal(stored.PayoutPreview)
	if err != nil {
		return StoredOfferTerms{}, err
	}
	if _, err := DecodeTerms(preview); err != nil {
		return StoredOfferTerms{}, err
	}
	return stored, nil
}

func decodeStrict(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return ErrInvalidArithmetic
		}
		return err
	}
	return nil
}

func AccumulateLifetimeValue(state *save.State, receipt economy.Receipt, valueResourceID string) error {
	if state == nil || !mechanicalIDPattern.MatchString(valueResourceID) {
		return ErrInvalidArithmetic
	}
	for _, change := range receipt.Changes {
		if change.ResourceID != valueResourceID {
			continue
		}
		delta, err := decimal.ParseCanonical(change.Delta)
		if err != nil || delta.Lt(decimal.Zero) {
			return ErrInvalidArithmetic
		}
		state.LifetimeValue = state.LifetimeValue.Add(delta).Quantize(decimal.CanonicalSignificantDigits)
		if !state.LifetimeValue.IsStateValue() {
			return ErrInvalidArithmetic
		}
	}
	return nil
}

func RecordOfflineSpan(state *save.State, from, to time.Time, catchupCeilingMS int64) error {
	from, to = save.CanonicalServerTime(from), save.CanonicalServerTime(to)
	if state == nil || catchupCeilingMS <= 0 || !to.After(from) {
		return ErrInvalidArithmetic
	}
	if to.Sub(from).Milliseconds() <= catchupCeilingMS {
		return nil
	}
	span := save.OfflineSpan{From: from, To: to}
	if count := len(state.OfflineSpans); count > 0 && !from.After(state.OfflineSpans[count-1].To) {
		if to.After(state.OfflineSpans[count-1].To) {
			state.OfflineSpans[count-1].To = to
		}
		return nil
	}
	state.OfflineSpans = append(state.OfflineSpans, span)
	if len(state.OfflineSpans) > 256 {
		state.OfflineSpans[1].From = state.OfflineSpans[0].From
		state.OfflineSpans = append([]save.OfflineSpan(nil), state.OfflineSpans[1:]...)
	}
	return nil
}

func AttendedMS(state *save.State, endedAt time.Time) (int64, error) {
	endedAt = save.CanonicalServerTime(endedAt)
	if state == nil || state.RunStartedAt.IsZero() || endedAt.Before(state.RunStartedAt) {
		return 0, ErrInvalidArithmetic
	}
	rta := endedAt.Sub(state.RunStartedAt).Milliseconds()
	offline := int64(0)
	for _, span := range state.OfflineSpans {
		from, to := span.From, span.To
		if from.Before(state.RunStartedAt) {
			from = state.RunStartedAt
		}
		if to.After(endedAt) {
			to = endedAt
		}
		if to.After(from) {
			duration := to.Sub(from).Milliseconds()
			if duration > rta-offline {
				offline = rta
				break
			}
			offline += duration
		}
	}
	return rta - offline, nil
}

func NewRunState(catalog *economy.Catalog, priorCompany, founder *save.State, now time.Time) (*save.State, error) {
	if catalog == nil || priorCompany == nil || founder == nil || priorCompany.RunSeq < 1 || priorCompany.RunSeq >= decimal.MaxExactInteger {
		return nil, ErrInvalidArithmetic
	}
	now = save.CanonicalServerTime(now)
	ledger, err := economy.NewLedger(catalog, economy.ScopeCompany)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int64)
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		counts[generator.ID] = 0
	}
	reseed, err := MoralReseed(founder.Notoriety)
	if err != nil {
		return nil, err
	}
	state := &save.State{Ledger: ledger, GeneratorCounts: counts, EvaluatedThrough: now,
		ManualTokenMilli: catalog.ManualPolicy().BucketCapMilli, ManualTokenRefilledAt: now,
		GatesCrossed: map[string]bool{}, RunSeq: priorCompany.RunSeq + 1, DoctrinesByTransition: map[string]string{},
		LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{"trust.regulators.standing": int(reseed), "trust.regulators.grievance": int(100 - reseed)},
		RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{}, CompactSamples: []save.CompactSample{},
		LifetimeValue: decimal.Zero, RunStartedAt: now, OfflineSpans: []save.OfflineSpan{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}}
	return state, nil
}

func FounderSeed(founderID string, runSeq int64) uint64 {
	return runidentity.Seed(founderID, runSeq)
}

func OfferID(founderID string, runSeq, tier, declineCount int64, at time.Time) string {
	random := NewSplitMix64(FounderSeed(founderID, runSeq) ^ uint64(tier)<<32 ^ uint64(declineCount))
	var value [16]byte
	milliseconds := uint64(save.CanonicalServerTime(at).UnixMilli()) & ((1 << 48) - 1)
	value[0], value[1], value[2] = byte(milliseconds>>40), byte(milliseconds>>32), byte(milliseconds>>24)
	value[3], value[4], value[5] = byte(milliseconds>>16), byte(milliseconds>>8), byte(milliseconds)
	binary.BigEndian.PutUint64(value[6:14], random.Next())
	value[14], value[15] = byte(random.Next()>>56), byte(random.Next()>>48)
	value[6], value[8] = value[6]&0x0f|0x70, value[8]&0x3f|0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", binary.BigEndian.Uint32(value[0:4]), binary.BigEndian.Uint16(value[4:6]), binary.BigEndian.Uint16(value[6:8]), binary.BigEndian.Uint16(value[8:10]), uint64(value[10])<<40|uint64(value[11])<<32|uint64(value[12])<<24|uint64(value[13])<<16|uint64(value[14])<<8|uint64(value[15]))
}

func OfferDraws(founderID string, runSeq, tier, declineCount int64) (spawn int64, exitType string, driftUp bool) {
	random := NewSplitMix64(FounderSeed(founderID, runSeq) ^ uint64(tier)<<32 ^ uint64(declineCount))
	spawn = random.PPM()
	if random.Next()&1 == 0 {
		exitType = "acquihire"
	} else {
		exitType = "acquisition"
	}
	driftUp = random.Next()&1 == 0
	return
}

func DriftTerms(terms Terms, declineCount, stepPPM int64, up bool) Terms {
	return ApplyMarketModifier(terms, MarketModifierPPM(declineCount, stepPPM, up))
}

func MarketModifierPPM(declineCount, stepPPM int64, up bool) int64 {
	if declineCount <= 0 || stepPPM <= 0 {
		return 1_000_000
	}
	steps := declineCount
	if steps > 10 {
		steps = 10
	}
	delta := steps * stepPPM
	factor := int64(1_000_000)
	if up {
		factor += delta
	} else if delta < factor {
		factor -= delta
	} else {
		factor = 0
	}
	return factor
}

func ApplyMarketModifier(terms Terms, factor int64) Terms {
	if factor < 0 || factor > 2_000_000 {
		factor = 0
	}
	terms.ReputationDelta = ppmFloorCapped(terms.ReputationDelta, factor)
	terms.RouteKnowledge = ppmFloorCapped(terms.RouteKnowledge, factor)
	return terms
}

func ppmFloorCapped(value, factor int64) int64 {
	product := new(big.Int).Mul(big.NewInt(value), big.NewInt(factor))
	product.Quo(product, big.NewInt(1_000_000))
	cap := big.NewInt(decimal.MaxExactInteger)
	if product.Cmp(cap) > 0 {
		return decimal.MaxExactInteger
	}
	return product.Int64()
}
