package nepse

import (
	"fmt"
	"github.com/oarkflow/errors"
	"github.com/oarkflow/log"
	"github.com/oarkflow/search"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/oarkflow/nepse/nepse/csv"
)

func InitCSVStock() {
	engine, err := search.SetEngine[map[string]any]("stock", &search.Config{})
	if err != nil {
		panic(err)
	}
	log.Info().Msg("Indexing stock")
	_, err = LoadAllCsvFiles("./data/date", func(data []map[string]any) {
		engine.InsertWithPool(data, runtime.NumCPU(), 1000)
	})
	if err != nil {
		panic(err)
	}
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

func ParseCSVFile(filename string, callback func([]map[string]any)) ([]map[string]any, error) {
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
	if callback != nil {
		callback(mapData)
	}
	return mapData, nil
}

type FileInfo struct {
	Path string
	Name string
}

func LoadAllCsvFiles(directory string, callback func([]map[string]any)) ([]map[string]interface{}, error) {
	var allData []map[string]interface{}
	var files []FileInfo
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".csv") {
			files = append(files, FileInfo{Path: path, Name: info.Name()})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name > files[j].Name
	})
	for _, path := range files {
		data, err := ParseCSVFile(path.Path, callback)
		if err != nil {
			return nil, err
		}
		if callback == nil {
			allData = append(allData, data...)
		}
		log.Info().Msgf("File %s parsed", path.Path)
	}
	return allData, nil
}

func LoadAllCsvFilesToMap(directory string) (map[string][]map[string]any, error) {
	allData := make(map[string][]map[string]any)
	var files []FileInfo
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".csv") {
			files = append(files, FileInfo{Path: path, Name: info.Name()})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name > files[j].Name
	})
	for _, path := range files {
		date := strings.ReplaceAll(strings.TrimSuffix(filepath.Base(path.Path), ".csv"), "_", "-")
		data, err := ParseCSVFile(path.Path, nil)
		if err != nil {
			return nil, err
		}
		allData[date] = data
		log.Info().Msgf("File %s parsed", path.Path)
	}
	return allData, nil
}
