package utilities

import (
	"log"
	"shwetaik-expense-management-api/configs"
)

type WebSocketMessagePayload struct {
	ID        uint   `json:"id"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	ExpenseID uint   `json:"expenseId"`
	IsRead    bool   `json:"isRead"`
	CreatedAt string `json:"createdAt"`
}

func SendWebSocketMessage(
	userID uint,
	payload WebSocketMessagePayload,
) {
	configs.WebSocketConnections.RLock()
	defer configs.WebSocketConnections.RUnlock()

	conn, ok := configs.WebSocketConnections.M[userID]
	if ok {
		conn.Mu.Lock()
		err := conn.Conn.WriteJSON(payload)
		conn.Mu.Unlock()
		if err != nil {
			log.Printf("Error sending WebSocket message to user %d: %v", userID, err)
			configs.WebSocketConnections.Lock()
			delete(configs.WebSocketConnections.M, userID)
			configs.WebSocketConnections.Unlock()
		}
	} else {
		log.Printf("No active WebSocket connection for user %d.", userID)
	}
}
