package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/kernel"

	"github.com/centrifugal/centrifuge"
)

var (
	ErrInvalidNode        = errors.New("invalid transport node")
	disconnectQueueFull   = centrifuge.Disconnect{Code: CloseQueueOverflow, Reason: "receipt queue overflow"}
	disconnectAuthExpired = centrifuge.Disconnect{Code: CloseAuthExpired, Reason: "access token expired"}
	disconnectReplaced    = centrifuge.Disconnect{Code: CloseReplaced, Reason: "older connection replaced"}
	disconnectServerDrain = centrifuge.Disconnect{Code: CloseServerDrain, Reason: "server draining"}
)

var configureSlowDisconnectOnce sync.Once

type Authenticator interface {
	Authenticate(context.Context, string) (account.Claims, error)
}

type Node struct {
	policy      Policy
	auth        Authenticator
	memberships Memberships
	node        *centrifuge.Node

	connectionsMu sync.Mutex
	worldMu       sync.Mutex
	worldPending  []byte
	worldRevision uint64

	startOnce sync.Once
	stopOnce  sync.Once
	stop      chan struct{}
	done      chan struct{}
	startErr  error
}

func NewNode(policy Policy, auth Authenticator, memberships Memberships) (*Node, error) {
	if !policy.valid() || auth == nil {
		return nil, ErrInvalidNode
	}
	// Centrifuge exposes its slow-writer disconnect as a package policy rather
	// than a per-node option. Configure it exactly once before any Node can run;
	// this avoids per-node mutation races while preserving the application wire
	// contract that overflow is recoverable close code 4000.
	configureSlowDisconnectOnce.Do(func() { centrifuge.DisconnectSlow = disconnectQueueFull })
	engine, err := centrifuge.New(centrifuge.Config{
		Version:                        kernel.Version,
		ClientConnectIncludeServerTime: true,
		ClientQueueMaxSize:             policy.PlayerQueueBytes,
		ClientChannelLimit:             policy.SubscriptionsPerConnection,
		RecoveryMaxPublicationLimit:    policy.PlayerHistorySize,
		HistoryMaxPublicationLimit:     policy.PlayerHistorySize,
	})
	if err != nil {
		return nil, err
	}
	result := &Node{policy: policy, auth: auth, memberships: memberships, node: engine, stop: make(chan struct{}), done: make(chan struct{})}
	result.bindHandlers()
	return result, nil
}

func (n *Node) bindHandlers() {
	n.node.OnConnecting(func(ctx context.Context, event centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
		if event.Token == "" {
			return centrifuge.ConnectReply{}, disconnectAuthExpired
		}
		claims, err := n.auth.Authenticate(ctx, event.Token)
		if err != nil {
			return centrifuge.ConnectReply{}, disconnectAuthExpired
		}
		identity := Identity{AccountID: claims.Subject, FounderID: claims.FounderID}
		info, err := json.Marshal(identity)
		if err != nil {
			return centrifuge.ConnectReply{}, centrifuge.DisconnectServerError
		}
		return centrifuge.ConnectReply{
			Credentials: &centrifuge.Credentials{UserID: claims.Subject, ExpireAt: claims.ExpiresAt, Info: info},
			Storage:     map[string]any{"access_token": event.Token},
		}, nil
	})

	n.node.OnConnect(func(client *centrifuge.Client) {
		n.replaceOldestConnections(client)
		client.OnSubscribe(func(event centrifuge.SubscribeEvent, callback centrifuge.SubscribeCallback) {
			identity, ok := identityFromClient(client)
			if !ok || !Authorized(identity, event.Channel, n.memberships) {
				callback(centrifuge.SubscribeReply{}, centrifuge.ErrorPermissionDenied)
				return
			}
			callback(centrifuge.SubscribeReply{Options: n.subscriptionOptions(event.Channel)}, nil)
		})
		client.OnPublish(func(_ centrifuge.PublishEvent, callback centrifuge.PublishCallback) {
			callback(centrifuge.PublishReply{}, centrifuge.ErrorPermissionDenied)
		})
		client.OnRefresh(func(_ centrifuge.RefreshEvent, callback centrifuge.RefreshCallback) {
			callback(centrifuge.RefreshReply{}, disconnectAuthExpired)
		})
		client.OnAlive(func() {
			storage, release := client.AcquireStorage()
			token, _ := storage["access_token"].(string)
			release(storage)
			if token == "" {
				client.Disconnect(disconnectAuthExpired)
				return
			}
			if _, err := n.auth.Authenticate(context.Background(), token); err != nil {
				client.Disconnect(disconnectAuthExpired)
			}
		})
	})
}

func (n *Node) Run() error {
	n.startOnce.Do(func() {
		n.startErr = n.node.Run()
		if n.startErr != nil {
			close(n.done)
			return
		}
		go n.worldLoop()
	})
	return n.startErr
}

func (n *Node) Handler() http.Handler {
	allowed := make(map[string]struct{}, len(n.policy.AllowedOrigins))
	for _, origin := range n.policy.AllowedOrigins {
		allowed[origin] = struct{}{}
	}
	return centrifuge.NewWebsocketHandler(n.node, centrifuge.WebsocketConfig{
		UseWriteBufferPool: true,
		MessageSizeLimit:   n.policy.MessageBytes,
		CheckOrigin: func(request *http.Request) bool {
			_, ok := allowed[request.Header.Get("Origin")]
			return ok
		},
	})
}

func (n *Node) Publish(envelope Envelope) error {
	data, err := Encode(envelope, n.policy.MessageBytes)
	if err != nil {
		return err
	}
	if envelope.Channel == "world" {
		n.worldMu.Lock()
		if uint64(envelope.Revision) >= n.worldRevision {
			n.worldPending = append(n.worldPending[:0], data...)
			n.worldRevision = uint64(envelope.Revision)
		}
		n.worldMu.Unlock()
		return nil
	}
	_, err = n.node.Publish(envelope.Channel, data, n.publishOptions(envelope.Channel)...)
	return err
}

func (n *Node) Drain(ctx context.Context, constantsHash string, now time.Time) error {
	if err := n.BroadcastDrain(constantsHash, now); err != nil {
		return err
	}
	n.CloseForDrain()
	return n.Shutdown(ctx)
}

func (n *Node) BroadcastDrain(constantsHash string, now time.Time) error {
	if !hashPattern.MatchString(constantsHash) || now.IsZero() {
		return ErrInvalidNode
	}
	payload, _ := json.Marshal(struct {
		Code          string `json:"code"`
		ResumeAfterMS int64  `json:"resume_after_ms"`
	}{Code: "server_restarting", ResumeAfterMS: n.policy.DrainTimeoutMS})
	channels := n.node.Hub().Channels()
	sort.Strings(channels)
	for _, channel := range channels {
		data, err := Encode(Envelope{Version: WireVersion, Channel: channel, Kind: "system", Revision: 0, ConstantsHash: constantsHash, Timestamp: now.UTC(), Payload: payload}, n.policy.MessageBytes)
		if err != nil {
			return err
		}
		if _, err := n.node.Publish(channel, data, n.publishOptions(channel)...); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) CloseForDrain() {
	for _, client := range n.node.Hub().Connections() {
		client.Disconnect(disconnectServerDrain)
	}
}

func (n *Node) Shutdown(ctx context.Context) error {
	n.stopOnce.Do(func() { close(n.stop) })
	select {
	case <-n.done:
	case <-ctx.Done():
		return errors.Join(ctx.Err(), n.node.Shutdown(ctx))
	}
	return n.node.Shutdown(ctx)
}

func (n *Node) ConnectionCount() int { return n.node.Hub().NumClients() }

func (n *Node) DrainTimeout() time.Duration {
	return time.Duration(n.policy.DrainTimeoutMS) * time.Millisecond
}

func (n *Node) worldLoop() {
	defer close(n.done)
	ticker := time.NewTicker(time.Second / time.Duration(n.policy.WorldHz))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			n.flushWorld()
		case <-n.stop:
			n.flushWorld()
			return
		}
	}
}

func (n *Node) flushWorld() {
	n.worldMu.Lock()
	data := append([]byte(nil), n.worldPending...)
	revision := n.worldRevision
	n.worldPending = nil
	n.worldMu.Unlock()
	if len(data) == 0 {
		return
	}
	_, _ = n.node.Publish("world", data,
		centrifuge.WithHistory(1, time.Duration(n.policy.PlayerHistoryTTLMS)*time.Millisecond),
		centrifuge.WithVersion(revision, kernel.Version))
}

func (n *Node) publishOptions(channel string) []centrifuge.PublishOption {
	ttl := time.Duration(n.policy.PlayerHistoryTTLMS) * time.Millisecond
	switch {
	case isPlayerChannel(channel):
		return []centrifuge.PublishOption{centrifuge.WithHistory(n.policy.PlayerHistorySize, ttl)}
	case channel == "feed" || strings.HasPrefix(channel, "guild:") || strings.HasPrefix(channel, "cohort:"):
		return []centrifuge.PublishOption{centrifuge.WithHistory(n.policy.FeedHistorySize, ttl)}
	case strings.HasPrefix(channel, "match:"):
		return []centrifuge.PublishOption{centrifuge.WithHistory(1, ttl)}
	default:
		return nil
	}
}

func (n *Node) subscriptionOptions(channel string) centrifuge.SubscribeOptions {
	options := centrifuge.SubscribeOptions{}
	switch {
	case isPlayerChannel(channel), channel == "feed", strings.HasPrefix(channel, "guild:"), strings.HasPrefix(channel, "cohort:"), strings.HasPrefix(channel, "match:"):
		options.EnableRecovery = true
		options.EnablePositioning = true
	case channel == "world":
		options.EnableRecovery = true
		options.EnablePositioning = true
		options.RecoveryMode = centrifuge.RecoveryModeCache
	}
	if channel == "feed" || strings.HasPrefix(channel, "guild:") || strings.HasPrefix(channel, "cohort:") {
		options.EmitPresence = true
		options.EmitJoinLeave = true
		options.PushJoinLeave = true
	}
	return options
}

func (n *Node) replaceOldestConnections(current *centrifuge.Client) {
	n.connectionsMu.Lock()
	defer n.connectionsMu.Unlock()
	connections := n.node.Hub().UserConnections(current.UserID())
	if len(connections) <= n.policy.ConnectionsPerAccount {
		return
	}
	ordered := make([]*centrifuge.Client, 0, len(connections))
	for _, client := range connections {
		if client.ID() != current.ID() {
			ordered = append(ordered, client)
		}
	}
	sort.Slice(ordered, func(left, right int) bool {
		if ordered[left].ConnectedAtMS() != ordered[right].ConnectedAtMS() {
			return ordered[left].ConnectedAtMS() < ordered[right].ConnectedAtMS()
		}
		return ordered[left].ID() < ordered[right].ID()
	})
	remove := len(connections) - n.policy.ConnectionsPerAccount
	for index := 0; index < remove && index < len(ordered); index++ {
		ordered[index].Disconnect(disconnectReplaced)
	}
}

func identityFromClient(client *centrifuge.Client) (Identity, bool) {
	var identity Identity
	if err := json.Unmarshal(client.Info(), &identity); err != nil || identity.AccountID != client.UserID() || identity.FounderID == "" {
		return Identity{}, false
	}
	return identity, true
}
