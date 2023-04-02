package chatgpt

import "github.com/sashabaranov/go-openai"

type Message = openai.ChatCompletionMessage

var MessagePreset = []Message{
	{
		Role:    openai.ChatMessageRoleSystem,
		Content: "You are a helpful assistant. Your answers must use the same language as the user messages. You are allowed to use the Internet data.",
	},
}
