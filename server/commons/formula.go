package commons

import (
	"errors"
	"math"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/multiplier"
)

var ErrInvalidInput = errors.New("invalid commons formula input")

func ppmDecimal(value int64) decimal.Decimal { return decimal.New(float64(value), -6) }

func EnclosureIndex(catalog *Catalog, contributions []multiplier.Contribution) (decimal.Decimal, error) {
	if catalog == nil {
		return decimal.Zero, ErrInvalidInput
	}
	all, clean := decimal.One, decimal.One
	seen := map[string]bool{}
	for _, contribution := range contributions {
		if contribution.Slot == multiplier.SlotCommons {
			continue
		}
		weight, declared := catalog.SourceWeight(contribution.SourceID)
		if !declared {
			continue
		}
		if seen[contribution.SourceID] || contribution.Slot != weight.Slot || !contribution.Factor.IsStateValue() || !contribution.Factor.Gt(decimal.Zero) {
			return decimal.Zero, ErrInvalidInput
		}
		seen[contribution.SourceID] = true
		weighted := contribution.Factor.Pow(float64(weight.WeightPPM) / float64(PPM))
		if !weighted.IsStateValue() || !weighted.Gt(decimal.Zero) {
			return decimal.Zero, ErrInvalidInput
		}
		all = all.Mul(weighted)
		if !weight.Forsworn {
			clean = clean.Mul(weighted)
		}
	}
	if !all.IsStateValue() || !clean.IsStateValue() {
		return decimal.Zero, ErrInvalidInput
	}
	value := decimal.One.Sub(clean.Div(all))
	if value.Lt(decimal.Zero) {
		value = decimal.Zero
	}
	if value.Gt(decimal.One) {
		value = decimal.One
	}
	return value.Quantize(decimal.CanonicalSignificantDigits), nil
}

func CompliancePPM(tithePPM, targetPPM int64, enclosure decimal.Decimal) (int64, error) {
	if tithePPM < 0 || tithePPM > PPM || targetPPM <= 0 || targetPPM > PPM || !enclosure.IsStateValue() || enclosure.Lt(decimal.Zero) || enclosure.Gt(decimal.One) {
		return 0, ErrInvalidInput
	}
	ratio := ppmDecimal(tithePPM).Div(ppmDecimal(targetPPM))
	if ratio.Gt(decimal.One) {
		ratio = decimal.One
	}
	value := ratio.Mul(decimal.One.Sub(enclosure))
	result := int64(math.Floor(value.Mantissa()*math.Pow10(int(value.Exponent()+6)) + 1e-9))
	if value.Eq(decimal.One) {
		result = PPM
	}
	if result < 0 {
		result = 0
	}
	if result > PPM {
		result = PPM
	}
	return result, nil
}

func EffectiveHealthPPM(catalog *Catalog, guildPPM, cohortPPM, serverPPM int64, hasGuild bool) (int64, error) {
	for _, value := range []int64{guildPPM, cohortPPM, serverPPM} {
		if value < 0 || value > PPM {
			return 0, ErrInvalidInput
		}
	}
	if !hasGuild {
		guildPPM = cohortPPM
	}
	return (guildPPM*catalog.GuildHealthWeightPPM + cohortPPM*catalog.CohortHealthWeightPPM + serverPPM*catalog.ServerHealthWeightPPM) / PPM, nil
}

func Modifier(catalog *Catalog, healthPPM, solidarityPPM int64) (decimal.Decimal, error) {
	if catalog == nil || healthPPM < 0 || healthPPM > PPM || solidarityPPM < 0 || solidarityPPM > PPM {
		return decimal.Zero, ErrInvalidInput
	}
	collective := decimal.Zero
	if healthPPM > catalog.CollapseHealthPPM {
		x := ppmDecimal(healthPPM - catalog.CollapseHealthPPM).Div(ppmDecimal(PPM - catalog.CollapseHealthPPM))
		collective = x.Mul(x.Pow(0.5))
	}
	personalWeight := PPM - catalog.CollectiveWeightPPM
	inside := ppmDecimal(catalog.CollectiveWeightPPM).Mul(collective).Add(ppmDecimal(personalWeight).Mul(ppmDecimal(solidarityPPM)))
	result := decimal.One.Add(catalog.MaximumBonus.Mul(inside)).Quantize(decimal.CanonicalSignificantDigits)
	if !result.IsStateValue() || result.Lt(decimal.One) {
		return decimal.Zero, ErrInvalidInput
	}
	return result, nil
}

func Contribution(catalog *Catalog, member bool, healthPPM, solidarityPPM int64) ([]multiplier.Contribution, error) {
	if !member {
		return nil, nil
	}
	factor, err := Modifier(catalog, healthPPM, solidarityPPM)
	if err != nil {
		return nil, err
	}
	return []multiplier.Contribution{{Slot: multiplier.SlotCommons, SourceID: "commons.member", Target: "all", Factor: factor}}, nil
}
