package techan

// MarketSnapshot encompasses useful market state info such as pricing and trading state, which is
// needed for decision making when using techan with a broker during live trading
type MarketSnapshot struct {
	Pricing      Pricing
	TradingState map[string]TradingState
}
