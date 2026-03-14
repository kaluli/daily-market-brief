package analyst

// AnalysisResult is the structured output of the AI investment analyst (STEP 10).
// Used for JSON response and optional persistence.
type AnalysisResult struct {
	Relevance               string            `json:"relevance"`                 // Market Moving | Potentially Relevant | Noise
	Category                string            `json:"category"`                  // Primary category from STEP 2
	Summary                 string            `json:"summary"`                   // Short summary (max 2 sentences)
	WhyItMatters            string            `json:"why_it_matters"`            // Why the news matters for markets
	ImpactLevel             string            `json:"impact_level"`              // High | Medium | Low
	ExpectedMarketReaction  string            `json:"expected_market_reaction"`  // Risk On | Risk Off | Sector Rotation | Neutral
	AffectedAssets          []string          `json:"affected_assets"`          // Stocks, sectors, indices, commodities, currencies, bonds
	DirectionalBias         map[string]string `json:"directional_bias"`         // asset -> Bullish | Bearish | Neutral
	InvestmentSignals       []string          `json:"investment_signals"`         // Long/short/pair/rotation/macro suggestions
	TimeHorizon             string            `json:"time_horizon"`              // Intraday | Short Term | Medium Term | Long Term
	SignalStrength          string            `json:"signal_strength"`            // 1-10 score as string
}

// Valid categories (STEP 2).
const (
	CategoryMonetaryPolicy     = "Monetary Policy & Central Banks"
	CategoryMacroData          = "Macroeconomic Data"
	CategoryCorporateEarnings  = "Corporate Earnings & Company News"
	CategoryGeopolitics        = "Geopolitics"
	CategoryCommoditiesEnergy = "Commodities & Energy"
	CategoryTechInnovation     = "Technology & Innovation"
	CategoryRegulationFiscal  = "Regulation & Fiscal Policy"
	CategoryMarketFlows       = "Market Flows & Positioning"
	CategoryMarketSentiment   = "Market Sentiment"
	CategoryIndustrySpecific  = "Industry-Specific Developments"
)

// Valid relevance values (STEP 1).
const (
	RelevanceMarketMoving   = "Market Moving"
	RelevancePotentiallyRel = "Potentially Relevant"
	RelevanceNoise          = "Noise"
)

// Valid impact levels (STEP 4).
const (
	ImpactHigh   = "High"
	ImpactMedium = "Medium"
	ImpactLow    = "Low"
)

// Valid market reactions (STEP 4).
const (
	ReactionRiskOn         = "Risk On"
	ReactionRiskOff        = "Risk Off"
	ReactionSectorRotation = "Sector Rotation"
	ReactionNeutral        = "Neutral"
)

// Valid time horizons (STEP 8).
const (
	HorizonIntraday   = "Intraday"
	HorizonShortTerm  = "Short Term (days to weeks)"
	HorizonMediumTerm = "Medium Term (weeks to months)"
	HorizonLongTerm   = "Long Term (structural)"
)
