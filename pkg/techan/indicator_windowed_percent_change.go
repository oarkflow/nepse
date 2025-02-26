package techan

import (
	"github.com/schmidthole/big"
)

type windowedPercentChangeIndicator struct {
	indicator Indicator
	window    int
}

// The window percent change indicator calculates the change percentage over a fixed
// lookback window.
func NewWindowedPercentChangeIndicator(indicator Indicator, window int) Indicator {
	return windowedPercentChangeIndicator{
		indicator: indicator,
		window:    window,
	}
}

func (s windowedPercentChangeIndicator) Calculate(index int) big.Decimal {
	if s.window < 2 {
		return big.ZERO
	}

	calcWindow := s.window - 1
	if index < calcWindow {
		calcWindow = index
	}

	end := s.indicator.Calculate(index)
	start := s.indicator.Calculate(index - calcWindow)

	if start.EQ(big.ZERO) {
		if end.GT(big.ZERO) {
			return big.NewDecimal(100.00)
		} else if end.LT(big.ZERO) {
			return big.NewDecimal(-100.00)
		}

		return big.ZERO
	}

	return end.Sub(start).Div(start).Mul(big.NewFromInt(100))
}
