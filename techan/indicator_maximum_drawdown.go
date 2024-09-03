package techan

import "github.com/oarkflow/nepse/big"

type maximumDrawdownIndicator struct {
	drawdowns []big.Decimal
}

// Revamped maximum drawdown indicator which uses a local extrema calculation
// to find the difference between almost absolute peaks and valleys over time.
//
// Each "peak" and "valley" are paired to introduce a single "drawdown"
func NewMaximumDrawdownIndicator(indicator Indicator, window int, length int) Indicator {
	extrema := NewLocalExtremaIndicator(indicator, window, length)

	drawdowns := []big.Decimal{}
	maxDrawdown := big.ZERO
	lastPeak := big.ZERO
	inDrawdown := false

	for i := 0; i < length; i++ {
		isExtrema := extrema.Calculate(i)

		if isExtrema.GT(big.ZERO) {
			lastPeak = indicator.Calculate(i)
			inDrawdown = true
		} else if isExtrema.LT(big.ZERO) && inDrawdown {
			drawdown := lastPeak.Sub(indicator.Calculate(i))

			if drawdown.GT(maxDrawdown) {
				maxDrawdown = drawdown
			}

			inDrawdown = false
		}

		drawdowns = append(drawdowns, maxDrawdown)
	}

	return maximumDrawdownIndicator{drawdowns: drawdowns}
}

func (mdi maximumDrawdownIndicator) Calculate(index int) big.Decimal {
	return mdi.drawdowns[index]
}
