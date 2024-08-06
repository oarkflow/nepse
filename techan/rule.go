package techan

import "github.com/oarkflow/nepse/big"

// Rule is an interface describing an algorithm by which a set of criteria may be satisfied
type Rule interface {
	IsSatisfied(index int, record *TradingRecord) bool
}

// And returns a new rule whereby BOTH of the passed-in rules must be satisfied for the rule to be satisfied
func And(rules ...Rule) Rule {
	return andRule{rules: rules}
}

type andRule struct {
	rules []Rule
}

func (ar andRule) IsSatisfied(index int, record *TradingRecord) bool {
	for _, r := range ar.rules {
		if !r.IsSatisfied(index, record) {
			return false
		}
	}
	return true
}

// Or returns a new rule whereby ONE OF the passed-in rules must be satisfied for the rule to be satisfied
func Or(rules ...Rule) Rule {
	return orRule{rules: rules}
}

type orRule struct {
	rules []Rule
}

func (or orRule) IsSatisfied(index int, record *TradingRecord) bool {
	for _, r := range or.rules {
		if r.IsSatisfied(index, record) {
			return true
		}
	}
	return false
}

// Under is a rule where the previous Indicators must be less than the following Indicators to be Satisfied
func Under(indicators ...Indicator) Rule {
	return underIndicatorRule{indicators: indicators}
}

type underIndicatorRule struct {
	indicators []Indicator
}

// IsSatisfied returns true when the previous Indicators are less than the following Indicators
func (uir underIndicatorRule) IsSatisfied(index int, record *TradingRecord) bool {
	for i := 0; i < len(uir.indicators)-1; i++ {
		if !uir.indicators[i].Calculate(index).LT(uir.indicators[i+1].Calculate(index)) {
			return false
		}
	}
	return true
}

// NewPercentChangeRule returns a rule whereby the given Indicator must have changed by a given percentage to be satisfied.
// You should specify percent as a float value between -1 and 1
func NewPercentChangeRule(indicator Indicator, percent float64) Rule {
	return percentChangeRule{
		indicator: NewPercentChangeIndicator(indicator),
		percent:   big.NewDecimal(percent),
	}
}

type percentChangeRule struct {
	indicator Indicator
	percent   big.Decimal
}

func (pcr percentChangeRule) IsSatisfied(index int, record *TradingRecord) bool {
	return pcr.indicator.Calculate(index).Abs().GT(pcr.percent.Abs())
}
