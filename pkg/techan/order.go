package techan

import (
	"time"

	"github.com/schmidthole/big"
)

// OrderSide is a simple enumeration representing the side of an Order (buy or sell)
type OrderSide string

// BUY and SELL enumerations
const (
	BUY  OrderSide = "BUY"
	SELL OrderSide = "SELL"
)

// OrderType defines common order types accepted by brokers
type OrderType string

// OrderType enumerations, we only support the basics at this time
// there are more complex types such as "stop limit" "trail" etc. that may be added later
const (
	MARKET OrderType = "MKT"
	LIMIT  OrderType = "LMT"
	STOP   OrderType = "STP"
)

// TimeInForce defines how long an order is good for before it is auto cancelled
type TimeInForce string

// TimeInForce option enumerations
const (
	GTC TimeInForce = "GTC"
	OPG TimeInForce = "OPG"
	DAY TimeInForce = "DAY"
	IOC TimeInForce = "IOC"
)

// OrderStatus defines the state of an order with a broker
type OrderStatus string

const (
	PENDING   OrderStatus = "Pending"
	FILLED    OrderStatus = "Filled"
	CANCELLED OrderStatus = "Cancelled"
	OTHER     OrderStatus = "Other"
)

// Order represents a trade execution (buy or sell) with associated metadata.
type Order struct {
	ID            string
	Side          OrderSide
	Security      string
	ContractID    int
	Price         big.Decimal
	Type          OrderType
	FilledAmount  big.Decimal
	Amount        big.Decimal
	TimeInForce   TimeInForce
	ExecutionTime time.Time
	Status        OrderStatus
}

// Return the total cost to execute the order.
func (o *Order) CostBasis() big.Decimal {
	return o.Amount.Mul(o.Price)
}
