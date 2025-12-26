package websocket

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = pongWait * 9 / 10
)

type Client struct {
	Conn   *websocket.Conn
	Send   chan []byte
	Topics map[string]bool
	Hub    *Hub
}

func NewClient(conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Topics: make(map[string]bool),
		Hub:    hub,
	}
}

func (client *Client) ReadLoop() {
	defer func() {
		client.Hub.unregister <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(5120)
	client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg Message
		if err := client.Conn.ReadJSON(&msg); err != nil {
			log.Println("read error:", err)
			break
		}
		Dispatch(client, msg)
	}
}

func (client *Client) WriteLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := client.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
