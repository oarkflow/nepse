package techan

import "github.com/oarkflow/nepse/big"

func NewAlligatorIndicator(series *TimeSeries, window, offset int) Indicator {
	mp := NewMedianPriceIndicator(series)
	mma := NewMMAIndicator(mp, window)

	return &alligatorIndicator{
		indicator: mma,
		offset:    offset,
	}
}

type alligatorIndicator struct {
	indicator Indicator
	offset    int
}

func (ai *alligatorIndicator) Calculate(index int) big.Decimal {
	return ai.indicator.Calculate(index - ai.offset)
}
