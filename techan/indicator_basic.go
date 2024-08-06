package techan

import "github.com/oarkflow/nepse/big"

// NewVolumeIndicator returns an indicator which returns the volume of a candle for a given index
func NewVolumeIndicator(series *TimeSeries) Indicator {
	return volumeIndicator{series}
}

type volumeIndicator struct {
	*TimeSeries
}

func (vi volumeIndicator) Calculate(index int) big.Decimal {
	return vi.Candles[index].Volume
}

// NewOpenPriceIndicator returns an Indicator which returns the open price of a candle for a given index
func NewOpenPriceIndicator(series *TimeSeries) Indicator {
	return openPriceIndicator{
		series,
	}
}

type openPriceIndicator struct {
	*TimeSeries
}

func (opi openPriceIndicator) Calculate(index int) big.Decimal {
	return opi.Candles[index].OpenPrice
}

// NewClosePriceIndicator returns an Indicator which returns the close price of a candle for a given index
func NewClosePriceIndicator(series *TimeSeries) Indicator {
	return closePriceIndicator{series}
}

type closePriceIndicator struct {
	*TimeSeries
}

func (cpi closePriceIndicator) Calculate(index int) big.Decimal {
	return cpi.Candles[index].ClosePrice
}

// NewHighPriceIndicator returns an Indicator which returns the high price of a candle for a given index
func NewHighPriceIndicator(series *TimeSeries) Indicator {
	return highPriceIndicator{
		series,
	}
}

type highPriceIndicator struct {
	*TimeSeries
}

func (hpi highPriceIndicator) Calculate(index int) big.Decimal {
	return hpi.Candles[index].MaxPrice
}

// NewLowPriceIndicator returns an Indicator which returns the low price of a candle for a given index
func NewLowPriceIndicator(series *TimeSeries) Indicator {
	return lowPriceIndicator{
		series,
	}
}

type lowPriceIndicator struct {
	*TimeSeries
}

func (lpi lowPriceIndicator) Calculate(index int) big.Decimal {
	return lpi.Candles[index].MinPrice
}

// NewTypicalPriceIndicator returns an Indicator which returns the typical price of a candle for a given index.
// The typical price is an average of the high, low, and close prices for a given candle.
func NewTypicalPriceIndicator(series *TimeSeries) Indicator {
	return typicalPriceIndicator{series}
}

type typicalPriceIndicator struct {
	*TimeSeries
}

func (tpi typicalPriceIndicator) Calculate(index int) big.Decimal {
	numerator := tpi.Candles[index].MaxPrice.Add(tpi.Candles[index].MinPrice).Add(tpi.Candles[index].ClosePrice)
	return numerator.Div(big.NewFromInt(3))
}

// NewMedianPriceIndicator returns an Indicator which returns the median price of a candle for a given index.
// The median price is an average of the high and low prices for a given candle.
func NewMedianPriceIndicator(series *TimeSeries) Indicator {
	return medianPriceIndicator{series}
}

type medianPriceIndicator struct {
	*TimeSeries
}

func (mpi medianPriceIndicator) Calculate(index int) big.Decimal {
	numerator := mpi.Candles[index].MaxPrice.Add(mpi.Candles[index].MinPrice)
	return numerator.Div(big.NewFromInt(2))
}
