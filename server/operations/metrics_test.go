package operations

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

type fakeDatabase struct {
	ping error
}

func (database fakeDatabase) PingContext(context.Context) error               { return database.ping }
func (fakeDatabase) QueryRowContext(context.Context, string, ...any) *sql.Row { return &sql.Row{} }

func TestRegistryUsesOnlyBoundedHTTPAndJobLabels(t *testing.T) {
	registry, err := NewRegistry(fakeDatabase{ping: errors.New("offline")})
	if err != nil {
		t.Fatal(err)
	}
	handler := registry.HTTP(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) { response.WriteHeader(http.StatusTeapot) }))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/users/private-account-id", nil))
	now := time.Unix(1_700_000_000, 0)
	if err := registry.ObserveJob("credential_cleanup", now, func() error { return errors.New("secret-bearing failure") }); err == nil {
		t.Fatal("failed cleanup reported success")
	}
	registry.Increment("attacker-controlled-kind")
	text := gatheredText(t, registry)
	for _, want := range []string{
		`cloud_clicker_http_requests_total{method="post",route="api",status_class="4xx"} 1`,
		`cloud_clicker_job_runs_total{job_name="credential_cleanup",result="failure"} 1`,
		`cloud_clicker_invariants_total{kind="unknown"} 1`,
		`cloud_clicker_postgres_reachable 0`,
		`cloud_clicker_database_collection_success 0`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing bounded metric %q in:\n%s", want, text)
		}
	}
	for _, forbidden := range []string{"private-account-id", "secret-bearing failure", "attacker-controlled-kind"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("unbounded/private value escaped into metrics: %q", forbidden)
		}
	}
}

func TestRegistryReadinessAndSuccessfulJobAreObservable(t *testing.T) {
	registry, err := NewRegistry(fakeDatabase{ping: errors.New("offline")})
	if err != nil {
		t.Fatal(err)
	}
	registry.SetReady(true)
	now := time.Unix(1_700_000_000, 0)
	if err := registry.ObserveJob("verification", now, func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	text := gatheredText(t, registry)
	for _, want := range []string{
		"cloud_clicker_ready 1",
		`cloud_clicker_job_runs_total{job_name="verification",result="success"} 1`,
		`cloud_clicker_job_last_success_timestamp_seconds{job_name="verification"} 1.7e+09`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing metric %q in:\n%s", want, text)
		}
	}
	registry.SetReady(false)
	if text = gatheredText(t, registry); !strings.Contains(text, "cloud_clicker_ready 0") {
		t.Fatalf("readiness did not clear:\n%s", text)
	}
}

func TestRegistryRejectsUnregisteredJobName(t *testing.T) {
	registry, _ := NewRegistry(fakeDatabase{ping: errors.New("offline")})
	called := false
	if err := registry.ObserveJob("player-supplied", time.Now(), func() error { called = true; return nil }); !errors.Is(err, ErrInvalid) || called {
		t.Fatalf("unregistered job accepted: err=%v called=%v", err, called)
	}
}

func gatheredText(t *testing.T, registry *Registry) string {
	t.Helper()
	families, err := registry.Prometheus().Gather()
	if err != nil {
		t.Fatal(err)
	}
	var builder strings.Builder
	for _, family := range families {
		for _, metric := range family.Metric {
			builder.WriteString(family.GetName())
			builder.WriteString("{")
			for index, label := range metric.Label {
				if index > 0 {
					builder.WriteString(",")
				}
				builder.WriteString(label.GetName() + "=\"" + label.GetValue() + "\"")
			}
			builder.WriteString("} ")
			switch {
			case metric.Counter != nil:
				builder.WriteString(fmtFloat(metric.Counter.GetValue()))
			case metric.Gauge != nil:
				builder.WriteString(fmtFloat(metric.Gauge.GetValue()))
			}
			builder.WriteString("\n")
		}
	}
	return strings.ReplaceAll(builder.String(), "{}", "")
}

func fmtFloat(value float64) string {
	if value == 1_700_000_000 {
		return "1.7e+09"
	}
	if value == float64(int64(value)) {
		return strconv.FormatInt(int64(value), 10)
	}
	return strconv.FormatFloat(value, 'g', -1, 64)
}

func TestRegistryAvoidsTargetLabelsAndPrecreatesJobFailureSeries(t *testing.T) {
	registry, err := NewRegistry(fakeDatabase{ping: errors.New("offline")})
	if err != nil {
		t.Fatal(err)
	}
	families, err := registry.Prometheus().Gather()
	if err != nil {
		t.Fatal(err)
	}
	failures := map[string]float64{}
	for _, family := range families {
		for _, metric := range family.GetMetric() {
			labels := map[string]string{}
			for _, label := range metric.GetLabel() {
				// Prometheus attaches job/instance to every scraped target; an
				// application label with either name is renamed exported_* and
				// silently stops matching rules that select it.
				if label.GetName() == "job" || label.GetName() == "instance" {
					t.Fatalf("%s exposes reserved target label %q", family.GetName(), label.GetName())
				}
				labels[label.GetName()] = label.GetValue()
			}
			if family.GetName() == "cloud_clicker_job_runs_total" && labels["result"] == "failure" {
				failures[labels["job_name"]] = metric.GetCounter().GetValue()
			}
		}
	}
	for name := range jobNames {
		value, ok := failures[name]
		if !ok || value != 0 {
			t.Fatalf("job %s failure series before any run present=%t value=%v", name, ok, value)
		}
	}
}
