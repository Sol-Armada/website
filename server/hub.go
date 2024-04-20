package main

import (
	"context"
	"encoding/json"
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

type Action func(ctx context.Context, c *Client, arg any)

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
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case command := <-h.broadcast:
			var access userAccess
			if command.Token != "" && command.Token != "null" {
				uAccessRaw, err := decrypt(command.Token)
				if err != nil {
					command.Client.send <- []byte(err.Error())
					return
				}
				if err := json.Unmarshal([]byte(uAccessRaw), &access); err != nil {
					command.Client.send <- []byte(err.Error())
					return
				}
			}

			ctx := context.WithValue(context.Background(), contextKeyAccess, access)

			switch command.Thing {
			case "members":
				membersActions[command.Action](ctx, command.Client, command.Arg)
			case "login":
				loginActions[command.Action](ctx, command.Client, command.Arg)
			}

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
