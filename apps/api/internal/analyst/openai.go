package analyst

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const openAIURL = "https://api.openai.com/v1/chat/completions"

// newOpenAIClient builds the shared HTTP client for OpenAI. If apiKey is
// empty, uses OPENAI_API_KEY. If model is empty, uses OPENAI_MODEL or gpt-4o-mini.
func newOpenAIClient(apiKey, model string) *chatCompletionsClient {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if model == "" {
		model = os.Getenv("OPENAI_MODEL")
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return newChatCompletionsClient(openAIURL, apiKey, model, 90*time.Second)
}

// OpenAIAnalyzer calls OpenAI Chat Completions with the investment analyst prompt and parses JSON.
type OpenAIAnalyzer struct {
	apiKey string
	client *chatCompletionsClient
}

// NewOpenAIAnalyzer returns an Analyzer that uses OpenAI.
// If apiKey is empty, uses OPENAI_API_KEY. If model is empty, uses OPENAI_MODEL or gpt-4o-mini.
func NewOpenAIAnalyzer(apiKey, model string) *OpenAIAnalyzer {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	return &OpenAIAnalyzer{
		apiKey: apiKey,
		client: newOpenAIClient(apiKey, model),
	}
}

// Analyze sends the news item to OpenAI and parses the response into AnalysisResult.
func (o *OpenAIAnalyzer) Analyze(ctx context.Context, input NewsInput) (*AnalysisResult, error) {
	if o.apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}
	systemContent := SystemPrompt + "\n\n" + AnalysisInstructions
	userContent := buildUserContent(input)
	content, err := o.client.complete(ctx, systemContent, userContent)
	if err != nil {
		return nil, fmt.Errorf("openai: %w", err)
	}
	var result AnalysisResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("openai: parse json: %w", err)
	}
	return &result, nil
}
