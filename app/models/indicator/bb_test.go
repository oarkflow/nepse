package indicator_test

import (
	"github.com/oarkflow/nepse/app/models/indicator"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBBBuyAndSell(t *testing.T) {
	assert := assert.New(t)

	signals := indicator.BBSignals{}
	// when empty
	assert.False(signals.Sell("VOO", 0, 100))
	assert.True(signals.Buy("VOO", 0, 100))

	// when last is BUY
	assert.False(signals.Buy("VOO", 1, 100))
	assert.True(signals.Sell("VOO", 1, 100))

	// when last is SELL
	assert.False(signals.Sell("VOO", 2, 100))
	assert.True(signals.Buy("VOO", 2, 100))
}

func TestBBProfit(t *testing.T) {
	assert := assert.New(t)

	signals := indicator.BBSignals{
		BBSignals: []indicator.BBSignal{
			indicator.BBSignal{
				Symbol: "VOO",
				Time:   0,
				Price:  100,
				Action: indicator.BUY,
			},
			indicator.BBSignal{
				Symbol: "VOO",
				Time:   1,
				Price:  150,
				Action: indicator.SELL,
			},
		},
	}

	// when buy at 100, sell at 150,
	// expected profit is 50
	assert.Equal(50.0, signals.Profit())

	signals.BBSignals = append(signals.BBSignals, indicator.BBSignal{
		Symbol: "VOO", Time: 2, Price: 100, Action: indicator.BUY,
	})

	// when buy at 100, sell at 150, buy at 100,
	// expected profit is 50
	assert.Equal(50.0, signals.Profit())
}
