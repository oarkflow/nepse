package techan

import "github.com/oarkflow/nepse/big"

type bbandWidthIndicator struct {
	bbandUpper Indicator
	bbandLower Indicator
	ma         Indicator
}

// NewBollingerBandWidthIndicator a a derivative indicator which returns width of a bollinger band
// on the underlying indicator
func NewBollingerBandWidthIndicator(indicator Indicator, window int, sigma float64) Indicator {
	return bbandWidthIndicator{
		ma:         NewSimpleMovingAverage(indicator, window),
		bbandLower: NewBollingerLowerBandIndicator(indicator, window, sigma),
		bbandUpper: NewBollingerUpperBandIndicator(indicator, window, sigma),
	}
}

func (b bbandWidthIndicator) Calculate(index int) big.Decimal {
	return b.bbandUpper.Calculate(index).Sub(b.bbandLower.Calculate(index)).Div(b.ma.Calculate(index))
}
