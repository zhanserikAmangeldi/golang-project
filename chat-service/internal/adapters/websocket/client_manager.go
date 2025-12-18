package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

type ClientManager struct {
	clients map[int64]*websocket.Conn
	lock    sync.RWMutex
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[int64]*websocket.Conn),
	}
}

func (manager *ClientManager) AddClient(userID int64, conn *websocket.Conn) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	manager.clients[userID] = conn
}

func (manager *ClientManager) RemoveClient(userID int64) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	if conn, ok := manager.clients[userID]; ok {
		conn.Close()
		delete(manager.clients, userID)
	}
}

func (manager *ClientManager) GetClient(userID int64) (*websocket.Conn, bool) {
	manager.lock.RLock()
	defer manager.lock.RUnlock()
	conn, ok := manager.clients[userID]
	return conn, ok
}
