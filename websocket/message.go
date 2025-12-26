package websocket

import "encoding/json"

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Topic   string          `json:"topic,omitempty"`
	Data    any             `json:"data,omitempty"`
}
