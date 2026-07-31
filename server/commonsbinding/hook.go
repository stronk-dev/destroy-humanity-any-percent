package commonsbinding

import (
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"cloud-clicker/server/accrualhook"
	"cloud-clicker/server/commons"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

type WeightResolver interface {
	CompactWeightPPM(founderID string) (int64, bool)
}

type Hook struct {
	Catalogs commons.CatalogSet
	Weights  WeightResolver
}

// ResolvedHook is the replay-safe form: both catalog and projection-derived
// participation weight are already frozen in ApplyLogged's arguments.
type ResolvedHook struct {
	Catalog   *commons.Catalog
	WeightPPM *int64
}

// ReplayPolicy is the amplitude-safe adapter exposed to the production
// replay bundle. Production sees only tithe bounds plus a neutral hook; the
// Commons catalog and arithmetic remain owned by this package.
type ReplayPolicy struct{ Catalog *commons.Catalog }

func (policy ReplayPolicy) MinimumTithePPM() int64 {
	if policy.Catalog == nil {
		return -1
	}
	return policy.Catalog.MinimumTithePPM
}

func (policy ReplayPolicy) MaximumTithePPM() int64 {
	if policy.Catalog == nil {
		return -1
	}
	return policy.Catalog.MaximumTithePPM
}

func (policy ReplayPolicy) ResolvedAccrualHook(weightPPM *int64) accrualhook.Hook {
	return ResolvedHook{Catalog: policy.Catalog, WeightPPM: weightPPM}
}

func (hook Hook) AfterAccrual(state *save.State, _ *economy.Catalog, revision save.Revision, result accrualhook.Result, contributions []multiplier.Contribution) ([]save.EventWrite, error) {
	if state == nil || !state.CompactMember || result.ProductionMS <= 0 {
		return nil, nil
	}
	catalog, ok := hook.Catalogs.ResolveCommons(revision.ConstantsHash)
	if !ok || hook.Weights == nil {
		return nil, errors.New("commons accrual policy unavailable")
	}
	weightPPM, ok := hook.Weights.CompactWeightPPM(revision.OwnerID)
	if !ok || weightPPM < 0 || weightPPM > commons.PPM {
		return nil, errors.New("commons participation weight unavailable")
	}
	return applyResolved(state, revision, result, contributions, catalog, weightPPM)
}

func (hook ResolvedHook) AfterAccrual(state *save.State, _ *economy.Catalog, revision save.Revision, result accrualhook.Result, contributions []multiplier.Contribution) ([]save.EventWrite, error) {
	if state == nil || !state.CompactMember || result.ProductionMS <= 0 {
		return nil, nil
	}
	if hook.Catalog == nil || hook.WeightPPM == nil || *hook.WeightPPM < 0 || *hook.WeightPPM > commons.PPM {
		return nil, errors.New("commons resolved input unavailable")
	}
	return applyResolved(state, revision, result, contributions, hook.Catalog, *hook.WeightPPM)
}

func applyResolved(state *save.State, revision save.Revision, result accrualhook.Result, contributions []multiplier.Contribution, catalog *commons.Catalog, weightPPM int64) ([]save.EventWrite, error) {
	enclosure, err := commons.EnclosureIndex(catalog, contributions)
	if err != nil {
		return nil, err
	}
	compliancePPM, err := commons.CompliancePPM(state.CompactTithePPM, catalog.DefaultTithePPM, enclosure)
	if err != nil {
		return nil, err
	}
	end := state.EvaluatedThrough
	start := end.Add(-time.Duration(result.ProductionMS) * time.Millisecond)
	state.CompactSamples = appendSamples(state.CompactSamples, start, end, compliancePPM, catalog.SolidarityWindowMS)
	state.CompactSolidarityPPM = solidarityPPM(state.CompactSamples, catalog.SolidarityWindowMS)
	capacity := decimal.Zero
	for _, change := range result.Receipt.Changes {
		delta, err := decimal.ParseCanonical(change.Delta)
		if err != nil || delta.Lt(decimal.Zero) {
			return nil, errors.New("invalid accrual receipt")
		}
		capacity = capacity.Add(delta.Mul(decimal.New(float64(state.CompactTithePPM), -6)))
	}
	capacity = capacity.Quantize(decimal.CanonicalSignificantDigits)
	if !capacity.IsStateValue() {
		return nil, errors.New("commons capacity outside Decimal domain")
	}
	payload, _ := json.Marshal(map[string]any{
		"founder_id": revision.OwnerID,
		"run_id":     map[string]any{"company_stream_id": revision.StreamID, "run_seq": state.RunSeq},
		"weight_ppm": weightPPM, "compliance_ppm": compliancePPM, "enclosure": enclosure.String(),
		"capacity": capacity.String(), "solidarity_ppm": state.CompactSolidarityPPM, "sampled_ms": result.ProductionMS,
	})
	return []save.EventWrite{{Kind: save.EventCompactSampled, SchemaVersion: 1, Payload: payload}}, nil
}

func appendSamples(samples []save.CompactSample, start, end time.Time, compliancePPM, windowMS int64) []save.CompactSample {
	for start.Before(end) {
		hour := start.UTC().Truncate(time.Hour)
		boundary := hour.Add(time.Hour)
		if boundary.After(end) {
			boundary = end
		}
		covered := boundary.Sub(start).Milliseconds()
		if covered > 0 {
			if len(samples) > 0 && samples[len(samples)-1].HourStart.Equal(hour) {
				last := &samples[len(samples)-1]
				numerator := new(big.Int).Mul(big.NewInt(last.CompliancePPM), big.NewInt(last.CoveredMS))
				numerator.Add(numerator, new(big.Int).Mul(big.NewInt(compliancePPM), big.NewInt(covered)))
				last.CoveredMS += covered
				last.CompliancePPM = new(big.Int).Quo(numerator, big.NewInt(last.CoveredMS)).Int64()
			} else {
				samples = append(samples, save.CompactSample{HourStart: hour, CompliancePPM: compliancePPM, CoveredMS: covered})
			}
		}
		start = boundary
	}
	cutoff := end.Add(-time.Duration(windowMS) * time.Millisecond).UTC().Truncate(time.Hour)
	first := 0
	for first < len(samples) && samples[first].HourStart.Before(cutoff) {
		first++
	}
	return append([]save.CompactSample(nil), samples[first:]...)
}

func solidarityPPM(samples []save.CompactSample, windowMS int64) int64 {
	numerator := new(big.Int)
	for _, sample := range samples {
		numerator.Add(numerator, new(big.Int).Mul(big.NewInt(sample.CompliancePPM), big.NewInt(sample.CoveredMS)))
	}
	value := numerator.Quo(numerator, big.NewInt(windowMS))
	if !value.IsInt64() {
		return commons.PPM
	}
	result := value.Int64()
	if result < 0 {
		return 0
	}
	if result > commons.PPM {
		return commons.PPM
	}
	return result
}
