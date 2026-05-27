package service

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	maxMessageSize = 512
)

type Client struct {
	Hub  *Hub
	conn *websocket.Conn
	send chan []byte
	log  *slog.Logger
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
				c.log.Error("unexpected websocket close", slog.String("error", err.Error()))
			} else {
				c.log.Info("client disconnected normally")
			}
			break
		}
		c.log.Debug("message received", slog.String("payload", string(message)))
		c.Hub.Broadcast <- message
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
				c.log.Error("failed to get next writer", slog.String("error", err.Error()))
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				c.log.Error("failed to close writer", slog.String("error", err.Error()))
				return
			}
		}
	}
}
