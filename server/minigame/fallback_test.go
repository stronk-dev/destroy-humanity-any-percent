package minigame

import (
	"errors"
	"testing"
)

func TestFallbackPolicyLoadsEveryClosedArm(t *testing.T) {
	cases := []struct {
		raw  string
		want FallbackPolicy
	}{
		{`{"kind":"solo"}`, FallbackPolicy{Kind: FallbackSolo}},
		{`{"kind":"bot","bot_ref":{"policy_id":"combat.basic","version":"1.2.3"},"rate_reduction_ppm":250000}`,
			FallbackPolicy{Kind: FallbackBot, BotRef: &PolicyIdentity{ID: "combat.basic", Version: "1.2.3"}, RateReductionPPM: 250000}},
		{`{"kind":"npc_partner","npc_profile":{"profile_id":"guild.virtual_partner","version":"2.0.0"},"rate_reduction_ppm":1000000}`,
			FallbackPolicy{Kind: FallbackNPCPartner, NPCProfile: &PolicyIdentity{ID: "guild.virtual_partner", Version: "2.0.0"}, RateReductionPPM: 1000000}},
	}
	for _, test := range cases {
		got, err := LoadFallbackPolicy([]byte(test.raw))
		if err != nil {
			t.Fatalf("%s: %v", test.raw, err)
		}
		if got.Kind != test.want.Kind || got.RateReductionPPM != test.want.RateReductionPPM || !sameIdentity(got.BotRef, test.want.BotRef) || !sameIdentity(got.NPCProfile, test.want.NPCProfile) {
			t.Fatalf("got=%+v want=%+v", got, test.want)
		}
	}
}

func TestFallbackPolicyRejectsInventedOrIncompleteRows(t *testing.T) {
	cases := map[string]string{
		"unknown arm":        `{"kind":"human_peer"}`,
		"solo extra":         `{"kind":"solo","rate_reduction_ppm":0}`,
		"bot missing rate":   `{"kind":"bot","bot_ref":{"policy_id":"combat.basic","version":"1.0.0"}}`,
		"bot wrong identity": `{"kind":"bot","bot_ref":{"profile_id":"combat.basic","version":"1.0.0"},"rate_reduction_ppm":1}`,
		"bot bad version":    `{"kind":"bot","bot_ref":{"policy_id":"combat.basic","version":"latest"},"rate_reduction_ppm":1}`,
		"npc bad profile":    `{"kind":"npc_partner","npc_profile":{"profile_id":"Bad Profile","version":"1.0.0"},"rate_reduction_ppm":1}`,
		"negative rate":      `{"kind":"bot","bot_ref":{"policy_id":"combat.basic","version":"1.0.0"},"rate_reduction_ppm":-1}`,
		"rate overflow":      `{"kind":"bot","bot_ref":{"policy_id":"combat.basic","version":"1.0.0"},"rate_reduction_ppm":1000001}`,
		"duplicate kind":     `{"kind":"solo","kind":"bot"}`,
		"nested duplicate":   `{"kind":"bot","bot_ref":{"policy_id":"combat.basic","policy_id":"combat.other","version":"1.0.0"},"rate_reduction_ppm":1}`,
		"trailing value":     `{"kind":"solo"} {}`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadFallbackPolicy([]byte(raw)); !errors.Is(err, ErrInvalidFallbackPolicy) {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func sameIdentity(left, right *PolicyIdentity) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}
