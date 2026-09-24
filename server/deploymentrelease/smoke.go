package deploymentrelease

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coder/websocket"
)

type SmokeSession struct {
	connection  *websocket.Conn
	client      *http.Client
	origin      string
	accessToken string
	founderID   string
	sequence    uint32
}

type wireReply struct {
	ID        uint32          `json:"id"`
	Error     json.RawMessage `json:"error"`
	Connect   json.RawMessage `json:"connect"`
	Subscribe json.RawMessage `json:"subscribe"`
	Push      *struct {
		Channel     string `json:"channel"`
		Publication *struct {
			Data json.RawMessage `json:"data"`
		} `json:"pub"`
	} `json:"push"`
}

func OpenAuthenticatedSmoke(ctx context.Context, client *http.Client, publicOrigin, expectedConstantsHash string) (*SmokeSession, error) {
	parsed, err := url.Parse(publicOrigin)
	if err != nil || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" ||
		parsed.Scheme != "http" && parsed.Scheme != "https" || client == nil || !hashPattern.MatchString(expectedConstantsHash) {
		return nil, ErrInvalid
	}
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	body := fmt.Sprintf(`{"idempotency_key":%q}`, hex.EncodeToString(key))
	request, _ := http.NewRequestWithContext(ctx, http.MethodPost, publicOrigin+"/api/v1/bootstrap", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	var bootstrap struct {
		Session struct {
			AccessToken string `json:"access_token"`
		} `json:"session"`
		GameUISnapshot json.RawMessage `json:"game_ui_snapshot"`
	}
	if response.StatusCode != http.StatusCreated || decodeResponse(response.Body, &bootstrap) != nil || bootstrap.Session.AccessToken == "" {
		return nil, ErrInvalid
	}
	var snapshot struct {
		ConstantsHash string `json:"constants_hash"`
		Run           struct {
			FounderID string `json:"founder_id"`
		} `json:"run"`
	}
	if json.Unmarshal(bootstrap.GameUISnapshot, &snapshot) != nil || snapshot.ConstantsHash != expectedConstantsHash || snapshot.Run.FounderID == "" {
		return nil, ErrInvalid
	}
	state, _ := http.NewRequestWithContext(ctx, http.MethodGet, publicOrigin+"/api/v1/founder/state", nil)
	state.Header.Set("Authorization", "Bearer "+bootstrap.Session.AccessToken)
	stateResponse, err := client.Do(state)
	if err != nil {
		return nil, err
	}
	defer stateResponse.Body.Close()
	var stateBody struct {
		ConstantsHash string `json:"constants_hash"`
	}
	if stateResponse.StatusCode != http.StatusOK || decodeResponse(stateResponse.Body, &stateBody) != nil || stateBody.ConstantsHash != expectedConstantsHash {
		return nil, ErrInvalid
	}
	websocketURL := *parsed
	if websocketURL.Scheme == "https" {
		websocketURL.Scheme = "wss"
	} else {
		websocketURL.Scheme = "ws"
	}
	websocketURL.Path = "/connection/websocket"
	connection, _, err := websocket.Dial(ctx, websocketURL.String(), &websocket.DialOptions{HTTPClient: client, HTTPHeader: http.Header{"Origin": []string{publicOrigin}}})
	if err != nil {
		return nil, err
	}
	session := &SmokeSession{connection: connection, client: client, origin: publicOrigin, accessToken: bootstrap.Session.AccessToken, founderID: snapshot.Run.FounderID}
	if err := session.command(ctx, map[string]any{"connect": map[string]any{"token": session.accessToken}}); err != nil {
		_ = connection.CloseNow()
		return nil, err
	}
	if err := session.command(ctx, map[string]any{"subscribe": map[string]any{"channel": "world"}}); err != nil {
		_ = connection.CloseNow()
		return nil, err
	}
	return session, nil
}

func (session *SmokeSession) Close() error {
	if session == nil || session.connection == nil {
		return nil
	}
	return session.connection.Close(websocket.StatusNormalClosure, "smoke complete")
}

func (session *SmokeSession) IntentRefused(ctx context.Context) (bool, error) {
	request, _ := http.NewRequestWithContext(ctx, http.MethodPost, session.origin+"/api/v1/intents", bytes.NewBufferString(`{}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+session.accessToken)
	response, err := session.client.Do(request)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	return response.StatusCode == http.StatusServiceUnavailable && string(data) == "{\"category\":\"server_draining\",\"detail\":\"retry_same_intent_id\"}\n", nil
}

func (session *SmokeSession) WaitForDrain(ctx context.Context) (bool, bool, error) {
	courtesy := false
	for {
		_, data, err := session.connection.Read(ctx)
		if err != nil {
			status := websocket.CloseStatus(err)
			return courtesy, status == 4003, err
		}
		var reply wireReply
		if json.Unmarshal(data, &reply) != nil || reply.Push == nil || reply.Push.Publication == nil {
			continue
		}
		if isRestartCourtesy(reply.Push.Publication.Data) {
			courtesy = true
		}
	}
}

func isRestartCourtesy(data []byte) bool {
	var envelope struct {
		Kind    string `json:"kind"`
		Payload struct {
			Code string `json:"code"`
		} `json:"payload"`
	}
	return json.Unmarshal(data, &envelope) == nil && envelope.Kind == "system" && envelope.Payload.Code == "server_restarting"
}

func (session *SmokeSession) command(ctx context.Context, value map[string]any) error {
	session.sequence++
	value["id"] = session.sequence
	data, _ := json.Marshal(value)
	if err := session.connection.Write(ctx, websocket.MessageText, data); err != nil {
		return err
	}
	for {
		_, incoming, err := session.connection.Read(ctx)
		if err != nil {
			return err
		}
		var reply wireReply
		if json.Unmarshal(incoming, &reply) != nil || reply.ID != session.sequence {
			continue
		}
		if len(reply.Error) != 0 && string(reply.Error) != "null" || len(reply.Connect) == 0 && len(reply.Subscribe) == 0 {
			return ErrInvalid
		}
		return nil
	}
}

func decodeResponse(reader io.Reader, target any) error {
	decoder := json.NewDecoder(io.LimitReader(reader, 4<<20))
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return ErrInvalid
	}
	return nil
}

// waitHTTPState waits for the gameserver's own readiness answer: 204 when
// ready, 503 while draining. A proxy 502/504 or a transport error after the
// process has gone is not evidence that readiness was withdrawn in order.
func waitHTTPState(ctx context.Context, client *http.Client, endpoint string, ready bool) bool {
	interval, want := 100*time.Millisecond, http.StatusNoContent
	if !ready {
		interval, want = 25*time.Millisecond, http.StatusServiceUnavailable
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		response, err := client.Do(request)
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == want {
				return true
			}
		}
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
		}
	}
}
