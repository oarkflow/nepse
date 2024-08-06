package techan

import "github.com/oarkflow/nepse/big"

type trueRangeIndicator struct {
	series *TimeSeries
}

// NewTrueRangeIndicator returns a base indicator
// which calculates the true range at the current point in time for a series
// https://www.investopedia.com/terms/a/atr.asp
func NewTrueRangeIndicator(series *TimeSeries) Indicator {
	return trueRangeIndicator{
		series: series,
	}
}

func (tri trueRangeIndicator) Calculate(index int) big.Decimal {
	if index-1 < 0 {
		return big.ZERO
	}

	candle := tri.series.Candles[index]
	high := candle.MaxPrice
	low := candle.MinPrice
	previousClose := tri.series.Candles[index-1].ClosePrice
	return big.MaxSlice(high.Sub(low), high.Sub(previousClose).Abs(), low.Sub(previousClose).Abs())
}
