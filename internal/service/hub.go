package service

import "log/slog"

type Hub struct {
	clients map[*Client]bool

	Broadcast chan []byte

	Register chan *Client

	Unregister chan *Client

	log *slog.Logger
}

func NewHub(log *slog.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		log:        log.With(slog.String("component", "Hub")), // Привязываем метку компонента
	}
}

func (h *Hub) Run() {
	h.log.Info("Hub event loop started")
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

		case message := <-h.Broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
					h.log.Warn("client send channel full, forced disconnect")
				}
			}
		}
	}
}
