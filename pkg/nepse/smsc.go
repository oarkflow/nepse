package nepse

import (
	"fmt"
	"math"
	"strings"

	"github.com/oarkflow/trading/pkg/techan"
)

type SMCAnalyzer struct {
	period int
}

func NewSMCAnalyzer(period int) *SMCAnalyzer {
	if period < 1 {
		period = smcPeriod
	}
	return &SMCAnalyzer{period: period}
}

func (a *SMCAnalyzer) detectBOS(ts *techan.TimeSeries) (int, string) {
	if len(ts.Candles) < a.period {
		return 0, "Not enough candles for BOS analysis"
	}
	recentHigh := ts.Candles[len(ts.Candles)-a.period].MaxPrice.Float()
	recentLow := ts.Candles[len(ts.Candles)-a.period].MinPrice.Float()
	for i := len(ts.Candles) - a.period; i < len(ts.Candles); i++ {
		if ts.Candles[i].MaxPrice.Float() > recentHigh {
			recentHigh = ts.Candles[i].MaxPrice.Float()
		}
		if ts.Candles[i].MinPrice.Float() < recentLow {
			recentLow = ts.Candles[i].MinPrice.Float()
		}
	}
	currentClose := ts.LastCandle().ClosePrice.Float()
	if currentClose > recentHigh {
		return 1, fmt.Sprintf("BOS: Price (%.2f) broke above recent high (%.2f)", currentClose, recentHigh)
	} else if currentClose < recentLow {
		return -1, fmt.Sprintf("BOS: Price (%.2f) broke below recent low (%.2f)", currentClose, recentLow)
	}
	return 0, ""
}

func (a *SMCAnalyzer) detectCHOCH(ts *techan.TimeSeries) (int, string) {
	if len(ts.Candles) < 2 {
		return 0, "Not enough candles for CHOCH analysis"
	}
	currentCandle := ts.LastCandle()
	prevCandle := ts.Candles[len(ts.Candles)-2]
	if currentCandle.ClosePrice.Float() > prevCandle.ClosePrice.Float() &&
		prevCandle.ClosePrice.Float() < prevCandle.OpenPrice.Float() {
		return 1, "CHOCH: Transition from bearish to bullish detected"
	} else if currentCandle.ClosePrice.Float() < prevCandle.ClosePrice.Float() &&
		prevCandle.ClosePrice.Float() > prevCandle.OpenPrice.Float() {
		return -1, "CHOCH: Transition from bullish to bearish detected"
	}
	return 0, ""
}

func (a *SMCAnalyzer) detectFVG(ts *techan.TimeSeries) (int, string) {
	if len(ts.Candles) < 2 {
		return 0, "Not enough candles for FVG analysis"
	}
	currentCandle := ts.LastCandle()
	prevCandle := ts.Candles[len(ts.Candles)-2]
	gap := math.Abs(currentCandle.OpenPrice.Float() - prevCandle.ClosePrice.Float())
	avgRange := math.Abs(prevCandle.MaxPrice.Float() - prevCandle.MinPrice.Float())
	if avgRange > 0 && gap > 0.5*avgRange {
		if currentCandle.OpenPrice.Float() > prevCandle.ClosePrice.Float() {
			return 1, "FVG: Significant upward gap detected"
		}
		return -1, "FVG: Significant downward gap detected"
	}
	return 0, ""
}

func (a *SMCAnalyzer) detectOrderBlock(ts *techan.TimeSeries) (int, string) {
	if len(ts.Candles) < a.period {
		return 0, "Not enough candles for Order Block analysis"
	}
	var sumBody float64
	for i := len(ts.Candles) - a.period; i < len(ts.Candles); i++ {
		body := math.Abs(ts.Candles[i].ClosePrice.Float() - ts.Candles[i].OpenPrice.Float())
		sumBody += body
	}
	avgBody := sumBody / float64(a.period)
	vote := 0
	var messages []string
	for i := len(ts.Candles) - a.period; i < len(ts.Candles); i++ {
		body := math.Abs(ts.Candles[i].ClosePrice.Float() - ts.Candles[i].OpenPrice.Float())
		if body > 1.5*avgBody {
			if ts.Candles[i].ClosePrice.Float() > ts.Candles[i].OpenPrice.Float() {
				vote++
				messages = append(messages, fmt.Sprintf("OB: Bullish order block at candle %d", i))
			} else {
				vote--
				messages = append(messages, fmt.Sprintf("OB: Bearish order block at candle %d", i))
			}
		}
	}
	if vote > 0 {
		return 1, strings.Join(messages, "; ")
	} else if vote < 0 {
		return -1, strings.Join(messages, "; ")
	}
	return 0, ""
}

func (a *SMCAnalyzer) detectLiquiditySweep(ts *techan.TimeSeries) (int, string) {
	currentCandle := ts.LastCandle()
	bodySize := math.Abs(currentCandle.ClosePrice.Float() - currentCandle.OpenPrice.Float())
	wickSize := math.Abs(currentCandle.ClosePrice.Float() - currentCandle.MinPrice.Float())
	if bodySize == 0 {
		return 0, "No body size to evaluate liquidity sweep"
	}
	if wickSize > 2*bodySize {
		if currentCandle.ClosePrice.Float()-currentCandle.OpenPrice.Float() < 0.1*bodySize {
			return 1, "Liquidity Sweep: Lower wick rejection suggests upward reversal"
		} else if currentCandle.OpenPrice.Float()-currentCandle.ClosePrice.Float() < 0.1*bodySize {
			return -1, "Liquidity Sweep: Upper wick rejection suggests downward reversal"
		}
	}
	return 0, ""
}

func (a *SMCAnalyzer) detectStopRun(ts *techan.TimeSeries) (int, string) {
	if len(ts.Candles) < a.period+1 {
		return 0, "Not enough candles for stop-run detection"
	}
	currentCandle := ts.LastCandle()

	var totalVolume float64
	for i := len(ts.Candles) - a.period - 1; i < len(ts.Candles)-1; i++ {
		totalVolume += ts.Candles[i].Volume.Float()
	}
	avgVolume := totalVolume / float64(a.period)
	if avgVolume == 0 {
		return 0, "Average volume zero, cannot detect stop-run"
	}
	if currentCandle.Volume.Float() > stopRunVolumeMultiplier*avgVolume {
		prevCandle := ts.Candles[len(ts.Candles)-2]
		gapPercent := (currentCandle.OpenPrice.Float() - prevCandle.ClosePrice.Float()) / prevCandle.ClosePrice.Float()
		if math.Abs(gapPercent) > stopRunGapThreshold {
			if gapPercent > 0 {
				return 1, fmt.Sprintf("Stop-Run: Upward gap of %.2f%% with high volume", gapPercent*100)
			}
			return -1, fmt.Sprintf("Stop-Run: Downward gap of %.2f%% with high volume", math.Abs(gapPercent)*100)
		}
	}
	return 0, ""
}

func (a *SMCAnalyzer) detectVolumeImbalance(ts *techan.TimeSeries) (int, string) {
	if len(ts.Candles) < a.period+1 {
		return 0, "Not enough candles for volume imbalance detection"
	}
	currentCandle := ts.LastCandle()
	var totalVolume float64
	for i := len(ts.Candles) - a.period - 1; i < len(ts.Candles)-1; i++ {
		totalVolume += ts.Candles[i].Volume.Float()
	}
	avgVolume := totalVolume / float64(a.period)
	if currentCandle.Volume.Float() > volumeImbalanceMultiplier*avgVolume {
		if currentCandle.ClosePrice.Float() > currentCandle.OpenPrice.Float() {
			return 1, "Volume Imbalance: High volume bullish candle detected"
		}
		return -1, "Volume Imbalance: High volume bearish candle detected"
	}
	return 0, ""
}

func (a *SMCAnalyzer) detectCandlestickPatterns(ts *techan.TimeSeries) (int, string) {
	if len(ts.Candles) < 2 {
		return 0, "Not enough candles for candlestick pattern detection"
	}
	current := ts.LastCandle()
	prev := ts.Candles[len(ts.Candles)-2]
	bodyCurrent := math.Abs(current.ClosePrice.Float() - current.OpenPrice.Float())
	bodyPrev := math.Abs(prev.ClosePrice.Float() - prev.OpenPrice.Float())

	if prev.ClosePrice.Float() < prev.OpenPrice.Float() &&
		current.ClosePrice.Float() > current.OpenPrice.Float() &&
		bodyCurrent > bodyPrev &&
		current.OpenPrice.Float() < prev.ClosePrice.Float() &&
		current.ClosePrice.Float() > prev.OpenPrice.Float() {
		return 1, "Bullish Engulfing pattern detected"
	}

	if prev.ClosePrice.Float() > prev.OpenPrice.Float() &&
		current.ClosePrice.Float() < current.OpenPrice.Float() &&
		bodyCurrent > bodyPrev &&
		current.OpenPrice.Float() > prev.ClosePrice.Float() &&
		current.ClosePrice.Float() < prev.OpenPrice.Float() {
		return -1, "Bearish Engulfing pattern detected"
	}
	return 0, ""
}

func (a *SMCAnalyzer) detectAccumulationDistribution(ts *techan.TimeSeries) (int, string) {
	current := ts.LastCandle()
	high := current.MaxPrice.Float()
	low := current.MinPrice.Float()
	closePrice := current.ClosePrice.Float()
	if high == low {
		return 0, "No price range for A/D calculation"
	}
	mfMultiplier := ((closePrice - low) - (high - closePrice)) / (high - low)
	mfVolume := mfMultiplier * current.Volume.Float()
	if mfVolume > adThresholdFraction*current.Volume.Float() {
		return 1, "Accumulation detected via A/D indicator"
	} else if mfVolume < -adThresholdFraction*current.Volume.Float() {
		return -1, "Distribution detected via A/D indicator"
	}
	return 0, ""
}

func (a *SMCAnalyzer) detectVWAPCluster(ts *techan.TimeSeries) (int, string) {
	vwap := computeVWAP(ts)
	currentClose := ts.LastCandle().ClosePrice.Float()
	diff := (currentClose - vwap) / vwap
	if diff > vwapDiffThreshold {
		return 1, fmt.Sprintf("Price is %.2f%% above VWAP (%.2f)", diff*100, vwap)
	} else if diff < -vwapDiffThreshold {
		return -1, fmt.Sprintf("Price is %.2f%% below VWAP (%.2f)", math.Abs(diff)*100, vwap)
	}
	return 0, ""
}

func (a *SMCAnalyzer) detectDivergence(ts *techan.TimeSeries) (int, string) {
	if len(ts.Candles) < 5 {
		return 0, "Not enough candles for divergence detection"
	}
	closePriceIndicator := techan.NewClosePriceIndicator(ts)
	rsiIndicator := techan.NewRelativeStrengthIndexIndicator(closePriceIndicator, rsiPeriod)

	var lowIndexes []int
	for i := 1; i < len(ts.Candles)-1; i++ {
		if ts.Candles[i].ClosePrice.Float() < ts.Candles[i-1].ClosePrice.Float() &&
			ts.Candles[i].ClosePrice.Float() < ts.Candles[i+1].ClosePrice.Float() {
			lowIndexes = append(lowIndexes, i)
		}
	}
	if len(lowIndexes) >= 2 {
		i1 := lowIndexes[len(lowIndexes)-2]
		i2 := lowIndexes[len(lowIndexes)-1]
		priceLow1 := closePriceIndicator.Calculate(i1).Float()
		priceLow2 := closePriceIndicator.Calculate(i2).Float()
		rsiLow1 := rsiIndicator.Calculate(i1).Float()
		rsiLow2 := rsiIndicator.Calculate(i2).Float()
		if priceLow2 < priceLow1 && rsiLow2 > rsiLow1 {
			return 1, fmt.Sprintf("Bullish divergence: Price lows from %.2f to %.2f, RSI lows from %.2f to %.2f", priceLow1, priceLow2, rsiLow1, rsiLow2)
		}
	}

	var highIndexes []int
	for i := 1; i < len(ts.Candles)-1; i++ {
		if ts.Candles[i].ClosePrice.Float() > ts.Candles[i-1].ClosePrice.Float() &&
			ts.Candles[i].ClosePrice.Float() > ts.Candles[i+1].ClosePrice.Float() {
			highIndexes = append(highIndexes, i)
		}
	}
	if len(highIndexes) >= 2 {
		j1 := highIndexes[len(highIndexes)-2]
		j2 := highIndexes[len(highIndexes)-1]
		priceHigh1 := closePriceIndicator.Calculate(j1).Float()
		priceHigh2 := closePriceIndicator.Calculate(j2).Float()
		rsiHigh1 := rsiIndicator.Calculate(j1).Float()
		rsiHigh2 := rsiIndicator.Calculate(j2).Float()
		if priceHigh2 > priceHigh1 && rsiHigh2 < rsiHigh1 {
			return -1, fmt.Sprintf("Bearish divergence: Price highs from %.2f to %.2f, RSI highs from %.2f to %.2f", priceHigh1, priceHigh2, rsiHigh1, rsiHigh2)
		}
	}

	return 0, ""
}

func (a *SMCAnalyzer) detectMarketProfile(ts *techan.TimeSeries) (int, string) {
	if len(ts.Candles) < a.period {
		return 0, "Not enough candles for market profile detection"
	}
	minPrice := ts.Candles[len(ts.Candles)-a.period].MinPrice.Float()
	maxPrice := ts.Candles[len(ts.Candles)-a.period].MaxPrice.Float()
	for i := len(ts.Candles) - a.period; i < len(ts.Candles); i++ {
		if ts.Candles[i].MinPrice.Float() < minPrice {
			minPrice = ts.Candles[i].MinPrice.Float()
		}
		if ts.Candles[i].MaxPrice.Float() > maxPrice {
			maxPrice = ts.Candles[i].MaxPrice.Float()
		}
	}
	binSize := (maxPrice - minPrice) / float64(marketProfileBinCount)
	if binSize == 0 {
		return 0, "Price range is zero for market profile"
	}
	volumeProfile := make([]float64, marketProfileBinCount)
	for i := len(ts.Candles) - a.period; i < len(ts.Candles); i++ {
		price := ts.Candles[i].ClosePrice.Float()
		binIndex := int((price - minPrice) / binSize)
		if binIndex >= marketProfileBinCount {
			binIndex = marketProfileBinCount - 1
		}
		volumeProfile[binIndex] += ts.Candles[i].Volume.Float()
	}
	pocIndex := 0
	maxVolume := volumeProfile[0]
	for i, vol := range volumeProfile {
		if vol > maxVolume {
			maxVolume = vol
			pocIndex = i
		}
	}
	pocPrice := minPrice + binSize*float64(pocIndex) + binSize/2.0
	currentClose := ts.LastCandle().ClosePrice.Float()
	diffPercent := math.Abs(currentClose-pocPrice) / pocPrice
	if diffPercent < 0.01 {
		return 0, fmt.Sprintf("Market Profile: Price near Point of Control (%.2f)", pocPrice)
	} else if currentClose > pocPrice {
		return 1, fmt.Sprintf("Market Profile: Price above Point of Control (%.2f)", pocPrice)
	}
	return -1, fmt.Sprintf("Market Profile: Price below Point of Control (%.2f)", pocPrice)
}

func (a *SMCAnalyzer) detectTrendStrength(ts *techan.TimeSeries) (int, string) {
	if len(ts.Candles) < a.period {
		return 0, "Not enough candles for trend strength detection"
	}
	sum := 0.0
	for i := len(ts.Candles) - a.period; i < len(ts.Candles); i++ {
		sum += ts.Candles[i].ClosePrice.Float()
	}
	sma := sum / float64(a.period)
	current := ts.LastCandle().ClosePrice.Float()
	diff := current - sma

	if diff > 0.01*sma {
		return 1, "Trend Strength: Uptrend confirmed"
	} else if diff < -0.01*sma {
		return -1, "Trend Strength: Downtrend confirmed"
	}
	return 0, ""
}

func (a *SMCAnalyzer) AnalyzeSMC(ts *techan.TimeSeries) (string, string) {
	var signals []int
	var reasons []string

	if sig, msg := a.detectBOS(ts); sig != 0 {
		signals = append(signals, sig)
		reasons = append(reasons, msg)
	}
	if sig, msg := a.detectCHOCH(ts); sig != 0 {
		signals = append(signals, sig)
		reasons = append(reasons, msg)
	}
	if sig, msg := a.detectFVG(ts); sig != 0 {
		signals = append(signals, sig)
		reasons = append(reasons, msg)
	}
	if sig, msg := a.detectOrderBlock(ts); sig != 0 {
		signals = append(signals, sig)
		reasons = append(reasons, msg)
	}
	if sig, msg := a.detectLiquiditySweep(ts); sig != 0 {
		signals = append(signals, sig)
		reasons = append(reasons, msg)
	}
	if sig, msg := a.detectStopRun(ts); sig != 0 {
		signals = append(signals, sig)
		reasons = append(reasons, msg)
	}
	if sig, msg := a.detectVolumeImbalance(ts); sig != 0 {
		signals = append(signals, sig)
		reasons = append(reasons, msg)
	}
	if sig, msg := a.detectCandlestickPatterns(ts); sig != 0 {
		signals = append(signals, sig)
		reasons = append(reasons, msg)
	}
	if sig, msg := a.detectAccumulationDistribution(ts); sig != 0 {
		signals = append(signals, sig)
		reasons = append(reasons, msg)
	}
	if sig, msg := a.detectVWAPCluster(ts); sig != 0 {
		signals = append(signals, sig)
		reasons = append(reasons, msg)
	}
	if sig, msg := a.detectDivergence(ts); sig != 0 {
		signals = append(signals, sig)
		reasons = append(reasons, msg)
	}
	if sig, msg := a.detectMarketProfile(ts); sig != 0 {
		signals = append(signals, sig)
		reasons = append(reasons, msg)
	}
	if sig, msg := a.detectTrendStrength(ts); sig != 0 {
		signals = append(signals, sig)
		reasons = append(reasons, msg)
	}
	if sig, msg := a.detectCandleVolumeSignals(ts); sig != 0 {
		signals = append(signals, sig)
		reasons = append(reasons, msg)
	}

	total := 0
	for _, s := range signals {
		total += s
	}
	finalSignal := "HOLD"
	if total > 0 {
		finalSignal = "BUY"
	} else if total < 0 {
		finalSignal = "SELL"
	}
	reasonText := strings.Join(reasons, "; ")
	if reasonText == "" {
		reasonText = "No significant SMC features detected"
	}
	return finalSignal, reasonText
}

type SMCDetailedSignal struct {
	Signal int
	Reason string
}

func (a *SMCAnalyzer) DetectMultipleSMCSignals(ts *techan.TimeSeries) map[string]SMCDetailedSignal {
	results := make(map[string]SMCDetailedSignal)

	if sig, msg := a.detectBOS(ts); sig != 0 {
		results["BOS"] = SMCDetailedSignal{Signal: sig, Reason: msg}
	}
	if sig, msg := a.detectCHOCH(ts); sig != 0 {
		results["CHOCH"] = SMCDetailedSignal{Signal: sig, Reason: msg}
	}
	if sig, msg := a.detectFVG(ts); sig != 0 {
		results["FVG"] = SMCDetailedSignal{Signal: sig, Reason: msg}
	}
	if sig, msg := a.detectOrderBlock(ts); sig != 0 {
		results["OrderBlock"] = SMCDetailedSignal{Signal: sig, Reason: msg}
	}
	if sig, msg := a.detectLiquiditySweep(ts); sig != 0 {
		results["LiquiditySweep"] = SMCDetailedSignal{Signal: sig, Reason: msg}
	}
	if sig, msg := a.detectStopRun(ts); sig != 0 {
		results["StopRun"] = SMCDetailedSignal{Signal: sig, Reason: msg}
	}
	if sig, msg := a.detectVolumeImbalance(ts); sig != 0 {
		results["VolumeImbalance"] = SMCDetailedSignal{Signal: sig, Reason: msg}
	}
	if sig, msg := a.detectCandlestickPatterns(ts); sig != 0 {
		results["CandlestickPatterns"] = SMCDetailedSignal{Signal: sig, Reason: msg}
	}
	if sig, msg := a.detectAccumulationDistribution(ts); sig != 0 {
		results["AccumulationDistribution"] = SMCDetailedSignal{Signal: sig, Reason: msg}
	}
	if sig, msg := a.detectVWAPCluster(ts); sig != 0 {
		results["VWAPCluster"] = SMCDetailedSignal{Signal: sig, Reason: msg}
	}
	if sig, msg := a.detectDivergence(ts); sig != 0 {
		results["Divergence"] = SMCDetailedSignal{Signal: sig, Reason: msg}
	}
	if sig, msg := a.detectMarketProfile(ts); sig != 0 {
		results["MarketProfile"] = SMCDetailedSignal{Signal: sig, Reason: msg}
	}
	if sig, msg := a.detectTrendStrength(ts); sig != 0 {
		results["TrendStrength"] = SMCDetailedSignal{Signal: sig, Reason: msg}
	}
	if sig, msg := a.detectCandleVolumeSignals(ts); sig != 0 {
		results["CandleVolume"] = SMCDetailedSignal{Signal: sig, Reason: msg}
	}

	return results
}

// detectCandleVolumeSignals analyzes the last 'period' candles to derive floorsheet-like signals.
func (a *SMCAnalyzer) detectCandleVolumeSignals(ts *techan.TimeSeries) (int, string) {
	if len(ts.Candles) < a.period {
		return 0, "Not enough candles for candle volume analysis"
	}

	start := len(ts.Candles) - a.period
	var bullishVolume, bearishVolume, totalVolume float64

	// Sum volumes and separate bullish vs bearish candles.
	for i := start; i < len(ts.Candles); i++ {
		candle := ts.Candles[i]
		vol := candle.Volume.Float()
		totalVolume += vol
		if candle.ClosePrice.Float() > candle.OpenPrice.Float() {
			bullishVolume += vol
		} else if candle.ClosePrice.Float() < candle.OpenPrice.Float() {
			bearishVolume += vol
		}
	}

	avgVolume := totalVolume / float64(a.period)
	signal := 0
	var reasons []string

	// Check for volume imbalance.
	if bullishVolume > floorSheetBuyImbalanceThreshold*bearishVolume {
		signal += 1
		reasons = append(reasons, fmt.Sprintf("Bullish volume imbalance: %.2f vs %.2f", bullishVolume, bearishVolume))
	} else if bearishVolume > floorSheetBuyImbalanceThreshold*bullishVolume {
		signal -= 1
		reasons = append(reasons, fmt.Sprintf("Bearish volume imbalance: %.2f vs %.2f", bearishVolume, bullishVolume))
	}

	// Check each candle for a large volume spike.
	for i := start; i < len(ts.Candles); i++ {
		candle := ts.Candles[i]
		if candle.Volume.Float() > floorSheetLargeTradeMultiplier*avgVolume {
			if candle.ClosePrice.Float() > candle.OpenPrice.Float() {
				signal += 1
				reasons = append(reasons, fmt.Sprintf("Large bullish candle at index %d", i))
			} else if candle.ClosePrice.Float() < candle.OpenPrice.Float() {
				signal -= 1
				reasons = append(reasons, fmt.Sprintf("Large bearish candle at index %d", i))
			}
		}
	}
	if len(reasons) > 0 {
		return signal, strings.Join(reasons, "; ")
	}
	return signal, ""
}
