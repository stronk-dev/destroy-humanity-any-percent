package operations

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

type profileGameserverProcess struct {
	command *exec.Cmd
	ended   <-chan error
}

func startOperationsGameserver(t *testing.T, ctx context.Context) *profileGameserverProcess {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	encoded := base64.StdEncoding.EncodeToString(key)
	command := exec.CommandContext(ctx, "/test/gameserver")
	// Only disposable development credentials, never inherited operator secrets.
	// Full source content is mounted read-only; this is not a release image.
	command.Env = []string{
		"DATABASE_URL=" + os.Getenv("TEST_DATABASE_URL"),
		"CLOUD_CLICKER_REPOSITORY_ROOT=/workspace",
		"CLOUD_CLICKER_SERVER_ID=01986666-a001-4000-8000-000000000001",
		"CLOUD_CLICKER_JWT_KEY=" + encoded,
		"CLOUD_CLICKER_BOOTSTRAP_KEY_ID=operations-fixture",
		"CLOUD_CLICKER_BOOTSTRAP_KEY=" + encoded,
		"CLOUD_CLICKER_CURSOR_KEY=" + encoded,
		"LISTEN_ADDR=:8080",
	}
	// Startup diagnostics may contain private data. Surface only process exit,
	// actual readiness/metrics and bounded test-stage diagnostics.
	if err := command.Start(); err != nil {
		t.Fatal("actual gameserver command could not start")
	}
	ended := make(chan error, 1)
	go func() { ended <- command.Wait(); close(ended) }()
	t.Cleanup(func() { _ = command.Process.Kill(); <-ended })
	client := &http.Client{Timeout: 2 * time.Second}
	for ctx.Err() == nil {
		select {
		case err := <-ended:
			t.Fatalf("actual gameserver exited before readiness: %v", err)
		default:
		}
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/readyz", nil)
		response, err := client.Do(request)
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusNoContent {
				t.Log("actual gameserver command reached readiness on the private fixture network")
				return &profileGameserverProcess{command: command, ended: ended}
			}
		}
		select {
		case err := <-ended:
			t.Fatalf("actual gameserver exited before readiness: %v", err)
		case <-ctx.Done():
		case <-time.After(100 * time.Millisecond):
		}
	}
	t.Fatal("actual gameserver startup observation did not complete")
	return nil
}

func runGameserverProfile(t *testing.T, ctx context.Context, client *http.Client, received <-chan []byte, game *profileGameserverProcess) {
	t.Helper()
	const alert = "CloudClickerPublicEndpointOrReadinessUnavailable"
	const endpoint = "http://caddy:8081/readyz"
	if !waitForHTTP(ctx, client, endpoint) || !waitForQueryValue(ctx, client, `up{job="gameserver"}`, "1") ||
		!waitForQueryValue(ctx, client, `cloud_clicker_ready{job="gameserver"}`, "1") || !waitForAlertState(ctx, client, alert, "inactive") {
		t.Fatal("actual gameserver route/metrics did not establish a healthy baseline")
	}
	failedAt := time.Now().UTC()
	if err := game.command.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal("owned gameserver did not accept termination")
	}
	select {
	case err := <-game.ended:
		if err != nil {
			t.Fatalf("actual gameserver failed its graceful stop: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("actual gameserver did not complete its stop")
	}
	t.Log("actual gameserver command stopped; proxy and monitoring remain live")
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	response, err := client.Do(request)
	if err != nil {
		t.Fatal("Caddy itself became unavailable during the game-process outage")
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusBadGateway || !waitForQueryValue(ctx, client, `up{job="caddy"}`, "1") ||
		!waitForQueryValue(ctx, client, `up{job="gameserver"}`, "0") {
		t.Fatal("gameserver stop did not isolate the actual upstream failure")
	}
	activeAt, err := waitForAvailabilityPending(ctx, client, "gameserver", failedAt)
	if err != nil {
		t.Fatalf("actual game-process outage pending observation failed: %v", err)
	}
	remaining := time.Until(activeAt.Add(4*time.Minute + 45*time.Second))
	if remaining <= 0 {
		t.Fatal("five-minute lower-bound observation began too late")
	}
	requireNoOperationsAlert(t, ctx, client, `ALERTS{alertname="CloudClickerPublicEndpointOrReadinessUnavailable",job="gameserver",alertstate="firing"}`, remaining)
	if !waitForAlertState(ctx, client, alert, "firing") || time.Now().Before(activeAt.Add(5*time.Minute)) {
		t.Fatal("actual game-process outage did not honor the five-minute firing bound")
	}
	firingStart := waitForOperationsNotification(t, ctx, received, alert, "gameserver", "firing", failedAt, time.Time{})
	t.Log("actual stopped gameserver reached the receiver after the unchanged five-minute rule")
	clearedAt := time.Now().UTC()
	startOperationsGameserver(t, ctx)
	if !waitForHTTP(ctx, client, endpoint) || !waitForQueryValue(ctx, client, `up{job="gameserver"}`, "1") ||
		!waitForQueryValue(ctx, client, `cloud_clicker_ready{job="gameserver"}`, "1") || !waitForAlertState(ctx, client, alert, "inactive") {
		t.Fatal("actual restarted gameserver did not recover authoritative readiness")
	}
	waitForOperationsNotification(t, ctx, received, alert, "gameserver", "resolved", clearedAt, firingStart)
	t.Logf("same game-process outage alert resolved after executable restart; outage started %s", failedAt.Format(time.RFC3339))
}
