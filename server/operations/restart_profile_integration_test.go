package operations

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"
)

func runRestartProfile(t *testing.T, ctx context.Context, client *http.Client, received <-chan []byte) {
	t.Helper()
	// This is the native production command's atomic file, already at three
	// actual Docker restarts before any monitoring service was started.
	before, err := os.Stat("/textfile/host.prom")
	if err != nil {
		t.Fatal("restart proof needs its actual host-observer textfile")
	}
	count := "cloud_clicker_gameserver_restart_count"
	if !waitForQueryValue(ctx, client, count, "3") ||
		!waitForQueryValue(ctx, client, "count_over_time("+count+"[10m]) >= bool 2", "1") ||
		!waitForQueryValue(ctx, client, "min_over_time("+count+"[10m])", "3") ||
		!waitForQueryValue(ctx, client, "increase("+count+"[10m])", "0") {
		t.Fatal("restart proof did not retain a flat first-scrape count3 without a zero baseline")
	}
	if !waitForAlertState(ctx, client, "CloudClickerGameserverRestartLoop", "firing") {
		t.Fatal("actual pre-scrape restarts never fired the restart rule")
	}
	firingStart := waitForOperationsNotification(t, ctx, received, "CloudClickerGameserverRestartLoop", "node-exporter", "firing", before.ModTime(), time.Time{})
	t.Log("actual Docker first-scrape count3 reached the configured receiver with increase0")
	clearedAt := time.Now().UTC()
	// A request is not a metric. The native owner must create a new real
	// container and publish its observed count0; this child never writes host.prom.
	request, err := os.OpenFile("/textfile/restart-recovery-request", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	_, writeErr := request.WriteString("recover\n")
	closeErr := request.Close()
	if writeErr != nil || closeErr != nil {
		t.Fatal("restart recovery request incomplete")
	}
	if !waitForQueryValue(ctx, client, count, "0") ||
		!waitForAlertState(ctx, client, "CloudClickerGameserverRestartLoop", "inactive") {
		t.Fatal("actual replacement observation did not resolve the restart rule")
	}
	after, err := os.Stat("/textfile/host.prom")
	if err != nil || !after.ModTime().After(before.ModTime()) {
		t.Fatal("restart recovery did not publish a fresh actual host file")
	}
	waitForOperationsNotification(t, ctx, received, "CloudClickerGameserverRestartLoop", "node-exporter", "resolved", clearedAt, firingStart)
	t.Log("same restart alert resolved after the actual replacement container observation")
}
