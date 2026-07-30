package guild

import (
	"errors"
	"testing"
)

func TestParseIntentClosedSurfaceAndStableHash(t *testing.T) {
	valid := `{"intent_id":"018f0000-0000-7000-8000-000000000001","kind":"create_guild","expected_revision":1,"name":"Small Systems","join_policy":"open"}`
	request, err := ParseIntent([]byte(valid))
	if err != nil || request.Kind != "create_guild" || request.Name != "Small Systems" || request.JoinPolicy != "open" || request.RequestHash == "" {
		t.Fatalf("request=%+v err=%v", request, err)
	}
	reordered := `{"join_policy":"open","name":"Small Systems","expected_revision":1,"kind":"create_guild","intent_id":"018f0000-0000-7000-8000-000000000001"}`
	second, err := ParseIntent([]byte(reordered))
	if err != nil || second.RequestHash != request.RequestHash {
		t.Fatalf("hash %q != %q err=%v", second.RequestHash, request.RequestHash, err)
	}
	invalid := []string{
		`{"intent_id":"018f0000-0000-7000-8000-000000000001","kind":"create_guild","expected_revision":1,"name":"Small Systems","join_policy":"open","extra":true}`,
		`{"intent_id":"018f0000-0000-7000-8000-000000000001","kind":"join_guild","expected_revision":1,"guild_id":"not-a-guild"}`,
		`{"intent_id":"018f0000-0000-7000-8000-000000000001","kind":"set_role","expected_revision":1,"account_id":"018f0000-0000-4000-8000-000000000001","role":"owner"}`,
		valid + `{}`,
	}
	for _, input := range invalid {
		if _, err := ParseIntent([]byte(input)); !errors.Is(err, ErrInvalidIntent) {
			t.Fatalf("input=%s err=%v", input, err)
		}
	}
}
