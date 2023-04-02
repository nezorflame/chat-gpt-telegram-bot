package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/nezorflame/chat-gpt-telegram-bot/internal/bolt"
	"github.com/nezorflame/chat-gpt-telegram-bot/internal/chatgpt"
)

func (b *bot) getUser(userID int64) (*User, error) {
	var user User

	userData, err := b.db.Get(userKey(userID))
	switch {
	case err == nil:
		// user is found in DB
		if err = json.Unmarshal(userData, &user); err != nil {
			return nil, fmt.Errorf("unable to unmarshal user data: %w", err)
		}
		b.log.Debug("Got user from DB", "user_id", userID)
	case errors.Is(err, bolt.ErrNilValue), errors.Is(err, bolt.ErrNotFound):
		// user is not in DB - create a new one
		user = newUser(userID)
	default:
		return nil, fmt.Errorf("unable to get user from DB: %w", err)
	}

	return &user, nil
}

func (b *bot) putUser(user *User) error {
	userData, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("unable to marshal user: %w", err)
	}

	if err = b.db.Put(userKey(user.ID), userData); err != nil {
		return fmt.Errorf("unable to put user to DB: %w", err)
	}

	b.log.Debug("Put user to DB", "user_id", user.ID)
	return nil
}

func (b *bot) getChatMessages(chatID int64) ([]chatgpt.Message, error) {
	var chatMessages []chatgpt.Message

	chatMessageData, err := b.db.Get(chatKey(chatID))
	switch {
	case err == nil:
		// chat is found in DB
		if err = json.Unmarshal(chatMessageData, &chatMessages); err != nil {
			return nil, fmt.Errorf("unable to unmarshal chat message data: %w", err)
		}
		b.log.Debug("Got chat from DB", "chat_id", chatID)
	case errors.Is(err, bolt.ErrNilValue), errors.Is(err, bolt.ErrNotFound):
		// chat is not in DB - create a new one
		chatMessages = newChatMessages()
	default:
		return nil, fmt.Errorf("unable to get chat messages from DB: %w", err)
	}

	return chatMessages, nil
}

func (b *bot) putChatMessages(chatID int64, chatMessages []chatgpt.Message) error {
	chatMessageData, err := json.Marshal(chatMessages)
	if err != nil {
		return fmt.Errorf("unable to marshal chat messages: %w", err)
	}

	if err = b.db.Put(chatKey(chatID), chatMessageData); err != nil {
		return fmt.Errorf("unable to put chat messages to DB: %w", err)
	}

	b.log.Debug("Put chat messages to DB", "chat_id", chatID)
	return nil
}

func userKey(userID int64) string {
	return userKeyPrefix + strconv.FormatInt(userID, 10)
}

func chatKey(chatID int64) string {
	return chatKeyPrefix + strconv.FormatInt(chatID, 10)
}
