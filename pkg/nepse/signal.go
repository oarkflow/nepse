package nepse

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/schmidthole/big"

	"github.com/oarkflow/trading/pkg/nepse/models"
	"github.com/oarkflow/trading/pkg/techan"
)

func loadDailyFloorsheet(dir, _ string, date ...*Range) ([]models.FloorSheet, error) {
	var dateRange []string
	if len(date) > 0 {
		dateRange = getDateRange(date[0].start, date[0].end)
	}
	var sheets []models.FloorSheet
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		fileName := strings.TrimSuffix(filepath.Base(file), ".json")
		if len(dateRange) > 0 {
			if slices.Contains(dateRange, fileName) {
				data, err := os.ReadFile(file)
				if err != nil {
					return nil, err
				}
				var sheet models.FloorSheet
				if err = json.Unmarshal(data, &sheet); err != nil {
					return nil, err
				}
				sheets = append(sheets, sheet)
			}
		} else {
			data, err := os.ReadFile(file)
			if err != nil {
				return nil, err
			}
			var sheet models.FloorSheet
			if err = json.Unmarshal(data, &sheet); err != nil {
				return nil, err
			}
			sheets = append(sheets, sheet)
		}
	}
	periods := getScaledSignalPeriods(len(sheets))
	macdSignalPeriod = periods.MACDSignalPeriod
	macdLongPeriod = periods.MACDLongPeriod
	macdShortPeriod = periods.MACDShortPeriod
	rsiPeriod = periods.RSIPeriod
	atrPeriod = periods.ATRPeriod
	smaPeriod = periods.SMAPeriod
	bollingerPeriod = periods.BollingerPeriod
	stochasticPeriod = periods.StochasticPeriod
	smcPeriod = periods.SMCPeriod
	emaLongPeriod = periods.EMALongPeriod
	emaShortPeriod = periods.EMAShortPeriod
	return sheets, nil
}

func loadLiveFloorsheet(dir, summaryFile string, _ ...*Range) ([]models.FloorSheet, error) {
	var ls models.LiveSummary
	data, err := os.ReadFile(summaryFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("File does not exist:", err)
			return nil, nil
		}
		return nil, err
	}
	if err = json.Unmarshal(data, &ls); err != nil {
		return nil, err
	}
	var sheets []models.FloorSheet
	for i := 1; i <= ls.CurrentPage; i++ {
		pageFile := fmt.Sprintf("%s/page-%d.json", dir, i)
		d, err := os.ReadFile(pageFile)
		if err != nil {
			return nil, err
		}
		var sheet models.LiveFloorsheet
		if err = json.Unmarshal(d, &sheet); err != nil {
			return nil, err
		}
		floorsheet := models.FloorSheet{
			Content:     sheet.Floorsheets.Content,
			TotalAmount: ls.TotalAmount,
			TotalQty:    ls.TotalQty,
			TotalTrades: ls.TotalTrades,
		}
		sheets = append(sheets, floorsheet)
	}
	return sheets, nil
}

func loadHoldings(filePath string) ([]models.Holding, error) {
	var holdings []models.Holding
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return holdings, nil
		}
		return nil, err
	}
	if err = json.Unmarshal(data, &holdings); err != nil {
		return nil, err
	}
	return holdings, nil
}

func buildTimeSeriesForStock(transactions []models.Transaction, interval time.Duration) *techan.TimeSeries {
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].TradeTime.Before(transactions[j].TradeTime.Time)
	})
	ts := techan.NewTimeSeries()
	var currentCandle *techan.Candle
	var currentPeriod time.Time
	for _, tx := range transactions {
		t := tx.TradeTime.Truncate(interval)
		if currentCandle == nil || !t.Equal(currentPeriod) {
			if currentCandle != nil {
				ts.AddCandle(currentCandle)
			}
			currentPeriod = t
			currentCandle = techan.NewCandle(techan.NewTimePeriod(t, interval))
			price := big.NewDecimal(tx.ContractRate)
			currentCandle.OpenPrice = price
			currentCandle.MaxPrice = price
			currentCandle.MinPrice = price
			currentCandle.ClosePrice = price
			currentCandle.Volume = big.NewDecimal(float64(tx.ContractQuantity))
		} else {
			price := big.NewDecimal(tx.ContractRate)
			currentCandle.ClosePrice = price
			if price.LT(currentCandle.MinPrice) {
				currentCandle.MinPrice = price
			}
			if price.GT(currentCandle.MaxPrice) {
				currentCandle.MaxPrice = price
			}
			vol := big.NewDecimal(float64(tx.ContractQuantity))
			currentCandle.Volume = currentCandle.Volume.Add(vol)
		}
	}
	if currentCandle != nil {
		ts.AddCandle(currentCandle)
	}
	return ts
}

func computeDailySummary(ts *techan.TimeSeries) (days int, highPrice, lowPrice, avgPrice float64) {
	days = len(ts.Candles)
	if days == 0 {
		return
	}
	highPrice = ts.Candles[0].MaxPrice.Float()
	lowPrice = ts.Candles[0].MinPrice.Float()
	var sumClose float64
	for _, candle := range ts.Candles {
		priceHigh := candle.MaxPrice.Float()
		priceLow := candle.MinPrice.Float()
		if priceHigh > highPrice {
			highPrice = priceHigh
		}
		if priceLow < lowPrice {
			lowPrice = priceLow
		}
		sumClose += candle.ClosePrice.Float()
	}
	avgPrice = sumClose / float64(days)
	return
}

func computeTotalVolume(ts *techan.TimeSeries) float64 {
	var totalVolume float64
	for _, candle := range ts.Candles {
		totalVolume += candle.Volume.Float()
	}
	return totalVolume
}

func determineTradeAction(dailySignal, macdSignal, atrSignal, emaCrossoverSignal, bollingerSignal, smcSignal string, ts *techan.TimeSeries, holding *models.Holding) string {
	buyVotes, sellVotes := 0, 0
	signals := []string{dailySignal, macdSignal, atrSignal, emaCrossoverSignal, bollingerSignal, smcSignal}
	for _, sig := range signals {
		if sig == "BUY" || sig == "VOLATILE_UP" || sig == "Bullish (Buy)" {
			buyVotes++
		} else if sig == "SELL" || sig == "VOLATILE_DOWN" || sig == "Bearish (Sell)" {
			sellVotes++
		}
	}
	var recommendation string
	if holding != nil {
		lastCandleTime := ts.LastCandle().Period.Start
		daysHeld := int(lastCandleTime.Sub(holding.EntryDate.Time).Hours() / 24)
		if daysHeld < settlementDays {
			recommendation = "HOLD (Locked)"
		} else if sellVotes > buyVotes {
			recommendation = "SELL"
		} else if buyVotes > sellVotes {
			recommendation = "BUY"
		} else {
			recommendation = "HOLD"
		}
	} else {
		if buyVotes > sellVotes {
			recommendation = "BUY"
		} else if sellVotes > buyVotes {
			recommendation = "SELL"
		} else {
			recommendation = "HOLD"
		}
	}
	return recommendation
}

func groupTransactions(sheets []models.FloorSheet) map[string][]models.Transaction {
	transactions := make(map[string][]models.Transaction)
	for _, sheet := range sheets {
		for _, tx := range sheet.Content {
			transactions[tx.StockSymbol] = append(transactions[tx.StockSymbol], tx)
		}
	}
	// Sort transactions for each stock symbol by ascending TradeTime.
	for symbol, txs := range transactions {
		sort.Slice(txs, func(i, j int) bool {
			// Using the embedded time.Time's Before method in CustomTime.
			return txs[i].TradeTime.Before(txs[j].TradeTime.Time)
		})
		transactions[symbol] = txs
	}
	return transactions
}

func analyzeStockSignals(stock string, tsDaily *techan.TimeSeries, holdingMap map[string]models.Holding) models.StockAnalysis {
	var analysis models.StockAnalysis
	analysis.Symbol = stock
	shortTermSignal, shortTermReason := generateShortTermSignal(tsDaily, smaPeriod, rsiPeriod, rsiThreshold)
	if shortTermSignal == "HOLD" {
		altSignal, altReason := generateMACDSignal(tsDaily, macdShortPeriod, macdLongPeriod, macdSignalPeriod)
		shortTermSignal = altSignal
		shortTermReason = "MACD alternative: " + altReason
	}
	analysis.ShortTermSignal = shortTermSignal
	analysis.ShortTermReason = shortTermReason
	macdSignal, macdReason := generateMACDSignal(tsDaily, macdShortPeriod, macdLongPeriod, macdSignalPeriod)
	analysis.MACDSignal = macdSignal
	analysis.MACDReason = macdReason
	atrSignal, atrReason := generateATRSignal(tsDaily, atrPeriod, atrThreshold)
	analysis.ATRSignal = atrSignal
	analysis.ATRReason = atrReason
	emaCrossoverSignal, emaCrossoverReason := generateEMACrossoverSignal(tsDaily, emaShortPeriod, emaLongPeriod)
	analysis.EMACrossoverSignal = emaCrossoverSignal
	analysis.EMACrossoverReason = emaCrossoverReason
	bollingerSignal, bollingerReason := generateBollingerSignal(tsDaily, bollingerPeriod, bollingerMultiplier)
	analysis.BollingerSignal = bollingerSignal
	analysis.BollingerReason = bollingerReason
	smcAnalyzer := NewSMCAnalyzer(smcPeriod)
	smcSignal, smcReason := smcAnalyzer.AnalyzeSMC(tsDaily)
	analysis.SMCSignal = smcSignal
	analysis.SMCReason = smcReason
	analysis.VWAP = computeVWAP(tsDaily)
	if len(tsDaily.Candles) >= bollingerPeriod {
		upperBB, middleBB, lowerBB := computeBollingerBands(tsDaily, bollingerPeriod, bollingerMultiplier)
		analysis.BollingerUpper = upperBB
		analysis.BollingerMiddle = middleBB
		analysis.BollingerLower = lowerBB
	}
	stoch, stochSignal := computeStochastic(tsDaily, stochasticPeriod)
	analysis.Stochastic = models.Stochastic{
		Index:  stoch,
		Signal: stochSignal,
	}
	var holding *models.Holding
	if h, ok := holdingMap[stock]; ok {
		holding = &h
	}
	analysis.Action = determineTradeAction(shortTermSignal, macdSignal, atrSignal, emaCrossoverSignal, bollingerSignal, smcSignal, tsDaily, holding)
	analysis.Trades, analysis.NetPL, analysis.AvgPL, analysis.WinRate = summarizeTrades(backtestStrategy(tsDaily, models.RiskParameters{
		StopLossPercent:   stopLossPercent,
		TakeProfitPercent: takeProfitPercent,
	}, smaPeriod, rsiPeriod, rsiThreshold))
	analysis.AvgRiskReward = computeAvgRiskReward(backtestStrategy(tsDaily, models.RiskParameters{
		StopLossPercent:   stopLossPercent,
		TakeProfitPercent: takeProfitPercent,
	}, smaPeriod, rsiPeriod, rsiThreshold))
	analysis.MaxDrawdown = computeMaxDrawdown(tsDaily)
	analysis.Days, analysis.HighPrice, analysis.LowPrice, analysis.AvgPrice = computeDailySummary(tsDaily)
	analysis.TotalVolume = computeTotalVolume(tsDaily)
	return analysis
}

func enhancedAnalysis(analysis *models.StockAnalysis, dailyTS *techan.TimeSeries, transactions []models.Transaction) {
	analysis.OBV, analysis.OBVSignal, analysis.OBVReason = generateOBVSignal(dailyTS)
	analysis.AD, analysis.ADSignal, analysis.ADReason = generateADSignal(dailyTS)
	analysis.CMF, analysis.CMFSignal, analysis.CMFReason = generateCMFSignal(dailyTS, 20)
	analysis.OrderImbalance, analysis.OrderImbalanceSignal, analysis.OrderImbalanceReason = generateOrderImbalanceSignal(transactions)
	analysis.CandlestickPattern, analysis.CandlestickSignal, analysis.CandlestickReason = generateCandlestickSignal(dailyTS)

	closePriceIndicator := techan.NewClosePriceIndicator(dailyTS)
	rsiIndicator := techan.NewRelativeStrengthIndexIndicator(closePriceIndicator, rsiPeriod)
	analysis.Divergence, analysis.DivergenceSignal, analysis.DivergenceReason = generateDivergenceSignal(dailyTS, rsiIndicator)

	bins, volumes, mpSignal, mpReason := generateMarketProfileSignal(dailyTS, marketProfileBinCount)
	analysis.MarketProfileBins = bins
	analysis.MarketProfileVolumes = volumes
	analysis.MarketProfileSignal = mpSignal
	analysis.MarketProfileReason = mpReason

	riskParams := models.RiskParameters{
		StopLossPercent:   stopLossPercent,
		TakeProfitPercent: takeProfitPercent,
	}
	analysis.OptimizedSMAPeriod = optimizeSMAPeriod(dailyTS, dailyTS, riskParams, 10, 30, rsiPeriod, rsiThreshold)

	// Advanced risk metrics.
	trades := backtestStrategy(dailyTS, riskParams, smaPeriod, rsiPeriod, rsiThreshold)
	analysis.SharpeRatio = computeSharpeRatio(trades, 0)
	analysis.SortinoRatio = computeSortinoRatio(trades, 0)
	analysis.Expectancy = computeExpectancy(trades)
	riskSignal, riskReason := generateRiskMetricsSignal(analysis.SharpeRatio, analysis.SortinoRatio, analysis.Expectancy)
	analysis.RiskMetricsSignal = riskSignal
	analysis.RiskMetricsReason = riskReason

	// External sentiment and macro data.
	analysis.Sentiment = fetchSentiment(analysis.Symbol)
	sentSignal, sentReason := generateSentimentSignal(analysis.Sentiment)
	analysis.SentimentSignal = sentSignal
	analysis.SentimentReason = sentReason

	analysis.MacroData = fetchMacroIndicator()
	macroSignal, macroReason := generateMacroSignal(analysis.MacroData)
	analysis.MacroSignal = macroSignal
	analysis.MacroReason = macroReason

	// Machine learning prediction.
	predSignal, predReason := generatePredictionSignal(dailyTS)
	analysis.PredictionSignal = predSignal
	analysis.PredictionReason = predReason

	// Anomaly detection.
	analysis.AnomalyDetected = detectAnomalies(dailyTS)
	analysis.AnomalySignal, analysis.AnomalyReason = generateAnomalySignal(dailyTS)
}

func analyzeStockTransactions(transactionsByStock map[string][]models.Transaction, holdingMap map[string]models.Holding, interval time.Duration) []models.StockAnalysis {
	var results []models.StockAnalysis
	for stock, transactions := range transactionsByStock {
		dailyTS := buildTimeSeriesForStock(transactions, interval)
		brokerStats := analyzeBrokerStats(transactions)
		topBrokers := brokerHoldingsSummary(brokerStats)
		boomAlert := "NO"
		if detectPossibleBoom(dailyTS, boomVolumeThreshold, boomPriceThreshold) {
			boomAlert = "YES"
		}
		analysis := analyzeStockSignals(stock, dailyTS, holdingMap)
		analysis.TopBrokers = topBrokers
		analysis.Boom = boomAlert
		enhancedAnalysis(&analysis, dailyTS, transactions)
		results = append(results, analysis)
	}
	return results
}

func GetHoldings() map[string]models.Holding {
	holdings, err := loadHoldings(holdingsFile)
	if err != nil {
		log.Printf("Error loading holdings: %v", err)
		holdings = []models.Holding{}
	}
	holdingMap := make(map[string]models.Holding)
	for _, h := range holdings {
		holdingMap[h.StockSymbol] = h
	}
	return holdingMap
}

type FloorSheetHandler func(string, string, ...*Range) ([]models.FloorSheet, error)

func getAnalysis(fn FloorSheetHandler, interval time.Duration, dir, file string, date ...*Range) ([]models.StockAnalysis, error) {
	sheets, err := fn(dir, file, date...)
	if err != nil {
		return nil, err
	}
	holdingMap := GetHoldings()
	transactions := groupTransactions(sheets)
	analysis := analyzeStockTransactions(transactions, holdingMap, interval)
	return analysis, nil
}

func GetDailyAnalysis(interval time.Duration, date ...*Range) ([]models.StockAnalysis, error) {
	return getAnalysis(loadDailyFloorsheet, interval, floorsheetDir, "", date...)
}

func GetIntraDayAnalysis(interval time.Duration, date ...*Range) ([]models.StockAnalysis, error) {
	if interval == 0 {
		interval = intradayInterval
	}
	return getAnalysis(loadDailyFloorsheet, interval, floorsheetDir, "", date...)
}

func GetLiveAnalysis(interval time.Duration) ([]models.StockAnalysis, error) {
	if interval == 0 {
		interval = liveInterval
	}
	return getAnalysis(loadLiveFloorsheet, interval, liveFloorsheetDir, liveSummaryFile)
}

func GetAnalysis(dateRange ...*Range) (models.AnalysisResult, error) {
	date, err := getDate(dateRange...)
	if err != nil {
		return models.AnalysisResult{}, err
	}
	dailyAnalysis, err := GetDailyAnalysis(dailyInterval, date)
	if err != nil {
		return models.AnalysisResult{}, err
	}
	result := models.AnalysisResult{
		StartDate: date.Start,
		EndDate:   date.End,
		Daily:     dailyAnalysis,
	}
	return result, nil
}
