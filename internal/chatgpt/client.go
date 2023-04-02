package chatgpt

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"golang.org/x/exp/slog"
)

const (
	chatModelVersion = openai.GPT3Dot5Turbo
	textModelVersion = openai.GPT3Curie
)

type client struct {
	api *openai.Client
	log *slog.Logger
}

func New(token, orgID string, timeout time.Duration, log *slog.Logger) (*client, error) {
	if token == "" {
		return nil, errors.New("empty token")
	}
	if token == "" {
		return nil, errors.New("empty token")
	}

	openAICfg := openai.DefaultConfig(token)
	openAICfg.OrgID = orgID
	openAICfg.HTTPClient.Timeout = timeout
	api := openai.NewClientWithConfig(openAICfg)

	return &client{api: api, log: log}, nil
}

func (c *client) GetChatCompletion(ctx context.Context, msg string, chatHistory ...Message) (string, []Message, error) {
	log := c.log.With("message", msg)
	log.DebugCtx(ctx, "Got new prompt")

	// make chat messages
	chatMessages := append(chatHistory, Message{
		Role:    openai.ChatMessageRoleUser,
		Content: msg,
	})

	// make the request
	req := openai.ChatCompletionRequest{
		Model:    chatModelVersion,
		Messages: chatMessages,
	}
	resp, err := c.api.CreateChatCompletion(ctx, req)
	apiErr := &openai.APIError{}
	switch {
	case errors.As(err, &apiErr):
		if apiErr.Type == "server_error" {
			return c.GetCompletion(ctx, msg, chatHistory...)
		}
	case err != nil:
		return "", chatMessages, fmt.Errorf("unable to get response from OpenAI: %w", err)
	case len(resp.Choices) == 0:
		return "", chatMessages, errors.New("unable to get response from OpenAI: not enough choices")
	}

	// parse the answers
	var results []string
	for _, choice := range resp.Choices {
		chatMessages = append(chatMessages, Message{
			Role:    openai.ChatMessageRoleAssistant,
			Content: choice.Message.Content,
		})
		log.DebugCtx(ctx, "Got the model response", "message", choice.Message.Content)
		results = append(results, choice.Message.Content)
	}
	return strings.Join(results, "\n"), chatMessages, nil
}

func (c *client) GetCompletion(ctx context.Context, msg string, chatHistory ...Message) (string, []Message, error) {
	log := c.log.With("message", msg)
	log.DebugCtx(ctx, "Got new prompt")

	// make chat messages
	chatMessages := append(chatHistory, Message{
		Role:    openai.ChatMessageRoleUser,
		Content: msg,
	})

	// make the request
	req := openai.CompletionRequest{
		Model:  textModelVersion,
		Prompt: msg,
	}
	resp, err := c.api.CreateCompletion(ctx, req)
	if err != nil {
		return "", chatMessages, fmt.Errorf("unable to get response from OpenAI: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", chatMessages, errors.New("unable to get response from OpenAI: not enough choices")
	}

	// parse the answers
	var results []string
	for _, choice := range resp.Choices {
		chatMessages = append(chatMessages, Message{
			Role:    openai.ChatMessageRoleAssistant,
			Content: choice.Text,
		})
		log.DebugCtx(ctx, "Got the model response", "message", choice.Text)
		results = append(results, choice.Text)
	}
	return strings.Join(results, "\n"), chatMessages, nil
}
