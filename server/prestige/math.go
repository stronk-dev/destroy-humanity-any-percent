package prestige

import (
	"errors"
	"math/big"

	"cloud-clicker/server/decimal"
)

var ErrInvalidArithmetic = errors.New("invalid prestige arithmetic")

func ReputationLevel(lifetimeValue, threshold decimal.Decimal) (int64, error) {
	if !lifetimeValue.IsStateValue() || lifetimeValue.Lt(decimal.Zero) || !threshold.IsStateValue() || !threshold.Gt(decimal.Zero) {
		return 0, ErrInvalidArithmetic
	}
	ratio := lifetimeValue.Div(threshold)
	if !ratio.IsStateValue() || ratio.Lt(decimal.One) {
		return 0, nil
	}
	low, high := int64(1), int64(decimal.MaxExactInteger)
	for low < high {
		mid := low + (high-low+1)/2
		candidate := decimal.FromFloat64(float64(mid))
		cube := candidate.Mul(candidate).Mul(candidate)
		if cube.IsStateValue() && cube.Lte(ratio) {
			low = mid
		} else {
			high = mid - 1
		}
	}
	return low, nil
}

func ReputationDelta(lifetimeValue, threshold decimal.Decimal, currentLevel, modifierPPM int64) (int64, error) {
	if currentLevel < 0 || currentLevel > decimal.MaxExactInteger || modifierPPM < 0 || modifierPPM > decimal.MaxExactInteger {
		return 0, ErrInvalidArithmetic
	}
	level, err := ReputationLevel(lifetimeValue, threshold)
	if err != nil {
		return 0, err
	}
	if level <= currentLevel {
		return 0, nil
	}
	delta := level - currentLevel
	product := new(big.Int).Mul(big.NewInt(delta), big.NewInt(modifierPPM))
	product.Quo(product, big.NewInt(1_000_000))
	if product.Cmp(big.NewInt(decimal.MaxExactInteger)) >= 0 {
		return decimal.MaxExactInteger, nil
	}
	return product.Int64(), nil
}

func MoralReseed(notoriety int64) (int64, error) {
	if notoriety < 0 || notoriety > decimal.MaxExactInteger {
		return 0, ErrInvalidArithmetic
	}
	value := int64(90)
	if notoriety > decimal.MaxExactInteger/35 {
		value = 55
	} else {
		value -= notoriety * 35 / 100
	}
	if value < 55 {
		return 55, nil
	}
	if value > 90 {
		return 90, nil
	}
	return value, nil
}

func AdvisorMultiplierPPM(completedRuns, perRunPPM, capPPM int64) (int64, error) {
	if completedRuns < 0 || completedRuns > decimal.MaxExactInteger || perRunPPM < 0 || perRunPPM > 1_000_000 || capPPM < 0 || capPPM > 1_000_000 {
		return 0, ErrInvalidArithmetic
	}
	bonus := capPPM
	if perRunPPM == 0 || completedRuns <= capPPM/perRunPPM {
		bonus = completedRuns * perRunPPM
	}
	if bonus > capPPM {
		bonus = capPPM
	}
	return 1_000_000 + bonus, nil
}

type SplitMix64 struct{ state uint64 }

func NewSplitMix64(seed uint64) *SplitMix64 { return &SplitMix64{state: seed} }

func (random *SplitMix64) Next() uint64 {
	random.state += 0x9e3779b97f4a7c15
	z := random.state
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

func (random *SplitMix64) PPM() int64 {
	bound := uint64(1_000_000)
	threshold := (-bound) % bound
	for {
		draw := random.Next()
		if draw >= threshold {
			return int64(draw % bound)
		}
	}
}
