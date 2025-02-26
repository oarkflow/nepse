package techan

import (
	"fmt"

	"github.com/schmidthole/big"
)

// A trade plan is a list of orders to be executed in the order they appear.
type TradePlan []Order

// A trade plan is created by calculating the diff between the current account positions and
// the desired allocations provided.
func CreateTradePlan(allocations Allocations, pricing Pricing, account *Account) (*TradePlan, error) {
	plan := TradePlan{}

	shareDiffs := map[string]big.Decimal{}
	for security, alloc := range allocations {
		cashValue := alloc.Mul(account.Equity())
		price, exists := pricing[security]
		if !exists {
			return nil, fmt.Errorf("no pricing data provided for %v, cannot create trade plan", security)
		}

		allocShares := CashToShares(price, cashValue)

		position, exists := account.OpenPosition(security)
		if exists {
			allocShares = allocShares.Sub(position.Amount)
		}

		if !allocShares.IsZero() {
			shareDiffs[security] = allocShares
		}
	}

	for security, pos := range account.Positions {
		_, exists := shareDiffs[security]
		if !exists {
			shareDiffs[security] = big.ZERO.Sub(pos.Amount)
		}
	}

	sells := []Order{}
	buys := []Order{}

	for security, shareDiff := range shareDiffs {
		orderSide := BUY
		if shareDiff.LT(big.ZERO) {
			orderSide = SELL
		}

		order := Order{
			Security: security,
			Side:     orderSide,
			Amount:   shareDiff.Abs(),
			Price:    pricing[security],
		}

		if orderSide == BUY {
			buys = append(buys, order)
		} else {
			sells = append(sells, order)
		}
	}

	plan = append(sells, buys...)

	return &plan, nil
}
