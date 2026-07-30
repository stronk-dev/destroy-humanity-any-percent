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

var ErrInvalidServer = errors.New("invalid gameserver")

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

type ReceiptRelay interface {
	Flush(context.Context) (int, error)
}

type Server struct {
	database      Database
	api           http.Handler
	realtime      Realtime
	relay         ReceiptRelay
	constantsHash string
	gate          *intentGate
	ready         atomic.Bool
	relayCancel   context.CancelFunc
	relayDone     chan struct{}
	startOnce     sync.Once
	startErr      error
}

func New(database Database, api http.Handler, realtime Realtime, relay ReceiptRelay, constantsHash string) (*Server, error) {
	if database == nil || api == nil || realtime == nil || relay == nil || !validConstantsHash(constantsHash) {
		return nil, ErrInvalidServer
	}
	server := &Server{database: database, api: api, realtime: realtime, relay: relay, constantsHash: constantsHash, gate: newIntentGate(), relayDone: make(chan struct{})}
	server.ready.Store(true)
	return server, nil
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
		server.startErr = server.realtime.Run()
		if server.startErr != nil {
			server.ready.Store(false)
			close(server.relayDone)
			return
		}
		var relayContext context.Context
		relayContext, server.relayCancel = context.WithCancel(ctx)
		go server.runRelay(relayContext)
	})
	return server.startErr
}

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
	server.ready.Store(false)
	zero := server.gate.beginDrain()
	if err := server.realtime.BroadcastDrain(server.constantsHash, now.UTC()); err != nil {
		return err
	}
	drainContext, cancel := context.WithTimeout(ctx, server.realtime.DrainTimeout())
	defer cancel()
	select {
	case <-zero:
	case <-drainContext.Done():
		return drainContext.Err()
	}
	for {
		count, err := server.relay.Flush(drainContext)
		if err != nil {
			return err
		}
		if count == 0 {
			break
		}
	}
	if server.relayCancel != nil {
		server.relayCancel()
		select {
		case <-server.relayDone:
		case <-drainContext.Done():
			return drainContext.Err()
		}
	}
	server.realtime.CloseForDrain()
	return server.realtime.Shutdown(drainContext)
}

func (server *Server) runRelay(ctx context.Context) {
	defer close(server.relayDone)
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		if _, err := server.relay.Flush(ctx); err != nil && ctx.Err() == nil {
			server.ready.Store(false)
		} else if err == nil && !server.gate.isDraining() {
			server.ready.Store(true)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (server *Server) handleReady(response http.ResponseWriter, request *http.Request) {
	if !server.ready.Load() || server.gate.isDraining() || server.database.PingContext(request.Context()) != nil {
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
