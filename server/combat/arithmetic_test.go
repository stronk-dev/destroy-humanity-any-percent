package combat

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"cloud-clicker/server/determinism"
)

type arithmeticFixture struct {
	Version int `json:"version"`
	Damage  []struct {
		Name        string      `json:"name"`
		BasePower   int32       `json:"base_power"`
		AttackerATK int32       `json:"attacker_atk"`
		Chart       ChartResult `json:"chart"`
		Critical    bool        `json:"critical"`
		Expected    int32       `json:"expected"`
	} `json:"damage"`
	RNG struct {
		MatchSeed  string            `json:"match_seed"`
		BattleSeed string            `json:"battle_seed"`
		Substreams map[string]string `json:"substreams"`
	} `json:"rng"`
}

func TestSharedArithmeticVectors(t *testing.T) {
	fixture := loadArithmeticFixture(t)
	if fixture.Version != 1 {
		t.Fatalf("fixture version=%d", fixture.Version)
	}
	for _, vector := range fixture.Damage {
		t.Run(vector.Name, func(t *testing.T) {
			actual, err := Damage(vector.BasePower, vector.AttackerATK, vector.Chart, vector.Critical)
			if err != nil || actual != vector.Expected {
				t.Fatalf("damage=%d want=%d err=%v", actual, vector.Expected, err)
			}
		})
	}
	if SaturateInt32(math.MaxInt64) != math.MaxInt32 || SaturateInt32(math.MinInt64) != math.MinInt32 {
		t.Fatal("int32 saturation failed")
	}
}

func TestTemperamentChartProperties(t *testing.T) {
	for _, attacker := range temperamentOrder {
		wins, losses := 0, 0
		for _, defender := range temperamentOrder {
			result, err := Chart(attacker, defender)
			if err != nil || result < Disadvantage || result > Advantage {
				t.Fatalf("chart %s/%s=%d err=%v", attacker, defender, result, err)
			}
			if result == Advantage {
				wins++
			}
			if result == Disadvantage {
				losses++
			}
		}
		if wins != 2 || losses != 2 {
			t.Fatalf("%s wins=%d losses=%d", attacker, wins, losses)
		}
	}
}

func TestLabeledSubstreamVectorsAndIsolation(t *testing.T) {
	fixture := loadArithmeticFixture(t)
	battle := determinism.BattleSeed(42)
	if formatUint(battle) != fixture.RNG.BattleSeed {
		t.Fatalf("battle seed=%d", battle)
	}
	for label, expected := range fixture.RNG.Substreams {
		if actual := determinism.Substream(battle, label).Next(); formatUint(actual) != expected {
			t.Fatalf("%s=%d want=%s", label, actual, expected)
		}
	}
	critBefore := determinism.Substream(battle, "crit").Next()
	_ = determinism.Substream(battle, "new_consumer").Next()
	critAfter := determinism.Substream(battle, "crit").Next()
	if critBefore != critAfter {
		t.Fatal("adding a consumer shifted the crit substream")
	}
}

func loadArithmeticFixture(t *testing.T) arithmeticFixture {
	t.Helper()
	data, err := os.ReadFile("../../testdata/combat/arithmetic-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture arithmeticFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func formatUint(value uint64) string {
	const digits = "0123456789"
	if value == 0 {
		return "0"
	}
	var buffer [20]byte
	index := len(buffer)
	for value > 0 {
		index--
		buffer[index] = digits[value%10]
		value /= 10
	}
	return string(buffer[index:])
}
