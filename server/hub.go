package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"slices"

	"github.com/sol-armada/sol-bot/members"
)

type CommandRequest struct {
	Thing  string
	Action string
	Arg    any
	Token  string
	Client *Client
}

type CommandResponse struct {
	Thing  string `json:"thing"`
	Action string `json:"action"`
	Result any    `json:"result"`
	Error  string `json:"error"`
}

func (c *CommandResponse) ToJsonBytes() []byte {
	j, _ := json.Marshal(c)
	return j
}

type Action func(ctx context.Context, c *Client, arg any) CommandResponse

type Hub struct {
	// registered clients
	clients map[*Client]bool

	// inbound commands from the clients
	broadcast chan *CommandRequest

	// register requests from the clients
	register chan *Client

	// unregister requests from clients
	unregister chan *Client

	ctx context.Context
}

type contextKey string

const (
	contextKeyAccess contextKey = "access"
	contextKeyMember contextKey = "member"
)

func newHub(ctx context.Context) *Hub {
	return &Hub{
		broadcast:  make(chan *CommandRequest),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		ctx:        ctx,
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			slog.Default().Info("registering client")
			h.clients[client] = true
		case client := <-h.unregister:
			slog.Default().Info("unregistering client")
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case command := <-h.broadcast:
			ctx := context.Background()

			logger := slog.Default()

			logger.Info("received command", slog.String("thing", command.Thing), slog.String("action", command.Action))

			excludeCommands := []string{"auth", "login", "rsi"}

			var access userAccess
			if command.Token != "" && command.Token != "null" && !slices.Contains(excludeCommands, command.Action) && !slices.Contains(excludeCommands, command.Thing) {
				uAccessRaw, err := decrypt(command.Token)
				if err != nil {
					command.Client.send <- []byte(err.Error())
					continue
				}
				if err := json.Unmarshal([]byte(uAccessRaw), &access); err != nil {
					command.Client.send <- []byte(err.Error())
					continue
				}

				ctx = context.WithValue(ctx, contextKeyAccess, access)

				member, err := members.Get(access.Id)
				if err != nil {
					logger.Error("failed to get user", "error", err)
					r := CommandResponse{Thing: command.Thing, Action: command.Action, Error: "internal_error"}
					command.Client.send <- r.ToJsonBytes()
					continue
				}

				ctx = context.WithValue(ctx, contextKeyMember, member)
			}

			var res CommandResponse

			switch command.Thing {
			case "rsi":
				res = rsiActions[command.Action](ctx, command.Client, command.Arg)
			case "members":
				res = membersActions[command.Action](ctx, command.Client, command.Arg)
			case "login":
				res = loginActions[command.Action](ctx, command.Client, command.Arg)
			case "contracts":
				res = contractsActions[command.Action](ctx, command.Client, command.Arg)
			case "attendance":
				res = attendanceActions[command.Action](ctx, command.Client, command.Arg)
			}

			if res.Error != "" {
				logger.Error("failed to execute command", "error", res.Error)
			}

			command.Client.send <- res.ToJsonBytes()

			// to broadcast to all connected clients
			// for client := range h.clients {
			// 	select {
			// 	case client.send <- res:
			// 	default:
			// 		close(client.send)
			// 		delete(h.clients, client)
			// 	}
			// }
		}
	}
}
