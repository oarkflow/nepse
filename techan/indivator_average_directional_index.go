package techan

import "github.com/oarkflow/nepse/big"

// NewAverageDirectionalIndexIndicator returns value of ADX for a given index
func NewAverageDirectionalIndexIndicator(series *TimeSeries, window int) Indicator {
	return averageDirectionalIndexIndicator{
		indicator: NewMMAIndicator(NewDirectionalIndexIndicator(series, window), window),
	}
}

type averageDirectionalIndexIndicator struct {
	indicator Indicator
}

func (adx averageDirectionalIndexIndicator) Calculate(index int) big.Decimal {
	return adx.indicator.Calculate(index)
}
