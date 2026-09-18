package analyst

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	defaultOllamaBaseURL = "http://localhost:11434"
	defaultOllamaModel   = "llama3.2:3b"
)

// newOllamaClient builds the shared HTTP client for a local/LAN Ollama
// instance. If baseURL is empty, uses OLLAMA_BASE_URL or http://localhost:11434.
// If model is empty, uses OLLAMA_MODEL or llama3.2:3b.
func newOllamaClient(baseURL, model string) *chatCompletionsClient {
	if baseURL == "" {
		baseURL = os.Getenv("OLLAMA_BASE_URL")
	}
	if baseURL == "" {
		baseURL = defaultOllamaBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if model == "" {
		model = os.Getenv("OLLAMA_MODEL")
	}
	if model == "" {
		model = defaultOllamaModel
	}
	// Ollama has no auth by default. OLLAMA_API_KEY is only needed if you've put
	// it behind a reverse proxy that requires one.
	apiKey := os.Getenv("OLLAMA_API_KEY")
	return newChatCompletionsClient(baseURL+"/v1/chat/completions", apiKey, model, 120*time.Second)
}

// OllamaAnalyzer calls a local (or LAN) Ollama server's OpenAI-compatible chat
// completions endpoint instead of a paid API — no API key, no per-token cost.
//
// Ollama binds to 127.0.0.1 only by default, so if the model runs on a different
// machine than the API server, Ollama there needs OLLAMA_HOST=0.0.0.0:11434 set
// before it starts (see docs/LOCAL_OLLAMA.md).
type OllamaAnalyzer struct {
	client *chatCompletionsClient
}

// NewOllamaAnalyzer returns an Analyzer backed by a local/LAN Ollama instance.
func NewOllamaAnalyzer(baseURL, model string) *OllamaAnalyzer {
	return &OllamaAnalyzer{client: newOllamaClient(baseURL, model)}
}

// Analyze sends the news item to the local Ollama model and parses the response.
func (o *OllamaAnalyzer) Analyze(ctx context.Context, input NewsInput) (*AnalysisResult, error) {
	systemContent := SystemPrompt + "\n\n" + AnalysisInstructions
	userContent := buildUserContent(input)
	content, err := o.client.complete(ctx, systemContent, userContent)
	if err != nil {
		return nil, fmt.Errorf("ollama: %w (is `ollama serve` running and reachable at %s?)", err, o.client.url)
	}
	var result AnalysisResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("ollama: parse json: %w (small local models sometimes reply with malformed JSON; raw content: %.200s)", err, content)
	}
	return &result, nil
}
