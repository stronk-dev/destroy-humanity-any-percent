package pitch

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"cloud-clicker/server/minigame"
)

func TestTenantRejectsMalformedJSONFields(t *testing.T) {
	data, _ := loadFixture(t)
	tenant := NewTenant()
	identity := minigame.CreateInput{Mode: minigame.ModeSolo, Seed: 2,
		ScalingInputs: map[string]int64{ScalingDestination: 1}, Content: data,
		ContentHash: ContentHash(data), ContentSchemaVersion: SchemaVersion}
	playing, err := tenant.Create(identity)
	if err != nil {
		t.Fatal(err)
	}
	apply := func(snapshot, command json.RawMessage, revision int64) (minigame.ApplyOutput, error) {
		return tenant.Apply(minigame.ApplyInput{Mode: identity.Mode, Seed: identity.Seed,
			ScalingInputs: identity.ScalingInputs, Content: data, ContentHash: identity.ContentHash,
			ContentSchemaVersion: SchemaVersion, Revision: revision, Snapshot: snapshot, Command: command})
	}
	entered, err := apply(playing, json.RawMessage(`{"kind":"play_hand","card_ids":["demo_day#1","patch_release#1","testimonial#2","uptime_nines#2"]}`), 1)
	var shop Snapshot
	if err != nil || json.Unmarshal(entered.Snapshot, &shop) != nil || shop.Phase != "shop" || len(shop.ShopOffers) == 0 {
		t.Fatalf("canonical shop setup: err=%v snapshot=%s", err, entered.Snapshot)
	}
	for _, snapshot := range []json.RawMessage{playing, entered.Snapshot} {
		if err := tenant.ValidateSnapshot(snapshot); err != nil {
			t.Fatalf("canonical snapshot rejected: %v", err)
		}
	}
	variants := []string{"missing", "null", "wrong_case", "case_collision"}
	for _, key := range []string{"phase", "round", "hands_remaining", "deck_count", "hand", "slotted_hacks",
		"run_currency", "shop_offers", "funding_target", "round_best_valuation", "revision", "pitch_content_hash", "pitch_schema_version"} {
		for _, variant := range variants {
			t.Run("snapshot/"+key+"/"+variant, func(t *testing.T) {
				bad := mutatePitchJSONField(t, playing, key, variant)
				if err := tenant.ValidateSnapshot(bad); err == nil {
					t.Fatalf("malformed snapshot admitted: %s", bad)
				}
				if _, err := apply(bad, json.RawMessage(`{"kind":"play_hand","card_ids":[]}`), 1); err == nil {
					t.Fatal("malformed snapshot reached tenant execution")
				}
			})
		}
	}
	for _, key := range []string{"hack_id", "offer_id", "price"} {
		for _, variant := range variants {
			t.Run("offer/"+key+"/"+variant, func(t *testing.T) {
				var root map[string]json.RawMessage
				var offers []json.RawMessage
				if json.Unmarshal(entered.Snapshot, &root) != nil || json.Unmarshal(root["shop_offers"], &offers) != nil {
					t.Fatal("invalid canonical shop")
				}
				offers[0] = mutatePitchJSONField(t, offers[0], key, variant)
				root["shop_offers"] = mustPitchJSON(t, offers)
				bad, err := json.Marshal(root)
				if err != nil {
					t.Fatal(err)
				}
				if err := tenant.ValidateSnapshot(bad); err == nil {
					t.Fatalf("malformed offer admitted: %s", bad)
				}
				if _, err := apply(bad, json.RawMessage(`{"kind":"end_shop"}`), 2); err == nil {
					t.Fatal("malformed offer reached tenant execution")
				}
			})
		}
	}
	for _, arm := range []struct {
		name     string
		command  json.RawMessage
		keys     []string
		snapshot json.RawMessage
		revision int64
	}{
		{"play_hand", json.RawMessage(`{"kind":"play_hand","card_ids":[]}`), []string{"kind", "card_ids"}, playing, 1},
		{"buy_hack", mustPitchJSON(t, map[string]string{"kind": "buy_hack", "offer_id": shop.ShopOffers[0].OfferID}), []string{"kind", "offer_id"}, entered.Snapshot, 2},
		{"end_shop", json.RawMessage(`{"kind":"end_shop"}`), []string{"kind"}, entered.Snapshot, 2},
	} {
		if err := tenant.ValidateCommand(arm.command); err != nil {
			t.Fatal(err)
		}
		if _, err := apply(arm.snapshot, arm.command, arm.revision); err != nil {
			t.Fatalf("canonical %s rejected: %v", arm.name, err)
		}
		for _, key := range arm.keys {
			for _, variant := range variants {
				t.Run("command/"+arm.name+"/"+key+"/"+variant, func(t *testing.T) {
					bad := mutatePitchJSONField(t, arm.command, key, variant)
					if err := tenant.ValidateCommand(bad); err == nil {
						t.Fatalf("malformed command admitted: %s", bad)
					}
					if _, err := apply(arm.snapshot, bad, arm.revision); err == nil {
						t.Fatal("malformed command executed")
					}
				})
			}
		}
	}
	for _, digest := range []string{strings.Repeat("A", 64), strings.Repeat("z", 64), strings.Repeat("/", 64)} {
		t.Run("hash/"+digest[:1], func(t *testing.T) {
			bad := bytes.Replace(playing, []byte(identity.ContentHash), []byte("sha256:"+digest), 1)
			if err := tenant.ValidateSnapshot(bad); err == nil {
				t.Fatal("noncanonical content hash admitted")
			}
			if _, err := apply(bad, json.RawMessage(`{"kind":"play_hand","card_ids":[]}`), 1); err == nil {
				t.Fatal("noncanonical content hash reached tenant execution")
			}
		})
	}
	// Equivalent decoded keys and explicitly present zero are not missing/null.
	escaped := bytes.Replace(playing, []byte(`"phase"`), []byte(`"\u0070hase"`), 1)
	if err := tenant.ValidateSnapshot(escaped); err != nil {
		t.Fatalf("equivalent escaped key rejected: %v", err)
	}
	var zero map[string]json.RawMessage
	if json.Unmarshal(playing, &zero) != nil {
		t.Fatal("invalid genesis")
	}
	zero["run_currency"] = json.RawMessage("0")
	if _, err := apply(mustPitchJSON(t, zero), json.RawMessage(`{"kind":"play_hand","card_ids":[]}`), 1); err != nil {
		t.Fatalf("explicit zero rejected: %v", err)
	}
}

func mutatePitchJSONField(t *testing.T, source []byte, key, variant string) json.RawMessage {
	t.Helper()
	var fields map[string]json.RawMessage
	if json.Unmarshal(source, &fields) != nil || fields[key] == nil {
		t.Fatalf("mutation has no canonical %s", key)
	}
	switch variant {
	case "missing":
		delete(fields, key)
	case "null":
		fields[key] = json.RawMessage("null")
	case "wrong_case", "case_collision":
		fields[strings.ToUpper(key)] = fields[key]
		if variant == "wrong_case" {
			delete(fields, key)
		}
	default:
		t.Fatalf("unknown mutation %s", variant)
	}
	return mustPitchJSON(t, fields)
}

func mustPitchJSON(t *testing.T, source any) json.RawMessage {
	t.Helper()
	encoded, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
