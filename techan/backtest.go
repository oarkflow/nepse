package techan

import "github.com/oarkflow/nepse/big"

// A Backtest is a generic runner which outputs a record of trades along with resulting equity.
type Backtest interface {
	Run(startingEquity big.Decimal) (big.Decimal, *TradingRecord)
}

type fixedEntryBacktest struct {
	security       string
	series         *TimeSeries
	priceIndicator Indicator
	TradingRecord  *TradingRecord
	strategy       Strategy
	orderPlan      OrderPlan
}

// FixedEntryBacktest runs a backtest based on a fixed price entry which can be passed via
// a price based indicator such as those defined by NewClosePriceIndicator and NewOpenPriceIndicator.
// This means the backtest can be setup to always enter/exit a position at the current open, typical,
// closing price, etc. of a series.
//
// For the purposes of this backtest framework, fractional trading of the security is always enabled
// even if the underlying instrument can only be traded in whole shares.
func NewFixedEntryBacktest(
	security string,
	series *TimeSeries,
	priceIndicator Indicator,
	strategy Strategy,
	orderPlan OrderPlan,
) Backtest {
	return fixedEntryBacktest{
		security:       security,
		series:         series,
		priceIndicator: priceIndicator,
		TradingRecord:  NewTradingRecord(),
		strategy:       strategy,
		orderPlan:      orderPlan,
	}
}

func (b fixedEntryBacktest) Run(startingEquity big.Decimal) (big.Decimal, *TradingRecord) {
	equity := startingEquity

	for i := 0; i <= b.series.LastIndex(); i++ {
		if b.strategy.ShouldEnter(i, b.TradingRecord) {
			price := b.priceIndicator.Calculate(i)
			percentEquityFraction := b.orderPlan.PercentEquity.Div(big.NewDecimal(100.0))
			allocation := equity.Mul(percentEquityFraction)
			amount := allocation.Div(price)

			side := b.orderPlan.Side

			entryOrder := Order{
				Side:          side,
				Security:      b.security,
				Price:         price,
				Amount:        amount,
				ExecutionTime: b.series.Candles[i].Period.Start,
			}

			b.TradingRecord.Operate(entryOrder)
			equity = equity.Sub(allocation)
		} else if b.strategy.ShouldExit(i, b.TradingRecord) {
			price := b.priceIndicator.Calculate(i)
			amount := b.TradingRecord.CurrentPosition().EntranceOrder().Amount

			side := BUY
			if b.TradingRecord.CurrentPosition().IsShort() {
				side = SELL
			}

			exitOrder := Order{
				Side:          side,
				Security:      b.security,
				Price:         price,
				Amount:        amount,
				ExecutionTime: b.series.Candles[i].Period.Start,
			}

			b.TradingRecord.Operate(exitOrder)
			equity = equity.Add(b.TradingRecord.LastTrade().ExitValue())
		}
	}

	return equity, b.TradingRecord
}
