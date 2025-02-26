package nepse

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/oarkflow/trading/pkg/nepse/models"
	"github.com/oarkflow/trading/pkg/utils"
)

var (
	TokenClearInterval    = time.Minute
	MaxPageSize           = 500
	DataPath              = "./data"
	CompanyFloorsheetPath = "floorsheet/company"
	DateFloorsheetPath    = "floorsheet/date"
	LiveFloorsheetPath    = "floorsheet/live"
	PricePath             = "price"
	DummyData             = []int32{
		147, 117, 239, 143, 157, 312, 161, 612, 512, 804,
		411, 527, 170, 511, 421, 667, 764, 621, 301, 106,
		133, 793, 411, 511, 312, 423, 344, 346, 653, 758,
		342, 222, 236, 811, 711, 611, 122, 447, 128, 199,
		183, 135, 489, 703, 800, 745, 152, 863, 134, 211,
		142, 564, 375, 793, 212, 153, 138, 153, 648, 611,
		151, 649, 318, 143, 117, 756, 119, 141, 717, 113,
		112, 146, 162, 660, 693, 261, 362, 354, 251, 641,
		157, 178, 631, 192, 734, 445, 192, 883, 187, 122,
		591, 731, 852, 384, 565, 596, 451, 772, 624, 691,
	}
)

// =============================================================================
// Global Variables & Constants
// =============================================================================

var (
	// floorsheetDir: Directory where daily floorsheet JSON files are stored.
	// These files contain historical transaction data that are grouped by day.
	floorsheetDir = DataPath + "/" + DateFloorsheetPath

	// liveFloorsheetDir: Directory containing live (or near-live) floorsheet data.
	liveFloorsheetDir = DataPath + "/" + LiveFloorsheetPath

	// liveSummaryFile: File path to the live summary JSON file that holds metadata (like latest page).
	liveSummaryFile = DataPath + "/" + LiveFloorsheetPath + "/summary.json"

	// holdingsFile: File path to the JSON file with current holdings information.
	holdingsFile = DataPath + "/holdings.json"

	// dailyInterval: Interval used for creating daily candles.
	// Since this is set to 24 hours, each candle represents one day.
	dailyInterval = 24 * time.Hour

	// intradayInterval: Interval for intraday analysis.
	// With a 15-minute interval, you get more granular data during a single day.
	intradayInterval                = 15 * time.Minute
	floorSheetBuyImbalanceThreshold = 1.5 // e.g. 50% higher buy vs sell volume
	floorSheetLargeTradeMultiplier  = 3.0 // trade volume >3x average trade volume
	liveInterval                    = 15 * time.Minute

	// smaPeriod: The number of candles (days, when using daily data) used to compute the Simple Moving Average.
	// For example, if you have daily candles, smaPeriod = 15 means a 15-day moving average.
	// If analyzing only the last 15 days of data, using 15 might be too long (using all available days),
	// so you might reduce it (e.g., to 5 or 7) for a more responsive average.
	smaPeriod = 15

	// rsiPeriod: The number of days used to compute the Relative Strength Index.
	// Standard RSI uses 14 periods; however, if you have only 15 days of data, you might consider a shorter period (e.g., 7)
	// to capture more meaningful momentum changes.
	rsiPeriod = 14

	// macdShortPeriod: The short-term period for the MACD calculation (typically in days).
	// Standard settings use 12 days, but if your overall dataset is short (15 days), you may need to scale this down.
	macdShortPeriod = 12

	// macdLongPeriod: The long-term period for the MACD calculation (in days).
	// Typically set to 26 days; if you have fewer days, a proportional reduction (e.g., to 11 or 13) might be necessary.
	macdLongPeriod = 26

	// macdSignalPeriod: The period used to compute the MACD signal line (in days).
	// Standard is 9, and it should be in proportion to the short and long periods.
	macdSignalPeriod = 9

	// rsiThreshold: The RSI threshold for signal generation.
	// An RSI above 50 typically indicates bullish momentum, while below 50 indicates bearish momentum.
	rsiThreshold = 50.0

	// boomVolumeThreshold: A threshold for trading volume to indicate an unusually high volume (a “boom”).
	// This is an absolute value, so higher volumes than this trigger additional signals.
	boomVolumeThreshold = float64(10000)

	// boomPriceThreshold: The percentage price change threshold that, when combined with volume,
	// may indicate a strong move (or boom) in price.
	boomPriceThreshold = 5.0

	// stopLossPercent: The percentage distance from the entry price used to set a stop loss.
	// A tighter stop loss (smaller percentage) reduces risk but may increase stop-outs.
	stopLossPercent = 2.0

	// takeProfitPercent: The percentage target for taking profit relative to the entry price.
	// A higher take profit allows for larger gains but might be less frequently reached.
	takeProfitPercent = 4.0

	// settlementDays: The minimum number of days to hold a position due to settlement requirements.
	// For example, 4 days might be required by the market structure or regulations.
	settlementDays = 4

	// bollingerPeriod: The number of days used to calculate Bollinger Bands.
	// For daily data, a typical value is 20 days. However, if you only have 15 days of data,
	// you may want to reduce this to better reflect the shorter timeframe (e.g., 10 days).
	bollingerPeriod = 20

	// bollingerMultiplier: The number of standard deviations used to set the upper and lower Bollinger Bands.
	// The standard multiplier is 2.0; this remains independent of the dataset length.
	bollingerMultiplier = 2.0

	// stochasticPeriod: The number of days for calculating the stochastic oscillator.
	// Standard value is 14 days, but with a shorter dataset, a lower period might be more responsive.
	stochasticPeriod = 14

	// atrPeriod: The number of days used to compute the Average True Range, a measure of volatility.
	// With a shorter dataset (e.g., 15 days), you might consider a lower period for a more reactive ATR.
	atrPeriod = 14

	// -----------------------------------------------------------------
	// smcPeriod: Period for Smart Money Concept analysis (BOS, CHOCH, OB, etc.).
	//   → Default is 5 days; used to scan recent price action for key institutional signals.
	smcPeriod = 5

	// emaShortPeriod: Period for the short-term Exponential Moving Average (EMA) in crossover analysis.
	//   → Default is 12 days; a standard value used in many technical analysis strategies.
	emaShortPeriod = 12

	// emaLongPeriod: Period for the long-term EMA in crossover analysis.
	//   → Default is 26 days; used in conjunction with emaShortPeriod to detect trend shifts.
	emaLongPeriod = 26

	// atrThreshold: The threshold for ATR to determine high or low volatility.
	// This is generally a relative measure and may not need adjustment based on the dataset length.
	atrThreshold = 2.0

	stopRunVolumeMultiplier   = 1.5  // Current volume must exceed this times average volume.
	stopRunGapThreshold       = 0.02 // 2% gap threshold.
	volumeImbalanceMultiplier = 1.5  // Factor for high volume imbalance.
	adThresholdFraction       = 0.5  // Fraction of volume used to determine strong accumulation/distribution.
	vwapDiffThreshold         = 0.02 // 2% difference from VWAP.
	marketProfileBinCount     = 10   // Number of bins for market profile.
)

func getCompanyPath(companyId int) string {
	return filepath.Join(DataPath, CompanyFloorsheetPath, fmt.Sprintf("%d", companyId))
}

func getCompanyFloorsheetPath(companyId int, date string) string {
	baseDir := filepath.Join(DataPath, CompanyFloorsheetPath, fmt.Sprintf("%d", companyId))
	if !utils.FileExists(baseDir) {
		os.MkdirAll(baseDir, os.ModePerm)
	}
	return filepath.Join(baseDir, date+".json")
}

func getDateFloorsheetPath(date string) string {
	baseDir := filepath.Join(DataPath, DateFloorsheetPath)
	if !utils.FileExists(baseDir) {
		os.MkdirAll(baseDir, os.ModePerm)
	}
	return filepath.Join(baseDir, date+".json")
}

func getLiveFloorsheetPath(page int) string {
	baseDir := filepath.Join(DataPath, LiveFloorsheetPath)
	if !utils.FileExists(baseDir) {
		os.MkdirAll(baseDir, os.ModePerm)
	}
	return filepath.Join(baseDir, fmt.Sprintf("page-%d.json", page))
}

func getLivePath() string {
	return filepath.Join(DataPath, LiveFloorsheetPath)
}

func getFloorsheetProgressPath() string {
	return filepath.Join(DataPath, DateFloorsheetPath, "progress.json")
}

func getCompanyProgressPath(companyId int) string {
	return filepath.Join(DataPath, fmt.Sprintf("%s/%d", CompanyFloorsheetPath, companyId), "progress.json")
}

func getLiveSummaryPath() string {
	baseDir := filepath.Join(DataPath, LiveFloorsheetPath)
	if !utils.FileExists(baseDir) {
		os.MkdirAll(baseDir, os.ModePerm)
	}
	return filepath.Join(baseDir, "summary.json")
}

func getPricePath(date string) string {
	return filepath.Join(DataPath, PricePath, date+".json")
}

const (
	AUTH_URI     = "/api/authenticate/prove"
	SUBINDEX     = "/api/nots"
	MARKET_OPEN  = "/api/nots/nepse-data/market-open"
	LIVE_MARKET  = "/api/nots/lives-market"
	BROKER_LIST  = "/api/nots/member"
	PRICE_VOLUME = "/api/nots/securityDailyTradeStat/58"

	SECURITIES_LISTING         = "/api/nots/security?nonDelisted=true"
	SECURITIES_FLOORSHEET      = "/api/nots/security/floorsheet"
	SECURITIES_DETAIL          = "/api/nots/security"
	SECURITIES_DAILY_GRAPH     = "/api/nots/market/graphdata/daily"
	SECURITIES_PRICE_VOLUME    = "/api/nots/market/security/price"
	SECURITIES_TRADING_HISTORY = "/api/nots/market/history/security"

	HOLIDAYS  = "/api/nots/holiday/list"
	FILE_LINK = "/api/nots/security/fetchFiles"

	BANKING_INDEX       = "/api/nots/graph/index/51"
	HOTEL_INDEX         = "/api/nots/graph/index/52"
	OTHERS_INDEX        = "/api/nots/graph/index/53"
	HYDRO_INDEX         = "/api/nots/graph/index/54"
	DEV_BANK_INDEX      = "/api/nots/graph/index/55"
	MANUFACTURING_INDEX = "/api/nots/graph/index/56"
	NON_LIFE_INS_INDEX  = "/api/nots/graph/index/59"
	GRAPH_INDEX         = "/api/nots/graph/index/58"
	FINANCE_INDEX       = "/api/nots/graph/index/60"
	TRADING_INDEX       = "/api/nots/graph/index/61"
	MICROFINANCE_INDEX  = "/api/nots/graph/index/64"
	LIFE_INS_INDEX      = "/api/nots/graph/index/65"
	MUTUAL_FUND_INDEX   = "/api/nots/graph/index/66"
	INVESTMENT_INDEX    = "/api/nots/graph/index/67"

	FLOORSHEET  = "/api/nots/nepse-data/floorsheet"
	NEPSE_INDEX = "/api/nots/nepse-index"
	SECTORWISE  = "/api/nots/sectorwise"

	COMPANY_LIST           = "/api/nots/company/list"
	COMPANY_PROFILE        = "/api/nots/security/profile"
	ALERTS                 = "/api/nots/news/media/news-and-alerts"
	COMPANY_NEWS           = "/api/nots/application/company-news"
	COMPANY_REPORTS        = "/api/nots/application/reports"
	COMPANY_DIVIDEND       = "/api/nots/application/dividend"
	AVERAGE_TRADING        = "/api/nots/nepse-data/trading-average"
	MARKET_SUMMARY_HISTORY = "/api/nots/market-summary-history"
	TODAY_PRICE            = "/api/nots/nepse-data/today-price"
	TOP_GAINER             = "/api/nots/top-ten/top-gainer?all=true"
	TOP_LOSER              = "/api/nots/top-ten/top-loser?all=true"
	TOP_TURNOVER           = "/api/nots/top-ten/turnover?all=true"
	TOP_VOLUME             = "/api/nots/top-ten/trade?all=true"
	TOP_TRANSACTION        = "/api/nots/top-ten/transaction?all=true"
	SUPPLY_DEMAND          = "/api/nots/nepse-data/supplydemand"
	MARKET_SUMMARY         = "/api/nots/market-summary"

	GET  = "GET"
	POST = "POST"
)

func requestURL(path string) string {
	return utils.BaseUrl + path
}

// =============================================================================
// Existing Utility Functions & Data Structures
// =============================================================================

// scalePeriod scales a period based on available days.
func scalePeriod(standardPeriod, availableDays, baseDays int) int {
	scaled := (standardPeriod * availableDays) / baseDays
	if scaled < 1 {
		return 1
	}
	return scaled
}

// getScaledSignalPeriods returns appropriate period values for all indicators based on availableDays.
// 'baseDays' represents the dataset length for which the default values are assumed.
// For example, if the defaults are designed for 30 days and you have only 15 days,
// each period will roughly be halved.
//
// Usage Example:
//
//	periods := getScaledSignalPeriods(15)
//	// periods.SMAPeriod  -> will be scaled from 15 days down to 7 (if baseDays = 30)
//	// periods.RSIPeriod  -> scaled from 14 days to 7
//	// periods.MACDShortPeriod -> scaled from 12 days to 6
//	// periods.MACDLongPeriod  -> scaled from 26 days to 13
//	// periods.MACDSignalPeriod -> scaled from 9 days to 4
//	// periods.BollingerPeriod  -> scaled from 20 days to 10
//	// periods.StochasticPeriod -> scaled from 14 days to 7
//	// periods.ATRPeriod        -> scaled from 14 days to 7
//	// periods.SMCPeriod        -> scaled from 5 days to 2 (if availableDays is 15)
//	// periods.EMAShortPeriod   -> scaled from 12 days to 6
//	// periods.EMALongPeriod    -> scaled from 26 days to 13
func getScaledSignalPeriods(availableDays int) models.SignalPeriods {
	baseDays := 30 // Default base period. Adjust if your standard values are designed for a different timeframe.
	return models.SignalPeriods{
		SMAPeriod:        scalePeriod(15, availableDays, baseDays), // Default SMA period is 15 days.
		RSIPeriod:        scalePeriod(14, availableDays, baseDays), // Default RSI period is 14 days.
		MACDShortPeriod:  scalePeriod(12, availableDays, baseDays), // Standard MACD short period is 12 days.
		MACDLongPeriod:   scalePeriod(26, availableDays, baseDays), // Standard MACD long period is 26 days.
		MACDSignalPeriod: scalePeriod(9, availableDays, baseDays),  // Standard MACD signal period is 9 days.
		BollingerPeriod:  scalePeriod(20, availableDays, baseDays), // Typical Bollinger period is 20 days.
		StochasticPeriod: scalePeriod(14, availableDays, baseDays), // Standard stochastic period is 14 days.
		ATRPeriod:        scalePeriod(14, availableDays, baseDays), // Standard ATR period is 14 days.
		SMCPeriod:        scalePeriod(5, availableDays, baseDays),  // Default SMC period is 5 days.
		EMAShortPeriod:   scalePeriod(12, availableDays, baseDays), // Short-term EMA period for crossover analysis, default is 12 days.
		EMALongPeriod:    scalePeriod(26, availableDays, baseDays), // Long-term EMA period for crossover analysis, default is 26 days.
	}
}

// Range defines a date range.
type Range struct {
	Start     string `json:"start"`
	End       string `json:"end"`
	StartTime time.Time
	EndTime   time.Time
}

// GetDateRange returns a slice of dates between StartTime and EndTime.
func GetDateRange(start, end time.Time) []string {
	var dateRange []string
	for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
		dateRange = append(dateRange, current.Format(time.DateOnly))
	}
	return dateRange
}

func GetDate(dateRange ...*Range) (*Range, error) {
	var date *Range
	if len(dateRange) == 0 {
		now := time.Now()
		start := now.AddDate(0, 0, -30)
		date = &Range{StartTime: start, EndTime: now, Start: start.Format(time.DateOnly), End: now.Format(time.DateOnly)}
		return date, nil
	}
	date = dateRange[0]
	var err error
	date.StartTime, err = time.Parse(time.DateOnly, date.Start)
	if err != nil {
		return nil, err
	}
	date.EndTime, err = time.Parse(time.DateOnly, date.End)
	return date, err
}
