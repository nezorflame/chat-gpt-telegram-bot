package bot

import (
	"context"

	"github.com/nezorflame/chat-gpt-telegram-bot/internal/chatgpt"
)

type ChatGPT interface {
	GetChatCompletion(ctx context.Context, msg string, chatHistory ...chatgpt.Message) (string, []chatgpt.Message, error)
}
