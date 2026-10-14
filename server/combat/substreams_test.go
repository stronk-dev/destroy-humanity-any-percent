package combat

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strconv"
	"testing"

	"cloud-clicker/server/determinism"
)

type streamSource interface {
	Next() uint64
	Bound(uint64) uint64
}
type streamFactory func(uint64, string) streamSource
type streamSeed struct {
	MatchSeed  string              `json:"match_seed"`
	BattleSeed string              `json:"battle_seed"`
	Streams    map[string][]string `json:"streams"`
}

func loadStreamSequences(t *testing.T) ([]string, int, []streamSeed) {
	t.Helper()
	data, err := os.ReadFile("../../testdata/combat/substream-sequences.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Version   int          `json:"version"`
		Labels    []string     `json:"labels"`
		DrawCount int          `json:"draw_count"`
		Seeds     []streamSeed `json:"seeds"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil || fixture.Version != 1 || fixture.DrawCount != 8 || len(fixture.Seeds) != 3 ||
		!slices.Equal(fixture.Labels, []string{"crit", "obedience", "bot_policy"}) {
		t.Fatalf("incomplete stream fixture: %v", err)
	}
	for index, row := range fixture.Seeds {
		if row.MatchSeed != []string{"0", "42", "18446744073709551615"}[index] || len(row.Streams) != len(fixture.Labels) {
			t.Fatal("wrong seed/label population")
		}
		for _, label := range fixture.Labels {
			if len(row.Streams[label]) != fixture.DrawCount {
				t.Fatal("truncated stream", label)
			}
		}
	}
	return fixture.Labels, fixture.DrawCount, fixture.Seeds
}

func observeStreamSequence(seed uint64, labels []string, draws int, pattern string, factory streamFactory) map[string][]string {
	streams := make(map[string]streamSource, len(labels))
	result := make(map[string][]string, len(labels))
	for _, label := range labels {
		streams[label] = factory(seed, label)
		result[label] = []string{}
	}
	var growth streamSource
	if pattern != "plain" {
		growth = factory(seed, "new_consumer")
	}
	consume := func(label string) {
		result[label] = append(result[label], formatUint(streams[label].Next()))
		if growth != nil {
			for _, bound := range []uint64{1, 7, 1<<63 + 1} {
				growth.Bound(bound)
			}
		}
	}
	order := slices.Clone(labels)
	if pattern == "reverse" {
		slices.Reverse(order)
	}
	if pattern == "chunked" {
		for _, label := range order {
			for range draws {
				consume(label)
			}
		}
	} else {
		for range draws {
			for _, label := range order {
				consume(label)
			}
		}
	}
	return result
}

func TestLabeledSubstreamSequenceIsolation(t *testing.T) {
	labels, draws, rows := loadStreamSequences(t)
	for _, row := range rows {
		match, err := strconv.ParseUint(row.MatchSeed, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		seed := determinism.BattleSeed(match)
		if formatUint(seed) != row.BattleSeed {
			t.Fatal("battle seed mismatch", row.MatchSeed)
		}
		for _, pattern := range []string{"plain", "growth", "reverse", "chunked"} {
			t.Run(row.MatchSeed+"/"+pattern, func(t *testing.T) {
				got := observeStreamSequence(seed, labels, draws, pattern, func(seed uint64, label string) streamSource {
					return determinism.Substream(seed, label)
				})
				if !reflect.DeepEqual(got, row.Streams) {
					t.Fatalf("stream draws changed: got %v want %v", got, row.Streams)
				}
			})
		}
	}
}

func TestSubstreamSequenceOracleRejectsSharedResetState(t *testing.T) {
	labels, draws, rows := loadStreamSequences(t)
	for _, row := range rows {
		seed, err := strconv.ParseUint(row.BattleSeed, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		shared := determinism.NewSplitMix64(0)
		badFactory := func(seed uint64, label string) streamSource {
			*shared = *determinism.Substream(seed, label)
			return shared
		}
		// This deliberate alias passes the earlier recreate-and-first-draw
		// pattern. It is not a claim that the production factory aliases state.
		for _, label := range labels {
			if formatUint(badFactory(seed, label).Next()) != row.Streams[label][0] {
				t.Fatal("invalid negative-control premise")
			}
		}
		before := badFactory(seed, "crit").Next()
		badFactory(seed, "new_consumer").Next()
		after := badFactory(seed, "crit").Next()
		if before != after || formatUint(before) != row.Streams["crit"][0] {
			t.Fatal("negative control does not pass the earlier isolation pattern")
		}
		got := observeStreamSequence(seed, labels, draws, "growth", badFactory)
		if reflect.DeepEqual(got, row.Streams) {
			t.Fatal("sequence oracle accepted shared resettable state")
		}
	}
}
