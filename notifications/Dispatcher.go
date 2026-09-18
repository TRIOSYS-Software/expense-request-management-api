package notifications

import (
	"fmt"
	"log"
	"time"

	"shwetaik-expense-management-api/utilities"
)

type Kind struct {
	idKey string
}

var (
	Expense = Kind{idKey: "expenseId"}
	Advance = Kind{idKey: "advanceId"}
)

// Notification is one message to deliver.
type Notification struct {
	UserID    uint
	RequestID uint
	Message   string
	Type      string
	PushTitle string
	Kind      Kind
}

func (n Notification) pushData() map[string]string {
	return map[string]string{
		n.Kind.idKey: fmt.Sprintf("%d", n.RequestID),
		"type":       n.Type,
	}
}

func (n Notification) socketPayload(id uint, createdAt time.Time) utilities.WebSocketMessagePayload {
	return utilities.WebSocketMessagePayload{
		ID:        id,
		Message:   n.Message,
		Type:      n.Type,
		ExpenseID: n.RequestID,
		IsRead:    false,
		CreatedAt: createdAt.Format(time.RFC3339),
	}
}

// TokenSource returns a user's registered device tokens.
type TokenSource interface {
	GetTokensByUserID(userID uint) ([]string, error)
}

// Pusher delivers a push notification to a set of device tokens.
type Pusher interface {
	SendPushNotification(tokens []string, title string, body string, data map[string]string)
}

// Dispatcher fans a notification out to a user's devices and open socket.
type Dispatcher struct {
	tokens TokenSource
	pusher Pusher
}

func NewDispatcher(tokens TokenSource, pusher Pusher) *Dispatcher {
	return &Dispatcher{tokens: tokens, pusher: pusher}
}

// FanOut delivers inline. Used by callers that are already running on their own
// goroutine after the transaction has committed.
func (d *Dispatcher) FanOut(n Notification, notificationID uint, createdAt time.Time) {
	tokens, err := d.tokens.GetTokensByUserID(n.UserID)
	if err != nil {
		log.Printf("Error fetching device tokens for user %d: %v", n.UserID, err)
	} else if len(tokens) > 0 {
		d.pusher.SendPushNotification(tokens, n.PushTitle, n.Message, n.pushData())
	}

	utilities.SendWebSocketMessage(n.UserID, n.socketPayload(notificationID, createdAt))
}

// FanOutAsync delivers on its own goroutines. Used from inside a transaction,
// so a slow push or socket write cannot hold the transaction open.
//
// The token lookup still happens synchronously — it is a local query, and doing
// it here keeps the "no tokens, no push" decision on the caller's goroutine.
func (d *Dispatcher) FanOutAsync(n Notification, notificationID uint, createdAt time.Time) {
	if tokens, err := d.tokens.GetTokensByUserID(n.UserID); err == nil && len(tokens) > 0 {
		go d.pusher.SendPushNotification(tokens, n.PushTitle, n.Message, n.pushData())
	}

	go utilities.SendWebSocketMessage(n.UserID, n.socketPayload(notificationID, createdAt))
}

// BroadcastRemovals tells each affected user's client to drop the actionable
// notifications that an approval decision just cleared.
//
// Callers must invoke this only *after* their transaction commits, so a
// removal is never announced for a change that was rolled back.
func BroadcastRemovals(requestID uint, perUser map[uint][]uint) {
	for userID, ids := range perUser {
		go utilities.SendWebSocketRemoval(userID, requestID, ids)
	}
}
