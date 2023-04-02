package bot

import (
	"time"
)

const (
	userKeyPrefix = "user:"

	defaultChatTimeout  = time.Hour
	defaultChatGPTLimit = 1000
)

type User struct {
	ID            int64 `json:"id"`
	ChatID        int64 `json:"chat_id"`
	ChatGPTLimit  int   `json:"chatgpt_limit"`
	MessageAmount int   `json:"message_amount"`
	LastMessageTS int64 `json:"last_message_ts"`
}

func newUser(id int64) User {
	return User{
		ID:            id,
		ChatGPTLimit:  defaultChatGPTLimit,
		MessageAmount: 0,
	}
}

// IsChatStale reports if the chat should be considered as stale for the user.
// It returns true if the last message was sent more than an hour ago
// or if the user has never sent a message.
func (u User) IsChatStale() bool {
	lastAcceptableTS := time.Now().Add(-defaultChatTimeout).Unix()
	return u.LastMessageTS < lastAcceptableTS
}

func (u User) IsChatGPTLimitReached() bool {
	if u.ChatGPTLimit == -1 {
		return false
	}
	return u.ChatGPTLimit <= u.MessageAmount
}
