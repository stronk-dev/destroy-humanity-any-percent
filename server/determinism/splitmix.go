package determinism

import "hash/fnv"

const splitMixIncrement uint64 = 0x9e3779b97f4a7c15

type SplitMix64 struct {
	state uint64
}

func NewSplitMix64(seed uint64) *SplitMix64 { return &SplitMix64{state: seed} }

func (random *SplitMix64) Next() uint64 {
	random.state += splitMixIncrement
	z := random.state
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

func (random *SplitMix64) Bound(bound uint64) uint64 {
	if bound == 0 {
		panic("SplitMix64 bound must be positive")
	}
	threshold := (-bound) % bound
	for {
		draw := random.Next()
		if draw >= threshold {
			return draw % bound
		}
	}
}

func BattleSeed(matchSeed uint64) uint64 { return NewSplitMix64(matchSeed).Next() }

func Substream(battleSeed uint64, label string) *SplitMix64 {
	if label == "" {
		panic("SplitMix64 substream label must not be empty")
	}
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(label))
	return NewSplitMix64(battleSeed ^ hash.Sum64())
}
