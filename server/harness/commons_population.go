package harness

import (
	"math"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/decimal"
)

type CommonsPopulationResult struct {
	Population      int   `json:"population"`
	MeanModifierPPM int64 `json:"mean_modifier_ppm"`
	ServerHealthPPM int64 `json:"server_health_ppm"`
}

func SimulateCommonsPopulation(catalog *commons.Catalog, population int, seed uint64) (CommonsPopulationResult, error) {
	baseSize := 200
	random := NewSplitMix64(seed)
	base := make([]commons.MemberSample, baseSize)
	for index := range base {
		base[index] = commons.MemberSample{WeightPPM: 500_000 + int64(random.Bound(500_001)), CompliancePPM: 600_000 + int64(random.Bound(400_001)), Capacity: decimal.One}
	}
	samples := make([]commons.MemberSample, population)
	for index := range samples {
		samples[index] = base[index%baseSize]
	}
	server, err := commons.AggregateHealth(catalog, samples)
	if err != nil {
		return CommonsPopulationResult{}, err
	}
	var modifierTotal int64
	for start := 0; start < population; start += catalog.CohortTargetSize {
		end := start + catalog.CohortTargetSize
		if end > population {
			end = population
		}
		cohort, err := commons.AggregateHealth(catalog, samples[start:end])
		if err != nil {
			return CommonsPopulationResult{}, err
		}
		health := (cohort.HealthPPM*800_000 + server.HealthPPM*200_000) / commons.PPM
		factor, err := commons.Modifier(catalog, health, commons.PPM)
		if err != nil {
			return CommonsPopulationResult{}, err
		}
		factorPPM := int64(math.Round(factor.Mantissa() * math.Pow10(int(factor.Exponent())) * float64(commons.PPM)))
		modifierTotal += factorPPM * int64(end-start)
	}
	return CommonsPopulationResult{Population: population, MeanModifierPPM: modifierTotal / int64(population), ServerHealthPPM: server.HealthPPM}, nil
}
