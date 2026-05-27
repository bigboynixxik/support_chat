package ws

import (
	"log/slog"
	"net/http"

	"support_chat/pkg/logger"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  256,
	WriteBufferSize: 256,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("failed to upgrade to websocket", slog.String("error", err.Error()))
		return
	}
	defer conn.Close()

	log.Info("new client connected")

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Warn("client disconnected or read error", slog.String("error", err.Error()))
			break
		}

		log.Debug("message received", slog.String("payload", string(message)))

		err = conn.WriteMessage(messageType, []byte("Echo: "+string(message)))
		if err != nil {
			log.Error("failed to write message", slog.String("error", err.Error()))
			break
		}
	}
}
