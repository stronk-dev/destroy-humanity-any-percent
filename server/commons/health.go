package commons

import (
	"math"
	"math/big"

	"cloud-clicker/server/decimal"
)

type MemberSample struct {
	WeightPPM     int64
	CompliancePPM int64
	Capacity      decimal.Decimal
}

type Aggregate struct {
	HealthPPM    int64
	Capacity     decimal.Decimal
	RealMembers  int
	NPCWeightPPM int64
	Numerator    *big.Int
	Denominator  *big.Int
}

func AggregateHealth(catalog *Catalog, samples []MemberSample) (Aggregate, error) {
	if catalog == nil {
		return Aggregate{}, ErrInvalidInput
	}
	numerator, denominator := new(big.Int), new(big.Int)
	capacities := make([]decimal.Decimal, 0, len(samples))
	for _, sample := range samples {
		if sample.WeightPPM < 0 || sample.WeightPPM > PPM || sample.CompliancePPM < 0 || sample.CompliancePPM > PPM || !sample.Capacity.IsStateValue() || sample.Capacity.Lt(decimal.Zero) {
			return Aggregate{}, ErrInvalidInput
		}
		numerator.Add(numerator, new(big.Int).Mul(big.NewInt(sample.WeightPPM), big.NewInt(sample.CompliancePPM)))
		denominator.Add(denominator, big.NewInt(sample.WeightPPM))
		capacities = append(capacities, sample.Capacity)
	}
	npcWeight := int64(0)
	if len(samples) < catalog.NPCPopulationFloor {
		npcWeight = int64(catalog.NPCPopulationFloor-len(samples)) * catalog.NPCWeightPPM
		numerator.Add(numerator, new(big.Int).Mul(big.NewInt(npcWeight), big.NewInt(catalog.NPCCompliancePPM)))
		denominator.Add(denominator, big.NewInt(npcWeight))
	}
	health := int64(0)
	if denominator.Sign() > 0 {
		health = new(big.Int).Quo(numerator, denominator).Int64()
	}
	capacity := decimal.SumDeterministic(capacities).Quantize(decimal.CanonicalSignificantDigits)
	if !capacity.IsStateValue() {
		return Aggregate{}, ErrInvalidInput
	}
	return Aggregate{HealthPPM: health, Capacity: capacity, RealMembers: len(samples), NPCWeightPPM: npcWeight, Numerator: new(big.Int).Set(numerator), Denominator: new(big.Int).Set(denominator)}, nil
}

func SmoothHealthPPM(catalog *Catalog, previousPPM, rawPPM int64, elapsedMS int64) (int64, error) {
	if catalog == nil || previousPPM < 0 || previousPPM > PPM || rawPPM < 0 || rawPPM > PPM || elapsedMS < 0 {
		return 0, ErrInvalidInput
	}
	if elapsedMS == 0 || previousPPM == rawPPM {
		return previousPPM, nil
	}
	rate := catalog.HealthRecoveryPPMPerHour
	if rawPPM < previousPPM {
		rate = catalog.HealthDecayPPMPerHour
	}
	fraction := 1 - math.Pow(1-float64(rate)/float64(PPM), float64(elapsedMS)/float64(timeHourMS))
	value := float64(previousPPM) + float64(rawPPM-previousPPM)*fraction
	result := int64(math.Floor(value + 1e-9))
	if result < 0 {
		result = 0
	}
	if result > PPM {
		result = PPM
	}
	return result, nil
}

const timeHourMS = int64(3_600_000)
