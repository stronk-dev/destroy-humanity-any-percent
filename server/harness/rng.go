package harness

import (
	"encoding/binary"
	"fmt"
)

type SplitMix64 struct {
	state uint64
}

func NewSplitMix64(seed uint64) *SplitMix64 { return &SplitMix64{state: seed} }

func (random *SplitMix64) Next() uint64 {
	random.state += 0x9e3779b97f4a7c15
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

type UUIDStream struct {
	random *SplitMix64
	seen   map[string]bool
}

func NewUUIDStream(seed uint64) *UUIDStream {
	return &UUIDStream{random: NewSplitMix64(seed ^ 0xd1b54a32d192ed03), seen: make(map[string]bool)}
}

func (stream *UUIDStream) Next(unixMS int64) (string, error) {
	var value [16]byte
	timestamp := uint64(unixMS) & ((1 << 48) - 1)
	value[0] = byte(timestamp >> 40)
	value[1] = byte(timestamp >> 32)
	value[2] = byte(timestamp >> 24)
	value[3] = byte(timestamp >> 16)
	value[4] = byte(timestamp >> 8)
	value[5] = byte(timestamp)
	binary.BigEndian.PutUint64(value[6:14], stream.random.Next())
	value[14] = byte(stream.random.Next() >> 56)
	value[15] = byte(stream.random.Next() >> 48)
	value[6] = value[6]&0x0f | 0x70
	value[8] = value[8]&0x3f | 0x80
	id := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(value[0:4]), binary.BigEndian.Uint16(value[4:6]),
		binary.BigEndian.Uint16(value[6:8]), binary.BigEndian.Uint16(value[8:10]),
		uint64(value[10])<<40|uint64(value[11])<<32|uint64(value[12])<<24|
			uint64(value[13])<<16|uint64(value[14])<<8|uint64(value[15]))
	if stream.seen[id] {
		return "", fmt.Errorf("deterministic UUID collision: %s", id)
	}
	stream.seen[id] = true
	return id, nil
}
