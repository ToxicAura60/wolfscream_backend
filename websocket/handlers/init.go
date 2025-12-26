package websocket_handlers

import (
	"wolfscream/websocket"
)

func InitHandlers() {
	websocket.RegisterHandler("subscribe", Subscribe)
}
