package realtime

import (
	"log/slog"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

const (
	TopicSystemHealth      = "system.health"
	TopicAdminMembers      = "admin.members.updated"
	TopicAdminAttendance   = "admin.attendance.updated"
	TopicAdminTokenLedger  = "admin.token_ledger.updated"
	TopicSystemSubscribeOK = "system.subscription"
)

type Envelope struct {
	Type      string    `json:"type"`
	Topic     string    `json:"topic"`
	Sequence  int64     `json:"sequence,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"payload,omitempty"`
}

type Client struct {
	Roles []string
	Send  chan Envelope

	mu            sync.RWMutex
	subscriptions map[string]struct{}
}

func NewClient(roles []string, sendBuffer int) *Client {
	if sendBuffer <= 0 {
		sendBuffer = 16
	}

	client := &Client{
		Roles:         append([]string(nil), roles...),
		Send:          make(chan Envelope, sendBuffer),
		subscriptions: make(map[string]struct{}),
	}
	client.subscriptions[TopicSystemHealth] = struct{}{}
	return client
}

func (c *Client) SetSubscriptions(topics []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	next := map[string]struct{}{TopicSystemHealth: {}}
	for _, topic := range topics {
		next[topic] = struct{}{}
	}
	c.subscriptions = next
}

func (c *Client) IsSubscribed(topic string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.subscriptions[topic]
	return ok
}

func (c *Client) Enqueue(event Envelope) bool {
	select {
	case c.Send <- event:
		return true
	default:
		return false
	}
}

type Hub struct {
	logger *slog.Logger

	mu      sync.RWMutex
	clients map[*Client]struct{}

	sequence int64
	closed   atomic.Bool
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		logger:  logger,
		clients: make(map[*Client]struct{}),
	}
}

func (h *Hub) Register(client *Client) {
	if h.closed.Load() {
		return
	}

	h.mu.Lock()
	h.clients[client] = struct{}{}
	count := len(h.clients)
	h.mu.Unlock()

	h.logger.Debug("ws client connected", "clients", count)
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	if _, exists := h.clients[client]; exists {
		delete(h.clients, client)
		close(client.Send)
	}
	count := len(h.clients)
	h.mu.Unlock()

	h.logger.Debug("ws client disconnected", "clients", count)
}

func (h *Hub) Publish(topic string, payload any) {
	if h.closed.Load() {
		return
	}

	event := Envelope{
		Type:      "event",
		Topic:     topic,
		Sequence:  atomic.AddInt64(&h.sequence, 1),
		Timestamp: time.Now().UTC(),
		Payload:   payload,
	}

	h.mu.RLock()
	for client := range h.clients {
		if !client.IsSubscribed(topic) {
			continue
		}
		if !client.Enqueue(event) {
			h.logger.Warn("dropping ws event for slow client", "topic", topic)
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) RunHealthHeartbeat(interval time.Duration) {
	if interval <= 0 {
		interval = 20 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if h.closed.Load() {
			return
		}

		h.Publish(TopicSystemHealth, map[string]any{
			"status": "ok",
		})

		<-ticker.C
	}
}

func (h *Hub) Close() {
	if !h.closed.CompareAndSwap(false, true) {
		return
	}

	h.mu.Lock()
	for client := range h.clients {
		close(client.Send)
		delete(h.clients, client)
	}
	h.mu.Unlock()
}

func AllowedTopicsForRoles(roles []string) map[string]bool {
	allowed := map[string]bool{
		TopicSystemHealth: true,
	}

	isAdmin := slices.Contains(roles, "admin")

	if isAdmin {
		allowed[TopicAdminMembers] = true
		allowed[TopicAdminAttendance] = true
		allowed[TopicAdminTokenLedger] = true
	}

	return allowed
}
