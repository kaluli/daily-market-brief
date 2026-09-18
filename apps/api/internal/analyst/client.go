package analyst

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// chatCompletionsClient calls an OpenAI-compatible /v1/chat/completions endpoint.
// Both the OpenAI analyzer and the local Ollama analyzer share this: Ollama exposes
// the same request/response shape at <base>/v1/chat/completions, just with no auth
// required and a different base URL/model.
type chatCompletionsClient struct {
	url        string
	apiKey     string // optional; if empty, no Authorization header is sent
	model      string
	httpClient *http.Client
}

func newChatCompletionsClient(url, apiKey, model string, timeout time.Duration) *chatCompletionsClient {
	return &chatCompletionsClient{
		url:        url,
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: timeout},
	}
}

// complete sends one chat-completions request in strict JSON mode and returns
// the extracted JSON content of the model's reply (markdown code fences, if
// any, are stripped). Used by the investment analyst (AnalysisResult schema).
func (c *chatCompletionsClient) complete(ctx context.Context, systemContent, userContent string) (string, error) {
	content, err := c.doComplete(ctx, systemContent, userContent, true)
	if err != nil {
		return "", err
	}
	return extractJSON(content), nil
}

// completeFreeform is like complete but without forcing JSON output — used for
// open-ended text answers (e.g. the agents feedback coach) rather than the
// strict AnalysisResult schema.
func (c *chatCompletionsClient) completeFreeform(ctx context.Context, systemContent, userContent string) (string, error) {
	return c.doComplete(ctx, systemContent, userContent, false)
}

func (c *chatCompletionsClient) doComplete(ctx context.Context, systemContent, userContent string, jsonMode bool) (string, error) {
	reqBody := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemContent},
			{"role": "user", "content": userContent},
		},
		"temperature": float64(0.2),
	}
	if jsonMode {
		reqBody["response_format"] = map[string]string{"type": "json_object"}
	}
	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var errBody struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		if errBody.Error.Message != "" {
			return "", fmt.Errorf("llm api %d: %s", resp.StatusCode, errBody.Error.Message)
		}
		return "", fmt.Errorf("llm api %d", resp.StatusCode)
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", err
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("llm: no choices in response")
	}
	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}

// buildUserContent formats the news item the same way for every LLM backend.
func buildUserContent(input NewsInput) string {
	userContent := fmt.Sprintf("Analyze this news item and return only the JSON object (no markdown, no explanation):\n\nTitle: %s\nURL: %s\nSource: %s\n", input.Title, input.URL, input.Source)
	if input.Summary != "" {
		userContent += fmt.Sprintf("Summary/Description: %s\n", input.Summary)
	}
	return userContent
}

// extractJSON pulls the first JSON object from content (handles ```json ... ``` wrappers).
func extractJSON(content string) string {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		re := regexp.MustCompile(`(?s)\x60\x60\x60(?:json)?\s*([\s\S]*?)\x60\x60\x60`)
		if m := re.FindStringSubmatch(content); len(m) > 1 {
			return strings.TrimSpace(m[1])
		}
	}
	return content
}

// ChatCompleteFunc answers a free-form system+user prompt with the model's
// text reply (no JSON schema enforced).
type ChatCompleteFunc func(ctx context.Context, systemPrompt, userPrompt string) (string, error)

// NewChatCompleterFromEnv returns a free-form chat function using whichever
// LLM provider is configured — same LLM_PROVIDER / OLLAMA_* / OPENAI_* env
// vars as the investment analyst — plus a human-readable provider name.
// Used by things like the agents feedback coach, which needs an open-ended
// answer rather than the strict AnalysisResult JSON schema. Falls back to a
// stub reply if no provider is configured.
func NewChatCompleterFromEnv() (ChatCompleteFunc, string) {
	provider := strings.ToLower(os.Getenv("LLM_PROVIDER"))
	if provider == "" {
		switch {
		case os.Getenv("OLLAMA_BASE_URL") != "":
			provider = "ollama"
		case os.Getenv("OPENAI_API_KEY") != "":
			provider = "openai"
		default:
			provider = "stub"
		}
	}
	switch provider {
	case "ollama":
		c := newOllamaClient("", "")
		return c.completeFreeform, "ollama"
	case "openai":
		c := newOpenAIClient("", "")
		return c.completeFreeform, "openai"
	default:
		return stubChatComplete, "stub"
	}
}

func stubChatComplete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return "No hay un modelo de LLM configurado (LLM_PROVIDER / OPENAI_API_KEY / OLLAMA_BASE_URL). Configuralo en .env para recibir feedback real.", nil
}
