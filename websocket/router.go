package websocket

import "encoding/json"

type HandlerFunc func(c *Client, payload json.RawMessage)

var handlers = map[string]HandlerFunc{}

func RegisterHandler(event string, handler HandlerFunc) {
	handlers[event] = handler
}

func Dispatch(c *Client, msg Message) {
	if handler, ok := handlers[msg.Type]; ok {
		handler(c, msg.Payload)
	}
}
