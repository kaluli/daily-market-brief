package analyst

import (
	"context"
)

// NewsInput is the minimal input for analyzing one news item.
type NewsInput struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Source  string `json:"source"`
	Summary string `json:"summary,omitempty"` // optional body/description for richer analysis
}

// Analyzer analyzes a single news item and returns structured investment insight.
// Implementations can be stub (placeholder), LLM-based (OpenAI/Anthropic), etc.
type Analyzer interface {
	Analyze(ctx context.Context, input NewsInput) (*AnalysisResult, error)
}

// StubAnalyzer returns a valid schema with placeholder values when no LLM is configured.
// Replace with an LLM client (e.g. OpenAI) that uses SystemPrompt + AnalysisInstructions
// and parses the model output into AnalysisResult.
type StubAnalyzer struct{}

func NewStub() *StubAnalyzer {
	return &StubAnalyzer{}
}

func (s *StubAnalyzer) Analyze(ctx context.Context, input NewsInput) (*AnalysisResult, error) {
	return &AnalysisResult{
		Relevance:              RelevancePotentiallyRel,
		Category:               CategoryMarketSentiment,
		Summary:                input.Title,
		WhyItMatters:           "Stub: no LLM configured. Set up an LLM analyzer to get real analysis.",
		ImpactLevel:            ImpactLow,
		ExpectedMarketReaction: ReactionNeutral,
		AffectedAssets:         []string{},
		DirectionalBias:        map[string]string{},
		InvestmentSignals:      []string{},
		TimeHorizon:            HorizonShortTerm,
		SignalStrength:         "1",
	}, nil
}
