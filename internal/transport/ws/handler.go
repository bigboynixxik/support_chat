package ws

import (
	"log/slog"
	"net/http"
	"strconv"

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

	roleParam := r.URL.Query().Get("role")
	idParam := r.URL.Query().Get("id")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("failed to upgrade to websocket", slog.String("error", err.Error()))
		return
	}

	log.Info("new client connection established",
		slog.String("mock_role", roleParam),
		slog.String("mock_id", idParam),
	)

	client := service.NewClient(h.hub, conn, log)

	if roleParam == "operator" {
		client.Role = service.RoleOperator
	} else {
		client.Role = service.RoleClient
	}

	if parsedID, err := strconv.ParseInt(idParam, 10, 64); err == nil {
		client.UserID = parsedID
	}
	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
