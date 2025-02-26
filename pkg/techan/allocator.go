package techan

import "github.com/schmidthole/big"

// An allocator is a functional component which provides a portfolio allocation across many Securities.
// A list of strategy structures is passed in to calculate the allocations based on rule analysis or
// timeseries data. All outputted allocation fractions will add up to 1.0 (big.ONE).
type Allocator interface {
	Allocate(index int, strategies []Strategy) Allocations
	AllocateWithAccount(index int, strategies []Strategy, account *Account) Allocations
}

// Allocations are simply a map of securities and their fraction of allocation. All item's fraction
// in the map should add up to 1.0.
type Allocations map[string]big.Decimal

// The naive allocator will proportionally allocate to all strategies which are triggered. A maximum
// position and total allocation can be provided in order to limit single position size and how much
// of a given portfolio is allocated.
type NaiveAllocator struct {
	maxSinglePositionFraction big.Decimal
	maxTotalPositionFraction  big.Decimal
}

// Sets up a new naive allocator. if any of the maximum fractions exceed 1.0, then 1.0 will be used.
// If the max single fraction exceeds the total fraction, then it will be set at the total fraction.
func NewNaiveAllocator(maxSinglePositionFraction big.Decimal, maxTotalPositionFraction big.Decimal) *NaiveAllocator {
	maxSingle := maxSinglePositionFraction
	if maxSinglePositionFraction.GT(big.ONE) {
		maxSingle = big.ONE
	}

	maxTotal := maxTotalPositionFraction
	if maxTotalPositionFraction.GT(big.ONE) {
		maxTotal = big.ONE
	}

	if maxSingle.GT(maxTotal) {
		maxSingle = maxTotal
	}

	return &NaiveAllocator{
		maxSinglePositionFraction: maxSingle,
		maxTotalPositionFraction:  maxTotal,
	}
}

// Perform a naive allocation which simply gives an equal portion of allocation to all strategies whose
// rules are satisfied.
func (na *NaiveAllocator) Allocate(index int, strategies []Strategy) Allocations {
	triggers := make([]string, 0)
	allocations := make(map[string]big.Decimal, 0)

	for _, s := range strategies {
		if s.Rule.IsSatisfied(index) {
			triggers = append(triggers, s.Security)
		}
	}

	if len(triggers) == 0 {
		return allocations
	}

	allocationFraction := na.maxTotalPositionFraction.Div(big.NewFromInt(len(triggers)))
	if allocationFraction.GT(na.maxSinglePositionFraction) {
		allocationFraction = na.maxSinglePositionFraction
	}

	for _, t := range triggers {
		allocations[t] = allocationFraction
	}

	return allocations
}

// Perform a naive allocation, but take into account which positions are currently open
// with an account. This will also only include an allocation for a new position if the
// strategy has an entry on this index. This avoids allocating a trade at a sub-optimal
// entrypoint.
func (na *NaiveAllocator) AllocateWithAccount(index int, strategies []Strategy, account *Account) Allocations {
	if index == 0 {
		return na.Allocate(index, strategies)
	}

	triggers := make([]string, 0)
	entries := make(map[string]bool, 0)
	allocations := make(map[string]big.Decimal, 0)

	for _, s := range strategies {
		if s.Rule.IsSatisfied(index) {
			triggers = append(triggers, s.Security)

			if s.Rule.IsSatisfied(index - 1) {
				entries[s.Security] = true
			}
		}
	}

	if len(triggers) == 0 {
		return allocations
	}

	allocationFraction := na.maxTotalPositionFraction.Div(big.NewFromInt(len(triggers)))
	if allocationFraction.GT(na.maxSinglePositionFraction) {
		allocationFraction = na.maxSinglePositionFraction
	}

	for _, t := range triggers {
		_, hasPosition := account.OpenPosition(t)
		_, isEntry := entries[t]

		if hasPosition || isEntry {
			allocations[t] = allocationFraction
		}
	}

	return allocations
}
