package techan

import "github.com/schmidthole/big"

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
		} else if inDrawdown {
			drawdown := lastPeak.Sub(indicator.Calculate(i)).Div(lastPeak).Mul(big.NewDecimal(100.00))

			if drawdown.GT(maxDrawdown) {
				maxDrawdown = drawdown
			}

			if isExtrema.LT(big.ZERO) {
				inDrawdown = false
			}
		}

		drawdowns = append(drawdowns, maxDrawdown)
	}

	return maximumDrawdownIndicator{drawdowns: drawdowns}
}

func (mdi maximumDrawdownIndicator) Calculate(index int) big.Decimal {
	return mdi.drawdowns[index]
}
