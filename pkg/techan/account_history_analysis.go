package techan

import (
	"math"
	"os"

	"github.com/schmidthole/big" // updated import path
	"gopkg.in/yaml.v3"
)

// Measures the return for a given period (such as a year, month etc) for analysis.
type ReturnPeriod struct {
	Period      TimePeriod  `yaml:"period"`
	PercentGain big.Decimal `yaml:"percent_gain"`
	TotalProfit big.Decimal `yaml:"total_profit"`
}

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

// Returns broken down by month
func (ah *AccountHistory) MonthlyPercentGains() []ReturnPeriod {
	monthlyReturns := []ReturnPeriod{}

	if len(ah.Snapshots) < 2 {
		return monthlyReturns
	}

	startIndex := 0

	lastIndex := ah.LastIndex()
	for i := 0; i <= lastIndex; i++ {
		startPeriod := ah.Snapshots[startIndex].Period.Start
		startEquity := ah.Snapshots[startIndex].Equity

		currentPeriod := ah.Snapshots[i].Period.Start
		if startPeriod.Month() != currentPeriod.Month() {
			endPeriod := ah.Snapshots[i-1].Period.Start
			endEquity := ah.Snapshots[i-1].Equity

			period := TimePeriod{Start: startPeriod, End: endPeriod}
			profit := endEquity.Sub(startEquity)

			var percentGain big.Decimal

			if !startEquity.Zero() {
				percentGain = profit.Div(startEquity)
			} else {
				percentGain = big.ZERO
			}

			monthReturn := ReturnPeriod{
				Period:      period,
				TotalProfit: profit,
				PercentGain: percentGain,
			}
			monthlyReturns = append(monthlyReturns, monthReturn)

			startIndex = i
		} else if i == lastIndex {
			endPeriod := ah.Snapshots[i].Period.Start
			endEquity := ah.Snapshots[i].Equity

			period := TimePeriod{Start: startPeriod, End: endPeriod}
			profit := endEquity.Sub(startEquity)

			var percentGain big.Decimal

			if !startEquity.Zero() {
				percentGain = profit.Div(startEquity)
			} else {
				percentGain = big.ZERO
			}

			monthReturn := ReturnPeriod{
				Period:      period,
				TotalProfit: profit,
				PercentGain: percentGain,
			}
			monthlyReturns = append(monthlyReturns, monthReturn)
		}
	}

	return monthlyReturns
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

// Exports the account snapshots in a readable yaml format for viewing and analysis.
func (ah *AccountHistory) ExportSnapshotsYaml(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	for _, snapshot := range ah.Snapshots {
		err = encoder.Encode(snapshot)
		if err != nil {
			return err
		}
	}

	return nil
}

// Exports an analysis summary to yaml for viewing and analysis.
func (ah *AccountHistory) ExportAnalysisSummaryYaml(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	analysis := map[string]interface{}{
		"start":             ah.Snapshots[0].Period.Start,
		"end":               ah.Snapshots[ah.LastIndex()].Period.Start,
		"total_profit":      ah.TotalProfit(),
		"percent_gain":      ah.PercentGain(),
		"annualized_return": ah.AnnualizedReturn(),
	}

	err = encoder.Encode(analysis)
	if err != nil {
		return err
	}

	return nil
}

// Exports the monthly gains to yaml for viewing and analysis.
func (ah *AccountHistory) ExportMonthlyGainsYaml(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	monthlyReturns := ah.MonthlyPercentGains()
	err = encoder.Encode(monthlyReturns)
	if err != nil {
		return err
	}

	return nil
}
