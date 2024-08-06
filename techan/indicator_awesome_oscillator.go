package techan

import "github.com/oarkflow/nepse/big"

// NewVolumeIndicator returns an indicator which returns the volume of a candle for a given index
func NewAwesomeOscillatorIndicator(series *TimeSeries) Indicator {
	medianPriceIndicator := NewMedianPriceIndicator(series)
	return awesomeOscillatorIndicator{
		sma5:  NewSimpleMovingAverage(medianPriceIndicator, 5),
		sma34: NewSimpleMovingAverage(medianPriceIndicator, 34),
	}
}

type awesomeOscillatorIndicator struct {
	sma5  Indicator
	sma34 Indicator
}

func (aoi awesomeOscillatorIndicator) Calculate(index int) big.Decimal {
	return aoi.sma5.Calculate(index).Sub(aoi.sma34.Calculate(index))
}
