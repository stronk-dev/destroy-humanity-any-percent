package operations

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

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
