package nepse

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/oarkflow/convert"
	"github.com/oarkflow/errors"
	"github.com/oarkflow/log"
	"github.com/oarkflow/search"

	"github.com/jumpei00/gostocktrade/nepse/csv"
)

func InitCSVStock() {
	files, err := loadAllCSVFiles("./data/date")
	if err != nil {
		panic(err)
	}
	engine, err := search.SetEngine[map[string]any]("stock", &search.Config{})
	log.Info().Msg("Indexing stock")
	engine.InsertWithPool(files, runtime.NumCPU(), 1000)
	log.Info().Msg("Indexed stock")
}

type StockData struct {
	Symbol               string  `csv:"Symbol"`
	Date                 string  `csv:"Date"`
	Confidence           float64 `csv:"Confidence"`
	OpenPrice            float64 `csv:"OpenPrice"`
	HighPrice            float64 `csv:"HighPrice"`
	LowPrice             float64 `csv:"LowPrice"`
	ClosePrice           float64 `csv:"ClosePrice"`
	VWAP                 float64 `csv:"VWAP"`
	Volume               float64 `csv:"Volume"`
	PreviousClose        float64 `csv:"PreviousClose"`
	Turnover             float64 `csv:"Turnover"`
	Transactions         int     `csv:"Transactions"`
	Difference           float64 `csv:"Difference"`
	Range                float64 `csv:"Range"`
	DifferencePercentage float64 `csv:"DifferencePercentage"`
	RangePercentage      float64 `csv:"RangePercentage"`
	VWAPPercentage       float64 `csv:"VWAPPercentage"`
	_120Days             float64 `csv:"120Days"`
	_180Days             float64 `csv:"180Days"`
	_52WeeksHigh         float64 `csv:"52WeeksHigh"`
	_52WeeksLow          float64 `csv:"52WeeksLow"`
}

func parseFloat(value string) (float64, error) {
	value = strings.ReplaceAll(value, ",", "")
	return strconv.ParseFloat(value, 64)
}

func parseInt(value string) (int64, error) {
	value = strings.ReplaceAll(value, ",", "")
	return strconv.ParseInt(value, 10, 64)
}

// RemoveCommas removes commas from numeric strings and converts them to float64.
func removeCommas(value string) string {
	return strings.ReplaceAll(value, ",", "")
}

// ConvertMapToStockData converts a map[string]any to a StockData instance.
func convertMapToStockData(dataMap map[string]any) (StockData, error) {
	var stock StockData

	// Helper function to convert values and handle errors
	parseFloat := func(key string) (float64, error) {
		switch val := dataMap[key].(type) {
		case string:
			cleanedVal := removeCommas(val)
			if cleanedVal != "-" {
				return strconv.ParseFloat(cleanedVal, 64)
			} else {
				return 0, nil
			}
		case float64:
			return val, nil
		default:
			return 0, fmt.Errorf("value for key %s is not a %v", key, reflect.TypeOf(val))
		}
	}

	// Convert map values to StockData fields
	var err error
	var ok bool
	if stock.Symbol, ok = dataMap["Symbol"].(string); !ok {
		return stock, fmt.Errorf("invalid type for Symbol")
	}

	stock.Confidence, err = parseFloat("Confidence")
	if err != nil {
		return stock, err
	}
	stock.OpenPrice, err = parseFloat("OpenPrice")
	if err != nil {
		return stock, err
	}
	stock.HighPrice, err = parseFloat("HighPrice")
	if err != nil {
		return stock, err
	}
	stock.LowPrice, err = parseFloat("LowPrice")
	if err != nil {
		return stock, err
	}
	stock.ClosePrice, err = parseFloat("ClosePrice")
	if err != nil {
		return stock, err
	}
	stock.VWAP, err = parseFloat("VWAP")
	if err != nil {
		return stock, err
	}
	stock.Volume, err = parseFloat("Volume")
	if err != nil {
		return stock, err
	}
	stock.PreviousClose, err = parseFloat("PreviousClose")
	if err != nil {
		return stock, err
	}
	stock.Turnover, err = parseFloat("Turnover")
	if err != nil {
		return stock, err
	}

	if transactions, ok := convert.ToInt(dataMap["Transactions"]); ok {
		stock.Transactions = transactions
	} else {
		return stock, fmt.Errorf("invalid type for Transactions")
	}

	stock.Difference, err = parseFloat("Difference")
	if err != nil {
		return stock, err
	}
	stock.Range, err = parseFloat("Range")
	if err != nil {
		return stock, err
	}
	stock.DifferencePercentage, err = parseFloat("DifferencePercentage")
	if err != nil {
		return stock, err
	}
	stock.RangePercentage, err = parseFloat("RangePercentage")
	if err != nil {
		return stock, err
	}
	stock.VWAPPercentage, err = parseFloat("VWAPPercentage")
	if err != nil {
		return stock, err
	}
	stock._120Days, err = parseFloat("120Days")
	if err != nil {
		return stock, err
	}
	stock._180Days, err = parseFloat("180Days")
	if err != nil {
		return stock, err
	}
	stock._52WeeksHigh, err = parseFloat("52WeeksHigh")
	if err != nil {
		return stock, err
	}
	stock._52WeeksLow, err = parseFloat("52WeeksLow")
	if err != nil {
		return stock, err
	}

	return stock, nil
}

func parseCSVFile(filename string) ([]map[string]any, error) {
	file := strings.ReplaceAll(strings.TrimSuffix(filepath.Base(filename), ".csv"), "_", "-")
	result, err := csv.QueryCsv(filename, "SELECT * FROM @file")
	if err != nil {
		return nil, err
	}
	mapData := csv.PrepareCsvResponseWithHeaderJson(result)
	for i, d := range mapData {
		var err error
		for key, val := range d {
			if !slices.Contains([]string{"Symbol", "Transactions"}, key) && fmt.Sprintf("%v", val) != "-" {
				d[key], err = parseFloat(fmt.Sprintf("%v", val))
				if err != nil {
					return nil, errors.Wrap(err, fmt.Sprintf(`%s + ":" + %v`, key, val), "")
				}
			} else if key == "Transactions" {
				d[key], err = parseInt(fmt.Sprintf("%v", val))
				if err != nil {
					return nil, errors.Wrap(err, fmt.Sprintf(`%s + ":" + %v`, key, val), "")
				}
			}
		}
		d["Date"] = file
		mapData[i] = d
	}
	return mapData, nil
}

func loadAllCSVFiles(directory string) ([]map[string]interface{}, error) {
	var allData []map[string]interface{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errList []error
	dataCh := make(chan []map[string]interface{})
	errCh := make(chan error)
	doneCh := make(chan struct{})

	// Walk through the directory
	go func() {
		err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".csv" {
				wg.Add(1)
				go func(path string) {
					defer wg.Done()
					data, err := parseCSVFile(path)
					if err != nil {
						errCh <- err
						return
					}
					dataCh <- data
				}(path)
			}
			return nil
		})

		if err != nil {
			errCh <- err
		}

		// Wait for all goroutines to finish
		wg.Wait()
		close(doneCh)
	}()

	for {
		select {
		case data := <-dataCh:
			mu.Lock()
			allData = append(allData, data...)
			mu.Unlock()
		case err := <-errCh:
			mu.Lock()
			errList = append(errList, err)
			mu.Unlock()
		case <-doneCh:
			if len(errList) > 0 {
				return nil, fmt.Errorf("errors occurred: %v", errList)
			}
			return allData, nil
		}
	}
}
