package prestige

import (
	"encoding/json"
	"os"
	"testing"

	"cloud-clicker/server/decimal"
)

type vectorFile struct {
	Version   int    `json:"version"`
	Threshold string `json:"threshold"`
	Vectors   []struct {
		Name          string `json:"name"`
		LifetimeValue string `json:"lifetime_value"`
		CurrentLevel  int64  `json:"current_level"`
		ModifierPPM   int64  `json:"modifier_ppm"`
		Level         int64  `json:"level"`
		Delta         int64  `json:"delta"`
	} `json:"vectors"`
}

func TestPrestigeVectors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/prestige-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture vectorFile
	if err := json.Unmarshal(data, &fixture); err != nil || fixture.Version != 1 || len(fixture.Vectors) < 8 {
		t.Fatalf("fixture=%+v err=%v", fixture, err)
	}
	threshold, _ := decimal.ParseCanonical(fixture.Threshold)
	for _, vector := range fixture.Vectors {
		t.Run(vector.Name, func(t *testing.T) {
			value, err := decimal.ParseCanonical(vector.LifetimeValue)
			if err != nil {
				t.Fatal(err)
			}
			level, err := ReputationLevel(value, threshold)
			if err != nil || level != vector.Level {
				t.Fatalf("level=%d want=%d err=%v", level, vector.Level, err)
			}
			delta, err := ReputationDelta(value, threshold, vector.CurrentLevel, vector.ModifierPPM)
			if err != nil || delta != vector.Delta {
				t.Fatalf("delta=%d want=%d err=%v", delta, vector.Delta, err)
			}
		})
	}
}

func TestMoralReseedAndAdvisorBounds(t *testing.T) {
	for _, test := range []struct{ notoriety, want int64 }{{0, 90}, {1, 90}, {100, 55}, {9007199254740991, 55}} {
		got, err := MoralReseed(test.notoriety)
		if err != nil || got != test.want {
			t.Fatalf("notoriety=%d got=%d want=%d err=%v", test.notoriety, got, test.want, err)
		}
	}
	got, err := AdvisorMultiplierPPM(100, 20_000, 500_000)
	if err != nil || got != 1_500_000 {
		t.Fatalf("advisor=%d err=%v", got, err)
	}
}

func TestSplitMix64Stable(t *testing.T) {
	random := NewSplitMix64(0)
	want := []uint64{0xe220a8397b1dcdaf, 0x6e789e6aa1b965f4, 0x06c45d188009454f}
	for index, expected := range want {
		if got := random.Next(); got != expected {
			t.Fatalf("draw %d=%x want=%x", index, got, expected)
		}
	}
}

func TestPhase0PolicyLoadsStrictly(t *testing.T) {
	data, err := os.ReadFile("../../balance/prestige/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	policy, err := LoadPolicy(data)
	if err != nil || policy.ThresholdValue().String() != "1e12" || policy.SpawnGatePPM[1] != 300_000 {
		t.Fatalf("policy=%+v err=%v", policy, err)
	}
	if _, err := LoadPolicy(append(data[:len(data)-2], []byte(",\"unknown\":true}\n")...)); err == nil {
		t.Fatal("unknown policy field accepted")
	}
}
