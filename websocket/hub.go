package websocket

import "encoding/json"

type Hub struct {
	clients map[*Client]bool
	topics  map[string]map[*Client]bool

	register   chan *Client
	unregister chan *Client
}

var hub *Hub

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		topics:     make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func InitHub() {
	hub = NewHub()
	go hub.Run()
}

func GetHub() *Hub {
	if hub == nil {
		panic("ws hub not initialized")
	}
	return hub
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = true

		case c := <-h.unregister:
			delete(h.clients, c)
			for topic := range c.Topics {
				delete(h.topics[topic], c)
			}
			close(c.Send)
		}
	}
}

func (h *Hub) Subscribe(c *Client, topic string) {
	if h.topics[topic] == nil {
		h.topics[topic] = make(map[*Client]bool)
	}
	h.topics[topic][c] = true
	c.Topics[topic] = true
}

func (h *Hub) Unsubscribe(c *Client, topic string) {
	delete(h.topics[topic], c)
	delete(c.Topics, topic)
}

func (h *Hub) Broadcast(topic string, data any) {
	msg, _ := json.Marshal(Message{
		Type:  "event",
		Topic: topic,
		Data:  data,
	})

	for c := range h.topics[topic] {
		select {
		case c.Send <- msg:
		default:
			// client lambat → drop message
		}
	}
}
