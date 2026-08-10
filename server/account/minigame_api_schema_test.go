package account

import (
	"bytes"
	"net/http"
	"testing"

	"cloud-clicker/server/publicapi"
)

const (
	testMinigameSessionID = "01986666-ca01-7000-8000-000000000010"
	testConstantsHash     = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	testPitchContentHash  = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func TestMinigameAPIRegistryValidatesExactWire(t *testing.T) {
	registry, err := newPrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	requests := map[string][]byte{
		"create_minigame_session":      []byte(`{"idempotency_key":"create-1"}`),
		"get_current_minigame_session": nil,
		"play_minigame_command":        []byte(`{"command":{"card_ids":["metric.activation#1"],"kind":"play_hand"},"command_id":"command-1","expected_revision":1}`),
		"resolve_minigame_session":     []byte(`{}`),
	}
	for operation, body := range requests {
		if err := registry.ValidateRequest(operation, body); err != nil {
			t.Fatalf("%s request: %v", operation, err)
		}
	}
	active := []byte(`{"constants_hash":"` + testConstantsHash + `","engine_ref":"pitch","engine_version":"1.0.0","minigame_id":"minigame.pitch","mode":"solo","revision":1,"session_id":"` + testMinigameSessionID + `","snapshot":{"deck_count":17,"funding_target":"1e3","hand":[],"hands_remaining":3,"phase":"playing","pitch_content_hash":"` + testPitchContentHash + `","pitch_schema_version":1,"revision":1,"round":1,"round_best_valuation":"0","run_currency":4,"shop_offers":[],"slotted_hacks":[]},"status":"active"}`)
	for _, operation := range []string{"create_minigame_session", "play_minigame_command"} {
		if err := registry.ValidateResponse(operation, http.StatusOK, active); err != nil {
			t.Fatalf("%s active response: %v", operation, err)
		}
	}
	current := append([]byte(`{"kind":"active","session":{"constants_hash":"`+testConstantsHash+`","engine_ref":"pitch","engine_version":"1.0.0","minigame_id":"minigame.pitch","mode":"solo","revision":1,"session_id":"`+testMinigameSessionID+`","status":"active"},"snapshot":`), activeSnapshot(active)...)
	current = append(current, '}')
	if err := registry.ValidateResponse("get_current_minigame_session", http.StatusOK, current); err != nil {
		t.Fatalf("current response: %v\n%s", err, current)
	}
	if err := registry.ValidateResponse("get_current_minigame_session", http.StatusOK, []byte(`{"kind":"none"}`)); err != nil {
		t.Fatal(err)
	}
	apiError := []byte(`{"category":"unauthorized","detail":"access_token"}`)
	if err := registry.ValidateResponse("create_minigame_session", http.StatusUnauthorized, apiError); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []struct {
		operation string
		body      []byte
	}{
		{"create_minigame_session", []byte(`{"idempotency_key":"create-1","founder_id":"attacker"}`)},
		{"get_current_minigame_session", []byte(`{}`)},
		{"play_minigame_command", []byte(`{"command":{"kind":"unknown"},"command_id":"command-1","expected_revision":1}`)},
	} {
		if err := registry.ValidateRequest(invalid.operation, invalid.body); err == nil {
			t.Fatalf("%s accepted invalid request %s", invalid.operation, invalid.body)
		}
	}
	broken := append([]byte(nil), active...)
	broken = append(broken[:len(broken)-1], []byte(`,"claim_token":"private"}`)...)
	if err := registry.ValidateResponse("create_minigame_session", http.StatusOK, broken); err == nil {
		t.Fatal("private claim token accepted by response schema")
	}
}

func TestMinigameEndpointPrivacyContractIsClosed(t *testing.T) {
	registry, err := newPrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	wantParameters := map[string][]string{
		"create_minigame_session":      {"minigame_id"},
		"get_current_minigame_session": nil,
		"play_minigame_command":        {"session_id"},
		"resolve_minigame_session":     {"session_id"},
	}
	valid := map[string][]byte{
		"create_minigame_session":      []byte(`{"idempotency_key":"create-1"}`),
		"get_current_minigame_session": nil,
		"play_minigame_command":        []byte(`{"command":{"kind":"end_shop"},"command_id":"command-1","expected_revision":1}`),
		"resolve_minigame_session":     []byte(`{}`),
	}
	for operationID, parameters := range wantParameters {
		operation, ok := registry.Operation(operationID)
		if !ok || operation.Public || operation.Surface != publicapi.SurfacePrivateV1 || operation.Auth != publicapi.AuthAccessToken {
			t.Fatalf("%s authority=%+v ok=%v", operationID, operation, ok)
		}
		if len(operation.Parameters) != len(parameters) {
			t.Fatalf("%s parameters=%+v", operationID, operation.Parameters)
		}
		for index, name := range parameters {
			if operation.Parameters[index].Name != name {
				t.Fatalf("%s parameter[%d]=%q", operationID, index, operation.Parameters[index].Name)
			}
		}
		if err := registry.ValidateRequest(operationID, valid[operationID]); err != nil {
			t.Fatalf("%s valid request: %v", operationID, err)
		}
		for _, privateField := range []string{"founder_id", "company_stream_id", "server_now_ms"} {
			body := append([]byte(nil), valid[operationID]...)
			if len(body) == 0 || string(body) == `{}` {
				body = []byte(`{"` + privateField + `":"attacker"}`)
			} else {
				body = append(body[:len(body)-1], []byte(`,"`+privateField+`":"attacker"}`)...)
			}
			if err := registry.ValidateRequest(operationID, body); err == nil {
				t.Fatalf("%s accepted private field %s", operationID, privateField)
			}
		}
	}
}

func TestSoulRecoveryAPIRegistryValidatesArchivedWire(t *testing.T) {
	registry, err := newPrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	requests := map[string][]byte{
		"cancel_soul_recovery":   []byte(`{"session_id":"` + testMinigameSessionID + `"}`),
		"progress_soul_recovery": []byte(`{"progress_token":"018f0000-0000-4000-8000-000000000202","session_id":"` + testMinigameSessionID + `"}`),
		"resolve_soul_recovery":  []byte(`{"session_id":"` + testMinigameSessionID + `"}`),
		"start_soul_recovery":    []byte(`{"activity_id":"touch_grass.fixture"}`),
	}
	for operation, body := range requests {
		if err := registry.ValidateRequest(operation, body); err != nil {
			t.Fatalf("%s request: %v", operation, err)
		}
	}
	responses := map[string][]byte{
		"start_soul_recovery":    []byte(`{"activity_id":"touch_grass.fixture","attended_progress_ms":0,"last_progress_server_ms":1,"progress_token":"018f0000-0000-4000-8000-000000000202","required_duration_attended_ms":300000,"session_id":"` + testMinigameSessionID + `","started_wall_ms":1}`),
		"progress_soul_recovery": []byte(`{"attended_progress_ms":300000,"eligible":true,"last_progress_server_ms":300001,"required_duration_attended_ms":300000,"session_id":"` + testMinigameSessionID + `"}`),
		"resolve_soul_recovery":  []byte(`{"action":"resolve","activity_id":"touch_grass.fixture","band_after":"whole","band_before":"dimming","company_revision":2,"founder_revision":2,"intent_id":"` + testMinigameSessionID + `","outcome":"applied","session_id":"` + testMinigameSessionID + `","soul_after":80,"soul_before":70}`),
		"cancel_soul_recovery":   []byte(`{"action":"cancel","activity_id":"touch_grass.fixture","band_after":"dimming","band_before":"dimming","cancelled_by":"watchdog","company_revision":2,"founder_revision":2,"intent_id":"` + testMinigameSessionID + `","outcome":"applied","session_id":"` + testMinigameSessionID + `","soul_after":70,"soul_before":70}`),
	}
	for operation, body := range responses {
		if err := registry.ValidateResponse(operation, http.StatusOK, body); err != nil {
			t.Fatalf("%s response: %v", operation, err)
		}
	}
	if err := registry.ValidateRequest("start_soul_recovery", []byte(`{"activity_id":"touch_grass.fixture","founder_id":"private"}`)); err == nil {
		t.Fatal("Founder authority accepted from recovery request")
	}
}

func activeSnapshot(response []byte) []byte {
	const marker = `"snapshot":`
	start := bytes.Index(response, []byte(marker))
	if start < 0 {
		return nil
	}
	start += len(marker)
	depth := 0
	for index := start; index < len(response); index++ {
		switch response[index] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return append([]byte(nil), response[start:index+1]...)
			}
		}
	}
	return nil
}
