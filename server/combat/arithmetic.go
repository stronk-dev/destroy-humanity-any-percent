package combat

import (
	"errors"
	"math"
)

var ErrInvalidArithmetic = errors.New("invalid combat arithmetic")

type Temperament string

const (
	Lazy    Temperament = "lazy"
	Playful Temperament = "playful"
	Curious Temperament = "curious"
	Sassy   Temperament = "sassy"
	Shy     Temperament = "shy"
	Chaotic Temperament = "chaotic"
)

type ChartResult int8

const (
	Disadvantage ChartResult = -1
	Neutral      ChartResult = 0
	Advantage    ChartResult = 1
)

var temperamentOrder = [...]Temperament{Lazy, Playful, Curious, Sassy, Shy, Chaotic}

func Chart(attacker, defender Temperament) (ChartResult, error) {
	attackerIndex, attackerOK := temperamentIndex(attacker)
	defenderIndex, defenderOK := temperamentIndex(defender)
	if !attackerOK || !defenderOK {
		return Neutral, ErrInvalidArithmetic
	}
	delta := (defenderIndex - attackerIndex + len(temperamentOrder)) % len(temperamentOrder)
	switch delta {
	case 1, 2:
		return Advantage, nil
	case 4, 5:
		return Disadvantage, nil
	default:
		return Neutral, nil
	}
}

func Damage(basePower, attackerATK int32, chart ChartResult, critical bool) (int32, error) {
	if basePower < 0 || attackerATK < 0 || chart < Disadvantage || chart > Advantage {
		return 0, ErrInvalidArithmetic
	}
	if basePower == 0 {
		return 0, nil
	}
	value := int64(basePower)
	value = floorRatio(value, int64(attackerATK), 64)
	switch chart {
	case Advantage:
		value = floorRatio(value, 13, 10)
	case Disadvantage:
		value = floorRatio(value, 10, 13)
	}
	if critical {
		value = floorRatio(value, 3, 2)
	}
	if value < 1 {
		value = 1
	}
	return SaturateInt32(value), nil
}

func SaturateInt32(value int64) int32 {
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	if value < math.MinInt32 {
		return math.MinInt32
	}
	return int32(value)
}

func Clamp(value, minimum, maximum int64) (int32, error) {
	if minimum < 0 || maximum < minimum || maximum > math.MaxInt32 {
		return 0, ErrInvalidArithmetic
	}
	if value < minimum {
		value = minimum
	}
	if value > maximum {
		value = maximum
	}
	return int32(value), nil
}

func floorRatio(value, numerator, denominator int64) int64 {
	if value <= 0 || numerator <= 0 || denominator <= 0 {
		return 0
	}
	return value * numerator / denominator
}

func temperamentIndex(value Temperament) (int, bool) {
	for index, candidate := range temperamentOrder {
		if value == candidate {
			return index, true
		}
	}
	return 0, false
}
