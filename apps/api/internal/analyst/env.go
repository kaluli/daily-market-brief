package analyst

import (
	"os"
	"strings"
)

// NewAnalyzerFromEnv builds an investment analyst Analyzer using whichever
// LLM provider is configured — same LLM_PROVIDER / OLLAMA_* / OPENAI_* env
// vars as the API server (see cmd/server/main.go) — plus a human-readable
// provider name. Lets standalone commands (e.g. cmd/simulate) use the same
// analyst the API server uses without duplicating its provider-selection logic.
func NewAnalyzerFromEnv() (Analyzer, string) {
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
		return NewOllamaAnalyzer("", ""), "ollama"
	case "openai":
		return NewOpenAIAnalyzer("", ""), "openai"
	default:
		return NewStub(), "stub"
	}
}
