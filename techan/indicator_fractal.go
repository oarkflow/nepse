package techan

import (
	"github.com/oarkflow/nepse/big"
	"math"
)

// NewUpFractalIndicator returns an indicator which returns the high price of the previous up fractal for a given index
func NewUpFractalIndicator(series *TimeSeries, window int) Indicator {
	return &upFractalIndicator{
		highPriceIndicator: NewHighPriceIndicator(series),
		window:             window,
		resultCache:        make([]*big.Decimal, 10000),
	}
}

type upFractalIndicator struct {
	highPriceIndicator Indicator
	window             int
	resultCache        resultCache
}

func (ufi *upFractalIndicator) Calculate(index int) big.Decimal {
	firstIndex := index - ufi.windowSize() + 1
	if firstIndex < 0 {
		return big.NewFromInt(math.MaxInt)
	}

	val := ufi.highPriceIndicator.Calculate(index - ufi.window)
	highest := val
	for i := firstIndex; i <= index; i++ {
		currentVal := ufi.highPriceIndicator.Calculate(i)
		if currentVal.GT(highest) {
			highest = currentVal
		}
	}

	result := highest

	if !result.EQ(val) {
		result = ufi.Calculate(index - 1)
		cacheResult(ufi, index, result)
		return result
	}

	for i := index - ufi.window + 1; i <= index; i++ {
		cacheResult(ufi, i, result)
	}

	return result
}

func (ufi upFractalIndicator) cache() resultCache {
	return ufi.resultCache
}

func (ufi *upFractalIndicator) setCache(cache resultCache) {
	ufi.resultCache = cache
}

func (ufi upFractalIndicator) windowSize() int {
	return ufi.window*2 + 1
}

// NewDownFractalIndicator returns an indicator which returns the low price of the previous down fractal for a given index
func NewDownFractalIndicator(series *TimeSeries, window int) Indicator {
	return &downFractalIndicator{
		lowPriceIndicator: NewLowPriceIndicator(series),
		window:            window,
		resultCache:       make([]*big.Decimal, 10000),
	}
}

type downFractalIndicator struct {
	lowPriceIndicator Indicator
	window            int
	resultCache       resultCache
}

func (dfi *downFractalIndicator) Calculate(index int) big.Decimal {
	firstIndex := index - dfi.windowSize() + 1
	if firstIndex < 0 {
		return big.NewFromInt(-math.MaxInt)
	}

	val := dfi.lowPriceIndicator.Calculate(index - dfi.window)
	lowest := val
	for i := firstIndex; i <= index; i++ {
		currentVal := dfi.lowPriceIndicator.Calculate(i)
		if currentVal.LT(lowest) {
			lowest = currentVal
		}
	}

	result := lowest

	if !result.EQ(val) {
		result = dfi.Calculate(index - 1)
		cacheResult(dfi, index, result)
		return result
	}

	for i := index - dfi.window + 1; i <= index; i++ {
		cacheResult(dfi, i, result)
	}

	return result
}

func (dfi downFractalIndicator) cache() resultCache {
	return dfi.resultCache
}

func (dfi *downFractalIndicator) setCache(cache resultCache) {
	dfi.resultCache = cache
}

func (dfi downFractalIndicator) windowSize() int {
	return dfi.window*2 + 1
}
