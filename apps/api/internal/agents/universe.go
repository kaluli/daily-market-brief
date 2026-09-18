package agents

import "strings"

// Asset is one tradeable instrument in the fixed universe the investor
// agents are allowed to operate on. News items mention assets in free text
// ("US Treasury Yields", "Dollar Index", ...); MapAssetName maps that text to
// one of these real, liquid, easy-to-price ETFs so trades can be backed by
// real historical prices (see docs/AGENTS.md).
type Asset struct {
	Ticker string
	Name   string
	Class  string
}

var Universe = []Asset{
	{Ticker: "SPY", Name: "S&P 500 (US large-cap stocks)", Class: "equities"},
	{Ticker: "QQQ", Name: "Nasdaq 100 (big tech)", Class: "equities"},
	{Ticker: "IWM", Name: "Russell 2000 (US small caps)", Class: "equities"},
	{Ticker: "XLE", Name: "Energy sector stocks", Class: "sector"},
	{Ticker: "XLF", Name: "Financials sector stocks", Class: "sector"},
	{Ticker: "TLT", Name: "20+ Year US Treasury Bonds", Class: "bonds"},
	{Ticker: "GLD", Name: "Gold", Class: "commodities"},
	{Ticker: "USO", Name: "Crude Oil", Class: "commodities"},
	{Ticker: "UUP", Name: "US Dollar Index", Class: "currencies"},
	{Ticker: "BITO", Name: "Bitcoin (futures-based)", Class: "crypto"},
}

// keyword -> ticker, checked in order (most specific first). The first match wins.
var assetKeywords = []struct {
	ticker   string
	keywords []string
}{
	{"BITO", []string{"bitcoin", "crypto", "btc"}},
	{"GLD", []string{"gold"}},
	{"USO", []string{"crude oil", "crude", "oil price", "opec", "barrel"}},
	{"XLE", []string{"energy sector", "energy stocks", "oil & gas companies", "oil and gas companies"}},
	{"XLF", []string{"bank", "financial sector", "financials", "lender"}},
	{"TLT", []string{"treasury", "treasuries", "bond market", "bonds", "yield", "interest rate", "rate hike", "rate cut", "fed funds"}},
	{"UUP", []string{"dollar index", "u.s. dollar", "us dollar", "usd", "currency", "forex", "fx market"}},
	{"IWM", []string{"small cap", "small-cap", "russell 2000"}},
	{"QQQ", []string{"nasdaq", "big tech", "tech stocks", "technology sector", "semiconductor", "chipmaker"}},
	{"SPY", []string{"s&p 500", "s&p500", "stock market", "equities", "equity market", "wall street", "broad market", "index"}},
}

// MapAssetName maps a free-text asset name (as produced by the investment
// analyst, e.g. "US Treasury Yields") to a ticker in Universe. ok is false if
// nothing in the fixed universe plausibly matches — callers should skip it
// rather than force a trade.
func MapAssetName(name string) (ticker string, ok bool) {
	lower := strings.ToLower(name)
	for _, k := range assetKeywords {
		for _, kw := range k.keywords {
			if strings.Contains(lower, kw) {
				return k.ticker, true
			}
		}
	}
	return "", false
}
