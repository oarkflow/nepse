package techan

import "github.com/schmidthole/big"

type localExtremaIndicator struct {
	extrema []int
}

// Defines an indicator to determine local extrema (minima/maxima). This indicator is really
// useful for post analysis and calculating things like max drawdown. It is *NOT* good for
// live trading as it uses a lookback to find the local extrema when an roc switches direction.
//
// This indicator can be useful for identification of "boxes" or "bases" in an asset/security
// where it is not critical that the extrema be identified immediately.
//
// Right now, the entire extrema slice is computed during creation, so we need the total length
// of data to compute as a param. This is required because for a given index to be considered
// a local extrema, we potentially need to look both in front of and behind it.
func NewLocalExtremaIndicator(indicator Indicator, window int, length int) Indicator {
	roc := NewRateOfChangeIndicator(indicator, window)

	extrema := make([]int, length)
	for i := range extrema {
		extrema[i] = 0
	}

	calcWindow := window - 1
	for i := calcWindow; i < length; i++ {
		roc0 := roc.Calculate(i)
		roc1 := roc.Calculate(i - 1)

		startIndex := i - calcWindow
		if startIndex < 0 {
			startIndex = 0
		}

		if roc0.GT(big.ZERO) && roc1.LTE(big.ZERO) {
			minimaIndex := findMinIndexInWindow(indicator, startIndex, i)
			extrema[minimaIndex] = -1
		} else if roc0.LTE(big.ZERO) && roc1.GT(big.ZERO) {
			maximaIndex := findMaxIndexInWindow(indicator, startIndex, i)
			extrema[maximaIndex] = 1
		}
	}

	return localExtremaIndicator{extrema: extrema}
}

func (ext localExtremaIndicator) Calculate(index int) big.Decimal {
	return big.NewFromInt(ext.extrema[index])
}

func findMaxIndexInWindow(ind Indicator, startIndex int, endIndex int) int {
	maxIndex := startIndex
	for i := startIndex; i <= endIndex; i++ {
		if ind.Calculate(i).GTE(ind.Calculate(maxIndex)) {
			maxIndex = i
		}
	}

	return maxIndex
}

func findMinIndexInWindow(ind Indicator, startIndex int, endIndex int) int {
	minIndex := startIndex
	for i := startIndex; i <= endIndex; i++ {
		if ind.Calculate(i).LTE(ind.Calculate(minIndex)) {
			minIndex = i
		}
	}

	return minIndex
}
