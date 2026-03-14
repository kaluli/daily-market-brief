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

const openAIURL = "https://api.openai.com/v1/chat/completions"

// OpenAIAnalyzer calls OpenAI Chat Completions with the investment analyst prompt and parses JSON.
type OpenAIAnalyzer struct {
	apiKey string
	model  string
	client *http.Client
}

// NewOpenAIAnalyzer returns an Analyzer that uses OpenAI.
// If apiKey is empty, uses OPENAI_API_KEY. If model is empty, uses OPENAI_MODEL or gpt-4o-mini.
func NewOpenAIAnalyzer(apiKey, model string) *OpenAIAnalyzer {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if model == "" {
		model = os.Getenv("OPENAI_MODEL")
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIAnalyzer{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 90 * time.Second},
	}
}

// Analyze sends the news item to OpenAI and parses the response into AnalysisResult.
func (o *OpenAIAnalyzer) Analyze(ctx context.Context, input NewsInput) (*AnalysisResult, error) {
	if o.apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}
	userContent := fmt.Sprintf("Analyze this news item and return only the JSON object (no markdown, no explanation):\n\nTitle: %s\nURL: %s\nSource: %s\n", input.Title, input.URL, input.Source)
	if input.Summary != "" {
		userContent += fmt.Sprintf("Summary/Description: %s\n", input.Summary)
	}
	systemContent := SystemPrompt + "\n\n" + AnalysisInstructions

	reqBody := map[string]interface{}{
		"model": o.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemContent},
			{"role": "user", "content": userContent},
		},
		"response_format": map[string]string{"type": "json_object"},
		"temperature":     float64(0.2),
	}
	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
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
			return nil, fmt.Errorf("openai api %d: %s", resp.StatusCode, errBody.Error.Message)
		}
		return nil, fmt.Errorf("openai api %d", resp.StatusCode)
	}

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, err
	}
	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("openai: no choices in response")
	}
	content := strings.TrimSpace(openAIResp.Choices[0].Message.Content)
	content = extractJSON(content)
	var result AnalysisResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("openai: parse json: %w", err)
	}
	return &result, nil
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
