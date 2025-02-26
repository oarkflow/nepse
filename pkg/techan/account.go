package techan

import (
	"fmt"

	"github.com/schmidthole/big"
)

// Account is an object describing a trading account, including trading record, open positions
// and current cash on hand.
type Account struct {
	Positions   map[string]*Position
	Cash        big.Decimal
	TradeRecord []*Order
}

// An AccountSnapshot provides the point in time state of an account and its positions.
type AccountSnapshot struct {
	Period    TimePeriod          `yaml:"period"`
	Equity    big.Decimal         `yaml:"equity"`
	Cash      big.Decimal         `yaml:"cash"`
	Positions []*PositionSnapshot `yaml:"positions"`
}

// NewAccount returns a new Account
func NewAccount() (a *Account) {
	a = new(Account)
	a.Positions = make(map[string]*Position)
	a.Cash = big.ZERO
	a.TradeRecord = make([]*Order, 0)
	return a
}

// CurrentPosition returns the current position in this record
func (a *Account) OpenPosition(security string) (*Position, bool) {
	pos, exists := a.Positions[security]
	return pos, exists
}

// Enters a cash deposit into the account structure
func (a *Account) Deposit(cash big.Decimal) {
	a.Cash = a.Cash.Add(cash)
}

// Withdraws a cash sum from the account structure
func (a *Account) Withdraw(cash big.Decimal) error {
	intermediate := a.Cash.Sub(cash)

	if intermediate.LT(big.ZERO) {
		return fmt.Errorf(
			"insufficient funds. cannot withdraw %v from %v",
			cash.String(),
			a.Cash.String(),
		)
	}

	a.Cash = intermediate
	return nil
}

// Updates prices for all open positions. If a price is not provided for a given position its price
// is not updated
func (a *Account) UpdatePrices(prices Pricing) {
	for key := range a.Positions {
		price, exists := prices[key]
		if exists {
			a.Positions[key].UpdatePrice(price)
		}
	}
}

// Checks whether the account has enough funds to execute an order.
// We assume there is always enough funds for a sell order since the system
// is currently long only.
func (a *Account) HasSufficientFunds(order *Order) bool {
	if order.Side == SELL {
		return true
	}

	return order.CostBasis().LTE(a.Cash)
}

// Calculates the total equity of the account including unrealized equity for open positions
func (a *Account) Equity() big.Decimal {
	equity := big.NewDecimal(0.0)
	equity = equity.Add(a.Cash)

	for _, pos := range a.Positions {
		equity = equity.Add(pos.UnrealizedEquity())
	}

	return equity
}

// Execute an order against the account
func (a *Account) ExecuteOrder(order *Order) error {
	if !a.HasSufficientFunds(order) {
		return fmt.Errorf(
			"insufficient funds to execute order for %v of %v. need %v, have %v",
			order.Amount,
			order.Security,
			order.CostBasis(),
			a.Cash,
		)
	}

	_, exists := a.Positions[order.Security]

	// operate on the position or create a new one
	if exists {
		err := a.Positions[order.Security].ExecuteOrder(order)
		if err != nil {
			return err
		}

		if a.Positions[order.Security].IsClosed() {
			delete(a.Positions, order.Security)
		}
	} else if !exists && (order.Side == BUY) {
		a.Positions[order.Security] = NewPosition(order)
	} else {
		return fmt.Errorf(
			"cannot enter a short position for %v shares of %v",
			order.Amount.String(),
			order.Security,
		)
	}

	// reflect the order in the account's cash
	if order.Side == BUY {
		a.Withdraw(order.CostBasis())
	} else {
		a.Deposit(order.CostBasis())
	}

	a.TradeRecord = append(a.TradeRecord, order)

	return nil
}

// Exports a snapshot of the current account state to be used in the account history
// and analysis tooling.
func (a *Account) ExportSnapshot(period TimePeriod) *AccountSnapshot {
	snapshot := new(AccountSnapshot)

	snapshot.Period = period
	snapshot.Cash = a.Cash
	snapshot.Equity = a.Equity()

	snapshot.Positions = make([]*PositionSnapshot, 0)
	for _, value := range a.Positions {
		snapshot.Positions = append(snapshot.Positions, value.ExportSnapshot())
	}

	return snapshot
}
