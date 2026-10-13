package combat

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestIdentityAxesAndChartVectors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/combat/identity-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Version  int       `json:"version"`
		Runtimes []Runtime `json:"runtimes"`
		Invalid  []struct {
			Name        string      `json:"name"`
			Temperament Temperament `json:"temperament"`
			Runtime     Runtime     `json:"runtime"`
		} `json:"invalid"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	wantRuntimes := []Runtime{RuntimeContainer, RuntimeVM, RuntimeBareMetal, RuntimeServerless, RuntimeEdge, RuntimeMainframe}
	if fixture.Version != 1 || !reflect.DeepEqual(fixture.Runtimes, wantRuntimes) || len(fixture.Invalid) != 12 {
		t.Fatal("incomplete shared identity vectors")
	}
	chart := loadArithmeticFixture(t).Chart
	type vectorIdentity struct {
		value      Identity
		chartIndex int
	}
	identities := make([]vectorIdentity, 0, 36)
	for chartIndex, temperament := range chart.Temperaments {
		for _, runtime := range fixture.Runtimes {
			value, err := NewIdentity(temperament, runtime)
			if err != nil || value != (Identity{Temperament: temperament, Runtime: runtime}) {
				t.Fatalf("identity=%+v, err=%v", value, err)
			}
			identities = append(identities, vectorIdentity{value, chartIndex})
		}
	}
	if len(identities) != 36 {
		t.Fatal("incomplete identity population")
	}
	for _, attacker := range identities {
		var counts [3]int
		for _, defender := range identities {
			actual, err := ChartIdentities(attacker.value, defender.value)
			want := chart.Rows[attacker.chartIndex][defender.chartIndex]
			if err != nil || actual != want {
				t.Fatalf("%+v -> %+v = %d, want %d, err=%v", attacker, defender, actual, want, err)
			}
			counts[int(actual)+1]++
		}
		if counts != ([3]int{12, 12, 12}) {
			t.Fatalf("runtime changed matchup population: %v", counts)
		}
	}
	valid := identities[0].value
	for _, vector := range fixture.Invalid {
		t.Run(vector.Name, func(t *testing.T) {
			value, err := NewIdentity(vector.Temperament, vector.Runtime)
			if !errors.Is(err, ErrInvalidIdentity) || value != (Identity{}) {
				t.Fatalf("invalid identity returned %+v, err=%v", value, err)
			}
			for _, pair := range [][2]Identity{{{vector.Temperament, vector.Runtime}, valid}, {valid, {vector.Temperament, vector.Runtime}}} {
				if result, err := ChartIdentities(pair[0], pair[1]); !errors.Is(err, ErrInvalidIdentity) || result != Neutral {
					t.Fatalf("invalid caller-built identity admitted: %d, err=%v", result, err)
				}
			}
		})
	}
}
