package nepse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"time"
)

type TraderHandler struct {
	Trader *Trader
}

func NewTraderHandler() *TraderHandler {
	return &TraderHandler{
		Trader: New(),
	}
}

func PanicHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered from panic:", r)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (h *TraderHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.IndexPage)
	mux.HandleFunc("/company/floorsheet", h.CompanyFloorSheet)
	mux.HandleFunc("/analyze/floorsheet/broker-buy-sell", h.AnalyzeTotalBuySellByBroker)
	mux.HandleFunc("/analyze/floorsheet/broker-sector", h.AnalyzeSectorWise)
	mux.HandleFunc("/analyze/floorsheet/broker", h.AnalyzeBrokerWise)
	mux.HandleFunc("/analyze/floorsheet/broker-transactions", h.AnalyzeBrokerTransactions)
	mux.HandleFunc("/analyze/floorsheet/top-brokers", h.AnalyzeTopBrokers)
	mux.HandleFunc("/analyze/floorsheet/traded-stocks", h.AnalyzeTradedStocks)
	mux.HandleFunc("/analyze/floorsheet/stock-volatility", h.AnalyzeStockVolatility)
	mux.HandleFunc("/analyze/floorsheet/buyer-seller", h.AnalyzeBuyerSellerMatching)
	mux.HandleFunc("/sector-companies", h.SectorCompanies)
	mux.HandleFunc("/company/price-volume", h.CompanyPriceVolume)
	mux.HandleFunc("/holidays", h.Holidays)
	mux.HandleFunc("/company/trading-history", h.CompanyTradingHistory)
	mux.HandleFunc("/company/detail", h.CompanyDetail)
	mux.HandleFunc("/company/daily-graph", h.CompanyDailyGraph)
	mux.HandleFunc("/sectorwise", h.SectorWise)
	mux.HandleFunc("/subindex", h.SubIndex)
	mux.HandleFunc("/alerts", h.Alerts)
	mux.HandleFunc("/market-date", h.ValidMarketDate)
	mux.HandleFunc("/company/news", h.CompanyNews)
	mux.HandleFunc("/company/dividend", h.CompanyDividend)
	mux.HandleFunc("/company/reports", h.CompanyReports)
	mux.HandleFunc("/broker-list", h.BrokerList)
	mux.HandleFunc("/average-trading", h.AverageTrading)
	mux.HandleFunc("/nepse-index", h.NepseIndex)
	mux.HandleFunc("/price-volume", h.PriceVolume)
	mux.HandleFunc("/live-market", h.LiveMarket)
	mux.HandleFunc("/market-open", h.MarketOpen)
	mux.HandleFunc("/graph-index", h.GraphIndex)
	mux.HandleFunc("/banking-index", h.BankingIndex)
	mux.HandleFunc("/dev-bank-index", h.DevBankIndex)
	mux.HandleFunc("/finance-index", h.FinanceIndex)
	mux.HandleFunc("/hotel-index", h.HotelIndex)
	mux.HandleFunc("/hydro-index", h.HydroIndex)
	mux.HandleFunc("/investment-index", h.InvestmentIndex)
	mux.HandleFunc("/life-ins-index", h.LifeInsuranceIndex)
	mux.HandleFunc("/non-life-ins-index", h.NonLifeInsuranceIndex)
	mux.HandleFunc("/manufacturing-index", h.ManufacturingIndex)
	mux.HandleFunc("/microfinance-index", h.MicrofinanceIndex)
	mux.HandleFunc("/trading-index", h.TradingIndex)
	mux.HandleFunc("/mutual-fund-index", h.MutualFundIndex)
	mux.HandleFunc("/others-index", h.OthersIndex)
	mux.HandleFunc("/supply-demand", h.SupplyDemand)
	mux.HandleFunc("/market-summary", h.MarketSummary)
	mux.HandleFunc("/market-summary-history", h.MarketSummaryHistory)
	mux.HandleFunc("/company-list", h.CompanyList)
	mux.HandleFunc("/security-list", h.SecurityLis)
	mux.HandleFunc("/company-profile", h.CompanyProfile)
	mux.HandleFunc("/top-gainer", h.TopGainer)
	mux.HandleFunc("/top-loser", h.TopLoser)
	mux.HandleFunc("/top-turnover", h.TopTurnover)
	mux.HandleFunc("/top-volume", h.TopVolume)
	mux.HandleFunc("/top-transaction", h.TopTransaction)
	mux.HandleFunc("/floor-sheet", h.FloorSheet)
	mux.HandleFunc("/day-floorsheet", h.GetFloorSheet)
	mux.HandleFunc("/todays-price", h.TodaysPrice)
}

func (h *TraderHandler) IndexPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("static", "index.html"))
}

func (h *TraderHandler) CompanyFloorSheet(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	data, err := h.Trader.CompanyFloorSheet(req.CompanyID, req.Size, req.Date)
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) SectorCompanies(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.SectorCompanies()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) ValidMarketDate(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	date = h.Trader.ValidMarketDate(date)
	writeJSONResponse(w, date, nil)
}

func (h *TraderHandler) AnalyzeTotalBuySellByBroker(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	floorsheet, err := h.Trader.GetFloorSheet(req.Date)
	if err != nil {
		writeJSONResponse(w, floorsheet, err)
		return
	}
	companyList, err := h.Trader.CompanyListMap()
	if err != nil {
		writeJSONResponse(w, floorsheet, err)
		return
	}
	data := h.Trader.AnalyzeTotalBuySellByBroker(floorsheet, companyList)
	writeJSONResponse(w, data, nil)
}

func (h *TraderHandler) AnalyzeSectorWise(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	floorsheet, err := h.Trader.GetFloorSheet(req.Date)
	if err != nil {
		writeJSONResponse(w, floorsheet, err)
		return
	}
	companyList, err := h.Trader.CompanyListMap()
	if err != nil {
		writeJSONResponse(w, floorsheet, err)
		return
	}
	data := h.Trader.AnalyzeSectorWise(floorsheet, companyList)
	writeJSONResponse(w, data, nil)
}

func (h *TraderHandler) AnalyzeBrokerWise(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	floorsheet, err := h.Trader.GetFloorSheet(req.Date)
	if err != nil {
		writeJSONResponse(w, floorsheet, err)
		return
	}
	data := h.Trader.AnalyzeBrokerWise(floorsheet)
	writeJSONResponse(w, data, nil)
}

func (h *TraderHandler) AnalyzeBrokerTransactions(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	floorsheet, err := h.Trader.GetFloorSheet(req.Date)
	if err != nil {
		writeJSONResponse(w, floorsheet, err)
		return
	}
	data := h.Trader.AnalyzeBrokerTransactions(floorsheet)
	writeJSONResponse(w, data, nil)
}

func (h *TraderHandler) AnalyzeTopBrokers(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	floorsheet, err := h.Trader.GetFloorSheet(req.Date)
	if err != nil {
		writeJSONResponse(w, floorsheet, err)
		return
	}
	data := h.Trader.AnalyzeTopBrokers(floorsheet)
	writeJSONResponse(w, data, nil)
}

func (h *TraderHandler) AnalyzeTradedStocks(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	floorsheet, err := h.Trader.GetFloorSheet(req.Date)
	if err != nil {
		writeJSONResponse(w, floorsheet, err)
		return
	}
	data := h.Trader.AnalyzeTradedStocks(floorsheet)
	writeJSONResponse(w, data, nil)
}

func (h *TraderHandler) AnalyzeStockVolatility(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	floorsheet, err := h.Trader.GetFloorSheet(req.Date)
	if err != nil {
		writeJSONResponse(w, floorsheet, err)
		return
	}
	data := h.Trader.AnalyzeStockVolatility(floorsheet)
	writeJSONResponse(w, data, nil)
}

func (h *TraderHandler) AnalyzeBuyerSellerMatching(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	floorsheet, err := h.Trader.GetFloorSheet(req.Date)
	if err != nil {
		writeJSONResponse(w, floorsheet, err)
		return
	}
	data := h.Trader.AnalyzeBuyerSellerMatching(floorsheet)
	writeJSONResponse(w, data, nil)
}

func (h *TraderHandler) CompanyPriceVolume(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	data, err := h.Trader.CompanyPriceVolume(req.CompanyID)
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) Holidays(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	data, err := h.Trader.Holidays(req.Year)
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) CompanyTradingHistory(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	data, err := h.Trader.CompanyTradingHistory(req.CompanyID, req.Size, req.StartDate, req.EndDate)
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) CompanyDetail(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	data, err := h.Trader.CompanyDetail(req.CompanyID)
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) CompanyDailyGraph(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	data, err := h.Trader.CompanyDailyGraph(req.CompanyID)
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) SectorWise(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.SectorWise()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) SubIndex(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.SubIndex()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) Alerts(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.Alerts()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) CompanyNews(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	data, err := h.Trader.CompanyNews(req.CompanyID)
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) CompanyReports(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	data, err := h.Trader.CompanyReports(req.CompanyID)
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) CompanyDividend(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	data, err := h.Trader.CompanyDividend(req.CompanyID)
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) BrokerList(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	data, err := h.Trader.BrokerList(req.Size)
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) AverageTrading(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	data, err := h.Trader.AverageTrading(req.Days)
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) NepseIndex(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.Index()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) PriceVolume(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.PriceVolume()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) LiveMarket(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.LiveMarket()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) MarketOpen(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.MarketOpen()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) SupplyDemand(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.SupplyDemand()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) MarketSummary(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.MarketSummary()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) MarketSummaryHistory(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.MarketSummaryHistory()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) CompanyList(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.CompanyList()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) SecurityLis(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.SecurityLis()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) CompanyProfile(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	data, err := h.Trader.CompanyProfile(req.CompanyID)
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) TopGainer(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.TopGainer()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) TopLoser(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.TopLoser()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) TopTurnover(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.TopTurnover()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) TopVolume(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.TopVolume()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) TopTransaction(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.TopTransaction()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) FloorSheet(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	data, err := h.Trader.LiveFloorsheet(req.Page, req.Size, req.Date)
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) GetFloorSheet(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	data, err := h.Trader.GetFloorSheet(req.Date)
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) TodaysPrice(w http.ResponseWriter, r *http.Request) {
	req := parseQueryParams(r, h.Trader.HolidayDates()...)
	data, err := h.Trader.TodayPrice(req.Date, req.Page, req.Size)
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) GraphIndex(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.GraphIndex()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) BankingIndex(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.BankingIndex()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) DevBankIndex(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.DevBankIndex()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) FinanceIndex(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.FinanceIndex()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) HotelIndex(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.HotelIndex()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) HydroIndex(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.HydroIndex()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) InvestmentIndex(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.InvestmentIndex()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) LifeInsuranceIndex(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.LifeInsuranceIndex()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) NonLifeInsuranceIndex(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.NonLifeInsuranceIndex()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) ManufacturingIndex(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.ManufacturingIndex()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) MicrofinanceIndex(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.MicrofinanceIndex()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) TradingIndex(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.TradingIndex()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) MutualFundIndex(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.MutualFundIndex()
	writeJSONResponse(w, data, err)
}

func (h *TraderHandler) OthersIndex(w http.ResponseWriter, _ *http.Request) {
	data, err := h.Trader.OtherIndex()
	writeJSONResponse(w, data, err)
}

func writeJSONResponse(w http.ResponseWriter, data any, err error) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if encodeErr := json.NewEncoder(w).Encode(data); encodeErr != nil {
		http.Error(w, encodeErr.Error(), http.StatusInternalServerError)
	}
}

type HTTPRequest struct {
	CompanyID int    `json:"company_id" query:"company_id" form:"company_id"`
	Page      int    `json:"page" query:"page" form:"page"`
	Size      int    `json:"size" query:"size" form:"size"`
	Symbol    string `json:"symbol" query:"symbol" form:"symbol"`
	StartDate string `json:"start_date" query:"start_date" form:"start_date"`
	EndDate   string `json:"end_date" query:"end_date" form:"end_date"`
	Date      string `json:"date" query:"date" form:"date"`
	Days      int    `json:"days" query:"days" form:"days"`
	Year      int    `json:"year" query:"year" form:"year"`
}

func parseQueryParams(r *http.Request, holidays ...string) HTTPRequest {
	now := getCurrentDate(time.Now(), holidays...)
	currentYear := getCurrentYear(time.Now())
	request := HTTPRequest{
		Date:      now,
		StartDate: now,
		EndDate:   now,
		Symbol:    r.URL.Query().Get("symbol"),
	}
	id, _ := strconv.Atoi(r.URL.Query().Get("company_id"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	date := r.URL.Query().Get("date")
	request.CompanyID = id
	request.Page = page
	request.Size = size
	request.Days = days
	if year == 0 {
		year = currentYear
	}
	request.Year = year
	if startDate != "" {
		request.StartDate = startDate
	}
	if endDate != "" {
		request.EndDate = endDate
	}
	if date != "" {
		request.Date = date
	}
	return request
}

func diffDays(date time.Time, days int) time.Time {
	return date.AddDate(0, 0, days)
}

func getCurrentYear(now time.Time) int {
	if now.Hour() < 11 {
		now = now.AddDate(0, 0, -1)
	}
	var previousDate time.Time
	switch now.Weekday() {
	case time.Friday:
		previousDate = now.AddDate(0, 0, -1)
	case time.Saturday:
		previousDate = now.AddDate(0, 0, -2)
	default:
		previousDate = now
	}
	return previousDate.Year()
}

func getCurrentDate(now time.Time, holidays ...string) string {
	var previousDate time.Time
	switch now.Weekday() {
	case time.Friday:
		previousDate = now.AddDate(0, 0, -1)
	case time.Saturday:
		previousDate = now.AddDate(0, 0, -2)
	default:
		previousDate = now
	}

	date := previousDate.Format(time.DateOnly)
	if len(holidays) == 0 {
		return date
	}
	if slices.Contains(holidays, date) {
		return getCurrentDate(previousDate.AddDate(0, 0, -1), holidays...)
	}
	return date
}

func isHoliday(now time.Time, holidays ...string) bool {
	var previousDate time.Time
	switch now.Weekday() {
	case time.Friday:
		return true
	case time.Saturday:
		return true
	default:
		previousDate = now
	}

	date := previousDate.Format(time.DateOnly)
	if len(holidays) == 0 {
		return false
	}
	return slices.Contains(holidays, date)
}

func isBetween11and3(now time.Time) bool {
	start := time.Date(now.Year(), now.Month(), now.Day(), 11, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, now.Location())
	return now.After(start) && now.Before(end)
}

func isBelow11(now time.Time) bool {
	start := time.Date(now.Year(), now.Month(), now.Day(), 11, 0, 0, 0, now.Location())
	return now.Before(start)
}

func isAfter3(now time.Time) bool {
	return now.After(time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, now.Location()))
}
