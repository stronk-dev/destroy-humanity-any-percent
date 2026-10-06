package production

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"
)

const reputationPlanRequestID = "01986666-7f10-7000-8000-000000000001"
const reputationPlanOfferID = "01986666-7f10-7000-8000-000000000002"

// Literal canonical field ordering is independent of canonicalRequest. The
// node IDs here test syntax only, not catalog membership or affordability.
func reputationPlanRequestBytes(kind, rawPlan string) ([]byte, []byte) {
	prefix := fmt.Sprintf(`{"expected_founder_revision":3,"expected_revision":2,"kind":%q`, kind)
	if kind == IntentAcceptExitOffer || kind == IntentDeclineExitOffer {
		prefix += fmt.Sprintf(`,"offer_id":%q`, reputationPlanOfferID)
	}
	if rawPlan != "" {
		prefix += `,"reputation_plan":` + rawPlan
	}
	canonical := []byte(prefix + "}")
	wire := []byte(prefix + fmt.Sprintf(`,"intent_id":%q}`, reputationPlanRequestID))
	return wire, canonical
}

func reputationPlanRequireCanonical(t *testing.T, request IntentRequest, expected []byte) {
	t.Helper()
	digest := sha256.Sum256(expected)
	expectedHash := "sha256:" + hex.EncodeToString(digest[:])
	if !bytes.Equal(request.CanonicalPayload, expected) || request.RequestHash != expectedHash {
		t.Fatalf("canonical identity differs: bytes=%s want=%s hash=%s want=%s", request.CanonicalPayload, expected, request.RequestHash, expectedHash)
	}
}

func TestReputationPlanRequestShapes(t *testing.T) {
	ids := make([]string, 65)
	for i := range ids {
		ids[i] = fmt.Sprintf("reputation.node_%02d", i)
	}
	plan64, err := json.Marshal(ids[:64])
	if err != nil {
		t.Fatal(err)
	}
	plan65, err := json.Marshal(ids)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, raw string
		want      []string
		invalid   bool
	}{
		{name: "absent"},
		{name: "empty", raw: `[]`, want: []string{}},
		{name: "single", raw: `["reputation.unlock.p05"]`, want: []string{"reputation.unlock.p05"}},
		{name: "ordered", raw: `["reputation.starter.cash_small","reputation.starter.generated_beige_tower"]`, want: []string{"reputation.starter.cash_small", "reputation.starter.generated_beige_tower"}},
		{name: "reversed", raw: `["reputation.starter.generated_beige_tower","reputation.starter.cash_small"]`, want: []string{"reputation.starter.generated_beige_tower", "reputation.starter.cash_small"}},
		{name: "64", raw: string(plan64), want: ids[:64]},
		{name: "65", raw: string(plan65), invalid: true},
		{name: "null", raw: `null`, invalid: true},
		{name: "string", raw: `"reputation.unlock.p05"`, invalid: true},
		{name: "object", raw: `{}`, invalid: true},
		{name: "number", raw: `1`, invalid: true},
		{name: "boolean", raw: `true`, invalid: true},
		{name: "number-element", raw: `["reputation.unlock.p05",1]`, invalid: true},
		{name: "null-element", raw: `["reputation.unlock.p05",null]`, invalid: true},
		{name: "empty-id", raw: `[""]`, invalid: true},
		{name: "uppercase-id", raw: `["Reputation.unlock.p05"]`, invalid: true},
		{name: "space-id", raw: `["reputation.unlock.p05 "]`, invalid: true},
		{name: "hyphen-id", raw: `["reputation.unlock.p-05"]`, invalid: true},
		{name: "empty-segment", raw: `["reputation..p05"]`, invalid: true},
		{name: "duplicate", raw: `["reputation.unlock.p05","reputation.unlock.p05"]`, invalid: true},
	}
	for _, kind := range []string{IntentWindDown, IntentAcceptExitOffer} {
		for _, tc := range cases {
			t.Run(kind+"/"+tc.name, func(t *testing.T) {
				wire, canonical := reputationPlanRequestBytes(kind, tc.raw)
				request, err := ParseIntent(wire)
				if err != nil {
					t.Fatal(err)
				}
				wantDetail := ""
				if tc.invalid {
					wantDetail = "reputation_plan"
				}
				if request.InvalidDetail != wantDetail || !slices.Equal(request.ReputationPlan, tc.want) {
					t.Fatalf("parsed plan/detail=%v/%q want=%v/%q", request.ReputationPlan, request.InvalidDetail, tc.want, wantDetail)
				}
				reputationPlanRequireCanonical(t, request, canonical)
				if !tc.invalid && (request.Kind != kind || request.IntentID != reputationPlanRequestID || request.ExpectedRevision != 2 || request.ExpectedFounderRevision != 3 || (kind == IntentAcceptExitOffer && request.OfferID != reputationPlanOfferID)) {
					t.Fatal("plan parsing lost other request fields")
				}
			})
		}
	}
}

func TestReputationPlanRequestIdentity(t *testing.T) {
	for _, kind := range []string{IntentWindDown, IntentAcceptExitOffer} {
		t.Run(kind, func(t *testing.T) {
			parse := func(raw string) IntentRequest {
				t.Helper()
				wire, canonical := reputationPlanRequestBytes(kind, raw)
				request, err := ParseIntent(wire)
				if err != nil || request.InvalidDetail != "" {
					t.Fatalf("request invalid: %v/%s", err, request.InvalidDetail)
				}
				reputationPlanRequireCanonical(t, request, canonical)
				return request
			}
			absent, empty := parse(""), parse(`[]`)
			if !slices.Equal(absent.ReputationPlan, empty.ReputationPlan) || absent.RequestHash == empty.RequestHash {
				t.Fatal("absent/empty semantics or distinct identity differs")
			}
			ordered := parse(`["reputation.starter.cash_small","reputation.starter.generated_beige_tower"]`)
			reversed := parse(`["reputation.starter.generated_beige_tower","reputation.starter.cash_small"]`)
			if ordered.RequestHash == reversed.RequestHash || slices.Equal(ordered.ReputationPlan, reversed.ReputationPlan) {
				t.Fatal("purchase order lost from parsed plan/hash")
			}
			wire, _ := reputationPlanRequestBytes(kind, `["reputation.starter.cash_small","reputation.starter.generated_beige_tower"]`)
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(wire, &fields); err != nil {
				t.Fatal(err)
			}
			// Put intent_id first and reverse lexical field order, with spaces.
			keys := []string{"intent_id", "reputation_plan", "offer_id", "kind", "expected_revision", "expected_founder_revision"}
			parts := []string{}
			for _, key := range keys {
				if raw, ok := fields[key]; ok {
					parts = append(parts, fmt.Sprintf("%q : %s", key, raw))
				}
			}
			reformatted, err := ParseIntent([]byte("{ " + strings.Join(parts, " , ") + " }"))
			if err != nil || reformatted.InvalidDetail != "" || reformatted.RequestHash != ordered.RequestHash || !bytes.Equal(reformatted.CanonicalPayload, ordered.CanonicalPayload) {
				t.Fatal("key order/whitespace changed canonical identity")
			}
			for _, tc := range []struct{ name, from, to string }{
				{"intent-id", reputationPlanRequestID, "01986666-7f10-7000-8000-000000000003"},
				{"company-revision", `"expected_revision":2`, `"expected_revision":4`},
				{"founder-revision", `"expected_founder_revision":3`, `"expected_founder_revision":5`},
				{"offer-id", reputationPlanOfferID, "01986666-7f10-7000-8000-000000000004"},
			} {
				if tc.name == "offer-id" && kind != IntentAcceptExitOffer {
					continue
				}
				t.Run(tc.name, func(t *testing.T) {
					changedWire := bytes.Replace(wire, []byte(tc.from), []byte(tc.to), 1)
					if bytes.Equal(changedWire, wire) {
						t.Fatal("identity probe did not alter wire")
					}
					changed, err := ParseIntent(changedWire)
					if err != nil || changed.InvalidDetail != "" {
						t.Fatalf("changed request invalid: %v/%s", err, changed.InvalidDetail)
					}
					wantSame := tc.name == "intent-id"
					if (changed.RequestHash == ordered.RequestHash) != wantSame || (bytes.Equal(changed.CanonicalPayload, ordered.CanonicalPayload)) != wantSame {
						t.Fatal("request identity bound the wrong field")
					}
				})
			}
		})
	}
}

func TestReputationPlanRequestClosedCommands(t *testing.T) {
	for _, kind := range []string{IntentFileIPO, IntentDeclineExitOffer, "scripted_first"} {
		for _, plan := range []string{`[]`, `["reputation.unlock.p05"]`} {
			t.Run(kind+"/"+plan, func(t *testing.T) {
				wire, _ := reputationPlanRequestBytes(kind, plan)
				if kind == IntentDeclineExitOffer {
					wire = bytes.Replace(wire, []byte(`"expected_founder_revision":3,`), nil, 1)
				}
				request, err := ParseIntent(wire)
				want := kind + ".fields"
				if kind == "scripted_first" {
					want = kind
				}
				if err != nil || request.InvalidDetail != want || len(request.ReputationPlan) != 0 {
					t.Fatalf("forbidden plan accepted/misclassified: detail=%q plan=%v err=%v", request.InvalidDetail, request.ReputationPlan, err)
				}
			})
		}
	}
	for _, kind := range []string{IntentWindDown, IntentAcceptExitOffer} {
		for _, failure := range []string{"extra-key", "founder-revision"} {
			t.Run(kind+"/"+failure, func(t *testing.T) {
				wire, _ := reputationPlanRequestBytes(kind, `["reputation.unlock.p05"]`)
				want := "expected_founder_revision"
				if failure == "extra-key" {
					wire = append(wire[:len(wire)-1], []byte(`,"unexpected":true}`)...)
					want = kind + ".fields"
				} else {
					wire = bytes.Replace(wire, []byte(`"expected_founder_revision":3`), []byte(`"expected_founder_revision":0`), 1)
				}
				request, err := ParseIntent(wire)
				if err != nil || request.InvalidDetail != want {
					t.Fatalf("other field validation lost: detail=%q want=%q err=%v", request.InvalidDetail, want, err)
				}
			})
		}
	}
}
