package nepse

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/oarkflow/trading/pkg/nepse/models"
	"github.com/oarkflow/trading/pkg/utils"
)

type Trader struct {
	PostPayloadID   int32
	APIAccessTokens struct {
		AccessToken  string
		RefreshToken string
	}
	SaltValues          []int32
	DummyID             int32
	LastDummyIDFetchDay int
	Headers             map[string]string
	Mutex               sync.Mutex
	AuthResponse        AuthenticateResponse
	Token               string
	ServerTime          time.Time
	MarketStatus        models.MarketOpen
	RefreshToken        string
}

// New - Helper to initialize Trader
func New() *Trader {
	return &Trader{
		LastDummyIDFetchDay: 0,
	}
}

func (n *Trader) Connect() (string, error) {
	authenticateResponse, err := utils.Request[AuthenticateResponse](requestURL(AUTH_URI), GET, nil)
	if err != nil {
		return "", err
	}
	n.AuthResponse = authenticateResponse
	n.SaltValues = []int32{n.AuthResponse.Salt1, n.AuthResponse.Salt2, n.AuthResponse.Salt3, n.AuthResponse.Salt4, n.AuthResponse.Salt5}
	n.Token = authenticateResponse.GetParsedAccessToken()
	n.RefreshToken = authenticateResponse.GetParsedRefreshToken()
	n.ServerTime = time.UnixMilli(authenticateResponse.ServerTime)
	fmt.Println("Reconnecting...")
	return n.Token, nil
}

func (n *Trader) getAuthHeader() (map[string]string, error) {
	var err error
	token := n.Token
	if n.Token == "" {
		token, err = n.Connect()
		if err != nil {
			return nil, err
		}
		go func() {
			<-time.After(TokenClearInterval)
			n.Token = ""
		}()
	}
	headers := map[string]string{"Authorization": "Salter " + token}
	return headers, nil
}

func (n *Trader) SectorCompanies() (map[string][]models.Company, error) {
	file := filepath.Join(DataPath, "sector_companies.json")
	if utils.FileExists(file) {
		return utils.ReadJsonFile[map[string][]models.Company](file)
	}
	sectorFile := filepath.Join(DataPath, "sector.json")
	companyFile := filepath.Join(DataPath, "companies.json")
	sectorCompanies := make(map[string][]models.Company)
	sectors, err := utils.ReadJsonFile[[]models.Sector](sectorFile)
	if err != nil {
		return nil, err
	}
	companies, err := utils.ReadJsonFile[[]models.Company](companyFile)
	if err != nil {
		return nil, err
	}
	for _, sector := range sectors {
		for _, company := range companies {
			if sector.SectorName == company.SectorName {
				sectorCompanies[sector.SectorName] = append(sectorCompanies[sector.SectorName], company)
			}
		}
	}
	err = utils.WriteJsonFile(file, sectorCompanies)
	return sectorCompanies, err
}

func (n *Trader) CompanyFloorSheet(id, size int, date string) (map[string]any, error) {
	date = getCurrentDate(parseDate(date), n.HolidayDates()...)
	file := getCompanyFloorsheetPath(id, date)
	if utils.FileExists(file) {
		return utils.ReadJsonFile[map[string]any](file)
	}
	size = utils.GetSize(size)
	uri := fmt.Sprintf(SECURITIES_FLOORSHEET+"/%d?businessDate=%s&size=%d&sort=contractid,desc", id, date, size)
	payload, headers, err := n.floorSheetPayload()
	if err != nil {
		return nil, err
	}
	data, err := nepseRequest[map[string]any](n, uri, POST, bytes.NewReader(payload), headers)
	if err != nil {
		return nil, err
	}
	err = utils.WriteJsonFile(file, data)
	return data, nil
}

func parseDate(date string) time.Time {
	d, err := time.Parse(time.DateOnly, date)
	if err != nil {
		fmt.Println("Error parsing date:", err)
		return time.Time{}
	}
	currentTime := time.Now().Format(time.TimeOnly)
	dateTimeStr := fmt.Sprintf("%s %s", d.Format(time.DateOnly), currentTime)
	dateTime, err := time.Parse(time.DateTime, dateTimeStr)
	if err != nil {
		fmt.Println("Error parsing combined date and time:", err)
		return time.Time{}
	}
	return dateTime
}

func (n *Trader) CompanyPriceVolume(id int) (map[string]any, error) {
	uri := fmt.Sprintf(SECURITIES_PRICE_VOLUME+"/%d", id)
	return nepseRequest[map[string]any](n, uri, GET, nil)
}

func (n *Trader) ValidMarketDate(date string) string {
	return getCurrentDate(parseDate(date), n.HolidayDates()...)
}

func (n *Trader) Holidays(year int) ([]models.Holiday, error) {
	file := filepath.Join(DataPath, "holidays.json")
	if utils.FileExists(file) {
		return utils.ReadJsonFile[[]models.Holiday](file)
	}
	uri := fmt.Sprintf(HOLIDAYS+"?year=%d", year)
	data, err := nepseRequest[[]models.Holiday](n, uri, GET, nil)
	if err != nil {
		return nil, err
	}
	err = utils.WriteJsonFile(file, data)
	return data, err
}

func (n *Trader) HolidayDates() (data []string) {
	holidays, err := n.Holidays(2025)
	if err != nil {
		return
	}

	for _, holiday := range holidays {
		data = append(data, holiday.Date)
	}
	return
}

func (n *Trader) GetFile(file string) ([]map[string]any, error) {
	uri := fmt.Sprintf(FILE_LINK+"?fileLocation=%s", file)
	data, err := nepseRequest[[]map[string]any](n, uri, GET, nil)
	if err != nil {
		return nil, err
	}
	err = utils.WriteJsonFile(file, data)
	return data, err
}

func (n *Trader) CompanyTradingHistory(id, size int, startDate, endDate string) (map[string]any, error) {
	startDate = getCurrentDate(parseDate(startDate), n.HolidayDates()...)
	endDate = getCurrentDate(parseDate(endDate), n.HolidayDates()...)
	size = utils.GetSize(size)
	uri := fmt.Sprintf(SECURITIES_TRADING_HISTORY+"/%d?size=%d&startDate=%s&endDate=%s", id, size, startDate, endDate)
	return nepseRequest[map[string]any](n, uri, GET, nil)
}

func (n *Trader) CompanyDetail(id int) (map[string]any, error) {
	payload, headers, err := n.payloadFromMarket()
	if err != nil {
		return nil, err
	}
	uri := fmt.Sprintf(SECURITIES_DETAIL+"/%d", id)
	return nepseRequest[map[string]any](n, uri, POST, bytes.NewReader(payload), headers)
}

func (n *Trader) CompanyDailyGraph(id int) ([]map[string]any, error) {
	uri := fmt.Sprintf(SECURITIES_DAILY_GRAPH+"/%d", id)
	return nepseRequest[[]map[string]any](n, uri, GET, nil)
}

func (n *Trader) SectorWise() ([]map[string]any, error) {
	file := filepath.Join(DataPath, "sector.json")
	if utils.FileExists(file) {
		return utils.ReadJsonFile[[]map[string]any](file)
	}
	data, err := nepseRequest[[]map[string]any](n, SECTORWISE, GET, nil)
	if err != nil {
		return nil, err
	}
	err = utils.WriteJsonFile(file, data)
	return data, err
}

func (n *Trader) SubIndex() ([]map[string]any, error) {
	return nepseRequest[[]map[string]any](n, SUBINDEX, GET, nil)
}

func (n *Trader) Alerts() ([]map[string]any, error) {
	return nepseRequest[[]map[string]any](n, ALERTS, GET, nil)
}

func (n *Trader) CompanyNews(id any) ([]map[string]any, error) {
	url := fmt.Sprintf("%s/%v", COMPANY_NEWS, id)
	return nepseRequest[[]map[string]any](n, url, GET, nil)
}

func (n *Trader) CompanyReports(id any) ([]map[string]any, error) {
	url := fmt.Sprintf("%s/%v", COMPANY_REPORTS, id)
	return nepseRequest[[]map[string]any](n, url, GET, nil)
}

func (n *Trader) CompanyDividend(id any) ([]map[string]any, error) {
	url := fmt.Sprintf("%s/%v", COMPANY_DIVIDEND, id)
	return nepseRequest[[]map[string]any](n, url, GET, nil)
}

func (n *Trader) BrokerList(size int) (models.BrokerData, error) {
	file := filepath.Join(DataPath, "broker_list.json")
	if utils.FileExists(file) {
		return utils.ReadJsonFile[models.BrokerData](file)
	}
	size = utils.GetSize(size)
	uri := BROKER_LIST + "?size=" + fmt.Sprintf("%d", size)
	data, err := nepseRequest[models.BrokerData](n, uri, GET, nil)
	if err != nil {
		return models.BrokerData{}, err
	}
	err = utils.WriteJsonFile(file, data)
	return data, err
}

func (n *Trader) AverageTrading(days int) ([]map[string]any, error) {
	if days == 0 {
		days = 1
	}
	uri := fmt.Sprintf(AVERAGE_TRADING+"?nDays=%d", days)
	return nepseRequest[[]map[string]any](n, uri, GET, nil)
}

func (n *Trader) Index() ([]map[string]any, error) {
	return nepseRequest[[]map[string]any](n, NEPSE_INDEX, GET, nil)
}

func (n *Trader) PriceVolume() ([]map[string]any, error) {
	return nepseRequest[[]map[string]any](n, PRICE_VOLUME, GET, nil)
}

func (n *Trader) index(uri string) ([]any, error) {
	payload, headers, err := n.indexPayload()
	if err != nil {
		return nil, err
	}
	return nepseRequest[[]any](n, uri, POST, bytes.NewReader(payload), headers)
}

func (n *Trader) GraphIndex() ([]any, error) {
	return n.index(GRAPH_INDEX)
}

func (n *Trader) BankingIndex() ([]any, error) {
	return n.index(BANKING_INDEX)
}

func (n *Trader) DevBankIndex() ([]any, error) {
	return n.index(DEV_BANK_INDEX)
}

func (n *Trader) FinanceIndex() ([]any, error) {
	return n.index(FINANCE_INDEX)
}

func (n *Trader) HotelIndex() ([]any, error) {
	return n.index(HOTEL_INDEX)
}

func (n *Trader) HydroIndex() ([]any, error) {
	return n.index(HYDRO_INDEX)
}

func (n *Trader) InvestmentIndex() ([]any, error) {
	return n.index(INVESTMENT_INDEX)
}

func (n *Trader) LifeInsuranceIndex() ([]any, error) {
	return n.index(LIFE_INS_INDEX)
}

func (n *Trader) NonLifeInsuranceIndex() ([]any, error) {
	return n.index(NON_LIFE_INS_INDEX)
}

func (n *Trader) ManufacturingIndex() ([]any, error) {
	return n.index(MANUFACTURING_INDEX)
}

func (n *Trader) MicrofinanceIndex() ([]any, error) {
	return n.index(MICROFINANCE_INDEX)
}

func (n *Trader) TradingIndex() ([]any, error) {
	return n.index(TRADING_INDEX)
}

func (n *Trader) OtherIndex() ([]any, error) {
	return n.index(OTHERS_INDEX)
}

func (n *Trader) MutualFundIndex() ([]any, error) {
	return n.index(MUTUAL_FUND_INDEX)
}

func (n *Trader) MarketOpen() (models.MarketOpen, error) {
	d, err := nepseRequest[models.MarketOpen](n, MARKET_OPEN, GET, nil)
	n.MarketStatus = d
	return d, err
}

func (n *Trader) SupplyDemand() (map[string]any, error) {
	return nepseRequest[map[string]any](n, SUPPLY_DEMAND, GET, nil)
}

func (n *Trader) MarketSummary() ([]map[string]any, error) {
	return nepseRequest[[]map[string]any](n, MARKET_SUMMARY, GET, nil)
}

func (n *Trader) MarketSummaryHistory() ([]map[string]any, error) {
	return nepseRequest[[]map[string]any](n, MARKET_SUMMARY_HISTORY, GET, nil)
}

func filterDeactivateCompanies(companies []models.Company) (res []models.Company) {
	for _, company := range companies {
		if company.Status == "A" {
			res = append(res, company)
		}
	}
	return
}

func (n *Trader) LiveMarket() ([]map[string]any, error) {
	return nepseRequest[[]map[string]any](n, LIVE_MARKET, GET, nil)
}

func (n *Trader) CompanyList() ([]models.Company, error) {
	file := filepath.Join(DataPath, "companies.json")
	if utils.FileExists(file) {
		data, err := utils.ReadJsonFile[[]models.Company](file)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	data, err := nepseRequest[[]models.Company](n, COMPANY_LIST, GET, nil)
	if err != nil {
		return nil, err
	}
	err = utils.WriteJsonFile(file, filterDeactivateCompanies(data))
	return filterDeactivateCompanies(data), err
}

func (n *Trader) SecurityLis() ([]map[string]any, error) {
	file := filepath.Join(DataPath, "securities.json")
	if utils.FileExists(file) {
		return utils.ReadJsonFile[[]map[string]any](file)
	}
	data, err := nepseRequest[[]map[string]any](n, COMPANY_LIST, GET, nil)
	if err != nil {
		return nil, err
	}
	err = utils.WriteJsonFile(file, data)
	return data, err
}

func (n *Trader) CompanyProfile(id any) (map[string]any, error) {
	file := filepath.Join(DataPath, fmt.Sprintf("profile-%v.json", id))
	if utils.FileExists(file) {
		return utils.ReadJsonFile[map[string]any](file)
	}
	uri := fmt.Sprintf(COMPANY_PROFILE+"/%v", id)
	data, err := nepseRequest[map[string]any](n, uri, GET, nil)
	if err != nil {
		return nil, err
	}
	err = utils.WriteJsonFile(file, data)
	return data, err
}

func (n *Trader) CompanyListMap() (map[string]models.Company, error) {
	companies, err := n.CompanyList()
	if err != nil {
		return nil, err
	}
	c := make(map[string]models.Company)
	for _, company := range companies {
		c[company.Symbol] = company
	}
	return c, nil
}

func (n *Trader) TopGainer() ([]map[string]any, error) {
	return nepseRequest[[]map[string]any](n, TOP_GAINER, GET, nil)
}

func (n *Trader) TopLoser() ([]map[string]any, error) {
	return nepseRequest[[]map[string]any](n, TOP_LOSER, GET, nil)
}

func (n *Trader) TopTurnover() ([]map[string]any, error) {
	return nepseRequest[[]map[string]any](n, TOP_TURNOVER, GET, nil)
}

func (n *Trader) TopVolume() ([]map[string]any, error) {
	return nepseRequest[[]map[string]any](n, TOP_VOLUME, GET, nil)
}

func (n *Trader) TopTransaction() ([]map[string]any, error) {
	return nepseRequest[[]map[string]any](n, TOP_TRANSACTION, GET, nil)
}

func (n *Trader) AggregateCompanyFloorsheets() error {
	baseDir := filepath.Join(DataPath, CompanyFloorsheetPath)
	companyDirs, err := os.ReadDir(baseDir)
	if err != nil {
		return err
	}
	dateWiseData := make(map[string]models.LiveFloorsheet)
	dateFiles, err := os.ReadDir(filepath.Join(DataPath, DateFloorsheetPath))
	var existingDateFiles []string
	for _, d := range dateFiles {
		if d.IsDir() {
			continue
		}
		if strings.HasSuffix(d.Name(), ".json") {
			existingDateFiles = append(existingDateFiles, d.Name())
		}
	}
	for _, company := range companyDirs {
		if !company.IsDir() {
			continue
		}
		companyPath := filepath.Join(baseDir, company.Name())

		files, err := os.ReadDir(companyPath)
		if err != nil {
			log.Printf("Error reading company directory %s: %v", companyPath, err)
			continue
		}

		dateFiles := make(map[string][]string)
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".json") && !strings.Contains(file.Name(), "progress.json") {
				date := strings.TrimSuffix(file.Name(), ".json")
				dateFiles[date] = append(dateFiles[date], file.Name())
			}
		}

		for date, jsonFiles := range dateFiles {
			sort.Strings(jsonFiles)
			for _, file := range jsonFiles {
				if slices.Contains(existingDateFiles, file) {
					continue
				}
				filePath := filepath.Join(companyPath, file)
				floorsheet, err := utils.ReadJsonFile[models.LiveFloorsheet](filePath)
				if err != nil {
					log.Printf("Error reading file %s: %v", filePath, err)
					continue
				}
				aggregatedData := dateWiseData[date]
				aggregatedData.Content = append(aggregatedData.Content, floorsheet.Content...)
				aggregatedData.TotalAmount += floorsheet.TotalAmount
				aggregatedData.TotalQty += floorsheet.TotalQty
				aggregatedData.TotalTrades += floorsheet.TotalTrades

				dateWiseData[date] = aggregatedData
			}
		}
	}
	for date, aggregated := range dateWiseData {
		if len(aggregated.Content) == 0 {
			continue
		}

		sort.SliceStable(aggregated.Content, func(i, j int) bool {
			timeI, timeJ := time.Time{}, time.Time{}

			if serverTimeI, ok := aggregated.Content[i]["serverTime"].(string); ok {
				parsedTime, err := time.Parse(time.RFC3339, serverTimeI)
				if err == nil {
					timeI = parsedTime
				}
			}

			if serverTimeJ, ok := aggregated.Content[j]["serverTime"].(string); ok {
				parsedTime, err := time.Parse(time.RFC3339, serverTimeJ)
				if err == nil {
					timeJ = parsedTime
				}
			}

			return timeI.After(timeJ)
		})

		if err := utils.WriteJsonFile(getDateFloorsheetPath(date), aggregated); err != nil {
			log.Printf("Error writing aggregated data for %s: %v", date, err)
		}
	}
	return nil
}

func (n *Trader) GetFloorSheet(date string) (models.FloorSheet, error) {
	date = getCurrentDate(parseDate(date), n.HolidayDates()...)
	file := getDateFloorsheetPath(date)
	if utils.FileExists(file) {
		data, err := utils.ReadJsonFile[models.FloorSheet](file)
		if err != nil {
			return data, err
		}
		if data.Date == "" {
			data.Date = date
		}
		return data, nil
	}
	return models.FloorSheet{}, nil
}

func (n *Trader) LiveFloorsheet(page, size int, dates ...string) (models.LiveFloorsheet, error) {
	var date string
	if len(dates) > 0 {
		date = dates[0]
	} else {
		date = getCurrentDate(time.Now(), n.HolidayDates()...)
	}
	size = utils.GetSize(size)
	lastPage, lastNumItems, err := loadSummary()
	if err != nil {
		return models.LiveFloorsheet{}, err
	}
	fetchPage := lastPage - page
	if lastNumItems < MaxPageSize {
		fetchPage--
	}
	file := getLiveFloorsheetPath(page)
	if utils.FileExists(file) {
		return utils.ReadJsonFile[models.LiveFloorsheet](file)
	}
	file = getDateFloorsheetPath(date)
	if utils.FileExists(file) {
		return utils.ReadJsonFile[models.LiveFloorsheet](file)
	}
	return models.LiveFloorsheet{}, nil
}

func (n *Trader) FloorSheet(page, size int) (models.LiveFloorsheet, error) {
	size = utils.GetSize(size)
	payload, headers, err := n.floorSheetPayload()
	if err != nil {
		return models.LiveFloorsheet{}, err
	}
	uri := fmt.Sprintf(FLOORSHEET+"?page=%d&size=%d&sort=contractId,desc", page, size)
	return nepseRequest[models.LiveFloorsheet](n, uri, POST, bytes.NewReader(payload), headers)
}

func loadSummary() (int, int, error) {
	filePath := getLiveSummaryPath()
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return 0, 0, nil
	}
	summary, err := utils.ReadJsonFile[map[string]any](filePath)
	if err != nil {
		return 0, 0, err
	}
	page, _ := summary["currentPage"].(float64)
	numItems, _ := summary["numberOfElements"].(float64)
	return int(page), int(numItems), nil
}

func (n *Trader) FetchLiveData() {
	for {
		now := time.Now()
		if isHoliday(now, n.HolidayDates()...) {
			break
		}
		if isAfter3(now) {
			break
		}
		if isBelow11(now) {
			waitUntil := time.Date(now.Year(), now.Month(), now.Day(), 11, 0, 0, 0, now.Location())
			waitDuration := waitUntil.Sub(now)
			fmt.Printf("Waiting till %s\n", waitDuration)
			time.Sleep(waitDuration)
			continue
		}
		n.FetchAndStoreFloorSheet()
	}
	now := time.Now()
	if isHoliday(now, n.HolidayDates()...) {
		return
	}
	if isAfter3(now) {
		date := now.Format(time.DateOnly)
		err := n.CollectAllCompanyFloorSheets(date)
		if err == nil {
			err = n.CollectAllFloorSheets()
			if err == nil {
				os.RemoveAll(getLivePath())
				fmt.Println("\nFloorsheet updated for", date)
			}
		}
	}
}

func (n *Trader) FetchAndStoreFloorSheet() {
	lastPage, lastNumItems, err := loadSummary()
	if err != nil {
		log.Printf("Error loading summary.json: %v", err)
		return
	}
	page := lastPage
	if lastNumItems < MaxPageSize {
		log.Printf("Re-fetching incomplete page %d", page)
	} else {
		page++
	}
	isLastPage := false
	for {
		if time.Now().Hour() > 15 || isLastPage {
			break
		}
		filePath := getLiveFloorsheetPath(page)
		var response models.LiveFloorsheet
		for {
			response, err = n.FloorSheet(page, MaxPageSize)
			if err != nil {
				log.Printf("Error fetching floorsheet page %d: %v", page, err)
				<-time.After(2 * time.Second)
				continue
			}
			if response.Floorsheets.Last {
				isLastPage = true
			}
			if len(response.Floorsheets.Content) < MaxPageSize {
				if err := utils.WriteJsonFile(filePath, response); err != nil {
					log.Printf("Error saving file %s: %v", filePath, err)
				}
				<-time.After(2 * time.Second)
				continue
			}
			break
		}
		if err := utils.WriteJsonFile(filePath, response); err != nil {
			log.Printf("Error saving file %s: %v", filePath, err)
		}

		summary := map[string]interface{}{
			"totalAmount":      response.TotalAmount,
			"totalQty":         response.TotalQty,
			"totalTrades":      response.TotalTrades,
			"totalPages":       response.Floorsheets.TotalPages,
			"currentPage":      response.Floorsheets.Number,
			"numberOfElements": response.Floorsheets.NumberOfItems,
		}
		if err := utils.WriteJsonFile(getLiveSummaryPath(), summary); err != nil {
			log.Printf("Error saving summary.json: %v", err)
		}
		page++
		<-time.After(2 * time.Second)
		log.Printf("Fetching next page: %d", page)
	}
}

func (n *Trader) CollectAllFloorSheets() error {
	var last bool
	page := 0
	var businessDate string
	size := 500
	size = utils.GetSize(size)
	payload, headers, err := n.floorSheetPayload()
	if err != nil {
		return err
	}
	existingFile := getDateFloorsheetPath(time.Now().Format(time.DateOnly))
	if utils.FileExists(existingFile) {
		fmt.Println("Floorsheet already exists")
		return nil
	}
	fileName := getFloorsheetProgressPath()
	var allData []map[string]any
	var totalAmount float64
	var totalQty float64
	var totalTrades float64

	// Resume from last saved page if exists
	if existingData, err := utils.ReadJsonFile[map[string]any](fileName); err == nil {
		if lastPage, ok := existingData["lastPage"].(float64); ok {
			page = int(lastPage) + 1
		}
		if content, ok := existingData["content"].([]map[string]any); ok {
			allData = content
		}
		totalAmount, _ = existingData["totalAmount"].(float64)
		totalQty, _ = existingData["totalQty"].(float64)
		totalTrades, _ = existingData["totalTrades"].(float64)
	}

	for !last {
		uri := fmt.Sprintf(FLOORSHEET+"?page=%d&size=%d&sort=contractId,desc", page, size)
		var response map[string]any
		for retry := 0; retry < 400; retry++ {
			response, err = nepseRequest[map[string]any](n, uri, POST, bytes.NewReader(payload), headers)
			if err == nil {
				break
			}
			if retry%5 == 0 {
				n.Token = ""
				fmt.Println("Reset Token")
			}
			log.Printf("Error fetching page %d: %v in CollectAllFloorSheets. Retrying...", page, err)
			time.Sleep(1 * time.Second)
		}
		if err != nil {
			return err
		}
		if response == nil {
			last = true
			continue
		}
		floorsheets, ok := response["floorsheets"].(map[string]any)
		if !ok {
			return fmt.Errorf("invalid response structure: missing 'floorsheets'")
		}

		content, ok := floorsheets["content"].([]any)
		if !ok {
			return fmt.Errorf("invalid response structure: missing 'content'")
		}

		if len(content) > 0 && businessDate == "" {
			firstItem, ok := content[0].(map[string]any)
			if ok {
				businessDate, _ = firstItem["businessDate"].(string)
			}
		}

		for _, item := range content {
			if dataMap, ok := item.(map[string]any); ok {
				allData = append(allData, dataMap)
			}
		}

		totalAmount, _ = response["totalAmount"].(float64)
		totalQty, _ = response["totalQty"].(float64)
		totalTrades, _ = response["totalTrades"].(float64)

		last, _ = floorsheets["last"].(bool)

		// Save progress after each page
		progress := map[string]any{
			"lastPage":    page,
			"content":     allData,
			"totalAmount": totalAmount,
			"totalQty":    totalQty,
			"totalTrades": totalTrades,
		}
		utils.WriteJsonFile(fileName, progress)
		fmt.Printf("%d, ", page)
		for k := range response {
			delete(response, k)
		}
		clear(response)
		page++
		<-time.After(time.Second)
	}

	if businessDate == "" {
		return fmt.Errorf("businessDate not found in response")
	}

	finalFileName := getDateFloorsheetPath(businessDate)
	err = utils.WriteJsonFile(finalFileName, map[string]any{
		"date":        businessDate,
		"content":     allData,
		"totalAmount": totalAmount,
		"totalQty":    totalQty,
		"totalTrades": totalTrades,
	})
	if err != nil {
		return err
	}
	return os.Remove(fileName)
}

func (n *Trader) CollectAllCompanyFloorSheets(dates ...string) error {
	companies, err := n.CompanyList()
	if err != nil {
		return err
	}
	for _, date := range dates {
		now, _ := time.Parse(time.DateOnly, date)
		if isHoliday(now, n.HolidayDates()...) {
			continue
		}
		for _, company := range companies {
			baseDir := getCompanyFloorsheetPath(company.Id, date)
			if _, err := os.Stat(baseDir); os.IsNotExist(err) {
				n.CollectCompanyFloorSheets(company.Id, date)
				<-time.After(1 * time.Second)
			}
		}
	}

	return nil
}

func (n *Trader) CollectCompanyFloorSheets(id int, dates ...string) error {
	var date string
	if len(dates) > 0 {
		date = dates[0]
	}
	var last bool
	page := 0
	var businessDate string
	size := 500
	size = utils.GetSize(size)
	fileName := getCompanyProgressPath(id)
	var allData []map[string]any
	var totalAmount float64
	var totalQty float64
	var totalTrades float64

	// Resume from last saved page if exists
	if existingData, err := utils.ReadJsonFile[map[string]any](fileName); err == nil {
		if lastPage, ok := existingData["lastPage"].(float64); ok {
			page = int(lastPage) + 1
		}
		if content, ok := existingData["content"].([]map[string]any); ok {
			allData = content
		}
		totalAmount, _ = existingData["totalAmount"].(float64)
		totalQty, _ = existingData["totalQty"].(float64)
		totalTrades, _ = existingData["totalTrades"].(float64)
	}

	fmt.Println("\nCompany", id, "Date", date, "Floorsheet")
	for !last {
		payload, headers, err := n.floorSheetPayload()
		if err != nil {
			return err
		}
		uri := fmt.Sprintf(SECURITIES_FLOORSHEET+"/%d?businessDate=%s&size=%d&page=%d&sort=contractid,desc", id, date, size, page)
		var response map[string]any
		for retry := 0; retry < 400; retry++ {
			response, err = nepseRequest[map[string]any](n, uri, POST, bytes.NewReader(payload), headers)
			if err == nil {
				break
			}
			if retry%5 == 0 {
				n.Token = ""
				fmt.Println("Reset Token")
			}
			log.Printf("Error fetching page %d: %v in CollectCompanyFloorSheets. Retrying...", page, err)
			time.Sleep(1 * time.Second)
		}
		if err != nil {
			return err
		}
		if response == nil {
			last = true
			continue
		}
		floorsheets, ok := response["floorsheets"].(map[string]any)
		if !ok {
			return fmt.Errorf("invalid response structure: missing 'floorsheets'")
		}

		content, ok := floorsheets["content"].([]any)
		if !ok {
			return fmt.Errorf("invalid response structure: missing 'content'")
		}

		if len(content) > 0 && businessDate == "" {
			firstItem, ok := content[0].(map[string]any)
			if ok {
				businessDate, _ = firstItem["businessDate"].(string)
			}
		}

		for _, item := range content {
			if dataMap, ok := item.(map[string]any); ok {
				allData = append(allData, dataMap)
			}
		}

		totalAmount, _ = response["totalAmount"].(float64)
		totalQty, _ = response["totalQty"].(float64)
		totalTrades, _ = response["totalTrades"].(float64)

		last, _ = floorsheets["last"].(bool)

		// Save progress after each page
		progress := map[string]any{
			"lastPage":    page,
			"content":     allData,
			"totalAmount": totalAmount,
			"totalQty":    totalQty,
			"totalTrades": totalTrades,
		}
		utils.WriteJsonFile(fileName, progress)
		fmt.Printf("%d, ", page)
		for k := range response {
			delete(response, k)
		}
		clear(response)
		page++
	}

	if businessDate == "" {
		businessDate = date
	}

	finalFileName := getCompanyFloorsheetPath(id, businessDate)
	err := utils.WriteJsonFile(finalFileName, map[string]any{
		"date":        businessDate,
		"content":     allData,
		"totalAmount": totalAmount,
		"totalQty":    totalQty,
		"totalTrades": totalTrades,
	})
	if err != nil {
		return err
	}
	return os.Remove(fileName)
}

func (n *Trader) TodayPrice(date string, page, size int) (models.TodayPrice, error) {
	date = getCurrentDate(parseDate(date), n.HolidayDates()...)
	var t models.TodayPrice
	file := getPricePath(date)
	if utils.FileExists(file) {
		return utils.ReadJsonFile[models.TodayPrice](file)
	}
	size = utils.GetSize(size)
	payload, headers, err := n.floorSheetPayload()
	if err != nil {
		return t, err
	}
	uri := fmt.Sprintf(TODAY_PRICE+"?page=%d&size=%d", page, size)
	data, err := nepseRequest[models.TodayPrice](n, uri, POST, bytes.NewReader(payload), headers)
	if err != nil {
		return t, err
	}
	if len(data.Content) > 0 {
		if data.Content[0].BusinessDate == date {
			err = utils.WriteJsonFile(file, data)
			return data, err
		}
	}
	return data, err
}

func (n *Trader) floorSheetPayload() ([]byte, map[string]string, error) {
	return n.getPayload(n.getPayloadForFloorSheet)
}

func (n *Trader) indexPayload() ([]byte, map[string]string, error) {
	return n.getPayload(n.getPayloadForIndex)
}

func (n *Trader) defaultPayload() ([]byte, map[string]string, error) {
	return n.getPayload(n.getDefaultPayload)
}

func (n *Trader) payloadFromMarket() ([]byte, map[string]string, error) {
	return n.getPayload(n.getPayloadFromMarketID)
}

func (n *Trader) getPayload(fn func() (int32, error)) ([]byte, map[string]string, error) {
	headers, err := n.getAuthHeader()
	if err != nil {
		return nil, nil, err
	}
	id, err := fn()
	if err != nil {
		return nil, nil, err
	}
	floorSheetPayload := []byte(fmt.Sprintf(`{"id":%d}`, id))
	headers["Content-Type"] = "application/json"
	headers["Content-Length"] = fmt.Sprintf("%d", len(floorSheetPayload))
	return floorSheetPayload, headers, nil
}

func nepseRequest[T any](n *Trader, uri, method string, body io.Reader, header ...map[string]string) (T, error) {
	var t T
	var err error
	var headers map[string]string
	if len(header) > 0 {
		headers = header[0]
	} else {
		headers, err = n.getAuthHeader()
		if err != nil {
			return t, err
		}
	}
	return utils.Request[T](requestURL(uri), method, body, headers)
}

func (n *Trader) getPayloadForIndex() (int32, error) {
	return n.identifyIndex(3, 5, 1)
}

func (n *Trader) getPayloadForFloorSheet() (int32, error) {
	return n.identifyIndex(1, 4, 3)
}

func (n *Trader) getDefaultPayload() (int32, error) {
	return n.identifyIndex(3, 5, 1)
}

func (n *Trader) getPayloadFromMarketID() (int32, error) {
	id, _, err := n.getIDForScrips()
	return id, err
}

func (n *Trader) identifyIndex(defVal, comp, assign int32) (int32, error) {
	e, day, err := n.getIDForScrips()
	if err != nil {
		return 0, err
	}
	index := utils.GetIndex(e, defVal, comp, assign)
	nValue := e + n.SaltValues[index]*day - n.SaltValues[index-1]
	n.PostPayloadID = nValue
	return nValue, nil
}

func (n *Trader) getIDForScrips() (int32, int32, error) {
	mo, err := n.MarketOpen()
	if err != nil {
		return 0, 0, err
	}
	day := n.ServerTime.Day()
	e := DummyData[mo.ID] + mo.ID + 2*int32(day)
	return e, int32(day), nil
}
