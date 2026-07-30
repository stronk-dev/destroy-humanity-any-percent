// Package testhttp provides a real net/http client/server exchange over
// in-memory pipes. Tests exercise HTTP upgrades and framing without binding an
// operating-system socket, which keeps the suite sandbox-safe.
package testhttp

import (
	"context"
	"net"
	"net/http"
	"sync"
)

type Server struct {
	URL       string
	Client    *http.Client
	server    *http.Server
	listener  *listener
	transport *http.Transport
}

func New(handler http.Handler) *Server {
	listener := newListener()
	server := &http.Server{Handler: handler}
	transport := &http.Transport{
		Proxy: nil,
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return listener.dial(ctx)
		},
	}
	result := &Server{URL: "http://memory", Client: &http.Client{Transport: transport}, server: server, listener: listener, transport: transport}
	go func() { _ = server.Serve(listener) }()
	return result
}

func (server *Server) Close() {
	_ = server.server.Close()
	server.transport.CloseIdleConnections()
	_ = server.listener.Close()
}

type listener struct {
	connections chan net.Conn
	closed      chan struct{}
	once        sync.Once
}

func newListener() *listener {
	return &listener{connections: make(chan net.Conn), closed: make(chan struct{})}
}

func (listener *listener) dial(ctx context.Context) (net.Conn, error) {
	client, server := net.Pipe()
	select {
	case listener.connections <- server:
		return client, nil
	case <-listener.closed:
		_ = client.Close()
		_ = server.Close()
		return nil, net.ErrClosed
	case <-ctx.Done():
		_ = client.Close()
		_ = server.Close()
		return nil, ctx.Err()
	}
}

func (listener *listener) Accept() (net.Conn, error) {
	select {
	case connection := <-listener.connections:
		return connection, nil
	case <-listener.closed:
		return nil, net.ErrClosed
	}
}

func (listener *listener) Close() error {
	listener.once.Do(func() { close(listener.closed) })
	return nil
}

func (*listener) Addr() net.Addr { return address("memory") }

type address string

func (value address) Network() string { return string(value) }
func (value address) String() string  { return string(value) }
