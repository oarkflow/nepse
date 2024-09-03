package techan

import (
	"math"

	"github.com/oarkflow/nepse/big"
)

// Total profit of the account history.
func (ah *AccountHistory) TotalProfit() big.Decimal {
	endEquity := ah.Snapshots[ah.LastIndex()].Equity
	startEquity := ah.Snapshots[0].Equity

	return endEquity.Sub(startEquity)
}

// Total percent gain of the account history.
func (ah *AccountHistory) PercentGain() big.Decimal {
	if ah.Snapshots[0].Equity.EQ(big.ZERO) {
		return big.ZERO
	}

	return ah.TotalProfit().Div(ah.Snapshots[0].Equity).Mul(big.NewDecimal(100.00))
}

// Get the annualized return of the account equity
func (ah *AccountHistory) AnnualizedReturn() big.Decimal {
	startTimestamp := ah.Snapshots[0].Period.Start
	endTimestamp := ah.Snapshots[ah.LastIndex()].Period.Start

	days := big.NewDecimal(endTimestamp.Sub(startTimestamp).Hours()).Div(big.NewDecimal(24.0))
	years := days.Div(big.NewDecimal(365.00))

	startEquity := ah.Snapshots[0].Equity
	endEquity := ah.Snapshots[ah.LastIndex()].Equity

	if years.IsZero() || startEquity.IsZero() {
		return big.ZERO
	}

	base := endEquity.Sub(startEquity).Div(startEquity).Add(big.ONE).Float()
	exponent := big.ONE.Div(years).Float()

	return big.NewDecimal(math.Pow(base, exponent)).Sub(big.ONE).Mul(big.NewDecimal(100.00))
}

// Calculate the annualized volatility of the account's equity.
func (ah *AccountHistory) AnnualizedVolatility() big.Decimal {
	if len(ah.Snapshots) <= 1 {
		return big.ZERO
	}

	dailyReturns := make([]big.Decimal, len(ah.Snapshots)-1)
	for i := 1; i < len(ah.Snapshots); i++ {
		dailyReturns[i-1] = ah.Snapshots[i].Equity.Div(ah.Snapshots[i-1].Equity).Sub(big.ONE)
	}

	sum := big.NewDecimal(0.0)
	for _, r := range dailyReturns {
		sum = sum.Add(r)
	}
	meanDailyReturn := sum.Div(big.NewFromInt(len(dailyReturns)))

	squaredDeviations := []big.Decimal{}
	for _, r := range dailyReturns {
		deviation := r.Sub(meanDailyReturn)
		squaredDeviations = append(squaredDeviations, deviation.Mul(deviation))
	}

	variance := big.NewDecimal(0.0)
	for _, sd := range squaredDeviations {
		variance = variance.Add(sd)
	}

	variance = variance.Div(big.NewFromInt(len(squaredDeviations) - 1))

	dailyVolatility := variance.Sqrt()

	return dailyVolatility.Mul(big.NewFromInt(252))
}
