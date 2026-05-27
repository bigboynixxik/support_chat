package service

import (
	"log/slog"
	"support_chat/internal/models"
)

type HubMessage struct {
	Client  *Client
	Payload []byte
	Message models.Message
}

type Hub struct {
	clients map[*Client]bool
	// Канал теперь принимает структуру HubMessage
	Broadcast  chan HubMessage
	Register   chan *Client
	Unregister chan *Client
	log        *slog.Logger
}

func NewHub(log *slog.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		Broadcast:  make(chan HubMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		log:        log.With(slog.String("component", "hub")),
	}
}

func (h *Hub) Run() {
	h.log.Info("hub event loop started")
	for {
		select {
		case client := <-h.Register:
			h.clients[client] = true
			h.log.Info("client registered", slog.Int("active_connections", len(h.clients)))

		case client := <-h.Unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.log.Info("client unregistered", slog.Int("active_connections", len(h.clients)))
			}

		case hubMsg := <-h.Broadcast:
			if hubMsg.Client.Role == RoleClient {
				for targetClient := range h.clients {
					if targetClient.Role == RoleOperator {
						h.sendToClient(targetClient, hubMsg.Payload)
					}
				}
			} else if hubMsg.Client.Role == RoleOperator {
				for targetClient := range h.clients {
					if targetClient.UserID == hubMsg.Message.TargetID {
						h.sendToClient(targetClient, hubMsg.Payload)
						break
					}
				}
			}
		}
	}
}

func (h *Hub) sendToClient(client *Client, payload []byte) {
	select {
	case client.send <- payload:
	default:
		close(client.send)
		delete(h.clients, client)
		h.log.Warn("client send channel full, forced disconnect")
	}
}
