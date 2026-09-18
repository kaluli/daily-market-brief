package analyst

import (
	"encoding/json"
	"strings"
)

// UnmarshalJSON makes AnalysisResult tolerant of LLMs (especially small local
// models) that don't always follow the exact schema — e.g. returning
// investment_signals or directional_bias values as JSON objects instead of
// plain strings, or signal_strength as a number instead of a string.
// Whatever shape arrives for those fields, each value is converted to a
// readable string instead of failing to parse.
func (r *AnalysisResult) UnmarshalJSON(data []byte) error {
	var alias struct {
		Relevance               string                     `json:"relevance"`
		Category                string                     `json:"category"`
		Summary                 string                     `json:"summary"`
		WhyItMatters             string                     `json:"why_it_matters"`
		ImpactLevel              string                     `json:"impact_level"`
		ExpectedMarketReaction   string                     `json:"expected_market_reaction"`
		AffectedAssets           []json.RawMessage          `json:"affected_assets"`
		DirectionalBias          map[string]json.RawMessage `json:"directional_bias"`
		InvestmentSignals        []json.RawMessage          `json:"investment_signals"`
		TimeHorizon              string                     `json:"time_horizon"`
		SignalStrength           json.RawMessage            `json:"signal_strength"`
	}
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}

	r.Relevance = alias.Relevance
	r.Category = alias.Category
	r.Summary = alias.Summary
	r.WhyItMatters = alias.WhyItMatters
	r.ImpactLevel = alias.ImpactLevel
	r.ExpectedMarketReaction = alias.ExpectedMarketReaction
	r.TimeHorizon = alias.TimeHorizon
	r.SignalStrength = rawToString(alias.SignalStrength)

	r.AffectedAssets = make([]string, 0, len(alias.AffectedAssets))
	for _, raw := range alias.AffectedAssets {
		if s := rawToString(raw); s != "" {
			r.AffectedAssets = append(r.AffectedAssets, s)
		}
	}

	r.InvestmentSignals = make([]string, 0, len(alias.InvestmentSignals))
	for _, raw := range alias.InvestmentSignals {
		if s := rawToString(raw); s != "" {
			r.InvestmentSignals = append(r.InvestmentSignals, s)
		}
	}

	r.DirectionalBias = make(map[string]string, len(alias.DirectionalBias))
	for k, raw := range alias.DirectionalBias {
		r.DirectionalBias[k] = rawToString(raw)
	}

	return nil
}

// rawToString converts a JSON value (string, number, bool, object, array, ...)
// into a plain, readable string. Plain JSON strings are unwrapped (no quotes);
// anything else (an object the model used instead of a string, a number, etc.)
// falls back to its compact JSON form so the information isn't lost.
func rawToString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return strings.TrimSpace(string(raw))
}
