package gameserver

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
)

var (
	ErrInvalidServer        = errors.New("invalid gameserver")
	ErrBackgroundJobStopped = errors.New("gameserver background job stopped unexpectedly")
)

type Database interface {
	PingContext(context.Context) error
}

type Realtime interface {
	Run() error
	Handler() http.Handler
	BroadcastDrain(string, time.Time) error
	CloseForDrain()
	Shutdown(context.Context) error
	DrainTimeout() time.Duration
}

type PlayerRelay interface {
	Flush(context.Context) (int, error)
}

type EpochSynchronizer interface {
	Sync(context.Context) (string, error)
}

type BackgroundJob interface {
	Prime(context.Context) error
	Run(context.Context) error
}

type Server struct {
	database      Database
	api           http.Handler
	realtime      Realtime
	relay         PlayerRelay
	epochs        EpochSynchronizer
	constantsHash string
	gate          *intentGate
	stateMu       sync.Mutex
	running       bool
	ready         atomic.Bool
	draining      atomic.Bool
	jobsHealthy   atomic.Bool
	relayCancel   context.CancelFunc
	relayDone     chan struct{}
	jobs          []BackgroundJob
	jobsCancel    context.CancelFunc
	jobsDone      chan struct{}
	failures      chan error
	failureOnce   sync.Once
	startOnce     sync.Once
	startErr      error
}

func New(database Database, api http.Handler, realtime Realtime, relay PlayerRelay, epochs EpochSynchronizer, constantsHash string) (*Server, error) {
	if database == nil || api == nil || realtime == nil || relay == nil || epochs == nil || !validConstantsHash(constantsHash) {
		return nil, ErrInvalidServer
	}
	return &Server{database: database, api: api, realtime: realtime, relay: relay, epochs: epochs, constantsHash: constantsHash, gate: newIntentGate(), relayDone: make(chan struct{}), jobsDone: make(chan struct{}), failures: make(chan error, 1)}, nil
}

func (server *Server) AttachJobs(jobs ...BackgroundJob) error {
	server.stateMu.Lock()
	defer server.stateMu.Unlock()
	if server.ready.Load() || server.draining.Load() || server.jobsCancel != nil {
		return ErrInvalidServer
	}
	for _, job := range jobs {
		if job == nil {
			return ErrInvalidServer
		}
	}
	server.jobs = append(server.jobs, jobs...)
	return nil
}

func validConstantsHash(value string) bool {
	if len(value) != 71 || value[:7] != "sha256:" {
		return false
	}
	_, err := hex.DecodeString(value[7:])
	return err == nil
}

func (server *Server) Start(ctx context.Context) error {
	server.startOnce.Do(func() {
		hash, err := server.epochs.Sync(ctx)
		if err != nil || hash != server.constantsHash {
			server.startErr = errors.Join(ErrInvalidServer, err)
			close(server.relayDone)
			close(server.jobsDone)
			return
		}
		server.startErr = server.realtime.Run()
		if server.startErr != nil {
			server.ready.Store(false)
			close(server.relayDone)
			close(server.jobsDone)
			return
		}
		if err := server.startJobs(ctx); err != nil {
			server.startErr = errors.Join(err, server.realtime.Shutdown(ctx))
			server.ready.Store(false)
			close(server.relayDone)
			return
		}
		if _, err := server.relay.Flush(ctx); err != nil {
			server.startErr = errors.Join(err, server.stopJobs(ctx), server.realtime.Shutdown(ctx))
			server.ready.Store(false)
			close(server.relayDone)
			return
		}
		var relayContext context.Context
		relayContext, server.relayCancel = context.WithCancel(ctx)
		go func() {
			<-ctx.Done()
			server.stateMu.Lock()
			server.running = false
			server.ready.Store(false)
			server.stateMu.Unlock()
		}()
		server.stateMu.Lock()
		server.running = ctx.Err() == nil
		server.stateMu.Unlock()
		server.markReadyIfRunning()
		go server.runRelay(relayContext)
	})
	return server.startErr
}

func (server *Server) Failures() <-chan error { return server.failures }

func (server *Server) Handler() http.Handler {
	router := chi.NewRouter()
	router.Get("/healthz", func(response http.ResponseWriter, _ *http.Request) { response.WriteHeader(http.StatusNoContent) })
	router.Get("/readyz", server.handleReady)
	router.Handle("/connection/websocket", server.realtime.Handler())
	router.Mount("/", server.intentAdmission(server.api))
	return router
}

func (server *Server) Drain(ctx context.Context, now time.Time) error {
	if now.IsZero() {
		return ErrInvalidServer
	}
	server.stateMu.Lock()
	server.running = false
	server.draining.Store(true)
	server.ready.Store(false)
	server.stateMu.Unlock()
	drainContext, cancel := context.WithTimeout(ctx, server.realtime.DrainTimeout())
	defer cancel()
	if err := server.realtime.BroadcastDrain(server.constantsHash, now.UTC()); err != nil {
		server.gate.beginDrain()
		jobsErr := server.stopJobs(drainContext)
		server.stopRelay()
		server.realtime.CloseForDrain()
		return errors.Join(jobsErr, err, server.realtime.Shutdown(drainContext))
	}
	zero := server.gate.beginDrain()
	select {
	case <-zero:
	case <-drainContext.Done():
		jobsErr := server.stopJobs(drainContext)
		server.stopRelay()
		server.realtime.CloseForDrain()
		return errors.Join(jobsErr, drainContext.Err(), server.realtime.Shutdown(drainContext))
	}
	jobsErr := server.stopJobs(drainContext)
	for {
		count, err := server.relay.Flush(drainContext)
		if err != nil {
			server.stopRelay()
			server.realtime.CloseForDrain()
			return errors.Join(jobsErr, err, server.realtime.Shutdown(drainContext))
		}
		if count == 0 {
			break
		}
	}
	if server.relayCancel != nil {
		server.stopRelay()
		select {
		case <-server.relayDone:
		case <-drainContext.Done():
			server.realtime.CloseForDrain()
			return errors.Join(jobsErr, drainContext.Err(), server.realtime.Shutdown(drainContext))
		}
	}
	server.realtime.CloseForDrain()
	return errors.Join(jobsErr, server.realtime.Shutdown(drainContext))
}

func (server *Server) startJobs(parent context.Context) error {
	ctx, cancel := context.WithCancel(parent)
	server.jobsCancel = cancel
	server.jobsHealthy.Store(true)
	for _, job := range server.jobs {
		if err := job.Prime(ctx); err != nil {
			server.jobsHealthy.Store(false)
			cancel()
			close(server.jobsDone)
			return err
		}
	}
	var wait sync.WaitGroup
	wait.Add(len(server.jobs))
	for _, job := range server.jobs {
		go func(job BackgroundJob) {
			defer wait.Done()
			err := job.Run(ctx)
			if ctx.Err() == nil {
				if err == nil {
					err = ErrBackgroundJobStopped
				}
				server.stateMu.Lock()
				server.jobsHealthy.Store(false)
				server.ready.Store(false)
				server.stateMu.Unlock()
				cancel()
				server.reportFailure(err)
			}
		}(job)
	}
	go func() {
		wait.Wait()
		close(server.jobsDone)
	}()
	return nil
}

func (server *Server) reportFailure(err error) {
	if err == nil {
		return
	}
	server.failureOnce.Do(func() { server.failures <- err })
}

func (server *Server) stopJobs(ctx context.Context) error {
	if server.jobsCancel == nil {
		return nil
	}
	server.jobsCancel()
	select {
	case <-server.jobsDone:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (server *Server) stopRelay() {
	if server.relayCancel != nil {
		server.relayCancel()
	}
}

func (server *Server) runRelay(ctx context.Context) {
	defer close(server.relayDone)
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			server.ready.Store(false)
			return
		default:
		}
		if _, err := server.relay.Flush(ctx); err != nil && ctx.Err() == nil {
			server.ready.Store(false)
		} else if err == nil && ctx.Err() == nil {
			server.markReadyIfRunning()
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (server *Server) markReadyIfRunning() {
	server.stateMu.Lock()
	defer server.stateMu.Unlock()
	if server.running && server.jobsHealthy.Load() && !server.draining.Load() && !server.gate.isDraining() {
		server.ready.Store(true)
	}
}

func (server *Server) handleReady(response http.ResponseWriter, request *http.Request) {
	if !server.ready.Load() || server.draining.Load() || server.gate.isDraining() || server.database.PingContext(request.Context()) != nil {
		response.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (server *Server) intentAdmission(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/api/v1/intents" {
			next.ServeHTTP(response, request)
			return
		}
		if !server.gate.enter() {
			response.Header().Set("Content-Type", "application/json")
			response.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(response).Encode(struct {
				Category string `json:"category"`
				Detail   string `json:"detail"`
			}{Category: "server_draining", Detail: "retry_same_intent_id"})
			return
		}
		defer server.gate.leave()
		next.ServeHTTP(response, request)
	})
}

type intentGate struct {
	mu       sync.Mutex
	draining bool
	active   int
	zero     chan struct{}
}

func newIntentGate() *intentGate {
	zero := make(chan struct{})
	close(zero)
	return &intentGate{zero: zero}
}

func (gate *intentGate) enter() bool {
	gate.mu.Lock()
	defer gate.mu.Unlock()
	if gate.draining {
		return false
	}
	if gate.active == 0 {
		gate.zero = make(chan struct{})
	}
	gate.active++
	return true
}

func (gate *intentGate) leave() {
	gate.mu.Lock()
	defer gate.mu.Unlock()
	if gate.active <= 0 {
		panic("gameserver: intent gate underflow")
	}
	gate.active--
	if gate.active == 0 {
		close(gate.zero)
	}
}

func (gate *intentGate) beginDrain() <-chan struct{} {
	gate.mu.Lock()
	defer gate.mu.Unlock()
	gate.draining = true
	return gate.zero
}

func (gate *intentGate) isDraining() bool {
	gate.mu.Lock()
	defer gate.mu.Unlock()
	return gate.draining
}
