package configs

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var Upgrader websocket.Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketConnection struct {
	Conn *websocket.Conn
	Mu   sync.Mutex
}

var WebSocketConnections = struct {
	sync.RWMutex
	M map[uint]*WebSocketConnection
}{M: make(map[uint]*WebSocketConnection)}
