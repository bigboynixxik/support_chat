package ws

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"

	"support_chat/internal/service"
	"support_chat/pkg/logger"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	hub *service.Hub
}

func NewHandler(hub *service.Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

func (h *Handler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("failed to upgrade to websocket", slog.String("error", err.Error()))
		return
	}

	log.Info("new client connection established")

	client := service.NewClient(h.hub, conn, log)

	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
