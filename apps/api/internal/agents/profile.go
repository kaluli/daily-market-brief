package agents

import "time"

// SimulationStart is when the fictitious monthly funding began for this
// simulation (real project start, not each portfolio row's DB creation
// time). Used to replay historical cash balances — see cmd/rewind-agents
// and BuildPortfolioViewRange's cash-after-trade calculation.
var SimulationStart = time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

// RiskProfile parametrizes how a trading agent turns a news analysis into
// buy/sell decisions: how confident the signal must be, how big a bet to
// place, and how many positions to hold at once.
type RiskProfile struct {
	Name                  string // matches portfolios.risk_profile in the DB
	Label                 string
	MinSignalStrength     int     // 1-10; ignore analyses below this (default for new portfolios — live value lives in the DB, see Portfolio.MinSignalStrength)
	MaxPositionPct        float64 // fraction of available cash to put into one new buy
	MaxOpenPositions      int     // cap on distinct tickers held at once
	MonthlyAllowanceCents int64   // fictitious cash credited once per calendar month
	MaxConcentrationPct   float64 // cap on one ticker's value as a fraction of total equity
	MaxDrawdownPct        float64 // pause new buys (sells still allowed) once equity falls this far below total cash funded so far
}

// Risky: acts on lower-confidence signals, bets a bigger slice of cash per
// trade, and holds more positions at once.
var Risky = RiskProfile{
	Name:                  "risky",
	Label:                 "Agresivo",
	MinSignalStrength:     5,
	MaxPositionPct:        0.30,
	MaxOpenPositions:      8,
	MonthlyAllowanceCents: 500000, // $5,000.00
	MaxConcentrationPct:   0.40,
	MaxDrawdownPct:        0.35,
}

// Conservative: only acts on high-confidence signals, bets a smaller slice
// of cash per trade, and holds fewer, more concentrated positions.
var Conservative = RiskProfile{
	Name:                  "conservative",
	Label:                 "Conservador",
	MinSignalStrength:     7,
	MaxPositionPct:        0.12,
	MaxOpenPositions:      4,
	MonthlyAllowanceCents: 500000, // $5,000.00
	MaxConcentrationPct:   0.25,
	MaxDrawdownPct:        0.20,
}

// All is every configured risk profile, used to loop over both agents.
var All = []RiskProfile{Risky, Conservative}

// ByName returns the profile matching name ("risky"/"conservative"), or ok=false.
func ByName(name string) (RiskProfile, bool) {
	for _, p := range All {
		if p.Name == name {
			return p, true
		}
	}
	return RiskProfile{}, false
}
