package techan

import (
	"fmt"

	"github.com/schmidthole/big"
)

// The PricingSnapshot provides point in time pricing information for all tracked securities,
// including the bechmark if used. The period for each snapshot should match a period supplied
// in an account snapshot.
type PricingSnapshot struct {
	Period TimePeriod
	Prices Pricing
}

// The AccountHistory contains a record of point in time account snapshots as well as a list of all
// of the Securities tracked by the account over time.
type AccountHistory struct {
	Securities []string
	Prices     []*PricingSnapshot
	Snapshots  []*AccountSnapshot
}

func NewAccountHistory() *AccountHistory {
	return &AccountHistory{
		Securities: make([]string, 0),
		Prices:     make([]*PricingSnapshot, 0),
		Snapshots:  make([]*AccountSnapshot, 0),
	}
}

// Add a new account and pricing snapshot to the history.
func (ah *AccountHistory) ApplySnapshot(accountSnapshot *AccountSnapshot, pricingSnapshot *PricingSnapshot) error {
	if !accountSnapshot.Period.Start.Equal(pricingSnapshot.Period.Start) {
		return fmt.Errorf(
			"start periods for account and pricing snapshots do not match: %v <-> %v",
			accountSnapshot.Period.Start,
			pricingSnapshot.Period.Start,
		)
	}

	ah.Snapshots = append(ah.Snapshots, accountSnapshot)
	ah.Prices = append(ah.Prices, pricingSnapshot)

	return nil
}

// Helper function to return the last index of snapshot data.
func (ah *AccountHistory) LastIndex() int {
	return len(ah.Snapshots) - 1
}

// Helper function to return a price at a given snapshot index.
func (ah *AccountHistory) PriceAtIndex(security string, index int) (big.Decimal, bool) {
	price, exists := ah.Prices[index].Prices[security]
	return price, exists
}

// Derive an Indicator from the account history equity.
func (ah *AccountHistory) AccountEquityAsIndicator() Indicator {
	equities := []float64{}
	for _, e := range ah.Snapshots {
		equities = append(equities, e.Equity.Float())
	}

	return NewFixedIndicator(equities...)
}
