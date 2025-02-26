package models

import (
	"time"
)

type BrokerVolume struct {
	Symbol string `json:"symbol"`
	Volume int    `json:"volume"`
}

type MarketRatio struct {
	Buy  int `json:"buy"`
	Sell int `json:"sell"`
}

type Pageable struct {
	Offset     int  `json:"offset"`
	PageNumber int  `json:"pageNumber"`
	PageSize   int  `json:"pageSize"`
	Paged      bool `json:"paged"`
	Sort       Sort `json:"sort"`
	Unpaged    bool `json:"unpaged"`
}

type Sort struct {
	Empty    bool `json:"empty"`
	Sorted   bool `json:"sorted"`
	Unsorted bool `json:"unsorted"`
}

type Paging struct {
	Empty            bool     `json:"empty"`
	First            bool     `json:"first"`
	Last             bool     `json:"last"`
	Number           int      `json:"number"`
	NumberOfElements int      `json:"numberOfElements"`
	Pageable         Pageable `json:"pageable"`
	Size             int      `json:"size"`
	Sort             Sort     `json:"sort"`
	TotalElements    int      `json:"totalElements"`
	TotalPages       int      `json:"totalPages"`
}

type TodayPrice struct {
	Content []StockPrice `json:"content"`
	Paging
}

type StockPrice struct {
	AverageTradedPrice    float64 `json:"averageTradedPrice"`
	BusinessDate          string  `json:"businessDate"`
	ClosePrice            float64 `json:"closePrice"`
	FiftyTwoWeekHigh      float64 `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow       float64 `json:"fiftyTwoWeekLow"`
	HighPrice             float64 `json:"highPrice"`
	LastUpdatedPrice      float64 `json:"lastUpdatedPrice"`
	LastUpdatedTime       string  `json:"lastUpdatedTime"`
	LowPrice              float64 `json:"lowPrice"`
	MarketCapitalization  float64 `json:"marketCapitalization"`
	OpenPrice             float64 `json:"openPrice"`
	PreviousDayClosePrice float64 `json:"previousDayClosePrice"`
	SecurityId            int     `json:"securityId"`
	SecurityName          string  `json:"securityName"`
	Symbol                string  `json:"symbol"`
	TotalTradedQuantity   int     `json:"totalTradedQuantity"`
	TotalTradedValue      float64 `json:"totalTradedValue"`
	TotalTrades           int     `json:"totalTrades"`
}

// SignalPeriods holds the period values for various technical indicators.
// All period values represent the number of days used in calculations when using daily candles.
type SignalPeriods struct {
	SMAPeriod        int // Number of days used to compute the Simple Moving Average. (Default: 15)
	RSIPeriod        int // Number of days for calculating the Relative Strength Index. (Default: 14)
	MACDShortPeriod  int // Short-term period for MACD calculation. (Default: 12)
	MACDLongPeriod   int // Long-term period for MACD calculation. (Default: 26)
	MACDSignalPeriod int // Period for the MACD signal line. (Default: 9)
	BollingerPeriod  int // Number of days for Bollinger Bands calculation. (Default: 20)
	StochasticPeriod int // Number of days for the stochastic oscillator. (Default: 14)
	ATRPeriod        int // Number of days for the Average True Range calculation. (Default: 14)
	SMCPeriod        int // Period for Smart Money Concept analysis (BOS, CHOCH, etc.). (Default: 5)
	EMAShortPeriod   int // Short-term period for EMA crossover analysis. (Default: 12)
	EMALongPeriod    int // Long-term period for EMA crossover analysis. (Default: 26)
}

type Holding struct {
	StockSymbol string     `json:"stockSymbol"`
	EntryDate   CustomTime `json:"entryDate"`
	EntryPrice  float64    `json:"entryPrice"`
	Quantity    int        `json:"quantity"`
}

type BrokerStats struct {
	TotalBuy  float64 `json:"totalBuy"`
	TotalSell float64 `json:"totalSell"`
	Net       float64 `json:"net"`
}

type Stochastic struct {
	Index  float64 `json:"index"`
	Signal string  `json:"signal"`
}

// Expanded to include reasons for each signal
type StockAnalysis struct {
	Symbol               string                 `json:"symbol"`
	ShortTermSignal      string                 `json:"shortTermSignal"`
	ShortTermReason      string                 `json:"shortTermReason"`
	MACDSignal           string                 `json:"macdSignal"`
	MACDReason           string                 `json:"macdReason"`
	ATRSignal            string                 `json:"atrSignal"`
	ATRReason            string                 `json:"atrReason"`
	EMACrossoverSignal   string                 `json:"emaCrossoverSignal"`
	EMACrossoverReason   string                 `json:"emaCrossoverReason"`
	BollingerSignal      string                 `json:"bollingerSignal"`
	BollingerReason      string                 `json:"bollingerReason"`
	VWAP                 float64                `json:"vwap"`
	BollingerUpper       float64                `json:"bollingerUpper"`
	BollingerMiddle      float64                `json:"bollingerMiddle"`
	BollingerLower       float64                `json:"bollingerLower"`
	Stochastic           Stochastic             `json:"stochastic"`
	Action               string                 `json:"action"`
	Trades               int                    `json:"trades"`
	AvgRiskReward        float64                `json:"avgRiskReward"`
	MaxDrawdown          float64                `json:"maxDrawdown"`
	NetPL                float64                `json:"netPL"`
	AvgPL                float64                `json:"avgPL"`
	WinRate              float64                `json:"winRate"`
	Boom                 string                 `json:"boom"`
	TopBrokers           map[string]BrokerStats `json:"topBrokers"`
	HoldingInfo          string                 `json:"holdingInfo"`
	Days                 int                    `json:"days"`
	HighPrice            float64                `json:"highPrice"`
	LowPrice             float64                `json:"lowPrice"`
	AvgPrice             float64                `json:"avgPrice"`
	TotalVolume          float64                `json:"totalVolume"`
	SMCReason            string                 `json:"smcReason"`
	SMCSignal            string                 `json:"smcSignal"`
	OBV                  []float64              `json:"obv"`
	OBVSignal            string
	OBVReason            string
	AD                   float64
	ADSignal             string
	ADReason             string
	CMF                  float64
	CMFSignal            string
	CMFReason            string
	OrderImbalance       float64
	OrderImbalanceSignal string
	OrderImbalanceReason string
	CandlestickPattern   string
	CandlestickSignal    string
	CandlestickReason    string
	Divergence           string
	DivergenceSignal     string
	DivergenceReason     string
	MarketProfileBins    []float64
	MarketProfileVolumes []float64
	MarketProfileSignal  string
	MarketProfileReason  string
	OptimizedSMAPeriod   int
	SharpeRatio          float64
	SortinoRatio         float64
	Expectancy           float64
	RiskMetricsSignal    string
	RiskMetricsReason    string
	Sentiment            float64
	SentimentSignal      string
	SentimentReason      string
	MacroData            map[string]float64
	MacroSignal          string
	MacroReason          string
	PredictionSignal     string
	PredictionReason     string
	AnomalyDetected      bool
	AnomalySignal        string
	AnomalyReason        string
}

type AnalysisResult struct {
	StartDate string          `json:"startDate"`
	EndDate   string          `json:"endDate"`
	Daily     []StockAnalysis `json:"daily"`
}

type LiveSummary struct {
	CurrentPage      int     `json:"currentPage"`
	NumberOfElements int     `json:"numberOfElements"`
	TotalAmount      float64 `json:"totalAmount"`
	TotalPages       int     `json:"totalPages"`
	TotalQty         int     `json:"totalQty"`
	TotalTrades      int     `json:"totalTrades"`
}

type RiskParameters struct {
	StopLossPercent   float64
	TakeProfitPercent float64
}

type Trade struct {
	EntryTime         time.Time
	EntryPrice        float64
	ExitTime          time.Time
	ExitPrice         float64
	Position          string
	StopLoss          float64
	TakeProfit        float64
	ProfitLossPercent float64
}
