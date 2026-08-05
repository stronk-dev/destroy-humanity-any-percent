package minigame

import (
	"errors"
	"testing"
)

var payoutDeclarations = PayoutDeclarations{
	ResourceIDs:  map[string]struct{}{"resource.compute": {}},
	ScoreFactIDs: map[string]struct{}{"score.total": {}},
	CopyKeys:     map[string]struct{}{"minigame.payout.cap": {}},
}

func TestPayoutPolicyLoadsExactDeclaredResourceRow(t *testing.T) {
	raw := `{"credited_resource_id":"resource.compute","sends_per_day":5,"per_send_cap":250,"conversion_ppm":125000,"payout_score_fact_id":"score.total","cap_reason_key":"minigame.payout.cap"}`
	policy, err := LoadPayoutPolicy([]byte(raw), payoutDeclarations)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := policy.MarshalJSON()
	if err != nil || string(encoded) != raw {
		t.Fatalf("encoded=%s err=%v", encoded, err)
	}
}

func TestPayoutPolicyRejectsUnknownResourceAndInventedGrammar(t *testing.T) {
	cases := map[string]string{
		"unknown resource": `{"credited_resource_id":"resource.unknown","sends_per_day":5,"per_send_cap":250,"conversion_ppm":125000,"payout_score_fact_id":"score.total","cap_reason_key":"minigame.payout.cap"}`,
		"old resource key": `{"resource_id":"resource.compute","sends_per_day":5,"per_send_cap":250,"conversion_ppm":125000}`,
		"extra window":     `{"credited_resource_id":"resource.compute","sends_per_day":5,"per_send_cap":250,"conversion_ppm":125000,"payout_score_fact_id":"score.total","cap_reason_key":"minigame.payout.cap","window_ms":86400000}`,
		"negative sends":   `{"credited_resource_id":"resource.compute","sends_per_day":-1,"per_send_cap":250,"conversion_ppm":125000,"payout_score_fact_id":"score.total","cap_reason_key":"minigame.payout.cap"}`,
		"conversion high":  `{"credited_resource_id":"resource.compute","sends_per_day":5,"per_send_cap":250,"conversion_ppm":1000001,"payout_score_fact_id":"score.total","cap_reason_key":"minigame.payout.cap"}`,
		"noninteger":       `{"credited_resource_id":"resource.compute","sends_per_day":5.0,"per_send_cap":250,"conversion_ppm":125000,"payout_score_fact_id":"score.total","cap_reason_key":"minigame.payout.cap"}`,
		"duplicate cap":    `{"credited_resource_id":"resource.compute","sends_per_day":5,"per_send_cap":250,"per_send_cap":500,"conversion_ppm":125000,"payout_score_fact_id":"score.total","cap_reason_key":"minigame.payout.cap"}`,
		"unknown score":    `{"credited_resource_id":"resource.compute","sends_per_day":5,"per_send_cap":250,"conversion_ppm":125000,"payout_score_fact_id":"score.wrong","cap_reason_key":"minigame.payout.cap"}`,
		"unknown copy":     `{"credited_resource_id":"resource.compute","sends_per_day":5,"per_send_cap":250,"conversion_ppm":125000,"payout_score_fact_id":"score.total","cap_reason_key":"minigame.payout.wrong"}`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadPayoutPolicy([]byte(raw), payoutDeclarations); !errors.Is(err, ErrInvalidPayoutPolicy) {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func TestSelectPayoutScoreUsesOnlyDeclaredCertifiedFact(t *testing.T) {
	policy, err := LoadPayoutPolicy([]byte(`{"credited_resource_id":"resource.compute","sends_per_day":5,"per_send_cap":250,"conversion_ppm":125000,"payout_score_fact_id":"score.total","cap_reason_key":"minigame.payout.cap"}`), payoutDeclarations)
	if err != nil {
		t.Fatal(err)
	}
	result := &Result{Outcome: "complete", ScoreFacts: []ScoreFact{{Kind: "score.bonus", Value: 9000}, {Kind: "score.total", Value: 42}}}
	if score, selectErr := SelectPayoutScore(result, policy); selectErr != nil || score != 42 {
		t.Fatalf("score=%d err=%v", score, selectErr)
	}
	for name, invalid := range map[string]*Result{
		"missing":  {Outcome: "complete", ScoreFacts: []ScoreFact{{Kind: "score.bonus", Value: 1}}},
		"negative": {Outcome: "complete", ScoreFacts: []ScoreFact{{Kind: "score.total", Value: -1}}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, selectErr := SelectPayoutScore(invalid, policy); !errors.Is(selectErr, ErrInvalidPayoutPolicy) {
				t.Fatalf("err=%v", selectErr)
			}
		})
	}
}

func TestConvertPayoutAppliesReductionThenCarriedPPM(t *testing.T) {
	result, err := ConvertPayout(11, 250000, 500000, 750000)
	if err != nil {
		t.Fatal(err)
	}
	if result.ReducedScore != 8 || result.ConvertedUnits != 4 || result.ConversionRemainderPPM != 750000 {
		t.Fatalf("result=%+v", result)
	}
	first, err := ConvertPayout(1, 0, 333333, 0)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ConvertPayout(2, 0, 333333, first.ConversionRemainderPPM)
	if err != nil {
		t.Fatal(err)
	}
	combined, err := ConvertPayout(3, 0, 333333, 0)
	if err != nil || first.ConvertedUnits+second.ConvertedUnits != combined.ConvertedUnits || second.ConversionRemainderPPM != combined.ConversionRemainderPPM {
		t.Fatalf("partition first=%+v second=%+v combined=%+v err=%v", first, second, combined, err)
	}
}

func TestConvertPayoutUsesExactIntermediateAtNumericBoundary(t *testing.T) {
	result, err := ConvertPayout(9_007_199_254_740_991, 0, 1_000_000, 999999)
	if err != nil {
		t.Fatal(err)
	}
	if result.ConvertedUnits != 9_007_199_254_740_991 || result.ConversionRemainderPPM != 999999 {
		t.Fatalf("result=%+v", result)
	}
	if _, err := ConvertPayout(1, 0, 1, 1_000_000); !errors.Is(err, ErrInvalidPayoutPolicy) {
		t.Fatalf("bad remainder err=%v", err)
	}
}
