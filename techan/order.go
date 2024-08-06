package techan

import (
	"time"

	"github.com/oarkflow/nepse/big"
)

// OrderSide is a simple enumeration representing the side of an Order (buy or sell)
type OrderSide int

// BUY and SELL enumerations
const (
	BUY OrderSide = iota
	SELL
)

// Order represents a trade execution (buy or sell) with associated metadata.
type Order struct {
	Side          OrderSide
	Security      string
	Price         big.Decimal
	Amount        big.Decimal
	ExecutionTime time.Time
}

// OrderPlan defines how to construct an Order object during execution of a Strategy.
// The `PercentEquity` field should be between 0.00 and 100.00, corresponding to the
// percent of the overall portfolio allocated to a given position.
type OrderPlan struct {
	Side          OrderSide
	PercentEquity big.Decimal
}
