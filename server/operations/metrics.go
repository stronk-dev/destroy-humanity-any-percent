package operations

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var ErrInvalid = errors.New("invalid operations instrumentation")

type Database interface {
	PingContext(context.Context) error
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type Registry struct {
	registry       *prometheus.Registry
	ready          prometheus.Gauge
	httpRequests   *prometheus.CounterVec
	httpDuration   *prometheus.HistogramVec
	jobRuns        *prometheus.CounterVec
	jobLastSuccess *prometheus.GaugeVec
	invariants     *prometheus.CounterVec
}

func NewRegistry(database Database) (*Registry, error) {
	if database == nil {
		return nil, ErrInvalid
	}
	registry := prometheus.NewRegistry()
	result := &Registry{
		registry: registry,
		ready: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "cloud_clicker", Name: "ready", Help: "Whether the gameserver is accepting new player work.",
		}),
		httpRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "cloud_clicker", Name: "http_requests_total", Help: "HTTP requests by bounded route, method, and status class.",
		}, []string{"route", "method", "status_class"}),
		httpDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "cloud_clicker", Name: "http_request_duration_seconds", Help: "HTTP request duration by bounded route and method.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		}, []string{"route", "method"}),
		jobRuns: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "cloud_clicker", Name: "job_runs_total", Help: "Composed background job executions by bounded name and result.",
		}, []string{"job", "result"}),
		jobLastSuccess: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "cloud_clicker", Name: "job_last_success_timestamp_seconds", Help: "Unix timestamp of the latest successful composed job execution.",
		}, []string{"job"}),
		invariants: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "cloud_clicker", Name: "invariants_total", Help: "Operational invariant reports by bounded kind.",
		}, []string{"kind"}),
	}
	collectors := []prometheus.Collector{
		prometheus.NewGoCollector(), prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}), result.ready,
		result.httpRequests, result.httpDuration, result.jobRuns, result.jobLastSuccess, result.invariants,
		newDatabaseCollector(database),
	}
	for _, collector := range collectors {
		if err := registry.Register(collector); err != nil {
			return nil, errors.Join(ErrInvalid, err)
		}
	}
	return result, nil
}

func (registry *Registry) Prometheus() *prometheus.Registry {
	if registry == nil {
		return nil
	}
	return registry.registry
}

func (registry *Registry) Handler() http.Handler {
	if registry == nil || registry.registry == nil {
		return http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
			response.WriteHeader(http.StatusServiceUnavailable)
		})
	}
	return promhttp.HandlerFor(registry.registry, promhttp.HandlerOpts{EnableOpenMetrics: false})
}

func (registry *Registry) SetReady(ready bool) {
	if registry == nil {
		return
	}
	if ready {
		registry.ready.Set(1)
	} else {
		registry.ready.Set(0)
	}
}

var jobNames = map[string]struct{}{
	"verification": {}, "presence": {}, "clearing": {}, "guild_sweep": {}, "credential_cleanup": {},
}

func (registry *Registry) ObserveJob(name string, now time.Time, run func() error) error {
	if registry == nil || run == nil || now.IsZero() {
		return ErrInvalid
	}
	if _, ok := jobNames[name]; !ok {
		return ErrInvalid
	}
	err := run()
	result := "success"
	if err != nil {
		result = "failure"
	}
	registry.jobRuns.WithLabelValues(name, result).Inc()
	if err == nil {
		registry.jobLastSuccess.WithLabelValues(name).Set(float64(now.UTC().Unix()))
	}
	return err
}

var invariantKinds = map[string]struct{}{
	"afford_fallback": {}, "residual_clamp": {}, "residual_abort": {},
	"player_message_dead_letter": {}, "player_resync_signal_failed": {},
	"verification_dead_letter": {}, "verification_poison_dead_letter": {},
}

func (registry *Registry) Increment(kind string) {
	if registry == nil {
		return
	}
	if _, ok := invariantKinds[kind]; !ok {
		kind = "unknown"
	}
	registry.invariants.WithLabelValues(kind).Inc()
}

func (registry *Registry) HTTP(next http.Handler) http.Handler {
	if registry == nil || next == nil {
		return next
	}
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		started := time.Now()
		capture := &statusWriter{ResponseWriter: response}
		next.ServeHTTP(capture, request)
		status := capture.status
		if status == 0 {
			status = http.StatusOK
		}
		route, method := boundedRoute(request.URL.Path), boundedMethod(request.Method)
		registry.httpRequests.WithLabelValues(route, method, fmt.Sprintf("%dxx", status/100)).Inc()
		registry.httpDuration.WithLabelValues(route, method).Observe(time.Since(started).Seconds())
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (writer *statusWriter) WriteHeader(status int) {
	if writer.status == 0 {
		writer.status = status
	}
	writer.ResponseWriter.WriteHeader(status)
}

func (writer *statusWriter) Write(data []byte) (int, error) {
	if writer.status == 0 {
		writer.WriteHeader(http.StatusOK)
	}
	return writer.ResponseWriter.Write(data)
}

func (writer *statusWriter) Unwrap() http.ResponseWriter { return writer.ResponseWriter }

func (writer *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := writer.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

func boundedMethod(method string) string {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodOptions:
		return strings.ToLower(method)
	default:
		return "other"
	}
}

func boundedRoute(path string) string {
	switch path {
	case "/healthz":
		return "health"
	case "/readyz":
		return "ready"
	case "/metrics":
		return "metrics"
	case "/connection/websocket":
		return "websocket"
	case "/api/v1/intents":
		return "intents"
	default:
		if strings.HasPrefix(path, "/api/") {
			return "api"
		}
		return "other"
	}
}

type databaseCollector struct {
	database    Database
	reachable   *prometheus.Desc
	collection  *prometheus.Desc
	outbox      *prometheus.Desc
	deadLetters *prometheus.Desc
}

func newDatabaseCollector(database Database) *databaseCollector {
	return &databaseCollector{
		database:    database,
		reachable:   prometheus.NewDesc("cloud_clicker_postgres_reachable", "Whether the authoritative Postgres database responds.", nil, nil),
		collection:  prometheus.NewDesc("cloud_clicker_database_collection_success", "Whether the current bounded database population query completed.", nil, nil),
		outbox:      prometheus.NewDesc("cloud_clicker_outbox_pending", "Unpublished non-dead player outbox rows.", nil, nil),
		deadLetters: prometheus.NewDesc("cloud_clicker_dead_letters", "Dead-letter rows by bounded queue family.", []string{"queue"}, nil),
	}
}

func (collector *databaseCollector) Describe(output chan<- *prometheus.Desc) {
	output <- collector.reachable
	output <- collector.collection
	output <- collector.outbox
	output <- collector.deadLetters
}

func (collector *databaseCollector) Collect(output chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if collector.database.PingContext(ctx) != nil {
		output <- prometheus.MustNewConstMetric(collector.reachable, prometheus.GaugeValue, 0)
		output <- prometheus.MustNewConstMetric(collector.collection, prometheus.GaugeValue, 0)
		return
	}
	output <- prometheus.MustNewConstMetric(collector.reachable, prometheus.GaugeValue, 1)
	var outbox, transportDead, verificationDead, poisonDead float64
	err := collector.database.QueryRowContext(ctx, `SELECT
		(SELECT count(*) FROM transport_player_outbox WHERE published_at IS NULL AND dead_lettered_at IS NULL),
		(SELECT count(*) FROM transport_player_outbox WHERE dead_lettered_at IS NOT NULL),
		(SELECT count(*) FROM verification_dead_letters),
		(SELECT count(*) FROM verification_poison_dead_letters)`).Scan(&outbox, &transportDead, &verificationDead, &poisonDead)
	if err != nil {
		output <- prometheus.MustNewConstMetric(collector.collection, prometheus.GaugeValue, 0)
		return
	}
	output <- prometheus.MustNewConstMetric(collector.collection, prometheus.GaugeValue, 1)
	output <- prometheus.MustNewConstMetric(collector.outbox, prometheus.GaugeValue, outbox)
	for queue, value := range map[string]float64{"transport": transportDead, "verification": verificationDead, "verification_poison": poisonDead} {
		output <- prometheus.MustNewConstMetric(collector.deadLetters, prometheus.GaugeValue, value, queue)
	}
}
