package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"survey-battle-backend-go/internal/config"
)

// LLMMessage mirrors the {role, content} chat messages passed to llm.invoke
// in service.ts.
type LLMMessage struct {
	Role    string
	Content string
}

// LLM is the minimal chat-completion surface the AI service needs, so
// generateQuestions/analyzeSurveyResponses stay provider-agnostic the same
// way llm.ts's BaseChatModel does.
type LLM interface {
	Invoke(ctx context.Context, messages []LLMMessage) (string, error)
}

// llmTemperature matches the fixed 0.7 used for both providers in llm.ts.
const llmTemperature = 0.7

// NewLLM selects a provider client the same way getLLM() does in llm.ts,
// failing fast on anything else.
func NewLLM(cfg config.AIConfig) (LLM, error) {
	switch cfg.Provider {
	case "openai":
		return &openAIClient{cfg: cfg, client: http.DefaultClient}, nil
	case "anthropic":
		return &anthropicClient{cfg: cfg, client: http.DefaultClient}, nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.Provider)
	}
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIClient struct {
	cfg    config.AIConfig
	client *http.Client
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Temperature float64         `json:"temperature"`
	Messages    []openAIMessage `json:"messages"`
}

type openAIResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *openAIClient) Invoke(ctx context.Context, messages []LLMMessage) (string, error) {
	reqMessages := make([]openAIMessage, len(messages))

	for i, m := range messages {
		reqMessages[i] = openAIMessage{Role: m.Role, Content: m.Content}
	}

	body, err := json.Marshal(openAIRequest{
		Model:       c.cfg.Model,
		Temperature: llmTemperature,
		Messages:    reqMessages,
	})
	if err != nil {
		return "", fmt.Errorf("encode OpenAI request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.openai.com/v1/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("build OpenAI request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call OpenAI: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read OpenAI response: %w", err)
	}

	var parsed openAIResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("decode OpenAI response: %w", err)
	}

	if parsed.Error != nil {
		return "", fmt.Errorf("OpenAI error: %s", parsed.Error.Message)
	}

	if resp.StatusCode != http.StatusOK || len(parsed.Choices) == 0 {
		return "", fmt.Errorf("OpenAI request failed with status %d", resp.StatusCode)
	}

	return parsed.Choices[0].Message.Content, nil
}

type anthropicClient struct {
	cfg    config.AIConfig
	client *http.Client
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	Temperature float64            `json:"temperature"`
	MaxTokens   int                `json:"max_tokens"`
	System      string             `json:"system,omitempty"`
	Messages    []anthropicMessage `json:"messages"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// anthropicMaxTokens has no TS equivalent (langchain's ChatAnthropic fills in
// a default); Anthropic's Messages API requires an explicit value.
const anthropicMaxTokens = 4096

func (c *anthropicClient) Invoke(ctx context.Context, messages []LLMMessage) (string, error) {
	var system string

	userMessages := make([]anthropicMessage, 0, len(messages))

	for _, m := range messages {
		if m.Role == "system" {
			system = m.Content
			continue
		}

		userMessages = append(userMessages, anthropicMessage{Role: m.Role, Content: m.Content})
	}

	body, err := json.Marshal(anthropicRequest{
		Model:       c.cfg.Model,
		Temperature: llmTemperature,
		MaxTokens:   anthropicMaxTokens,
		System:      system,
		Messages:    userMessages,
	})
	if err != nil {
		return "", fmt.Errorf("encode Anthropic request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.anthropic.com/v1/messages",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("build Anthropic request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call Anthropic: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read Anthropic response: %w", err)
	}

	var parsed anthropicResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("decode Anthropic response: %w", err)
	}

	if parsed.Error != nil {
		return "", fmt.Errorf("Anthropic error: %s", parsed.Error.Message)
	}

	if resp.StatusCode != http.StatusOK || len(parsed.Content) == 0 {
		return "", fmt.Errorf("Anthropic request failed with status %d", resp.StatusCode)
	}

	return parsed.Content[0].Text, nil
}
