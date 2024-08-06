package techan

import (
	"github.com/oarkflow/nepse/big"
	"testing"
	"time"
)

var startingEquity = big.NewDecimal(10000.00)

func createTestTimeSeries(t *testing.T) *TimeSeries {
	t.Helper()

	series := NewTimeSeries()

	dataset := [][]string{
		// Timestamp, Open, Close, High, Low, volume
		{"2020-01-01", "100.00", "102.50", "105.00", "98.50", "1000"},
		{"2020-01-02", "105.75", "108.00", "108.00", "103.25", "1200"},
		{"2020-01-03", "103.50", "107.25", "110.50", "103.50", "1500"},
		{"2020-01-04", "107.25", "109.00", "113.50", "106.75", "1800"},
		{"2020-01-05", "108.00", "111.50", "116.25", "109.50", "2000"},
		{"2020-01-06", "112.00", "109.00", "119.00", "111.75", "1600"},
		{"2020-01-07", "108.50", "112.00", "123.25", "115.75", "1700"},
		{"2020-01-08", "111.00", "113.25", "126.00", "118.75", "1400"}, // entry
		{"2020-01-09", "112.25", "115.00", "130.00", "120.50", "1800"},
		{"2020-01-10", "115.00", "112.75", "134.25", "124.50", "2000"}, // exit
		{"2020-01-11", "112.75", "111.00", "138.00", "127.75", "2100"},
		{"2020-01-12", "116.00", "111.75", "142.00", "131.25", "2200"},
		{"2020-01-13", "114.75", "112.25", "146.50", "134.75", "2300"},
		{"2020-01-14", "117.25", "115.50", "150.00", "138.75", "2100"}, // entry
		{"2020-01-15", "115.50", "118.25", "153.75", "142.50", "1900"},
		{"2020-01-16", "118.25", "116.00", "158.00", "146.75", "1800"},
		{"2020-01-17", "116.00", "114.75", "162.00", "150.25", "1700"},
		{"2020-01-18", "119.75", "113.50", "165.00", "153.25", "1600"},
		{"2020-01-19", "117.50", "112.00", "169.25", "157.50", "1500"}, // exit
		{"2020-01-20", "121.00", "110.75", "172.50", "161.25", "1400"},
		{"2020-01-21", "118.75", "102.25", "176.00", "164.75", "1300"},
		{"2020-01-22", "122.25", "100.50", "179.50", "168.75", "1200"},
		{"2020-01-23", "120.50", "103.25", "183.25", "172.75", "1100"},
		{"2020-01-24", "123.25", "101.00", "186.75", "176.50", "1000"},
		{"2020-01-25", "121.00", "104.50", "190.00", "180.25", "900"},
		{"2020-01-26", "124.50", "102.75", "193.50", "183.50", "800"},
		{"2020-01-27", "122.75", "106.00", "197.25", "186.75", "700"},
		{"2020-01-28", "126.00", "103.75", "200.50", "191.75", "600"},
		{"2020-01-29", "123.75", "107.50", "203.75", "195.50", "500"},
		{"2020-01-30", "127.50", "105.25", "207.00", "199.25", "400"},
	}

	for _, datum := range dataset {
		timestamp, err := time.Parse("2006-01-02", datum[0])
		if err != nil {
			panic("failed to parse test data")
		}

		period := NewTimePeriod(timestamp, time.Hour*24)

		candle := NewCandle(period)
		candle.OpenPrice = big.NewFromString(datum[1])
		candle.ClosePrice = big.NewFromString(datum[2])
		candle.MaxPrice = big.NewFromString(datum[3])
		candle.MinPrice = big.NewFromString(datum[4])
		candle.Volume = big.NewFromString(datum[5])

		series.AddCandle(candle)
	}

	return series
}

func createTestPriceIndicator(t *testing.T, series *TimeSeries) Indicator {
	t.Helper()

	return NewClosePriceIndicator(series)
}

func createTestStrategy(t *testing.T, indicator Indicator) Strategy {
	t.Helper()

	entryConstant := NewConstantIndicator(113.00)
	exitConstant := NewConstantIndicator(113.00)

	entryRule := And(
		NewCrossUpIndicatorRule(entryConstant, indicator, 100),
		PositionNewRule{}) // Is satisfied when the price ema moves above 30 and the current position is new

	exitRule := And(
		NewCrossDownIndicatorRule(indicator, exitConstant, 100),
		PositionOpenRule{}) // Is satisfied when the price ema moves below 10 and the current position is open

	strategy := RuleStrategy{
		UnstablePeriod: 2,
		EntryRule:      entryRule,
		ExitRule:       exitRule,
	}

	return strategy
}

func createTestOrderPlan(t *testing.T) OrderPlan {
	t.Helper()

	return OrderPlan{
		Side:          BUY,
		PercentEquity: big.NewFromString("100.0"),
	}
}

func TestNewFixedEntryBacktest(t *testing.T) {
	ts := createTestTimeSeries(t)
	priceInd := createTestPriceIndicator(t, ts)
	strat := createTestStrategy(t, priceInd)
	op := createTestOrderPlan(t)

	bt := NewFixedEntryBacktest("TEST", ts, priceInd, strat, op)

	endingEquity, tradeRec := bt.Run(startingEquity)

	if endingEquity.GT(big.NewDecimal(9700.00)) {
		t.Errorf("ending equity greater than estimated. expected: <9700, got: %v", endingEquity)
	}

	if len(tradeRec.Trades) != 2 {
		t.Errorf("expected 2 trades, found %v", len(tradeRec.Trades))
	}
}
