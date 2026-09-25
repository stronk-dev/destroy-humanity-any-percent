package leaderboard

import (
	"errors"
	"testing"
)

func TestResolveRankingKindRequiresAgreementAcrossStoredCatalogs(t *testing.T) {
	catalog := func(timer CategoryTimer) *CategoryCatalog {
		return &CategoryCatalog{Categories: []Category{{ID: "any_percent", Timer: timer}, {ID: "valuation", Timer: TimerNone}}}
	}
	for _, test := range []struct {
		name     string
		category string
		catalogs []*CategoryCatalog
		want     RankingKind
		err      error
	}{
		{"rta ranks by time", "any_percent", []*CategoryCatalog{catalog(TimerRTA)}, RankingTimeMS, nil},
		{"attended ranks by time", "any_percent", []*CategoryCatalog{catalog(TimerAttended), catalog(TimerRTA)}, RankingTimeMS, nil},
		{"untimed ranks by magnitude", "valuation", []*CategoryCatalog{catalog(TimerRTA)}, RankingMagnitude, nil},
		{"disagreement is an invariant", "any_percent", []*CategoryCatalog{catalog(TimerRTA), catalog(TimerNone)}, "", ErrInvalidEpoch},
		{"unknown timer is an invariant", "any_percent", []*CategoryCatalog{catalog(CategoryTimer("sundial"))}, "", ErrInvalidEpoch},
		{"undeclared is unknown", "low_percent", []*CategoryCatalog{catalog(TimerRTA)}, "", ErrUnknownPublicCategory},
		{"no stored catalog is unknown", "any_percent", nil, "", ErrUnknownPublicCategory},
	} {
		kind, err := resolveRankingKind(test.category, 8, test.catalogs)
		if kind != test.want || (test.err == nil) != (err == nil) || test.err != nil && !errors.Is(err, test.err) {
			t.Fatalf("%s: kind=%q err=%v", test.name, kind, err)
		}
	}
}
