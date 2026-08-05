package production

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

type founderAttendanceFixture struct {
	Cases []struct {
		Name                    string                  `json:"name"`
		FounderAgeMS            int64                   `json:"founder_age_ms"`
		ActualFounderRevision   int64                   `json:"actual_founder_revision"`
		ExpectedFounderRevision int64                   `json:"expected_founder_revision"`
		Sample                  FounderAttendanceSample `json:"sample"`
		Valid                   bool                    `json:"valid"`
	} `json:"cases"`
}

func TestFounderAttendanceSharedVectors(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/founder-attendance-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture founderAttendanceFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	_, bundle := foundationTestBundles(t)
	for _, testCase := range fixture.Cases {
		t.Run(testCase.Name, func(t *testing.T) {
			founder := foundationScopeState(t, bundle.Economy, economy.ScopeFounder)
			founder.AgeMS = testCase.FounderAgeMS
			err := ValidateFounderAttendanceSample(founder, testCase.ActualFounderRevision, testCase.ExpectedFounderRevision, testCase.Sample)
			if (err == nil) != testCase.Valid {
				t.Fatalf("valid=%v err=%v sample=%+v", testCase.Valid, err, testCase.Sample)
			}
		})
	}
}

func TestResolveFounderAttendanceClassifiesUnresolvedGap(t *testing.T) {
	_, bundle := foundationTestBundles(t)
	start := time.Date(2026, 8, 5, 8, 0, 0, 0, time.UTC)
	founder := foundationScopeState(t, bundle.Economy, economy.ScopeFounder)
	founder.AgeMS = 700
	founderRevision := save.Revision{StreamID: "01986666-9100-7000-8000-000000000001", OwnerID: "01986666-9200-7000-8000-000000000001", Number: 3, ConstantsHash: bundle.ConstantsHash}
	companyRevision := save.Revision{StreamID: "01986666-9300-7000-8000-000000000001", OwnerID: founderRevision.OwnerID, Number: 5, ConstantsHash: bundle.ConstantsHash}

	t.Run("sub-ceiling reconnect remains attended", func(t *testing.T) {
		company := replayFixtureState(t, bundle.Economy, start)
		company.EvaluatedThrough = start.Add(time.Second)
		company.ManualTokenRefilledAt = company.EvaluatedThrough
		now := company.EvaluatedThrough.Add(time.Duration(bundle.Prestige.CatchupCeilingMS-1) * time.Millisecond)
		sample, err := ResolveFounderAttendanceSample(founder, founderRevision, company, companyRevision, bundle, now)
		if err != nil || sample.CurrentRunPartialAttendedMS != now.Sub(start).Milliseconds() ||
			sample.EffectiveFounderAttendedMS != founder.AgeMS+sample.CurrentRunPartialAttendedMS {
			t.Fatalf("sample=%+v err=%v", sample, err)
		}
	})

	t.Run("twenty-five-hour dormant gap is offline", func(t *testing.T) {
		company := replayFixtureState(t, bundle.Economy, start)
		company.EvaluatedThrough = start.Add(time.Second)
		company.ManualTokenRefilledAt = company.EvaluatedThrough
		sample, err := ResolveFounderAttendanceSample(founder, founderRevision, company, companyRevision, bundle, company.EvaluatedThrough.Add(25*time.Hour))
		if err != nil || sample.CurrentRunPartialAttendedMS != 1_000 || sample.EffectiveFounderAttendedMS != 1_700 {
			t.Fatalf("sample=%+v err=%v", sample, err)
		}
		if len(company.OfflineSpans) != 0 {
			t.Fatalf("resolver persisted clone-only offline span: %+v", company.OfflineSpans)
		}
	})

	t.Run("pinned bundle required", func(t *testing.T) {
		company := replayFixtureState(t, bundle.Economy, start)
		wrong := bundle
		wrong.ConstantsHash = "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
		if _, err := ResolveFounderAttendanceSample(founder, founderRevision, company, companyRevision, wrong, start); !errors.Is(err, ErrFounderAttendanceContext) {
			t.Fatalf("unresolved pinned bundle err=%v", err)
		}
	})
}
