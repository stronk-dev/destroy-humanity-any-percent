package account

import (
	"bytes"
	"net/http"
	"strings"
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
	apiError := []byte("{\"category\":\"unauthorized\",\"detail\":\"access_token\"}\n")
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

func TestGameUISnapshotAPIRegistryPinsTheProjectionEnvelope(t *testing.T) {
	registry, err := newPrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	body := []byte(`{"constants_hash":"` + testConstantsHash + `","evaluated_through_ms":1800000000000,"facts":[{"fact_id":"bootstrap.needed","value":false}],"founder_revision":7,"generators":[{"generator_id":"generator.beige_tower","max_affordable":2,"next_cost":"1e1","next_cost_resource_id":"company.cash","owned":1,"provisioned":0,"rate_contribution":"1e0"}],"manual_action":{"action_id":"manual.click","bucket_cap_milli":50000,"refill_milli_per_ms":25,"refilled_at_ms":1800000000000,"tokens_milli":50000},"progress":[{"current":"5e-1","stage_id":"progress.tier","target":"1e0"}],"resources":[{"amount":"1e2","cap":{"amount":"1e1000","reason_key":"resource.company_cash.cap.phase0"},"rate_per_second":"1e0","resource_id":"company.cash"}],"revision":1,"run":{"category":"any_percent","exit_count":0,"founder_id":"01985555-1111-7111-8111-111111111111","run_seq":1,"run_started_at_ms":1799999000000,"tier":0},"schema_version":2,"server_now_ms":1800000000000,"upgrades":[]}`)
	if err := registry.ValidateRequest("get_game_ui_snapshot", nil); err != nil {
		t.Fatal(err)
	}
	if err := registry.ValidateResponse("get_game_ui_snapshot", http.StatusOK, body); err != nil {
		t.Fatalf("valid snapshot: %v", err)
	}
	legacy := bytes.ReplaceAll(body, []byte(`,"founder_revision":7`), nil)
	legacy = bytes.ReplaceAll(legacy, []byte(`"schema_version":2`), []byte(`"schema_version":1`))
	if err := registry.ValidateResponse("get_game_ui_snapshot", http.StatusOK, legacy); err == nil {
		t.Fatal("live snapshot endpoint accepted legacy schema v1")
	}
	unknown := append([]byte(nil), body...)
	unknown = append(unknown[:len(unknown)-1], []byte(`,"save_state":{}}`)...)
	if err := registry.ValidateResponse("get_game_ui_snapshot", http.StatusOK, unknown); err == nil {
		t.Fatal("raw save-state escape accepted")
	}
	operation, ok := registry.Operation("get_game_ui_snapshot")
	if !ok || operation.Public || operation.Auth != publicapi.AuthAccessToken || len(operation.Parameters) != 0 {
		t.Fatalf("snapshot authority=%+v ok=%v", operation, ok)
	}
}

func TestBootstrapAPIRegistryPinsCredentialSafeWire(t *testing.T) {
	registry, err := newPrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	request := []byte(`{"idempotency_key":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`)
	if err := registry.ValidateRequest("create_bootstrap", request); err != nil {
		t.Fatal(err)
	}
	operation, ok := registry.Operation("create_bootstrap")
	if !ok || operation.Public || operation.Auth != publicapi.AuthNone || operation.Method != http.MethodPost || operation.Path != "/api/v1/bootstrap" {
		t.Fatalf("bootstrap authority=%+v ok=%v", operation, ok)
	}
	snapshot := `{"constants_hash":"` + testConstantsHash + `","evaluated_through_ms":1800000000000,"facts":[],"generators":[],"manual_action":{"action_id":"manual.click","bucket_cap_milli":1,"refill_milli_per_ms":1,"refilled_at_ms":1800000000000,"tokens_milli":0},"progress":[],"resources":[],"revision":1,"run":{"category":"any_percent","exit_count":0,"founder_id":"01985555-1111-7111-8111-111111111111","run_seq":1,"run_started_at_ms":1800000000000,"tier":0},"schema_version":1,"server_now_ms":1800000000000,"upgrades":[]}`
	response := []byte(`{"account":{"account_id":"01985555-1111-7111-8111-111111111110","created_at":"2026-08-10T12:00:00.000Z","recovery_code":"recovery"},"session":{"access_token":"access","refresh_token":"refresh"},"game_ui_snapshot":` + snapshot + `}`)
	if err := registry.ValidateResponse("create_bootstrap", http.StatusCreated, response); err != nil {
		t.Fatalf("valid bootstrap response: %v", err)
	}
	currentSnapshot := strings.Replace(snapshot, `"facts":[]`, `"facts":[],"founder_revision":1`, 1)
	currentSnapshot = strings.Replace(currentSnapshot, `"schema_version":1`, `"schema_version":2`, 1)
	currentResponse := []byte(`{"account":{"account_id":"01985555-1111-7111-8111-111111111110","created_at":"2026-08-10T12:00:00.000Z","recovery_code":"recovery"},"session":{"access_token":"access","refresh_token":"refresh"},"game_ui_snapshot":` + currentSnapshot + `}`)
	if err := registry.ValidateResponse("create_bootstrap", http.StatusCreated, currentResponse); err != nil {
		t.Fatalf("current bootstrap response: %v", err)
	}
	for _, invalid := range [][]byte{
		[]byte(`{"idempotency_key":"a","account_id":"private"}`),
		append(append([]byte(nil), response[:len(response)-1]...), []byte(`,"receipt_ciphertext":"private"}`)...),
	} {
		if bytes.Contains(invalid, []byte("account_id\":\"private")) {
			if err := registry.ValidateRequest("create_bootstrap", invalid); err == nil {
				t.Fatalf("accepted private bootstrap request %s", invalid)
			}
		} else if err := registry.ValidateResponse("create_bootstrap", http.StatusCreated, invalid); err == nil {
			t.Fatalf("accepted private bootstrap response %s", invalid)
		}
	}
}

func TestMinigameAndRecoveryEndpointPrivacyContractIsClosed(t *testing.T) {
	registry, err := newPrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	wantParameters := map[string][]string{
		"cancel_soul_recovery":         nil,
		"create_minigame_session":      {"minigame_id"},
		"get_current_minigame_session": nil,
		"play_minigame_command":        {"session_id"},
		"progress_soul_recovery":       nil,
		"resolve_minigame_session":     {"session_id"},
		"resolve_soul_recovery":        nil,
		"start_soul_recovery":          nil,
	}
	validRequests := map[string][]byte{
		"cancel_soul_recovery":         []byte(`{"session_id":"` + testMinigameSessionID + `"}`),
		"create_minigame_session":      []byte(`{"idempotency_key":"create-1"}`),
		"get_current_minigame_session": nil,
		"play_minigame_command":        []byte(`{"command":{"kind":"end_shop"},"command_id":"command-1","expected_revision":1}`),
		"progress_soul_recovery":       []byte(`{"progress_token":"018f0000-0000-4000-8000-000000000202","session_id":"` + testMinigameSessionID + `"}`),
		"resolve_minigame_session":     []byte(`{}`),
		"resolve_soul_recovery":        []byte(`{"session_id":"` + testMinigameSessionID + `"}`),
		"start_soul_recovery":          []byte(`{"activity_id":"touch_grass.fixture"}`),
	}
	active := []byte(`{"constants_hash":"` + testConstantsHash + `","engine_ref":"pitch","engine_version":"1.0.0","minigame_id":"minigame.pitch","mode":"solo","revision":1,"session_id":"` + testMinigameSessionID + `","snapshot":{"deck_count":17,"funding_target":"1e3","hand":[],"hands_remaining":3,"phase":"playing","pitch_content_hash":"` + testPitchContentHash + `","pitch_schema_version":1,"revision":1,"round":1,"round_best_valuation":"0","run_currency":4,"shop_offers":[],"slotted_hacks":[]},"status":"active"}`)
	terminal := []byte(`{"constants_hash":"` + testConstantsHash + `","engine_ref":"pitch","engine_version":"1.0.0","minigame_id":"minigame.pitch","mode":"solo","resolution_receipt":{"cap_reason_key":"resource.company_cash.cap.phase0","certified_result_hash":"` + testConstantsHash + `","company_revision":2,"configured_cap_forfeit_units":0,"credited_delta":"1e0","credited_resource_id":"company.cash","founder_revision":2,"intent_id":"` + testMinigameSessionID + `","minigame_id":"minigame.pitch","outcome":"applied","quality_change":{"new":{"decay_remainder_ppm":0,"grade_ppm":1000000,"last_founder_attended_ms":1},"old":{"decay_remainder_ppm":0,"grade_ppm":1000000,"last_founder_attended_ms":1}},"rating_change":{"games_after":1,"games_before":0,"new_elo":1000,"old_elo":1000,"rated":false,"season_member":""},"session_id":"` + testMinigameSessionID + `"},"revision":2,"session_id":"` + testMinigameSessionID + `","snapshot":{"deck_count":17,"funding_target":"1e3","hand":[],"hands_remaining":0,"phase":"terminal","pitch_content_hash":"` + testPitchContentHash + `","pitch_schema_version":1,"revision":2,"round":1,"round_best_valuation":"1e3","run_currency":4,"shop_offers":[],"slotted_hacks":[]},"status":"resolved"}`)
	validResponses := map[string][]byte{
		"cancel_soul_recovery":         []byte(`{"action":"cancel","activity_id":"touch_grass.fixture","band_after":"dimming","band_before":"dimming","cancelled_by":"watchdog","company_revision":2,"founder_revision":2,"intent_id":"` + testMinigameSessionID + `","outcome":"applied","session_id":"` + testMinigameSessionID + `","soul_after":70,"soul_before":70}`),
		"create_minigame_session":      active,
		"get_current_minigame_session": []byte(`{"kind":"none"}`),
		"play_minigame_command":        active,
		"progress_soul_recovery":       []byte(`{"attended_progress_ms":300000,"eligible":true,"last_progress_server_ms":300001,"required_duration_attended_ms":300000,"session_id":"` + testMinigameSessionID + `"}`),
		"resolve_minigame_session":     terminal,
		"resolve_soul_recovery":        []byte(`{"action":"resolve","activity_id":"touch_grass.fixture","band_after":"whole","band_before":"dimming","company_revision":2,"founder_revision":2,"intent_id":"` + testMinigameSessionID + `","outcome":"applied","session_id":"` + testMinigameSessionID + `","soul_after":80,"soul_before":70}`),
		"start_soul_recovery":          []byte(`{"activity_id":"touch_grass.fixture","attended_progress_ms":0,"last_progress_server_ms":1,"progress_token":"018f0000-0000-4000-8000-000000000202","required_duration_attended_ms":300000,"session_id":"` + testMinigameSessionID + `","started_wall_ms":1}`),
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
		if err := registry.ValidateRequest(operationID, validRequests[operationID]); err != nil {
			t.Fatalf("%s valid request: %v", operationID, err)
		}
		for privateField, privateValue := range map[string]string{
			"company_stream_id": `"01986666-ca01-7000-8000-000000000011"`,
			"founder_id":        `"01985555-1111-7111-8111-111111111111"`,
			"server_now_ms":     `1800000000000`,
		} {
			body := append([]byte(nil), validRequests[operationID]...)
			if len(body) == 0 || string(body) == `{}` {
				body = []byte(`{"` + privateField + `":` + privateValue + `}`)
			} else {
				body = append(body[:len(body)-1], []byte(`,"`+privateField+`":`+privateValue+`}`)...)
			}
			if err := registry.ValidateRequest(operationID, body); err == nil {
				t.Fatalf("%s accepted private field %s", operationID, privateField)
			}
		}
		if err := registry.ValidateResponse(operationID, http.StatusOK, validResponses[operationID]); err != nil {
			t.Fatalf("%s valid response: %v", operationID, err)
		}
		body := append([]byte(nil), validResponses[operationID]...)
		body = append(body[:len(body)-1], []byte(`,"founder_id":"01985555-1111-7111-8111-111111111111"}`)...)
		if err := registry.ValidateResponse(operationID, http.StatusOK, body); err == nil {
			t.Fatalf("%s response exposed hidden Founder authority", operationID)
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
