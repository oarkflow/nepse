package techan

import (
	"github.com/schmidthole/big"
)

type supertrendIndicator struct {
	supertrend []big.Decimal
}

// NewSupertrendIndicator returns a derivative indicator which calculates the well-known
// supertrend indicator which is based on trend detection using a combination of the average
// true range (ATR) and the high/low prices of an asset over a defined window.
//
// The typical accepted multipler value is 3.
//
// https://www.investopedia.com/supertrend-indicator-7976167
func NewSupertrendIndicator(series *TimeSeries, window int, multiplier int) Indicator {
	atr := NewAverageTrueRangeIndicator(series, window)
	highs := NewHighPriceIndicator(series)
	lows := NewLowPriceIndicator(series)
	closes := NewClosePriceIndicator(series)

	multiplierAsDecimal := big.NewDecimal(float64(multiplier))
	avgDivisorAsDecimal := big.NewDecimal(2.0)

	basicUpperBand := make([]big.Decimal, series.LastIndex()+1)
	basicLowerBand := make([]big.Decimal, series.LastIndex()+1)
	finalUpperBand := make([]big.Decimal, series.LastIndex()+1)
	finalLowerBand := make([]big.Decimal, series.LastIndex()+1)

	supertrend := make([]big.Decimal, series.LastIndex()+1)

	for i := 0; i <= series.LastIndex(); i++ {
		avgPrice := highs.Calculate(i).Add(lows.Calculate(i)).Div(avgDivisorAsDecimal)
		atrValue := atr.Calculate(i)

		atrDiff := multiplierAsDecimal.Mul(atrValue)

		basicUpperBand[i] = avgPrice.Add(atrDiff)
		basicLowerBand[i] = avgPrice.Sub(atrDiff)

		if i == 0 {
			finalUpperBand[i] = basicUpperBand[i]
			finalLowerBand[i] = basicLowerBand[i]
			supertrend[i] = basicLowerBand[i] // Initialize to first final lower band
			continue
		}

		close := closes.Calculate(i)
		lastClose := closes.Calculate(i - 1)

		if basicUpperBand[i].LT(finalUpperBand[i-1]) || lastClose.GT(finalUpperBand[i-1]) {
			finalUpperBand[i] = basicUpperBand[i]
		} else {
			finalUpperBand[i] = finalUpperBand[i-1]
		}

		if basicLowerBand[i].GT(finalLowerBand[i-1]) || lastClose.LT(finalLowerBand[i-1]) {
			finalLowerBand[i] = basicLowerBand[i]
		} else {
			finalLowerBand[i] = finalLowerBand[i-1]
		}

		if lastClose.LTE(finalUpperBand[i-1]) && close.GT(finalUpperBand[i]) {
			supertrend[i] = finalLowerBand[i]
		} else if lastClose.GTE(finalLowerBand[i-1]) && close.LT(finalLowerBand[i]) {
			supertrend[i] = finalUpperBand[i]
		} else {
			if supertrend[i-1].EQ(finalUpperBand[i-1]) {
				supertrend[i] = finalUpperBand[i]
			} else {
				supertrend[i] = finalLowerBand[i]
			}
		}
	}

	return supertrendIndicator{supertrend: supertrend}
}

func (s supertrendIndicator) Calculate(index int) big.Decimal {
	return s.supertrend[index]
}
