package main

import (
	"fmt"
	"github.com/oarkflow/nepse/big"
	"github.com/sdcoffey/techan"
	"math"
	"math/rand"
	"strconv"
	"time"
)

var candleIndex int
var mockedTimeSeries = mockTimeSeriesFl(
	64.75, 63.79, 63.73,
	63.73, 63.55, 63.19,
	63.91, 63.85, 62.95,
	63.37, 61.33, 61.51)

func randomTimeSeries(size int) *techan.TimeSeries {
	vals := make([]string, size)
	for i := 0; i < size; i++ {
		val := rand.Float64() * 100
		if i == 0 {
			vals[i] = fmt.Sprint(val)
		} else {
			last, _ := strconv.ParseFloat(vals[i-1], 64)
			if i%2 == 0 {
				vals[i] = fmt.Sprint(last + (val / 10))
			} else {
				vals[i] = fmt.Sprint(last - (val / 10))
			}
		}
	}

	return mockTimeSeries(vals...)
}

func mockTimeSeriesOCHL(values ...[]float64) *techan.TimeSeries {
	ts := techan.NewTimeSeries()
	for i, ochl := range values {
		candle := techan.NewCandle(techan.NewTimePeriod(time.Unix(int64(i), 0), time.Second))
		candle.OpenPrice = big.NewDecimal(ochl[0])
		candle.ClosePrice = big.NewDecimal(ochl[1])
		candle.MaxPrice = big.NewDecimal(ochl[2])
		candle.MinPrice = big.NewDecimal(ochl[3])
		candle.Volume = big.NewDecimal(float64(i))

		ts.AddCandle(candle)
	}

	return ts
}

func mockTimeSeries(values ...string) *techan.TimeSeries {
	ts := techan.NewTimeSeries()
	for _, val := range values {
		candle := techan.NewCandle(techan.NewTimePeriod(time.Unix(int64(candleIndex), 0), time.Second))
		candle.OpenPrice = big.NewFromString(val)
		candle.ClosePrice = big.NewFromString(val)
		candle.MaxPrice = big.NewFromString(val).Add(big.ONE)
		candle.MinPrice = big.NewFromString(val).Sub(big.ONE)
		candle.Volume = big.NewFromString(val)

		ts.AddCandle(candle)

		candleIndex++
	}

	return ts
}

func mockTimeSeriesFl(values ...float64) *techan.TimeSeries {
	strVals := make([]string, len(values))

	for i, val := range values {
		strVals[i] = fmt.Sprint(val)
	}

	return mockTimeSeries(strVals...)
}

func decimalEquals(expected float64, actual big.Decimal) {
	fmt.Println(fmt.Sprintf("%.4f", expected), fmt.Sprintf("%.4f", actual.Float()))
}

func dump(indicator techan.Indicator) (values []float64) {
	precision := 4.0
	m := math.Pow(10, precision)

	defer func() {
		recover()
	}()

	var index int
	for {
		values = append(values, math.Round(indicator.Calculate(index).Float()*m)/m)
		index++
	}
}

func indicatorEquals(expected []float64, indicator techan.Indicator) {
	actualValues := dump(indicator)
	fmt.Println(expected, actualValues)
}

func generateTestTimeSeries() *techan.TimeSeries {
	series := techan.NewTimeSeries()

	start := time.Now().Truncate(time.Minute)
	d := 59 * time.Second

	data := [][]float64{
		// OpenPrice ClosePrice MaxPrice MinPrice Volume
		{28123.2, 28100.6, 28123.2, 28082.8, 396.541},
		{28100.6, 28124.2, 28142, 28100.6, 363.009},
		{28124.2, 28139.4, 28145, 28118.8, 613.82},
		{28139.4, 28183.8, 28183.9, 28139.4, 172.826},
		{28183.8, 28172.9, 28188.5, 28168.2, 458.278},
		{28172.9, 28176.1, 28186.3, 28155.1, 519.877},
		{28176.1, 28223.8, 28227.6, 28176.1, 256.998},
		{28223.8, 28195.1, 28228, 28187.5, 525.257},
		{28195.1, 28176.9, 28220, 28176.9, 498.021},
		{28176.9, 28174.8, 28188, 28152, 460.77},
		{28174.8, 28193, 28208, 28171.3, 472.718},
		{28193, 28175.7, 28205, 28175.7, 408.305},
		{28175.7, 28224.9, 28224.9, 28170.4, 549.905},
		{28224.9, 28252.8, 28252.8, 28210.6, 414.02},
		{28252.8, 28303.9, 28303.9, 28252.8, 331.801},
		{28303.9, 28397, 28433, 28303.9, 251.904},
		{28397, 28307.3, 28397, 28295, 293.59},
		{28307.3, 28249.9, 28307.3, 28233, 240.608},
		{28249.9, 28228.8, 28256.6, 28200, 412.379},
		{28228.8, 28185, 28228.8, 28185, 418.372},
		{28185, 28210, 28213, 28167.2, 360.216},
		{28210, 28168.1, 28239.4, 28168.1, 472.133},
		{28168.1, 28188, 28200, 28165, 483.285},
		{28188, 30754.6, 31100, 25734, 5995.132},
		{30754.6, 28217.8, 32400, 28195.9, 1556.42},
		{28217.8, 28217.7, 28217.8, 28182, 436.21},
		{28217.7, 28251.8, 28260.1, 28217.6, 431.981},
		{28251.8, 28208, 28257.1, 28208, 352.108},
		{28208, 28187.9, 28208, 28175, 347.064},
		{28187.9, 28206.1, 28207, 28176.4, 501.79},
		{28206.1, 28223.9, 28225, 28199.6, 429.412},
		{28223.9, 28241.8, 28244.4, 28215, 348.231},
		{28241.8, 28255, 28286.4, 28241.8, 449.594},
		{28255, 28296.4, 28303, 28236, 446.123},
		{28296.4, 28291, 28301.1, 28272.2, 562.012},
		{28291, 28311.1, 28330.8, 28285.8, 523.107},
		{28311.1, 28327.6, 28346, 28307.2, 347.319},
		{28327.6, 28306.1, 28333, 28295.3, 432.015},
		{28306.1, 28283, 28314.3, 28283, 529.542},
		{28283, 28293.1, 28293.1, 28260.1, 327.786},
		{28293.1, 28349.4, 28350, 28293.1, 625.926},
		{28349.4, 28365.9, 28368, 28336, 453.088},
		{28365.9, 28431.4, 28440, 28361.5, 117.198},
		{28431.4, 28384.7, 28444.6, 28363, 208.805},
		{28384.7, 28293, 28388.6, 28293, 128.871},
		{28293, 28222.6, 28293, 28215, 328.041},
		{28222.6, 28242, 28266.1, 28222.6, 500.92},
		{28242, 28222.1, 28243.1, 28208, 399.99},
		{28222.1, 28247.2, 28250, 28222.1, 566.124},
		{28247.2, 28254.5, 28259.9, 28232.7, 538.374},
		{28254.5, 28258.4, 28271.2, 28252, 571.833},
		{28258.4, 28215, 28258.4, 28215, 426.631},
		{28215, 28235.4, 28240, 28214.6, 548.354},
	}

	candles := make([]techan.Candle, len(data))
	for i, values := range data {
		period := techan.NewTimePeriod(start.Add(time.Duration(i)*time.Minute), d)
		candles[i] = *techan.NewCandle(period)
		candles[i].OpenPrice = big.NewDecimal(values[0])
		candles[i].ClosePrice = big.NewDecimal(values[1])
		candles[i].MaxPrice = big.NewDecimal(values[2])
		candles[i].MinPrice = big.NewDecimal(values[3])
		candles[i].Volume = big.NewDecimal(values[4])
	}

	for i := range candles {
		series.AddCandle(&candles[i])
	}
	return series
}
