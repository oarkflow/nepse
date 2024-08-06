package techan

// IncreaseRule is satisfied when the given Indicator has increasing values from index-window+1 to index.
func NewIncreaseRule(indicator Indicator, window int) Rule {
	return orderedRule{
		indicator:  indicator,
		window:     window,
		isIncrease: true,
	}
}

// IncreaseRule is satisfied when the given Indicator has descreaing values from index-window+1 to index.
func NewDecreaseRule(indicator Indicator, window int) Rule {
	return orderedRule{
		indicator:  indicator,
		window:     window,
		isIncrease: false,
	}
}

type orderedRule struct {
	indicator  Indicator
	window     int
	isIncrease bool
}

func (or orderedRule) IsSatisfied(index int, record *TradingRecord) bool {
	if index < or.window-1 {
		return false
	}

	for i := index - or.window + 1; i < index; i++ {
		current := or.indicator.Calculate(i)
		next := or.indicator.Calculate(i + 1)
		if or.isIncrease && current.GTE(next) || !or.isIncrease && current.LTE(next) {
			return false
		}
	}

	return true
}
