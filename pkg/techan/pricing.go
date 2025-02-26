package techan

import "github.com/schmidthole/big"

// The Pricing type is a simple alias for a map of the security symbol to a price.
type Pricing map[string]big.Decimal
