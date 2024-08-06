package techan

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/oarkflow/nepse/big"
)

// Candle represents basic market information for a security over a given time period
type Candle struct {
	Period     TimePeriod
	OpenPrice  big.Decimal
	ClosePrice big.Decimal
	MaxPrice   big.Decimal
	MinPrice   big.Decimal
	Volume     big.Decimal
	TradeCount uint
	CTime      time.Time
	Confirm    int
}

var candlePool = sync.Pool{
	New: func() interface{} {
		return &Candle{}
	},
}

func (c *Candle) ReturnToPool() {
	if c == nil {
		return
	}

	c.Volume.ReturnToPool()
	c.MaxPrice.ReturnToPool()
	c.MinPrice.ReturnToPool()
	c.ClosePrice.ReturnToPool()
	c.OpenPrice.ReturnToPool()

	*c = Candle{} //nolint:exhaustivestruct
	candlePool.Put(c)
}

func CandleFromPool() *Candle {
	return candlePool.Get().(*Candle)
}

// NewCandle returns a new *Candle for a given time period
func NewCandle(period TimePeriod) (c *Candle) {
	return &Candle{
		Period:     period,
		OpenPrice:  big.ZERO,
		ClosePrice: big.ZERO,
		MaxPrice:   big.ZERO,
		MinPrice:   big.ZERO,
		Volume:     big.ZERO,
	}
}

// AddTrade adds a trade to this candle. It will determine if the current price is higher or lower than the min or max
// price and increment the tradecount.
func (c *Candle) AddTrade(tradeAmount, tradePrice big.Decimal) {
	if c.OpenPrice.Zero() {
		c.OpenPrice = tradePrice
	}
	c.ClosePrice = tradePrice

	if c.MaxPrice.Zero() {
		c.MaxPrice = tradePrice
	} else if tradePrice.GT(c.MaxPrice) {
		c.MaxPrice = tradePrice
	}

	if c.MinPrice.Zero() {
		c.MinPrice = tradePrice
	} else if tradePrice.LT(c.MinPrice) {
		c.MinPrice = tradePrice
	}

	if c.Volume.Zero() {
		c.Volume = tradeAmount
	} else {
		c.Volume = c.Volume.Add(tradeAmount)
	}

	c.TradeCount++
}

func (c *Candle) String() string {
	return strings.TrimSpace(fmt.Sprintf(
		`
Time:	%s
Open:	%s
Close:	%s
High:	%s
Low:	%s
Volume:	%s
	`,
		c.Period,
		c.OpenPrice.FormattedString(2),
		c.ClosePrice.FormattedString(2),
		c.MaxPrice.FormattedString(2),
		c.MinPrice.FormattedString(2),
		c.Volume.FormattedString(2),
	))
}

// Complete one candle fields with another one.
func (c *Candle) Complete(candle *Candle) {
	if c.Volume.NaN() {
		c.Volume = candle.Volume
	}

	if c.MaxPrice.NaN() {
		c.MaxPrice = candle.MaxPrice
	}

	if c.MinPrice.NaN() {
		c.MinPrice = candle.MinPrice
	}

	if c.ClosePrice.NaN() {
		c.ClosePrice = candle.ClosePrice
	}

	if c.OpenPrice.NaN() {
		c.OpenPrice = candle.OpenPrice
	}
}

func MergeCandle(begin time.Time, dur time.Duration) func(*Candle) Candle {
	var lastCandle *Candle
	return func(c *Candle) Candle {
		if lastCandle == nil {
			if c.Period.Start.After(begin.Add(dur)) {
				panic(fmt.Sprintf("the first candle not between %s and %s", begin, begin.Add(dur)))
			}
			lastCandle = &Candle{
				Period:     NewTimePeriod(begin, dur),
				OpenPrice:  c.OpenPrice,
				ClosePrice: c.ClosePrice,
				MaxPrice:   c.MaxPrice,
				MinPrice:   c.MinPrice,
				Volume:     c.Volume,
				TradeCount: c.TradeCount,
				CTime:      c.CTime,
				Confirm:    0,
			}
			return *lastCandle
		}
		if c.Period.Start.Before(lastCandle.Period.End) {
			//merge into lastCandle
			lastCandle.CTime = c.CTime
			lastCandle.ClosePrice = c.ClosePrice
			lastCandle.Volume = c.Volume.Add(lastCandle.Volume)
			if lastCandle.MaxPrice.LT(c.MaxPrice) {
				lastCandle.MaxPrice = c.MaxPrice
			}
			if lastCandle.MinPrice.GT(c.MinPrice) {
				lastCandle.MinPrice = c.MinPrice
			}
			lastCandle.TradeCount = c.TradeCount + lastCandle.TradeCount
			return *lastCandle
		}
		lastCandle = &Candle{
			Period:     NewTimePeriod(c.Period.Start, dur),
			OpenPrice:  c.OpenPrice,
			ClosePrice: c.ClosePrice,
			MaxPrice:   c.MaxPrice,
			MinPrice:   c.MinPrice,
			Volume:     c.Volume,
			TradeCount: c.TradeCount,
			CTime:      c.CTime,
			Confirm:    0,
		}

		return *lastCandle
	}
}
