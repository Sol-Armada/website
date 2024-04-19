package main

type CommandRequest struct {
	Thing  string
	Action string
	Arg    any
	Client *Client
}

type CommandResponse struct {
	Thing  string `json:"thing"`
	Result any    `json:"result"`
	Error  string `json:"error"`
}

type Action func(c *Client, arg any)

type Hub struct {
	// registered clients
	clients map[*Client]bool

	// inbound commands from the clients
	broadcast chan *CommandRequest

	// register requests from the clients
	register chan *Client

	// unregister requests from clients
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan *CommandRequest),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
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

			switch command.Thing {
			case "members":
				membersActions[command.Action](command.Client, command.Arg)
			case "login":
				loginActions[command.Action](command.Client, command.Arg)
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
