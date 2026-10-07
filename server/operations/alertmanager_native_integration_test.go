package operations

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// This supplemental local lane runs a real Alertmanager, but does not replace
// the private Compose, Prometheus scrape or clean-host release populations.
type nativeAlertMessage struct {
	Alerts []struct {
		Status      string            `json:"status"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
	} `json:"alerts"`
}

func TestNativeAlertmanagerDeliveryIntegration(t *testing.T) {
	binary := os.Getenv("CLOUD_CLICKER_ALERTMANAGER_TEST_BINARY")
	if binary == "" {
		t.Skip("use make test-operations-alertmanager-native ALERTMANAGER_BINARY=<absolute path>")
	}
	if !filepath.IsAbs(binary) {
		t.Fatal("Alertmanager test binary must be an absolute path")
	}
	versionCtx, versionCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer versionCancel()
	version, err := exec.CommandContext(versionCtx, binary, "--version").CombinedOutput()
	if err != nil || !strings.Contains(string(version), "alertmanager, version 0.32.1 (") {
		t.Fatalf("requires pinned Alertmanager 0.32.1: %s (%v)", version, err)
	}
	t.Logf("real evaluator: %s", strings.TrimSpace(string(version)))

	for _, reject := range []bool{false, true} {
		name := "firing_and_resolution"
		if reject {
			name = "healthy_but_rejecting_receiver"
		}
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
			t.Cleanup(cancel)
			var lock sync.Mutex
			var messages []nativeAlertMessage
			receiver := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				if request.URL.Path == "/healthz" {
					response.WriteHeader(http.StatusNoContent)
					return
				}
				if request.URL.Path != "/alerts" || request.Method != http.MethodPost {
					response.WriteHeader(http.StatusNotFound)
					return
				}
				lock.Lock()
				defer lock.Unlock()
				message := len(messages)
				messages = append(messages, nativeAlertMessage{})
				if err := json.NewDecoder(http.MaxBytesReader(response, request.Body, 64<<10)).Decode(&messages[message]); err != nil {
					t.Errorf("real receiver payload: %v", err)
					response.WriteHeader(http.StatusBadRequest)
					return
				}
				if reject {
					response.WriteHeader(http.StatusInternalServerError)
					return
				}
				response.WriteHeader(http.StatusNoContent)
			}))
			t.Cleanup(receiver.Close)
			alertmanager := startNativeAlertmanager(t, ctx, binary, receiver.URL)
			client := &http.Client{Timeout: 2 * time.Second}
			proofCtx := ctx
			if reject {
				var proofCancel context.CancelFunc
				proofCtx, proofCancel = context.WithTimeout(ctx, 45*time.Second)
				defer proofCancel()
			}
			nonce, err := VerifyAlertDelivery(proofCtx, client, alertmanager, receiver.URL+"/healthz")
			if reject {
				if !errors.Is(err, ErrInvalid) || !strings.Contains(err.Error(), "failed notification") || proofCtx.Err() != nil {
					t.Fatalf("rejecting receiver must fail on actual notification failure, not a timeout: %v", err)
				}
				lock.Lock()
				defer lock.Unlock()
				if len(messages) == 0 {
					t.Fatal("rejection was not exercised at the real receiver")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			manifest := "sha256:" + strings.Repeat("a", 64) // diagnostic identity, not a release artifact
			observation, err := ObserveReleaseFloorAlertDelivery(ctx, client, alertmanager, receiver.URL+"/healthz", manifest, time.Now)
			if err != nil || ValidateAlertDeliveryObservation(observation) != nil {
				t.Fatalf("real firing/resolution delivery: %+v (%v)", observation, err)
			}
			lock.Lock()
			defer lock.Unlock()
			preflightDelivered := false
			seen := map[string]map[string]bool{}
			for _, message := range messages {
				for _, alert := range message.Alerts {
					if alert.Labels["alertname"] == "CloudClickerReceiverTest" && alert.Status == "firing" && alert.Annotations["proof_nonce"] == nonce {
						preflightDelivered = true
					}
					proofNonce := alert.Labels["proof_nonce"]
					if proofNonce == "" {
						continue
					}
					if seen[proofNonce] == nil {
						seen[proofNonce] = map[string]bool{}
					}
					seen[proofNonce][alert.Labels["alertname"]+":"+alert.Status] = true
				}
			}
			if !preflightDelivered || len(seen) != 1 {
				t.Fatalf("preflight nonce or unique seven-family nonce absent: preflight=%v nonces=%d", preflightDelivered, len(seen))
			}
			for _, states := range seen {
				for _, name := range ReleaseFloorAlertNames {
					if !states[name+":firing"] || !states[name+":resolved"] {
						t.Errorf("receiver lacks firing/resolution for %s", name)
					}
				}
			}
		})
	}
}

func startNativeAlertmanager(t *testing.T, ctx context.Context, binary, receiverURL string) string {
	t.Helper()
	configuration, err := os.ReadFile(filepath.Join("..", "..", "deployment", "operations", "alertmanager.test.yml"))
	if err != nil {
		t.Fatal(err)
	}
	const receiver = "http://operations-test:8090/alerts"
	if strings.Count(string(configuration), receiver) != 1 {
		t.Fatal("declared operations test receiver changed")
	}
	root := t.TempDir()
	configPath := filepath.Join(root, "alertmanager.yml")
	// Preserve the declared grouping, intervals and send_resolved policy.
	if err := os.WriteFile(configPath, []byte(strings.Replace(string(configuration), receiver, receiverURL+"/alerts", 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	_ = listener.Close()
	logPath := filepath.Join(root, "alertmanager.log")
	log, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}
	processCtx, stop := context.WithCancel(ctx)
	command := exec.CommandContext(processCtx, binary, "--config.file="+configPath, "--storage.path="+filepath.Join(root, "data"), "--web.listen-address="+address, "--cluster.listen-address=")
	command.Stdout, command.Stderr = log, log
	command.WaitDelay = 2 * time.Second
	if err := command.Start(); err != nil {
		stop()
		_ = log.Close()
		t.Fatal(err)
	}
	done := make(chan struct{})
	var waitErr error
	go func() { waitErr = command.Wait(); close(done) }()
	t.Cleanup(func() {
		stop()
		<-done
		_ = log.Close()
		if t.Failed() {
			data, _ := os.ReadFile(logPath)
			t.Logf("Alertmanager process log: %s", data)
		}
	})
	client := &http.Client{Timeout: 2 * time.Second}
	endpoint := "http://" + address
	for ctx.Err() == nil {
		select {
		case <-done:
			t.Fatalf("Alertmanager exited before readiness: %v", waitErr)
		default:
		}
		response, err := client.Get(endpoint + "/-/ready")
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return endpoint
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("Alertmanager readiness: %v", ctx.Err())
	return ""
}
