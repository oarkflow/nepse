package techan

import (
	"github.com/schmidthole/big"
)

type binaryFrequencyIndicator struct {
	indicators []Indicator
	window     int
	threshold  big.Decimal
}

// Binary Frequency is a specialized indicator which returns the number of times an
// indicator was greater than a defined threshold over a lookback window.
func NewBinaryFrequencyIndicator(indicators []Indicator, window int, threshold float64) Indicator {
	return binaryFrequencyIndicator{
		indicators: indicators,
		window:     window,
		threshold:  big.NewDecimal(threshold),
	}
}

func (s binaryFrequencyIndicator) Calculate(index int) big.Decimal {
	calcWindow := s.window - 1
	if index < s.window+1 {
		calcWindow = index
	}

	frequency := 0
	for i := index - calcWindow; i <= index; i++ {
		for _, indicator := range s.indicators {
			if indicator.Calculate(i).GT(big.ZERO) {
				frequency++
			}
		}
	}

	return big.NewFromInt(frequency)
}
