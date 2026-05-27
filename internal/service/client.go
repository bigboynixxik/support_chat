package service

import (
	"encoding/json"
	"log/slog"
	"support_chat/internal/models"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	maxMessageSize = 512
	RoleClient     = 1
	RoleOperator   = 2
)

type Client struct {
	Hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	log    *slog.Logger
	UserID int64
	Role   int
}

func NewClient(hub *Hub, conn *websocket.Conn, log *slog.Logger) *Client {
	return &Client{
		Hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
		log:  log,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.log.Error("client.ReadPump unexpected websocket close", slog.String("error", err.Error()))
			} else {
				c.log.Info("client.ReadPump client disconnected normally")
			}
			break
		}

		var msgParsed models.Message
		if err := json.Unmarshal(message, &msgParsed); err != nil {
			c.log.Warn("failed to parse message, ignoring", slog.String("error", err.Error()))
			continue
		}

		c.log.Debug("message parsed successfully",
			slog.String("action", msgParsed.Action),
			slog.String("text", msgParsed.Text),
		)

		hubMsg := HubMessage{
			Client:  c,
			Payload: message,
			Message: msgParsed,
		}
		c.Hub.Broadcast <- hubMsg
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.log.Error("client.WritePump failed to get next writer", slog.String("error", err.Error()))
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				c.log.Error("client.WritePump failed to close writer", slog.String("error", err.Error()))
				return
			}
		}
	}
}
