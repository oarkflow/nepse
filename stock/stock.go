package stock

import (
	"fmt"
	"time"

	"github.com/markcheno/go-quote"
	"github.com/oarkflow/search"
)

const timeFormat = "2006-01-02"

// GetStockData dawnloads daily stockdata for symbol(GOOGL, FB...etc) during today ~ before dayPeriod.
// dayPeriod must be day(1day, 30days...etc).
// If stock data is not dawnloaded due to bad symbol, output panic.
func GetStockData(symbol string, dayPeriod int, adj bool) (*quote.Quote, error) {
	endDay := time.Now()
	startDay := endDay.AddDate(0, 0, -dayPeriod)
	engine, err := search.GetEngine[map[string]any]("stock")
	if err != nil {
		return nil, err
	}
	result, err := engine.Search(&search.Params{
		Query:      symbol,
		Properties: []string{"Symbol"},
		Condition:  fmt.Sprintf("Date BETWEEN '%s' AND %s", startDay.Format(timeFormat), endDay.Format(timeFormat)),
	})
	if err != nil {
		return nil, err
	}
	return GetQuote[map[string]any](symbol, result), nil
}

func GetQuote[T any](symbol string, result search.Result[T]) *quote.Quote {
	numrows := result.FilteredTotal
	qt := quote.NewQuote(symbol, numrows)
	for i, row := range result.Hits {
		switch row := any(row.Data).(type) {
		case map[string]any:
			// Parse row of data
			d, _ := time.Parse("2006-01-02", row["Date"].(string))
			o, _ := row["OpenPrice"].(float64)
			h, _ := row["HighPrice"].(float64)
			l, _ := row["LowPrice"].(float64)
			c, _ := row["ClosePrice"].(float64)
			v, _ := row["Volume"].(float64)

			qt.Date[i] = d
			qt.Open[i] = o
			qt.High[i] = h
			qt.Low[i] = l
			qt.Close[i] = c
			qt.Volume[i] = v
		}
	}
	return &qt
}
