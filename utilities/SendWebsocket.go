package utilities

import "shwetaik-expense-management-api/configs"

func SendWebSocketMessage(userID uint, message string) {
	configs.WebSocketConnections.RLock()
	defer configs.WebSocketConnections.RUnlock()
	conn, ok := configs.WebSocketConnections.M[userID]
	if ok {
		conn.Mu.Lock()
		conn.Conn.WriteJSON(map[string]string{"message": message})
		conn.Mu.Unlock()
	}
}
