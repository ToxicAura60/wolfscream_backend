package websocket_handlers

import (
	"encoding/json"
	"log"
	"wolfscream/websocket"
)

type SubscribePayload struct {
	Topic string `json:"topic"`
}

func Subscribe(c *websocket.Client, payload json.RawMessage) {
	log.Println("SUBSCRIBE HANDLER CALLED")

	
	var data SubscribePayload
	if err := json.Unmarshal(payload, &data); err != nil {
		return
	}

	c.Hub.Subscribe(c, data.Topic)

	// initial snapshot (contoh)
	msg, _ := json.Marshal(websocket.Message{
		Type:  "event",
		Topic: data.Topic,
		Data: map[string]any{
			"status": "subscribed",
		},
	})

	c.Send <- msg
}
