package techan

import "github.com/oarkflow/nepse/big"

// NewPositiveDirectionalIndicator returns value of +DI for a given index
func NewPositiveDirectionalIndicator(series *TimeSeries, window int) Indicator {
	positionDirectionalMovement := positiveDirectionalMovement{highPriceIndicator: NewHighPriceIndicator(series)}

	return positiveDirectionalIndicator{
		smoothedPositiveDirectionalMovement: NewMMAIndicator(positionDirectionalMovement, window),
		averageTrueRangeIndicator:           NewAverageTrueRangeIndicator(series, window),
		window:                              window,
	}
}

type positiveDirectionalIndicator struct {
	smoothedPositiveDirectionalMovement Indicator
	averageTrueRangeIndicator           Indicator
	window                              int
}

func (pdi positiveDirectionalIndicator) Calculate(index int) big.Decimal {
	if index < pdi.window {
		return big.ZERO
	}

	return pdi.smoothedPositiveDirectionalMovement.Calculate(index).Div(pdi.averageTrueRangeIndicator.Calculate(index)).Mul(big.NewFromInt(100))
}

type positiveDirectionalMovement struct {
	highPriceIndicator Indicator
}

func (pdm positiveDirectionalMovement) Calculate(index int) big.Decimal {
	if index-1 < 0 {
		return big.ZERO
	}

	val := pdm.highPriceIndicator.Calculate(index).Sub(pdm.highPriceIndicator.Calculate(index - 1))
	if val.LT(big.ZERO) {
		return big.ZERO
	}

	return val
}

// NewNegativeDirectionalIndicator returns value of -DI for a given index
func NewNegativeDirectionalIndicator(series *TimeSeries, window int) Indicator {
	negativeDirectionalMovement := negativeDirectionalMovement{lowPriceIndicator: NewLowPriceIndicator(series)}

	return negativeDirectionalIndicator{
		smoothedNegativeDirectionalMovement: NewMMAIndicator(negativeDirectionalMovement, window),
		averageTrueRangeIndicator:           NewAverageTrueRangeIndicator(series, window),
		window:                              window,
	}
}

type negativeDirectionalIndicator struct {
	smoothedNegativeDirectionalMovement Indicator
	averageTrueRangeIndicator           Indicator
	window                              int
}

func (ndi negativeDirectionalIndicator) Calculate(index int) big.Decimal {
	if index < ndi.window {
		return big.ZERO
	}

	return ndi.smoothedNegativeDirectionalMovement.Calculate(index).Div(ndi.averageTrueRangeIndicator.Calculate(index)).Mul(big.NewFromInt(100))
}

type negativeDirectionalMovement struct {
	lowPriceIndicator Indicator
}

func (ndm negativeDirectionalMovement) Calculate(index int) big.Decimal {
	if index-1 < 0 {
		return big.ZERO
	}

	val := ndm.lowPriceIndicator.Calculate(index - 1).Sub(ndm.lowPriceIndicator.Calculate(index))
	if val.LT(big.ZERO) {
		return big.ZERO
	}

	return val
}
