package techan

// The Backtest is a holder struct to run a simulated trading strategy against an account.
type Backtest struct {
	tick       int
	strategies []Strategy
	allocator  Allocator
	account    *Account
	history    *AccountHistory
}

// Create a new backtest with the provided strategies, allocator, and starting account.
// All strategies are assumed to be normalized and have the same length, indexes, and periods.
func NewBacktest(strategies []Strategy, allocator Allocator, account *Account) *Backtest {
	backtest := Backtest{
		tick:       0,
		strategies: strategies,
		allocator:  allocator,
		account:    account,
		history:    NewAccountHistory(),
	}

	// setup the initial account state to be one period before all data.
	// this will serve as a starting point before orders are executed.
	if (len(strategies) > 0) && (len(strategies[0].Timeseries.Candles) > 0) {
		initialPeriod := strategies[0].Timeseries.Candles[0].Period.Advance(-1)
		backtest.history.ApplySnapshot(
			backtest.account.ExportSnapshot(initialPeriod),
			&PricingSnapshot{Period: initialPeriod, Prices: Pricing{}},
		)
	}

	return &backtest
}

// Run the backtest from start to finish.
func (b *Backtest) Run() (*AccountHistory, error) {
	for {
		err := b.executeTick()
		if err != nil {
			return nil, err
		}

		if b.tick == b.lastTick() {
			break
		}

		b.advanceTick()
	}

	return b.history, nil
}

func (b *Backtest) executeTick() error {
	prices := Pricing{}
	for _, strat := range b.strategies {
		prices[strat.Security] = strat.Timeseries.Candles[b.tick].ClosePrice
	}

	allocations := b.allocator.Allocate(b.tick, b.strategies)

	tradePlan, err := CreateTradePlan(allocations, prices, b.account)
	if err != nil {
		return err
	}

	for _, order := range *tradePlan {
		b.account.ExecuteOrder(&order)
	}

	period := b.strategies[0].Timeseries.Candles[b.tick].Period
	b.history.ApplySnapshot(
		b.account.ExportSnapshot(period),
		&PricingSnapshot{Period: period, Prices: prices},
	)

	return nil
}

func (b *Backtest) advanceTick() {
	if b.tick >= b.lastTick() {
		return
	}

	b.tick = b.tick + 1
}

func (b *Backtest) lastTick() int {
	return b.strategies[0].Timeseries.LastIndex()
}
