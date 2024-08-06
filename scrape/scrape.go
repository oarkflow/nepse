package scrape

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/oarkflow/anonymizer"
	"github.com/oarkflow/search"
)

var headerMapping = map[string]string{
	"Symbol":        "Symbol",
	"Conf.":         "Confidence",
	"Open":          "OpenPrice",
	"High":          "HighPrice",
	"Low":           "LowPrice",
	"Close":         "ClosePrice",
	"VWAP":          "VWAP",
	"Vol":           "Volume",
	"Prev. Close":   "PreviousClose",
	"Turnover":      "Turnover",
	"Trans.":        "Transactions",
	"Diff":          "Difference",
	"Range":         "Range",
	"Diff %":        "DifferencePercentage",
	"Range %":       "RangePercentage",
	"VWAP %":        "VWAPPercentage",
	"120 Days":      "120Days",
	"180 Days":      "180Days",
	"52 Weeks High": "52WeeksHigh",
	"52 Weeks Low":  "52WeeksLow",
}

func renameDir(dir, fromPattern, toPattern string) error {
	dirInfos, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, dirInfo := range dirInfos {
		output, err := anonymizer.Transform(fromPattern, toPattern, dirInfo.Name())
		if err != nil {
			return err
		}
		err = os.Rename(filepath.Join(dir, dirInfo.Name()), filepath.Join(dir, output))
		if err != nil {
			return err
		}
	}
	return nil
}

func renameCSVFileHeaders(dir string, newHeaders map[string]string) error {
	dirInfos, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, dirInfo := range dirInfos {
		if filepath.Ext(dirInfo.Name()) == ".csv" {
			path := filepath.Join(dir, dirInfo.Name())
			err := RenameHeaders(path, path, newHeaders)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// RenameHeaders renames the headers in the CSV file based on the provided mapping
func RenameHeaders(inputFile, outputFile string, headMap ...map[string]string) error {
	var mapping map[string]string
	if len(headMap) > 0 {
		mapping = headMap[0]
	} else {
		mapping = headerMapping
	}
	// Open the input CSV file
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Read the CSV data
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV data: %v", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("no data in CSV file")
	}

	// Get the header row
	oldHeaders := records[0]
	newHeaders := make([]string, len(oldHeaders))

	// Create a map for quick lookup of old header names
	oldHeaderMap := make(map[string]int)
	for i, header := range oldHeaders {
		oldHeaderMap[header] = i
	}

	// Map old headers to new headers
	for oldHeader, newHeader := range mapping {
		if index, found := oldHeaderMap[oldHeader]; found {
			newHeaders[index] = newHeader
		} else {
			return fmt.Errorf("old header '%s' not found in CSV file", oldHeader)
		}
	}

	// Fill in any headers that are not in the mapping with their original names
	for i, header := range oldHeaders {
		if newHeaders[i] == "" {
			newHeaders[i] = header
		}
	}

	// Replace the header row with the new headers
	records[0] = newHeaders

	// Open the output CSV file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	// Write the modified CSV data
	writer := csv.NewWriter(outFile)
	err = writer.WriteAll(records)
	if err != nil {
		return fmt.Errorf("failed to write CSV data: %v", err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("error during flush: %v", err)
	}

	return nil
}

func Scrape() error {
	engine, err := search.GetEngine[map[string]any]("stock")
	if err != nil {
		return err
	}
	now := time.Now()
	result, err := engine.Search(&search.Params{Query: now.Format(time.DateOnly), Properties: []string{"Date"}})
	if err != nil {
		return err
	}
	dateStr := now.Format("01/02/2006")
	if result.Count == 0 {
		err = parseDate(now)
		if err != nil {
			return err
		}
	}
	return RenameHeaders(dateStr, dateStr)
}

/*err := renameDir("./data/date2", "<year>_<month>_<date>.csv", "<year>-<month>-<date>.csv")
if err != nil {
	panic(err)
}*/
// Define the old and new header mappings
/*err := renameCSVFileHeaders("./data", headerMapping)
if err != nil {
	panic(err)
}*/

func parseDate(date time.Time) error {
	dateStr := date.Format("01/02/2006")
	url := "https://www.sharesansar.com/today-share-price"
	df := scrapeData(url, dateStr)
	finalDf := cleanDf(df)
	path := fmt.Sprintf("data/date/%s.csv", date.Format(time.DateOnly))
	saveCSV(finalDf, path)
	return RenameHeaders(path, path)
}

func scrapeData(url string, date string) [][]string {
	c := colly.NewCollector(
		colly.AllowedDomains("www.sharesansar.com"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.115 Safari/537.36"),
	)

	var df [][]string
	var headers []string

	c.OnHTML("table.table-bordered", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			var row []string
			el.ForEach("th, td", func(_ int, cell *colly.HTMLElement) {
				text := strings.TrimSpace(cell.Text)
				row = append(row, text)
			})
			if len(headers) == 0 {
				headers = row
			} else {
				df = append(df, row)
			}
		})
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
		r.Ctx.Put("date", date)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Visited", r.Request.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Error:", r.StatusCode, err)
	})

	// Visit the target URL
	c.Visit(url)

	return append([][]string{headers}, df...)
}

func cleanDf(df [][]string) [][]string {
	unique := make(map[string]bool)
	var newDf [][]string
	for _, row := range df {
		key := strings.Join(row, ",")
		if _, exists := unique[key]; !exists {
			unique[key] = true
			newDf = append(newDf, row)
		}
	}
	newDf = newDf[1:]
	header := df[0]
	for i := range newDf {
		newDf[i] = append([]string{}, newDf[i]...)
	}
	newDf = append([][]string{header}, newDf...)
	return newDf
}

func saveCSV(data [][]string, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Could not create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range data {
		err := writer.Write(row)
		if err != nil {
			log.Fatalf("Could not write to CSV file: %v", err)
		}
	}
}
