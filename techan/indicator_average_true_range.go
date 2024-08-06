package techan

import "github.com/oarkflow/nepse/big"

type averageTrueRangeIndicator struct {
	indicator Indicator
}

// NewAverageTrueRangeIndicator returns a base indicator that calculates the average true range of the
// underlying over a window
// https://www.investopedia.com/terms/a/atr.asp
func NewAverageTrueRangeIndicator(series *TimeSeries, window int) Indicator {
	return averageTrueRangeIndicator{
		indicator: NewMMAIndicator(NewTrueRangeIndicator(series), window),
	}
}

func (atr averageTrueRangeIndicator) Calculate(index int) big.Decimal {
	return atr.indicator.Calculate(index)
}
