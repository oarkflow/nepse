package techan

import "github.com/schmidthole/big"

type rateOfChangeIndicator struct {
	indicator Indicator
	window    int
}

// The rate of change indicator calculates the total change over a window divided by the number
// of candles in the window.
func NewRateOfChangeIndicator(indicator Indicator, window int) Indicator {
	return rateOfChangeIndicator{indicator: indicator, window: window}
}

func (roc rateOfChangeIndicator) Calculate(index int) big.Decimal {
	if (index < (roc.window - 1)) || (roc.window < 2) {
		return big.ZERO
	}

	start := roc.indicator.Calculate(index - (roc.window - 1))
	end := roc.indicator.Calculate(index)

	return end.Sub(start).Div(big.NewFromInt(roc.window))
}
