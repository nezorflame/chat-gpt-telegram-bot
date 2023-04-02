package bot

import (
	"context"
	"encoding/json"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nezorflame/chat-gpt-telegram-bot/internal/chatgpt"
)

const chatKeyPrefix = "chat:"

func (b *bot) newChat(msg *tgbotapi.Message) {
	chatID := strconv.FormatInt(msg.Chat.ID, 10)
	log := b.log.With("chat_id", chatID, "user_id", msg.From.ID)
	log.Debug("Creating a new chat")

	// marshal and save the chat preset
	chatMessageData, _ := json.Marshal(chatgpt.MessagePreset)
	chatIDKey := chatKeyPrefix + strconv.FormatInt(msg.Chat.ID, 10)
	if err := b.db.Put(chatIDKey, chatMessageData); err != nil {
		log.Error("Unable to save new chat", "error", err)
		b.reply(msg.Chat.ID, 0, b.cfg.GetString("messages.chatgpt.new_chat_error"))
		return
	}

	// inform the user
	log.Debug("Saved the new chat")
	b.reply(msg.Chat.ID, msg.MessageID, b.cfg.GetString("messages.chatgpt.new_chat_created"))
}

func (b *bot) parseChatMessage(ctx context.Context, msg *tgbotapi.Message) {
	log := b.log.With("chat_id", strconv.FormatInt(msg.Chat.ID, 10), "user_id", msg.From.ID)
	log.DebugCtx(ctx, "Parsing new chat message", "message", msg.Text)

	// get user from DB
	userID := msg.From.ID
	if !msg.Chat.IsPrivate() {
		userID = msg.Chat.ID
	}
	user, err := b.getUser(userID)
	if err != nil {
		log.ErrorCtx(ctx, "Unable to get user from DB", "error", err)
		b.reply(msg.Chat.ID, 0, b.cfg.GetString("messages.chatgpt.error"))
		return
	}
	if user.ChatID == 0 {
		user.ChatID = msg.Chat.ID
	}

	// check monthly limit
	if user.IsChatGPTLimitReached() {
		log.WarnCtx(ctx, "User has reached the monthly limit")
		b.reply(msg.Chat.ID, 0, b.cfg.GetString("messages.chatgpt.limit_reached"))
		return
	}

	// get the chat history from DB
	chatMessages, err := b.getChatMessages(msg.Chat.ID)
	if err != nil {
		log.ErrorCtx(ctx, "Unable to get chat history from DB", "error", err)
		b.reply(msg.Chat.ID, 0, b.cfg.GetString("messages.chatgpt.error"))
		return
	}
	if user.IsChatStale() {
		log.WarnCtx(ctx, "Chat is new or stale - making a new one")
		chatMessages = newChatMessages()
	}
	user.LastMessageTS = int64(msg.Date)

	// make user aware of accepted prompt
	b.reply(msg.Chat.ID, msg.MessageID, b.cfg.GetString("messages.chatgpt.sent"))

	// get ChatGPT response
	response, chatMessages, err := b.chatGPT.GetChatCompletion(ctx, msg.Text, chatMessages...)
	if err != nil {
		log.ErrorCtx(ctx, "Unable to get response from OpenAI", "prompt", msg.Text, "error", err)
		b.reply(msg.Chat.ID, 0, b.cfg.GetString("messages.chatgpt.error"))
		return
	}
	// if it succeeded - increase user's message amount
	user.MessageAmount++

	// send the result to user
	b.reply(msg.Chat.ID, 0, response)

	// save the chat to DB
	if err = b.putChatMessages(msg.Chat.ID, chatMessages); err != nil {
		log.ErrorCtx(ctx, "Unable to save chat messages to DB", "error", err)
	}
	log.DebugCtx(ctx, "Saved the chat to DB")

	// save the user to DB
	if err = b.putUser(user); err != nil {
		log.ErrorCtx(ctx, "Unable to save user to DB", "error", err)
	}
	log.DebugCtx(ctx, "Saved the user to DB")
}

func newChatMessages() []chatgpt.Message {
	return chatgpt.MessagePreset
}
