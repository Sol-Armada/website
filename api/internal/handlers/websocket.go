package handlers

import (
	"net/http"
	"time"

	"log/slog"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"github.com/sol-armada/website/internal/realtime"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

type socketSubscribeRequest struct {
	Type   string   `json:"type"`
	Topics []string `json:"topics"`
}

type WebSocketHandler struct {
	hub      *realtime.Hub
	logger   *slog.Logger
	upgrader websocket.Upgrader
}

func NewWebSocketHandler(hub *realtime.Hub, logger *slog.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		hub:    hub,
		logger: logger,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// TODO: tighten this with explicit frontend origins per environment.
			CheckOrigin: func(_ *http.Request) bool { return true },
		},
	}
}

func (h *WebSocketHandler) Handle(c echo.Context) error {
	roles, _ := c.Get("roles").([]string)
	allowedTopics := realtime.AllowedTopicsForRoles(roles)

	conn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		h.logger.Warn("websocket upgrade failed", "error", err)
		return nil
	}

	client := realtime.NewClient(roles, 32)
	h.hub.Register(client)
	defer h.hub.Unregister(client)

	go h.writePump(conn, client)

	client.Enqueue(realtime.Envelope{
		Type:      "connected",
		Topic:     realtime.TopicSystemHealth,
		Timestamp: time.Now().UTC(),
		Payload: map[string]any{
			"topics": allowedTopicList(allowedTopics),
		},
	})

	h.readPump(conn, client, allowedTopics)
	return nil
}

func (h *WebSocketHandler) readPump(conn *websocket.Conn, client *realtime.Client, allowedTopics map[string]bool) {
	defer func() {
		_ = conn.Close()
	}()

	conn.SetReadLimit(maxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var request socketSubscribeRequest
		if err := conn.ReadJSON(&request); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Warn("websocket read error", "error", err)
			}
			break
		}

		if request.Type != "subscribe" {
			continue
		}

		granted := make([]string, 0, len(request.Topics)+1)
		granted = append(granted, realtime.TopicSystemHealth)
		for _, topic := range request.Topics {
			if !allowedTopics[topic] {
				continue
			}
			granted = append(granted, topic)
		}
		client.SetSubscriptions(granted)
		client.Enqueue(realtime.Envelope{
			Type:      "subscribed",
			Topic:     realtime.TopicSystemSubscribeOK,
			Timestamp: time.Now().UTC(),
			Payload: map[string]any{
				"topics": granted,
			},
		})
	}
}

func (h *WebSocketHandler) writePump(conn *websocket.Conn, client *realtime.Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteJSON(message); err != nil {
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func allowedTopicList(allowed map[string]bool) []string {
	topics := make([]string, 0, len(allowed))
	for topic := range allowed {
		topics = append(topics, topic)
	}
	return topics
}
