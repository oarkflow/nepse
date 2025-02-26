package nepse

import (
	"fmt"
	"math"
	"sort"

	"github.com/schmidthole/big"

	"github.com/oarkflow/trading/pkg/nepse/models"
	"github.com/oarkflow/trading/pkg/techan"
)

// generateShortTermSignal uses SMA and RSI to provide a BUY/SELL/HOLD signal.
func generateShortTermSignal(ts *techan.TimeSeries, smaPeriod, rsiPeriod int, rsiThreshold float64) (string, string) {
	if len(ts.Candles) < smaPeriod {
		return "HOLD", "Not enough candles for short-term signal"
	}
	threshold := big.NewDecimal(rsiThreshold)
	closePriceIndicator := techan.NewClosePriceIndicator(ts)
	smaIndicator := techan.NewSimpleMovingAverage(closePriceIndicator, smaPeriod)
	rsiIndicator := techan.NewRelativeStrengthIndexIndicator(closePriceIndicator, rsiPeriod)
	lastIndex := ts.LastIndex()
	lastPrice := closePriceIndicator.Calculate(lastIndex)
	sma := smaIndicator.Calculate(lastIndex)
	rsi := rsiIndicator.Calculate(lastIndex)

	if lastPrice.GT(sma) && rsi.GTE(threshold) {
		return "BUY", fmt.Sprintf("Price (%.2f) is above SMA (%.2f) and RSI (%.2f) meets/exceeds threshold (%.2f)", lastPrice.Float(), sma.Float(), rsi.Float(), rsiThreshold)
	} else if lastPrice.LT(sma) && rsi.LTE(threshold) {
		return "SELL", fmt.Sprintf("Price (%.2f) is below SMA (%.2f) and RSI (%.2f) meets/is below threshold (%.2f)", lastPrice.Float(), sma.Float(), rsi.Float(), rsiThreshold)
	}
	return "HOLD", "SMA and RSI do not provide a clear signal"
}

// generateMACDSignal compares MACD to its signal line.
func generateMACDSignal(ts *techan.TimeSeries, shortPeriod, longPeriod, signalPeriod int) (string, string) {
	if len(ts.Candles) < longPeriod {
		return "HOLD", "Not enough candles for MACD calculation"
	}
	closePriceIndicator := techan.NewClosePriceIndicator(ts)
	macdIndicator := techan.NewMACDIndicator(closePriceIndicator, shortPeriod, longPeriod)
	signalLine := techan.NewEMAIndicator(macdIndicator, signalPeriod)
	lastIndex := ts.LastIndex()
	macdValue := macdIndicator.Calculate(lastIndex)
	signalValue := signalLine.Calculate(lastIndex)
	if macdValue.GT(signalValue) {
		return "BUY", fmt.Sprintf("MACD (%.2f) is above Signal Line (%.2f)", macdValue.Float(), signalValue.Float())
	} else if macdValue.LT(signalValue) {
		return "SELL", fmt.Sprintf("MACD (%.2f) is below Signal Line (%.2f)", macdValue.Float(), signalValue.Float())
	}
	return "HOLD", "MACD and Signal Line are too close to call"
}

// generateBollingerSignal uses Bollinger Bands to generate a signal.
func generateBollingerSignal(ts *techan.TimeSeries, period int, multiplier float64) (string, string) {
	if len(ts.Candles) < period {
		return "HOLD", "Not enough candles for Bollinger Bands"
	}
	upper, _, lower := computeBollingerBands(ts, period, multiplier)
	lastPrice := ts.LastCandle().ClosePrice.Float()
	if lastPrice > upper {
		return "SELL", fmt.Sprintf("Price (%.2f) is above the upper Bollinger Band (%.2f) indicating overbought conditions", lastPrice, upper)
	} else if lastPrice < lower {
		return "BUY", fmt.Sprintf("Price (%.2f) is below the lower Bollinger Band (%.2f) indicating oversold conditions", lastPrice, lower)
	}
	return "HOLD", fmt.Sprintf("Price (%.2f) is within Bollinger Bands (%.2f - %.2f)", lastPrice, lower, upper)
}

// generateATRSignal now returns BUY/SELL/HOLD by considering both ATR value and recent price direction.
func generateATRSignal(ts *techan.TimeSeries, period int, threshold float64) (string, string) {
	// Need one extra candle to compare the price change.
	if len(ts.Candles) < period+1 {
		return "HOLD", "Not enough candles for ATR calculation"
	}
	atrIndicator := techan.NewAverageTrueRangeIndicator(ts, period)
	atrValue := atrIndicator.Calculate(ts.LastIndex()).Float()
	lastCandle := ts.LastCandle()
	prevCandle := ts.Candles[len(ts.Candles)-2]
	priceRising := lastCandle.ClosePrice.Float() > prevCandle.ClosePrice.Float()

	if atrValue > threshold {
		if priceRising {
			return "BUY", fmt.Sprintf("ATR (%.2f) exceeds threshold (%.2f) and price is rising", atrValue, threshold)
		}
		return "SELL", fmt.Sprintf("ATR (%.2f) exceeds threshold (%.2f) and price is falling", atrValue, threshold)
	}
	return "HOLD", fmt.Sprintf("ATR (%.2f) is below threshold (%.2f), indicating low volatility", atrValue, threshold)
}

// generateEMACrossoverSignal returns BUY/SELL when the short EMA crosses over the long EMA.
func generateEMACrossoverSignal(ts *techan.TimeSeries, shortPeriod, longPeriod int) (string, string) {
	if len(ts.Candles) < longPeriod {
		return "HOLD", "Not enough candles for EMA crossover"
	}
	closePriceIndicator := techan.NewClosePriceIndicator(ts)
	shortEMA := techan.NewEMAIndicator(closePriceIndicator, shortPeriod)
	longEMA := techan.NewEMAIndicator(closePriceIndicator, longPeriod)
	lastIndex := ts.LastIndex()
	shortEMAVal := shortEMA.Calculate(lastIndex)
	longEMAVal := longEMA.Calculate(lastIndex)
	if shortEMAVal.GT(longEMAVal) {
		return "BUY", fmt.Sprintf("Short EMA (%.2f) is above Long EMA (%.2f)", shortEMAVal.Float(), longEMAVal.Float())
	} else if shortEMAVal.LT(longEMAVal) {
		return "SELL", fmt.Sprintf("Short EMA (%.2f) is below Long EMA (%.2f)", shortEMAVal.Float(), longEMAVal.Float())
	}
	return "HOLD", "EMAs are nearly equal, no crossover"
}

// computeVWAP calculates the Volume-Weighted Average Price.
func computeVWAP(ts *techan.TimeSeries) float64 {
	var totalPV, totalVolume float64
	for _, candle := range ts.Candles {
		price := candle.ClosePrice.Float()
		vol := candle.Volume.Float()
		totalPV += price * vol
		totalVolume += vol
	}
	if totalVolume == 0 {
		return 0
	}
	return totalPV / totalVolume
}

// computeBollingerBands calculates the upper, middle, and lower bands.
func computeBollingerBands(ts *techan.TimeSeries, period int, multiplier float64) (upper, middle, lower float64) {
	closePriceIndicator := techan.NewClosePriceIndicator(ts)
	sma := techan.NewSimpleMovingAverage(closePriceIndicator, period)
	stdDev := techan.NewStandardDeviationIndicator(closePriceIndicator)
	lastIndex := ts.LastIndex()
	middleVal := sma.Calculate(lastIndex)
	stdDevVal := stdDev.Calculate(lastIndex)
	upperVal := middleVal.Add(stdDevVal.Mul(big.NewDecimal(multiplier)))
	lowerVal := middleVal.Sub(stdDevVal.Mul(big.NewDecimal(multiplier)))
	return upperVal.Float(), middleVal.Float(), lowerVal.Float()
}

// DetermineTrend analyzes two indicators (typically %K and %D for stochastic) to provide a trend signal.
func DetermineTrend(k techan.Indicator, d techan.Indicator, index int) (big.Decimal, big.Decimal, string) {
	percentK := k.Calculate(index)
	percentD := d.Calculate(index)
	signal := "Neutral"
	if percentK.GT(big.NewFromInt(80)) {
		signal = "Overbought (Possible Sell)"
	} else if percentK.LT(big.NewFromInt(20)) {
		signal = "Oversold (Possible Buy)"
	}
	if index > 0 {
		prevK := k.Calculate(index - 1)
		prevD := d.Calculate(index - 1)
		if prevK.LT(prevD) && percentK.GT(percentD) {
			signal = "Bullish (Buy)"
		}
		if prevK.GT(prevD) && percentK.LT(percentD) {
			signal = "Bearish (Sell)"
		}
	}
	return percentK, percentD, signal
}

// computeStochastic calculates the slow stochastic indicator and returns its value with a signal.
func computeStochastic(ts *techan.TimeSeries, period int) (float64, string) {
	fastStoch := techan.NewFastStochasticIndicator(ts, period)
	slowStoch := techan.NewSlowStochasticIndicator(fastStoch, 3)
	_, _, finalSignal := DetermineTrend(fastStoch, slowStoch, ts.LastIndex())
	lastSlowStoch := slowStoch.Calculate(ts.LastIndex())
	return lastSlowStoch.Float(), finalSignal
}

// analyzeBrokerStats computes broker statistics from a slice of transactions.
func analyzeBrokerStats(transactions []models.Transaction) map[string]models.BrokerStats {
	stats := make(map[string]models.BrokerStats)
	for _, tx := range transactions {
		if tx.BuyerMemberId != "" {
			b := stats[tx.BuyerMemberId]
			b.TotalBuy += float64(tx.ContractQuantity)
			b.Net += float64(tx.ContractQuantity)
			stats[tx.BuyerMemberId] = b
		}
		if tx.SellerMemberId != "" {
			s := stats[tx.SellerMemberId]
			s.TotalSell += float64(tx.ContractQuantity)
			s.Net -= float64(tx.ContractQuantity)
			stats[tx.SellerMemberId] = s
		}
	}
	return stats
}

// brokerHoldingsSummary sorts and returns broker stats.
func brokerHoldingsSummary(stats map[string]models.BrokerStats) map[string]models.BrokerStats {
	type pair struct {
		Broker string
		Stats  models.BrokerStats
	}
	var pairs []pair
	for broker, stat := range stats {
		pairs = append(pairs, pair{Broker: broker, Stats: stat})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Stats.Net > pairs[j].Stats.Net
	})
	topList := make(map[string]models.BrokerStats)
	for i := 0; i < len(pairs); i++ {
		topList[pairs[i].Broker] = pairs[i].Stats
	}
	return topList
}

// detectPossibleBoom returns true if a candle shows a high volume and a significant price jump.
func detectPossibleBoom(ts *techan.TimeSeries, volumeThreshold, priceThreshold float64) bool {
	if len(ts.Candles) < 2 {
		return false
	}
	lastCandle := ts.LastCandle()
	previousCandle := ts.Candles[len(ts.Candles)-2]
	priceChangePercent := (lastCandle.ClosePrice.Float() - previousCandle.ClosePrice.Float()) / previousCandle.ClosePrice.Float() * 100
	return lastCandle.Volume.Float() > volumeThreshold && priceChangePercent > priceThreshold
}

// computeAvgRiskReward calculates the average reward-to-risk ratio from a slice of trades.
func computeAvgRiskReward(trades []models.Trade) float64 {
	if len(trades) == 0 {
		return 0
	}
	var sumRatio float64
	for _, t := range trades {
		risk := math.Abs(t.EntryPrice - t.StopLoss)
		reward := math.Abs(t.TakeProfit - t.EntryPrice)
		if risk > 0 {
			sumRatio += reward / risk
		}
	}
	return sumRatio / float64(len(trades))
}

// computeMaxDrawdown calculates the maximum drawdown (in percentage) from the timeseries.
func computeMaxDrawdown(ts *techan.TimeSeries) float64 {
	if len(ts.Candles) == 0 {
		return 0
	}
	peak := ts.Candles[0].ClosePrice.Float()
	maxDD := 0.0
	for _, candle := range ts.Candles {
		price := candle.ClosePrice.Float()
		if price > peak {
			peak = price
		}
		dd := (peak - price) / peak * 100
		if dd > maxDD {
			maxDD = dd
		}
	}
	return maxDD
}

// summarizeTrades is a helper to compute a simple trade summary.
// (This dummy implementation returns total trades, net profit/loss, average P/L, and win rate.)
func summarizeTrades(trades []models.Trade) (int, float64, float64, float64) {
	totalTrades := len(trades)
	netPL := 0.0
	winCount := 0
	for _, trade := range trades {
		netPL += trade.ProfitLossPercent
		if trade.ProfitLossPercent > 0 {
			winCount++
		}
	}
	avgPL := 0.0
	if totalTrades > 0 {
		avgPL = netPL / float64(totalTrades)
	}
	winRate := 0.0
	if totalTrades > 0 {
		winRate = float64(winCount) / float64(totalTrades)
	}
	return totalTrades, netPL, avgPL, winRate
}

// backtestStrategy simulates trades based on the short-term signal (and optionally MACD) over a daily timeseries.
func backtestStrategy(tsDaily *techan.TimeSeries, risk models.RiskParameters, smaPeriod, rsiPeriod int, rsiThreshold float64) []models.Trade {
	var trades []models.Trade
	var ts *techan.TimeSeries
	if len(tsDaily.Candles) >= smaPeriod+2 {
		ts = tsDaily
	} else {
		return trades
	}
	closePriceIndicator := techan.NewClosePriceIndicator(ts)
	smaIndicator := techan.NewSimpleMovingAverage(closePriceIndicator, smaPeriod)
	rsiIndicator := techan.NewRelativeStrengthIndexIndicator(closePriceIndicator, rsiPeriod)
	threshold := big.NewDecimal(rsiThreshold)
	for i := smaPeriod; i < len(ts.Candles)-1; i++ {
		lastPrice := closePriceIndicator.Calculate(i)
		sma := smaIndicator.Calculate(i)
		rsi := rsiIndicator.Calculate(i)
		signal := "HOLD"
		if lastPrice.GT(sma) && rsi.GT(threshold) {
			signal = "BUY"
		} else if lastPrice.LT(sma) && rsi.LT(threshold) {
			signal = "SELL"
		} else {
			// Example fallback using MACD with typical periods (12, 26, 9)
			altSignal, _ := generateMACDSignal(ts, 12, 26, 9)
			if altSignal != "HOLD" {
				signal = altSignal
			}
		}
		if signal == "HOLD" {
			continue
		}
		entryCandle := ts.Candles[i+1]
		entryPrice := entryCandle.OpenPrice.Float()
		var trade models.Trade
		trade.EntryTime = entryCandle.Period.Start
		trade.EntryPrice = entryPrice
		if signal == "BUY" {
			trade.Position = "LONG"
			trade.StopLoss = entryPrice * (1 - risk.StopLossPercent/100)
			trade.TakeProfit = entryPrice * (1 + risk.TakeProfitPercent/100)
		} else if signal == "SELL" {
			trade.Position = "SHORT"
			trade.StopLoss = entryPrice * (1 + risk.StopLossPercent/100)
			trade.TakeProfit = entryPrice * (1 - risk.TakeProfitPercent/100)
		}

		exitPrice := entryPrice
		exitTime := entryCandle.Period.Start
		exited := false
		for j := i + 1; j < len(ts.Candles); j++ {
			candle := ts.Candles[j]
			if trade.Position == "LONG" {
				if candle.MinPrice.Float() <= trade.StopLoss {
					exitPrice = trade.StopLoss
					exitTime = candle.Period.Start
					exited = true
					break
				}
				if candle.MaxPrice.Float() >= trade.TakeProfit {
					exitPrice = trade.TakeProfit
					exitTime = candle.Period.Start
					exited = true
					break
				}
			} else {
				if candle.MaxPrice.Float() >= trade.StopLoss {
					exitPrice = trade.StopLoss
					exitTime = candle.Period.Start
					exited = true
					break
				}
				if candle.MinPrice.Float() <= trade.TakeProfit {
					exitPrice = trade.TakeProfit
					exitTime = candle.Period.Start
					exited = true
					break
				}
			}
		}
		if !exited {
			lastCandle := ts.LastCandle()
			exitPrice = lastCandle.ClosePrice.Float()
			exitTime = lastCandle.Period.Start
		}
		trade.ExitTime = exitTime
		trade.ExitPrice = exitPrice
		if trade.Position == "LONG" {
			trade.ProfitLossPercent = (exitPrice - entryPrice) / entryPrice * 100
		} else {
			trade.ProfitLossPercent = (entryPrice - exitPrice) / entryPrice * 100
		}
		trades = append(trades, trade)
	}
	return trades
}

// computeOBV calculates the On-Balance Volume.
func computeOBV(ts *techan.TimeSeries) []float64 {
	obv := make([]float64, len(ts.Candles))
	if len(ts.Candles) == 0 {
		return obv
	}
	obv[0] = 0
	closePriceIndicator := techan.NewClosePriceIndicator(ts)
	for i := 1; i < len(ts.Candles); i++ {
		currentPrice := closePriceIndicator.Calculate(i).Float()
		previousPrice := closePriceIndicator.Calculate(i - 1).Float()
		volume := ts.Candles[i].Volume.Float()
		if currentPrice > previousPrice {
			obv[i] = obv[i-1] + volume
		} else if currentPrice < previousPrice {
			obv[i] = obv[i-1] - volume
		} else {
			obv[i] = obv[i-1]
		}
	}
	return obv
}

// generateOBVSignal returns a signal based on OBV trends.
func generateOBVSignal(ts *techan.TimeSeries) ([]float64, string, string) {
	obv := computeOBV(ts)
	if len(obv) < 2 {
		return obv, "HOLD", "Insufficient OBV data"
	}
	if obv[len(obv)-1] > obv[len(obv)-2] {
		return obv, "BUY", fmt.Sprintf("OBV increased from %.2f to %.2f", obv[len(obv)-2], obv[len(obv)-1])
	} else if obv[len(obv)-1] < obv[len(obv)-2] {
		return obv, "SELL", fmt.Sprintf("OBV decreased from %.2f to %.2f", obv[len(obv)-2], obv[len(obv)-1])
	}
	return obv, "HOLD", "OBV is stable"
}

// computeAD calculates the Accumulation/Distribution line.
func computeAD(ts *techan.TimeSeries) []float64 {
	ad := make([]float64, len(ts.Candles))
	if len(ts.Candles) == 0 {
		return ad
	}
	for i, candle := range ts.Candles {
		high := candle.MaxPrice.Float()
		low := candle.MinPrice.Float()
		closePrice := candle.ClosePrice.Float()
		volume := candle.Volume.Float()
		var moneyFlowMultiplier float64
		if high != low {
			moneyFlowMultiplier = ((closePrice - low) - (high - closePrice)) / (high - low)
		} else {
			moneyFlowMultiplier = 0
		}
		moneyFlowVolume := moneyFlowMultiplier * volume
		if i == 0 {
			ad[i] = moneyFlowVolume
		} else {
			ad[i] = ad[i-1] + moneyFlowVolume
		}
	}
	return ad
}

// generateADSignal returns a signal based on the A/D line.
func generateADSignal(ts *techan.TimeSeries) (float64, string, string) {
	adValues := computeAD(ts)
	if len(adValues) < 2 {
		return 0, "HOLD", "Insufficient A/D data"
	}
	ad := adValues[len(adValues)-1]
	if adValues[len(adValues)-1] > adValues[len(adValues)-2] {
		return ad, "BUY", fmt.Sprintf("A/D increased from %.2f to %.2f", adValues[len(adValues)-2], adValues[len(adValues)-1])
	} else if adValues[len(adValues)-1] < adValues[len(adValues)-2] {
		return ad, "SELL", fmt.Sprintf("A/D decreased from %.2f to %.2f", adValues[len(adValues)-2], adValues[len(adValues)-1])
	}
	return ad, "HOLD", "A/D is stable"
}

// computeCMF calculates the Chaikin Money Flow.
func computeCMF(ts *techan.TimeSeries, period int) float64 {
	if len(ts.Candles) < period {
		return 0
	}
	var sumMF, sumVolume float64
	for i := len(ts.Candles) - period; i < len(ts.Candles); i++ {
		candle := ts.Candles[i]
		high := candle.MaxPrice.Float()
		low := candle.MinPrice.Float()
		closePrice := candle.ClosePrice.Float()
		volume := candle.Volume.Float()
		var mfMultiplier float64
		if high != low {
			mfMultiplier = ((closePrice - low) - (high - closePrice)) / (high - low)
		} else {
			mfMultiplier = 0
		}
		sumMF += mfMultiplier * volume
		sumVolume += volume
	}
	if sumVolume == 0 {
		return 0
	}
	return sumMF / sumVolume
}

// generateCMFSignal returns a signal based on the CMF value.
func generateCMFSignal(ts *techan.TimeSeries, period int) (float64, string, string) {
	cmf := computeCMF(ts, period)
	threshold := 0.05 // adjust sensitivity as needed
	if cmf > threshold {
		return cmf, "BUY", fmt.Sprintf("CMF is positive (%.4f) exceeding threshold (%.2f), indicating buying pressure", cmf, threshold)
	} else if cmf < -threshold {
		return cmf, "SELL", fmt.Sprintf("CMF is negative (%.4f) below threshold (-%.2f), indicating selling pressure", cmf, threshold)
	}
	// If near neutral, use recent price momentum as a tiebreaker.
	if len(ts.Candles) >= 2 {
		lastCandle := ts.LastCandle()
		prevCandle := ts.Candles[len(ts.Candles)-2]
		if lastCandle.ClosePrice.Float() > prevCandle.ClosePrice.Float() {
			return cmf, "BUY", fmt.Sprintf("CMF is neutral (%.4f) but price is rising", cmf)
		} else if lastCandle.ClosePrice.Float() < prevCandle.ClosePrice.Float() {
			return cmf, "SELL", fmt.Sprintf("CMF is neutral (%.4f) but price is falling", cmf)
		}
	}
	return cmf, "HOLD", fmt.Sprintf("CMF is neutral (%.4f)", cmf)
}

// computeOrderImbalanceFromTransactions calculates order imbalance from transaction data.
func computeOrderImbalanceFromTransactions(transactions []models.Transaction) float64 {
	var totalBuy, totalSell float64
	for _, tx := range transactions {
		if tx.BuyerMemberId != "" {
			totalBuy += float64(tx.ContractQuantity)
		}
		if tx.SellerMemberId != "" {
			totalSell += float64(tx.ContractQuantity)
		}
	}
	denom := totalBuy + totalSell
	if denom == 0 {
		return 0
	}
	return (totalBuy - totalSell) / denom
}

// generateOrderImbalanceSignal returns a signal based on order imbalance.
func generateOrderImbalanceSignal(transactions []models.Transaction) (float64, string, string) {
	imbalance := computeOrderImbalanceFromTransactions(transactions)
	threshold := 0.01 // lower threshold for sensitivity
	if imbalance > threshold {
		return imbalance, "BUY", fmt.Sprintf("Order imbalance is %.2f, indicating net buying", imbalance)
	} else if imbalance < -threshold {
		return imbalance, "SELL", fmt.Sprintf("Order imbalance is %.2f, indicating net selling", imbalance)
	}
	return imbalance, "HOLD", fmt.Sprintf("Order imbalance is %.2f, market is balanced", imbalance)
}

// detectCandlestickPattern checks for simple patterns like Doji or Hammer.
func detectCandlestickPattern(ts *techan.TimeSeries) string {
	if len(ts.Candles) == 0 {
		return "No Data"
	}
	candle := ts.LastCandle()
	openPrice := candle.OpenPrice.Float()
	closePrice := candle.ClosePrice.Float()
	highPrice := candle.MaxPrice.Float()
	lowPrice := candle.MinPrice.Float()
	bodySize := math.Abs(closePrice - openPrice)
	rangeSize := highPrice - lowPrice
	if rangeSize == 0 {
		return "No Pattern"
	}
	// Detect Doji: body is less than 10% of total range.
	if bodySize < 0.1*rangeSize {
		return "Doji"
	}
	// Compute shadows.
	var lowerShadow, upperShadow float64
	if openPrice < closePrice {
		lowerShadow = openPrice - lowPrice
		upperShadow = highPrice - closePrice
	} else {
		lowerShadow = closePrice - lowPrice
		upperShadow = highPrice - openPrice
	}
	// Detect Hammer: long lower shadow and small upper shadow.
	if lowerShadow > 2*bodySize && upperShadow < bodySize {
		return "Hammer"
	}
	if openPrice > closePrice && upperShadow > 2*bodySize && lowerShadow < bodySize {
		return "ShootingStar"
	}
	return "No Clear Pattern"
}

// generateCandlestickSignal returns a signal based on candlestick pattern detection.
func generateCandlestickSignal(ts *techan.TimeSeries) (string, string, string) {
	pattern := detectCandlestickPattern(ts)
	switch pattern {
	case "Doji":
		return pattern, "HOLD", "Doji pattern detected, market indecision"
	case "Hammer":
		return pattern, "BUY", "Hammer pattern detected, potential upward reversal"
	case "ShootingStar":
		return pattern, "SELL", "Shooting Star pattern detected, potential downward reversal"
	default:
		return pattern, "HOLD", "No clear candlestick pattern detected"
	}
}

// detectDivergence compares price and a chosen indicator for divergence.
func detectDivergence(ts *techan.TimeSeries, indicator techan.Indicator) string {
	if len(ts.Candles) < 3 {
		return "Insufficient Data"
	}
	closePriceIndicator := techan.NewClosePriceIndicator(ts)
	lastPriceDiff := closePriceIndicator.Calculate(ts.LastIndex()).Float() - closePriceIndicator.Calculate(ts.LastIndex()-1).Float()
	lastIndicatorDiff := indicator.Calculate(ts.LastIndex()).Float() - indicator.Calculate(ts.LastIndex()-1).Float()
	divergenceThreshold := 0.001
	if lastPriceDiff > 0 && lastIndicatorDiff < -divergenceThreshold {
		return "Bullish Divergence"
	}
	if lastPriceDiff < 0 && lastIndicatorDiff > divergenceThreshold {
		return "Bearish Divergence"
	}
	return "No Divergence"
}

// generateDivergenceSignal returns a signal based on divergence between price and an indicator.
func generateDivergenceSignal(ts *techan.TimeSeries, indicator techan.Indicator) (string, string, string) {
	div := detectDivergence(ts, indicator)
	if div == "Bullish Divergence" {
		return div, "BUY", "Bullish divergence detected between price and indicator"
	} else if div == "Bearish Divergence" {
		return div, "SELL", "Bearish divergence detected between price and indicator"
	}
	return div, "HOLD", "No divergence detected"
}

// computeMarketProfile builds bins of price versus volume.
func computeMarketProfile(ts *techan.TimeSeries, binCount int) ([]float64, []float64) {
	if len(ts.Candles) == 0 {
		return nil, nil
	}
	minPrice := ts.Candles[0].MinPrice.Float()
	maxPrice := ts.Candles[0].MaxPrice.Float()
	for _, candle := range ts.Candles {
		if candle.MinPrice.Float() < minPrice {
			minPrice = candle.MinPrice.Float()
		}
		if candle.MaxPrice.Float() > maxPrice {
			maxPrice = candle.MaxPrice.Float()
		}
	}
	binWidth := (maxPrice - minPrice) / float64(binCount)
	bins := make([]float64, binCount)
	volumes := make([]float64, binCount)
	for i := 0; i < binCount; i++ {
		bins[i] = minPrice + binWidth*float64(i) + binWidth/2
	}
	for _, candle := range ts.Candles {
		midPrice := candle.ClosePrice.Float()
		binIndex := int((midPrice - minPrice) / binWidth)
		if binIndex >= binCount {
			binIndex = binCount - 1
		}
		volumes[binIndex] += candle.Volume.Float()
	}
	return bins, volumes
}

// generateMarketProfileSignal returns a signal if the price is near the extremes of the market profile.
func generateMarketProfileSignal(ts *techan.TimeSeries, binCount int) ([]float64, []float64, string, string) {
	bins, volumes := computeMarketProfile(ts, binCount)
	if len(bins) == 0 {
		return bins, volumes, "HOLD", "Insufficient market profile data"
	}
	lastPrice := ts.LastCandle().ClosePrice.Float()
	minBin := bins[0]
	maxBin := bins[len(bins)-1]
	rangePrice := maxBin - minBin
	// Define lower and upper thresholds at 20% and 80% of the profile range.
	lowerBound := minBin + 0.2*rangePrice
	upperBound := minBin + 0.8*rangePrice
	if lastPrice <= lowerBound {
		return bins, volumes, "BUY", fmt.Sprintf("Price (%.2f) is in the lower 20%% of market profile (<= %.2f)", lastPrice, lowerBound)
	} else if lastPrice >= upperBound {
		return bins, volumes, "SELL", fmt.Sprintf("Price (%.2f) is in the upper 20%% of market profile (>= %.2f)", lastPrice, upperBound)
	} else {
		// If in the middle, use recent price momentum.
		if len(ts.Candles) < 2 {
			return bins, volumes, "HOLD", "Not enough candles for momentum analysis"
		}
		lastCandle := ts.LastCandle()
		prevCandle := ts.Candles[len(ts.Candles)-2]
		if lastCandle.ClosePrice.Float() > prevCandle.ClosePrice.Float() {
			return bins, volumes, "BUY", fmt.Sprintf("Price (%.2f) is in the middle and rising", lastPrice)
		} else if lastCandle.ClosePrice.Float() < prevCandle.ClosePrice.Float() {
			return bins, volumes, "SELL", fmt.Sprintf("Price (%.2f) is in the middle and falling", lastPrice)
		}
		return bins, volumes, "HOLD", "Price is in the middle with no clear momentum"
	}
}

// optimizeSMAPeriod tests different SMA periods to choose the one with the best backtested profit.
func optimizeSMAPeriod(tsDaily, tsIntraday *techan.TimeSeries, risk models.RiskParameters, minPeriod, maxPeriod int, rsiPeriod int, rsiThreshold float64) int {
	bestPeriod := minPeriod
	bestProfit := -math.MaxFloat64
	for period := minPeriod; period <= maxPeriod; period++ {
		trades := backtestStrategy(tsDaily, risk, period, rsiPeriod, rsiThreshold)
		_, netPL, _, _ := summarizeTrades(trades)
		if netPL > bestProfit {
			bestProfit = netPL
			bestPeriod = period
		}
	}
	return bestPeriod
}

// meanAndStdDev computes the mean and standard deviation of a slice of float64 values.
func meanAndStdDev(data []float64) (mean, stdDev float64) {
	if len(data) == 0 {
		return 0, 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	mean = sum / float64(len(data))
	varianceSum := 0.0
	for _, v := range data {
		varianceSum += (v - mean) * (v - mean)
	}
	stdDev = math.Sqrt(varianceSum / float64(len(data)))
	return
}

// computeSharpeRatio calculates the Sharpe ratio for a series of trades.
func computeSharpeRatio(trades []models.Trade, riskFreeRate float64) float64 {
	if len(trades) == 0 {
		return 0
	}
	var returns []float64
	for _, t := range trades {
		excessReturn := t.ProfitLossPercent - riskFreeRate
		returns = append(returns, excessReturn)
	}
	mean, stdDev := meanAndStdDev(returns)
	if stdDev == 0 {
		return 0
	}
	return mean / stdDev
}

// computeSortinoRatio calculates the Sortino ratio for a series of trades.
func computeSortinoRatio(trades []models.Trade, riskFreeRate float64) float64 {
	if len(trades) == 0 {
		return 0
	}
	var returns, downsideReturns []float64
	for _, t := range trades {
		excessReturn := t.ProfitLossPercent - riskFreeRate
		returns = append(returns, excessReturn)
		if excessReturn < 0 {
			downsideReturns = append(downsideReturns, excessReturn)
		}
	}
	mean, _ := meanAndStdDev(returns)
	_, downsideStd := meanAndStdDev(downsideReturns)
	if downsideStd == 0 {
		return 0
	}
	return mean / downsideStd
}

// computeExpectancy calculates the expectancy of a trading strategy based on historical trades.
func computeExpectancy(trades []models.Trade) float64 {
	if len(trades) == 0 {
		return 0
	}
	var sumWin, sumLoss float64
	var wins, losses int
	for _, t := range trades {
		if t.ProfitLossPercent > 0 {
			sumWin += t.ProfitLossPercent
			wins++
		} else {
			sumLoss += math.Abs(t.ProfitLossPercent)
			losses++
		}
	}
	winRate := float64(wins) / float64(len(trades))
	lossRate := float64(losses) / float64(len(trades))
	avgWin := 0.0
	if wins > 0 {
		avgWin = sumWin / float64(wins)
	}
	avgLoss := 0.0
	if losses > 0 {
		avgLoss = sumLoss / float64(losses)
	}
	return avgWin*winRate - avgLoss*lossRate
}

// fetchSentiment is a stub for obtaining a stock sentiment score.
func fetchSentiment(stock string) float64 {
	return 0.0
}

// fetchMacroIndicator is a stub for retrieving macroeconomic indicators.
func fetchMacroIndicator() map[string]float64 {
	return map[string]float64{
		"GDPGrowth":    2.5,
		"Inflation":    3.2,
		"InterestRate": 5.0,
	}
}

// predictPriceMovement uses a simple heuristic based on the last two candles.
func predictPriceMovement(ts *techan.TimeSeries) string {
	if len(ts.Candles) < 2 {
		return "HOLD"
	}
	lastCandle := ts.LastCandle()
	prevCandle := ts.Candles[len(ts.Candles)-2]
	if lastCandle.ClosePrice.Float() > prevCandle.ClosePrice.Float() {
		return "BUY"
	} else if lastCandle.ClosePrice.Float() < prevCandle.ClosePrice.Float() {
		return "SELL"
	}
	return "HOLD"
}

// generatePredictionSignal returns a signal based on the simple price movement heuristic.
func generatePredictionSignal(ts *techan.TimeSeries) (string, string) {
	signal := predictPriceMovement(ts)
	return signal, fmt.Sprintf("Predicted signal based on simple heuristic: %s", signal)
}

// detectAnomalies checks for unusually high volume (more than 3× average).
func detectAnomalies(ts *techan.TimeSeries) bool {
	if len(ts.Candles) == 0 {
		return false
	}
	totalVolume := computeTotalVolume(ts)
	avgVolume := totalVolume / float64(len(ts.Candles))
	threshold := avgVolume * 3
	for _, candle := range ts.Candles {
		if candle.Volume.Float() > threshold {
			return true
		}
	}
	return false
}

// generateRiskMetricsSignal returns a signal based on risk metrics.
func generateRiskMetricsSignal(sharpe, sortino, expectancy float64) (string, string) {
	if expectancy > 0 {
		return "BUY", fmt.Sprintf("Positive expectancy (%.2f) with Sharpe %.2f and Sortino %.2f", expectancy, sharpe, sortino)
	} else if expectancy < 0 {
		return "SELL", fmt.Sprintf("Negative expectancy (%.2f) with Sharpe %.2f and Sortino %.2f", expectancy, sharpe, sortino)
	}
	return "HOLD", fmt.Sprintf("Neutral expectancy (%.2f)", expectancy)
}

// generateSentimentSignal returns a signal based on a sentiment score.
func generateSentimentSignal(sentiment float64) (string, string) {
	if sentiment > 0.1 {
		return "BUY", fmt.Sprintf("Positive sentiment score of %.2f", sentiment)
	} else if sentiment < -0.1 {
		return "SELL", fmt.Sprintf("Negative sentiment score of %.2f", sentiment)
	}
	return "HOLD", fmt.Sprintf("Neutral sentiment score of %.2f", sentiment)
}

// generateMacroSignal returns a signal based on macroeconomic indicators.
func generateMacroSignal(macro map[string]float64) (string, string) {
	gdp := macro["GDPGrowth"]
	inflation := macro["Inflation"]
	if gdp > 3 && inflation < 4 {
		return "BUY", fmt.Sprintf("Strong GDP growth (%.2f%%) and moderate inflation (%.2f%%)", gdp, inflation)
	} else if inflation > 4 {
		return "SELL", fmt.Sprintf("High inflation (%.2f%%) detected", inflation)
	}
	return "HOLD", fmt.Sprintf("GDP growth (%.2f%%) and inflation (%.2f%%) are moderate", gdp, inflation)
}

// generateAnomalySignal returns an alert if any anomalies are detected in volume.
func generateAnomalySignal(ts *techan.TimeSeries) (string, string) {
	if detectAnomalies(ts) {
		return "ALERT", "Anomalous volume detected"
	}
	return "HOLD", "No anomalies detected"
}
