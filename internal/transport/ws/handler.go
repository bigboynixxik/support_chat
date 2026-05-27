package ws

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"support_chat/internal/service"
	clients "support_chat/internal/transport/clients"
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
	hub        *service.Hub
	authClient *clients.Client
}

func NewHandler(hub *service.Hub, authClient *clients.Client) *Handler {
	return &Handler{
		hub:        hub,
		authClient: authClient,
	}
}

func (h *Handler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("failed to upgrade to websocket", slog.String("error", err.Error()))
		return
	}
	log.Info("connection initialized, waiting for auth frame")

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	_, msgBytes, err := conn.ReadMessage()
	if err != nil {
		log.Warn("failed to read auth message or timeout reached", slog.String("error", err.Error()))
		conn.Close()
		return
	}

	var authMsg struct {
		Action string `json:"action"`
		Token  string `json:"token"`
	}
	if err := json.Unmarshal(msgBytes, &authMsg); err != nil || authMsg.Action != "auth" || authMsg.Token == "" {
		log.Warn("invalid auth frame received")
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "invalid auth frame"}`))
		conn.Close()
		return
	}

	authResp, err := h.authClient.Validate(r.Context(), authMsg.Token)
	if err != nil {
		log.Error("grpc validation failed", slog.String("error", err.Error()))
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "internal server error"}`))
		conn.Close()
		return
	}

	if !authResp.IsValid {
		log.Warn("invalid token", slog.String("reason", authResp.ErrorMessage))
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "`+authResp.ErrorMessage+`"}`))
		conn.Close()
		return
	}

	conn.SetReadDeadline(time.Time{})

	client := service.NewClient(h.hub, conn, log)
	client.UserID = authResp.User.Id

	if authResp.User.Role.String() == "ROLE_OPERATOR" {
		client.Role = service.RoleOperator
	} else {
		client.Role = service.RoleClient
	}

	log.Info("client authenticated successfully",
		slog.Int64("user_id", client.UserID),
		slog.Int("role", client.Role),
	)

	conn.WriteMessage(websocket.TextMessage, []byte(`{"action": "system", "status": "authenticated"}`))

	client.Hub.Register <- client
	go client.WritePump()
	go client.ReadPump()
}
