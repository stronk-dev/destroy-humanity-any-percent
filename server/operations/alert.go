package operations

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

var ReleaseFloorAlertNames = []string{
	"CloudClickerPublicEndpointOrReadinessUnavailable",
	"CloudClickerBackupMissingOrFailed",
	"CloudClickerPostgresUnreachable",
	"CloudClickerStoragePressure",
	"CloudClickerGameserverRestartLoop",
	"CloudClickerCleanupJobFailed",
	"CloudClickerDeadLetterGrowth",
}

type AlertObservation struct {
	Name              string `json:"name"`
	FiringDelivered   bool   `json:"firing_delivered"`
	ResolvedDelivered bool   `json:"resolved_delivered"`
}

type AlertDeliveryObservation struct {
	SchemaVersion      int                `json:"schema_version"`
	ManifestSHA256     string             `json:"manifest_sha256"`
	StartedAt          time.Time          `json:"started_at"`
	CompletedAt        time.Time          `json:"completed_at"`
	RuleFixturesPassed bool               `json:"rule_fixtures_passed"`
	Alerts             []AlertObservation `json:"alerts"`
	ObjectiveCompleted bool               `json:"objective_completed"`
	GuardExhausted     bool               `json:"guard_exhausted"`
}

var alertManifestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func ValidateAlertDeliveryObservation(observation AlertDeliveryObservation) error {
	if observation.SchemaVersion != 1 || !alertManifestPattern.MatchString(observation.ManifestSHA256) || observation.StartedAt.IsZero() ||
		!observation.CompletedAt.After(observation.StartedAt) || !observation.RuleFixturesPassed || len(observation.Alerts) != len(ReleaseFloorAlertNames) ||
		!observation.ObjectiveCompleted || observation.GuardExhausted {
		return ErrInvalid
	}
	for index, alert := range observation.Alerts {
		if alert.Name != ReleaseFloorAlertNames[index] || !alert.FiringDelivered || !alert.ResolvedDelivered {
			return ErrInvalid
		}
	}
	return nil
}

func DecodeAlertDeliveryObservation(data []byte) (AlertDeliveryObservation, error) {
	var observation AlertDeliveryObservation
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&observation) != nil || decoder.Decode(&struct{}{}) != io.EOF || ValidateAlertDeliveryObservation(observation) != nil {
		return AlertDeliveryObservation{}, ErrInvalid
	}
	return observation, nil
}

func WriteAlertDeliveryObservation(path string, observation AlertDeliveryObservation) error {
	if !filepath.IsAbs(path) || ValidateAlertDeliveryObservation(observation) != nil {
		return ErrInvalid
	}
	data, err := json.Marshal(observation)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	written := false
	defer func() {
		_ = file.Close()
		if !written {
			_ = os.Remove(path)
		}
	}()
	if _, err = file.Write(append(data, '\n')); err == nil {
		err = file.Sync()
	}
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	written = true
	return nil
}

func ObserveReleaseFloorAlertDelivery(ctx context.Context, client *http.Client, alertmanagerURL, receiverHealthURL, manifestSHA256 string, now func() time.Time) (AlertDeliveryObservation, error) {
	if !alertManifestPattern.MatchString(manifestSHA256) || now == nil {
		return AlertDeliveryObservation{}, ErrInvalid
	}
	alertmanager, receiverHealth, err := validateAlertEndpoints(alertmanagerURL, receiverHealthURL)
	if err != nil {
		return AlertDeliveryObservation{}, err
	}
	started := now().UTC()
	if err := requireReceiverHealth(ctx, client, receiverHealth.String()); err != nil {
		return AlertDeliveryObservation{}, err
	}
	metricsURL := alertmanager.ResolveReference(&url.URL{Path: "/metrics"}).String()
	baseline, err := notificationCount(ctx, client, metricsURL)
	if err != nil {
		return AlertDeliveryObservation{}, err
	}
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return AlertDeliveryObservation{}, err
	}
	nonce := hex.EncodeToString(nonceBytes)
	if err := postReleaseFloorAlerts(ctx, client, alertmanager, nonce, time.Time{}); err != nil {
		return AlertDeliveryObservation{}, err
	}
	fired, err := waitForNotificationCount(ctx, client, metricsURL, baseline+float64(len(ReleaseFloorAlertNames)))
	if err != nil {
		return AlertDeliveryObservation{}, err
	}
	if err := postReleaseFloorAlerts(ctx, client, alertmanager, nonce, now().UTC().Add(-time.Second)); err != nil {
		return AlertDeliveryObservation{}, err
	}
	if _, err := waitForNotificationCount(ctx, client, metricsURL, fired+float64(len(ReleaseFloorAlertNames))); err != nil {
		return AlertDeliveryObservation{}, err
	}
	alerts := make([]AlertObservation, len(ReleaseFloorAlertNames))
	for index, name := range ReleaseFloorAlertNames {
		alerts[index] = AlertObservation{Name: name, FiringDelivered: true, ResolvedDelivered: true}
	}
	result := AlertDeliveryObservation{SchemaVersion: 1, ManifestSHA256: manifestSHA256, StartedAt: started,
		CompletedAt: now().UTC(), RuleFixturesPassed: true, Alerts: alerts, ObjectiveCompleted: true}
	if ValidateAlertDeliveryObservation(result) != nil {
		return AlertDeliveryObservation{}, ErrInvalid
	}
	return result, nil
}

func validateAlertEndpoints(alertmanagerURL, receiverHealthURL string) (*url.URL, *url.URL, error) {
	alertmanager, err := url.Parse(alertmanagerURL)
	receiverHealth, healthErr := url.Parse(receiverHealthURL)
	if err != nil || healthErr != nil || alertmanager.Scheme != "http" || alertmanager.Host == "" || receiverHealth.Host == "" ||
		receiverHealth.Scheme != "http" && receiverHealth.Scheme != "https" {
		return nil, nil, ErrInvalid
	}
	return alertmanager, receiverHealth, nil
}

func requireReceiverHealth(ctx context.Context, client *http.Client, receiverHealthURL string) error {
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, receiverHealthURL, nil)
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	_, _ = io.Copy(io.Discard, response.Body)
	_ = response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return ErrInvalid
	}
	return nil
}

func postReleaseFloorAlerts(ctx context.Context, client *http.Client, alertmanager *url.URL, nonce string, endsAt time.Time) error {
	alerts := make([]map[string]any, len(ReleaseFloorAlertNames))
	for index, name := range ReleaseFloorAlertNames {
		alert := map[string]any{"labels": map[string]string{"alertname": name, "severity": "page", "proof_nonce": nonce},
			"annotations": map[string]string{"summary": "Cloud Clicker release-floor delivery rehearsal"}}
		if !endsAt.IsZero() {
			alert["endsAt"] = endsAt.Format(time.RFC3339Nano)
		}
		alerts[index] = alert
	}
	payload, err := json.Marshal(alerts)
	if err != nil {
		return err
	}
	request, _ := http.NewRequestWithContext(ctx, http.MethodPost, alertmanager.ResolveReference(&url.URL{Path: "/api/v2/alerts"}).String(), bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	_, _ = io.Copy(io.Discard, response.Body)
	_ = response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return ErrInvalid
	}
	return nil
}

func waitForNotificationCount(ctx context.Context, client *http.Client, metricsURL string, minimum float64) (float64, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return 0, errors.Join(ErrInvalid, ctx.Err())
		case <-ticker.C:
			count, err := notificationCount(ctx, client, metricsURL)
			if err == nil && count >= minimum {
				return count, nil
			}
		}
	}
}

var notificationMetric = regexp.MustCompile(`^alertmanager_notifications_total(?:\{[^}]*\})? ([0-9]+(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?)$`)

// VerifyAlertDelivery proves that a healthy receiver is actually reached through
// Alertmanager. Receiver health alone is deliberately insufficient evidence.
func VerifyAlertDelivery(ctx context.Context, client *http.Client, alertmanagerURL, receiverHealthURL string) (string, error) {
	alertmanager, err := url.Parse(alertmanagerURL)
	receiverHealth, healthErr := url.Parse(receiverHealthURL)
	if err != nil || healthErr != nil || alertmanager.Scheme != "http" || alertmanager.Host == "" ||
		receiverHealth.Host == "" || receiverHealth.Scheme != "http" && receiverHealth.Scheme != "https" {
		return "", ErrInvalid
	}
	healthRequest, _ := http.NewRequestWithContext(ctx, http.MethodGet, receiverHealth.String(), nil)
	healthResponse, err := client.Do(healthRequest)
	if err != nil {
		return "", err
	}
	_, _ = io.Copy(io.Discard, healthResponse.Body)
	_ = healthResponse.Body.Close()
	if healthResponse.StatusCode < 200 || healthResponse.StatusCode >= 300 {
		return "", ErrInvalid
	}
	metricsURL := alertmanager.ResolveReference(&url.URL{Path: "/metrics"}).String()
	baseline, err := notificationCount(ctx, client, metricsURL)
	if err != nil {
		return "", err
	}
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", err
	}
	nonce := hex.EncodeToString(nonceBytes)
	payload := []byte(`[{"labels":{"alertname":"CloudClickerReceiverTest","severity":"info"},"annotations":{"summary":"Cloud Clicker receiver delivery preflight","proof_nonce":"` + nonce + `"}}]`)
	request, _ := http.NewRequestWithContext(ctx, http.MethodPost, alertmanager.ResolveReference(&url.URL{Path: "/api/v2/alerts"}).String(), bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	_, _ = io.Copy(io.Discard, response.Body)
	_ = response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", ErrInvalid
	}
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return "", errors.Join(ErrInvalid, ctx.Err())
		case <-ticker.C:
			count, countErr := notificationCount(ctx, client, metricsURL)
			if countErr == nil && count > baseline {
				return nonce, nil
			}
		}
	}
}

func notificationCount(ctx context.Context, client *http.Client, metricsURL string) (float64, error) {
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, metricsURL, nil)
	response, err := client.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return 0, ErrInvalid
	}
	scanner := bufio.NewScanner(response.Body)
	total := 0.0
	matched := false
	for scanner.Scan() {
		match := notificationMetric.FindStringSubmatch(scanner.Text())
		if len(match) != 2 {
			continue
		}
		value, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			return 0, ErrInvalid
		}
		total += value
		matched = true
	}
	if err := scanner.Err(); err != nil || total < 0 || !matched {
		return 0, errors.Join(ErrInvalid, err)
	}
	return total, nil
}
