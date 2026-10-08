package combat

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
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
	Saturation []struct {
		Value    string `json:"value"`
		Expected int32  `json:"expected"`
	} `json:"saturation"`
	Clamp []struct {
		Name     string `json:"name"`
		Value    string `json:"value"`
		Minimum  int64  `json:"minimum"`
		Maximum  int64  `json:"maximum"`
		Expected int32  `json:"expected"`
	} `json:"clamp"`
	InvalidClamp []struct {
		Name    string `json:"name"`
		Value   string `json:"value"`
		Minimum int64  `json:"minimum"`
		Maximum int64  `json:"maximum"`
	} `json:"invalid_clamp"`
	InvalidDamage []struct {
		Name        string      `json:"name"`
		BasePower   int32       `json:"base_power"`
		AttackerATK int32       `json:"attacker_atk"`
		Chart       ChartResult `json:"chart"`
		Critical    bool        `json:"critical"`
	} `json:"invalid_damage"`
	Chart struct {
		Temperaments []Temperament   `json:"temperaments"`
		Rows         [][]ChartResult `json:"rows"`
	} `json:"chart"`
	RNG struct {
		MatchSeed  string            `json:"match_seed"`
		BattleSeed string            `json:"battle_seed"`
		Substreams map[string]string `json:"substreams"`
		Bounded    []struct {
			Label    string `json:"label"`
			Bound    string `json:"bound"`
			Expected string `json:"expected"`
			Draws    int    `json:"draws"`
		} `json:"bounded"`
	} `json:"rng"`
}

func TestSharedArithmeticVectors(t *testing.T) {
	fixture := loadArithmeticFixture(t)
	if fixture.Version != 1 {
		t.Fatalf("fixture version=%d", fixture.Version)
	}
	if len(fixture.Damage) != 13 {
		t.Fatal("shared damage population missing")
	}
	for _, vector := range fixture.Damage {
		t.Run(vector.Name, func(t *testing.T) {
			actual, err := Damage(vector.BasePower, vector.AttackerATK, vector.Chart, vector.Critical)
			if err != nil || actual != vector.Expected {
				t.Fatalf("damage=%d want=%d err=%v", actual, vector.Expected, err)
			}
		})
	}
}

func TestSharedCombatStoreAndRefusalVectors(t *testing.T) {
	fixture := loadArithmeticFixture(t)
	if len(fixture.Saturation) != 9 || len(fixture.Clamp) != 13 || len(fixture.InvalidClamp) != 5 || len(fixture.InvalidDamage) != 6 {
		t.Fatal("shared boundary population missing")
	}
	parse := func(t *testing.T, value string) int64 {
		t.Helper()
		number, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		return number
	}
	for _, vector := range fixture.Saturation {
		t.Run("saturation_"+vector.Value, func(t *testing.T) {
			if got := SaturateInt32(parse(t, vector.Value)); got != vector.Expected {
				t.Fatalf("store=%d want=%d", got, vector.Expected)
			}
		})
	}
	for _, vector := range fixture.Clamp {
		t.Run("clamp_"+vector.Name, func(t *testing.T) {
			got, err := Clamp(parse(t, vector.Value), vector.Minimum, vector.Maximum)
			if err != nil || got != vector.Expected {
				t.Fatalf("clamp=%d want=%d err=%v", got, vector.Expected, err)
			}
		})
	}
	for _, vector := range fixture.InvalidClamp {
		t.Run("invalid_clamp_"+vector.Name, func(t *testing.T) {
			if _, err := Clamp(parse(t, vector.Value), vector.Minimum, vector.Maximum); !errors.Is(err, ErrInvalidArithmetic) {
				t.Fatalf("invalid bounds accepted: %v", err)
			}
		})
	}
	for _, vector := range fixture.InvalidDamage {
		t.Run("invalid_damage_"+vector.Name, func(t *testing.T) {
			if _, err := Damage(vector.BasePower, vector.AttackerATK, vector.Chart, vector.Critical); !errors.Is(err, ErrInvalidArithmetic) {
				t.Fatalf("invalid damage accepted: %v", err)
			}
		})
	}
}

func TestSharedTemperamentCycleVectors(t *testing.T) {
	fixture := loadArithmeticFixture(t)
	if len(fixture.Chart.Temperaments) != 6 || len(fixture.Chart.Rows) != 6 {
		t.Fatal("incomplete chart")
	}
	for row, attacker := range fixture.Chart.Temperaments {
		if len(fixture.Chart.Rows[row]) != 6 {
			t.Fatal("incomplete chart row")
		}
		for column, defender := range fixture.Chart.Temperaments {
			got, err := Chart(attacker, defender)
			if err != nil || got != fixture.Chart.Rows[row][column] {
				t.Fatalf("chart %s/%s=%d want=%d err=%v", attacker, defender, got, fixture.Chart.Rows[row][column], err)
			}
		}
	}
	for _, pair := range [][2]Temperament{{"unknown", Lazy}, {Lazy, "unknown"}, {"", ""}} {
		if _, err := Chart(pair[0], pair[1]); !errors.Is(err, ErrInvalidArithmetic) {
			t.Fatal("unknown temperament admitted")
		}
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
	matchSeed, err := strconv.ParseUint(fixture.RNG.MatchSeed, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	battle := determinism.BattleSeed(matchSeed)
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
	for _, vector := range fixture.RNG.Bounded {
		t.Run("bound_"+vector.Label+"_"+vector.Bound, func(t *testing.T) {
			bound, err := strconv.ParseUint(vector.Bound, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			random := determinism.Substream(battle, vector.Label)
			actual := random.Bound(bound)
			if formatUint(actual) != vector.Expected {
				t.Fatalf("bound result=%d want=%s", actual, vector.Expected)
			}
			threshold := (-bound) % bound
			replay := determinism.Substream(battle, vector.Label)
			draws := 0
			for {
				draws++
				if replay.Next() >= threshold {
					break
				}
			}
			if draws != vector.Draws {
				t.Fatalf("draws=%d want=%d", draws, vector.Draws)
			}
		})
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
	return strconv.FormatUint(value, 10)
}
