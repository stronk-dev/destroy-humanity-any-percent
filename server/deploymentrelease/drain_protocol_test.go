package deploymentrelease

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"cloud-clicker/server/deploymentconfig"
	"github.com/coder/websocket"
)

// This drives the actual release helper through physical HTTP/WebSocket I/O.
// The peer serves a synthetic bootstrap/state/protocol, and Docker commands are
// recorded, not executed. It is not account auth, SQL, Caddy or process-exit proof.
func TestDockerRuntimeDrainCurrentObservesProtocolAndCommandBoundaries(t *testing.T) {
	for _, fault := range []string{"none", "unmarked readiness", "wrong readiness marker", "missing courtesy", "wrong close code",
		"socket left open", "stop failure", "inspection failure", "nonzero exit", "bootstrap identity", "state identity"} {
		t.Run(fault, func(t *testing.T) {
			// The parent cancellation makes withheld observations finite. Leave
			// the actual helper's default20s drain bound unchanged; this is not
			// a measurement of supported release latency or the process bound.
			ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
			defer cancel()
			bundle := dockerFixtureBundle(t)
			stopStarted, downObserved := make(chan struct{}), make(chan struct{})
			var downOnce sync.Once
			var traceLock sync.Mutex
			trace := []string{}
			observe := func(event string) {
				traceLock.Lock()
				trace = append(trace, event)
				traceLock.Unlock()
			}
			peerDone := make(chan error, 1)
			peerStarted := false
			var startedLock sync.Mutex
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				switch request.URL.Path {
				case "/api/v1/bootstrap":
					if request.Method != http.MethodPost {
						t.Error("bootstrap method changed")
					}
					var body struct {
						IdempotencyKey string `json:"idempotency_key"`
					}
					if json.NewDecoder(request.Body).Decode(&body) != nil || len(body.IdempotencyKey) != 64 {
						t.Error("bootstrap omitted its real generated idempotency key")
					}
					observe("bootstrap")
					hash := bundle.Manifest.ConstantsHash
					if fault == "bootstrap identity" {
						hash = "sha256:" + strings.Repeat("f", 64)
					}
					response.Header().Set("Content-Type", "application/json")
					response.WriteHeader(http.StatusCreated)
					_ = json.NewEncoder(response).Encode(map[string]any{"session": map[string]string{"access_token": "fixture-token"},
						"game_ui_snapshot": map[string]any{"constants_hash": hash, "run": map[string]string{"founder_id": "fixture-founder"}}})
				case "/api/v1/founder/state":
					if request.Method != http.MethodGet || request.Header.Get("Authorization") != "Bearer fixture-token" {
						t.Error("state read lost the bootstrap credential")
					}
					observe("state")
					hash := bundle.Manifest.ConstantsHash
					if fault == "state identity" {
						hash = "sha256:" + strings.Repeat("f", 64)
					}
					response.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(response).Encode(map[string]string{"constants_hash": hash})
				case "/readyz":
					select {
					case <-stopStarted:
						if fault != "unmarked readiness" {
							marker := deploymentconfig.DrainReadinessValue
							if fault == "wrong readiness marker" {
								marker = "wrong"
							}
							response.Header().Set(deploymentconfig.DrainReadinessHeader, marker)
						}
						response.WriteHeader(http.StatusServiceUnavailable)
						downOnce.Do(func() { observe("readiness"); close(downObserved) })
					default:
						response.WriteHeader(http.StatusNoContent)
					}
				case "/connection/websocket":
					startedLock.Lock()
					peerStarted = true
					startedLock.Unlock()
					peerDone <- serveDrainProtocolPeer(ctx, response, request, stopStarted, downObserved, fault, observe)
				default:
					t.Errorf("unexpected smoke endpoint %s", request.URL.Path)
					http.NotFound(response, request)
				}
			}))
			// Stop/wait for the upgraded peer before closing its HTTP listener.
			// Every failure path shares this cleanup; no WebSocket goroutine leaks.
			t.Cleanup(func() {
				cancel()
				startedLock.Lock()
				started := peerStarted
				startedLock.Unlock()
				if started {
					if err := <-peerDone; err != nil {
						t.Errorf("synthetic protocol peer failed: %v", err)
					}
				}
				server.Close()
			})
			runner := &commandFixture{output: func(call []string) ([]byte, error) {
				switch {
				case slices.Contains(call, "ps"):
					observe("container")
					return []byte("fixture-container\n"), nil
				case slices.Contains(call, "stop"):
					observe("stop")
					close(stopStarted)
					select {
					case <-downObserved:
					case <-ctx.Done():
						return nil, ctx.Err()
					}
					if fault == "stop failure" {
						return nil, errors.New("fixture stop failure")
					}
					return nil, nil
				case slices.Contains(call, "inspect"):
					observe("inspect")
					if fault == "inspection failure" {
						return nil, errors.New("fixture inspection failure")
					}
					if fault == "nonzero exit" {
						return []byte("137\n"), nil
					}
					return []byte("0\n"), nil
				default:
					return nil, errors.New("unexpected runtime command")
				}
			}}
			runtime := dockerFixtureRuntime(runner, t.TempDir(), "")
			runtime.DrainTimeout = 0 // actual production normalization default
			runtime.Client, runtime.PublicOrigin = server.Client(), server.URL
			evidence, err := runtime.DrainCurrent(ctx, bundle)
			traceLock.Lock()
			observations := slices.Clone(trace)
			traceLock.Unlock()
			if fault == "bootstrap identity" || fault == "state identity" {
				if !errors.Is(err, ErrInvalid) || evidence != (DrainEvidence{}) || len(runner.calls) != 0 {
					t.Fatalf("failed smoke identity reached runtime commands: evidence=%+v error=%v commands=%d", evidence, err, len(runner.calls))
				}
				return
			}
			wantReadiness := fault != "unmarked readiness" && fault != "wrong readiness marker"
			wantCourtesy := fault != "missing courtesy"
			wantClosed := fault != "wrong close code" && fault != "socket left open"
			wantCleanExit := fault != "stop failure" && fault != "inspection failure" && fault != "nonzero exit"
			if evidence.ReadinessDown != wantReadiness || evidence.CourtesyFrame != wantCourtesy || evidence.SocketsClosed != wantClosed ||
				evidence.IntentsRefused != wantCleanExit || evidence.AdmittedComplete != wantCleanExit || evidence.JobsFlushed != wantCleanExit || !evidence.WithinBound {
				t.Fatalf("helper evidence did not match its actual inputs: fault=%s evidence=%+v observations=%v", fault, evidence, observations)
			}
			if fault == "none" {
				if err != nil || !evidence.Valid() {
					t.Fatalf("complete protocol drain refused: evidence=%+v error=%v", evidence, err)
				}
			} else if !errors.Is(err, ErrInvalid) || evidence.Valid() {
				t.Fatalf("damaged protocol/command evidence accepted: fault=%s evidence=%+v error=%v", fault, evidence, err)
			}
			if len(runner.calls) != 3 || !slices.Contains(runner.calls[0], "ps") || !slices.Contains(runner.calls[1], "stop") ||
				!slices.Contains(runner.calls[1], "20") || !slices.Contains(runner.calls[2], "--format={{.State.ExitCode}}") ||
				runner.calls[2][len(runner.calls[2])-1] != "fixture-container" {
				t.Fatalf("helper did not bind stop and inspected exit to the observed container: %v", runner.calls)
			}
			wantOrder := []string{"bootstrap", "state", "connect", "subscribe", "container", "stop", "readiness", "inspect"}
			ordered := []string{}
			for _, event := range observations {
				if slices.Contains(wantOrder, event) {
					ordered = append(ordered, event)
				}
			}
			if !slices.Equal(ordered, wantOrder) {
				t.Errorf("helper's actual smoke/stop/inspection order changed: %v", observations)
			}
		})
	}
}

func serveDrainProtocolPeer(ctx context.Context, response http.ResponseWriter, request *http.Request, stopStarted, downObserved <-chan struct{},
	fault string, observe func(string)) error {
	connection, err := websocket.Accept(response, request, nil)
	if err != nil {
		return err
	}
	defer connection.CloseNow()
	for index, kind := range []string{"connect", "subscribe"} {
		_, data, err := connection.Read(ctx)
		if err != nil {
			return err
		}
		var command map[string]json.RawMessage
		if json.Unmarshal(data, &command) != nil || command["id"] == nil || command[kind] == nil || len(command) != 2 {
			return fmt.Errorf("unexpected %s protocol command", kind)
		}
		var commandID uint32
		if json.Unmarshal(command["id"], &commandID) != nil || commandID != uint32(index+1) {
			return errors.New("protocol commands lost sequential correlation IDs")
		}
		if kind == "connect" {
			var value struct {
				Token string `json:"token"`
			}
			if json.Unmarshal(command[kind], &value) != nil || value.Token != "fixture-token" {
				return errors.New("connect lost its bootstrap token")
			}
		} else {
			var value struct {
				Channel string `json:"channel"`
			}
			if json.Unmarshal(command[kind], &value) != nil || value.Channel != "world" {
				return errors.New("smoke did not subscribe to world")
			}
		}
		observe(kind)
		reply, err := json.Marshal(map[string]any{"id": command["id"], kind: map[string]any{}})
		if err != nil || connection.Write(ctx, websocket.MessageText, reply) != nil {
			return errors.New("protocol acknowledgement failed")
		}
	}
	for _, boundary := range []<-chan struct{}{stopStarted, downObserved} {
		select {
		case <-boundary:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if fault != "missing courtesy" {
		frame := []byte(`{"push":{"channel":"world","pub":{"data":{"kind":"system","payload":{"code":"server_restarting"}}}}}`)
		if err := connection.Write(ctx, websocket.MessageText, frame); err != nil {
			return err
		}
		observe("courtesy")
	}
	if fault == "socket left open" {
		<-ctx.Done()
		return nil
	}
	code := websocket.StatusCode(4003)
	if fault == "wrong close code" {
		code = websocket.StatusNormalClosure
	}
	if err := connection.Close(code, "fixture drain"); err != nil {
		return err
	}
	observe("close")
	return nil
}
