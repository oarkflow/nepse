package techan

import "github.com/oarkflow/nepse/big"

// NewDirectionalMovementIndicator returns value of DX for a given index
func NewDirectionalIndexIndicator(series *TimeSeries, window int) Indicator {
	return directionalIndexIndicator{
		positiveDirectionIndicator: NewPositiveDirectionalIndicator(series, window),
		negativeDirectionIndicator: NewNegativeDirectionalIndicator(series, window),
		window:                     window,
	}
}

type directionalIndexIndicator struct {
	positiveDirectionIndicator Indicator
	negativeDirectionIndicator Indicator
	window                     int
}

func (dxi directionalIndexIndicator) Calculate(index int) big.Decimal {
	if index < dxi.window {
		return big.ZERO
	}

	absDiff := dxi.positiveDirectionIndicator.Calculate(index).Sub(dxi.negativeDirectionIndicator.Calculate(index)).Abs()
	sum := dxi.positiveDirectionIndicator.Calculate(index).Add(dxi.negativeDirectionIndicator.Calculate(index))

	return absDiff.Div(sum).Mul(big.NewFromInt(100))
}
