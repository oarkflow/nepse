package main

import (
	"fmt"
	"net/http"

	"github.com/oarkflow/trading/pkg/nepse"
)

func main() {
	/*date, err := nepse.GetDate(&nepse.Range{Start: "2023-01-01", End: "2023-01-27"})
	if err != nil {
		panic(err)
	}*/
	server := nepse.NewTraderHandler()
	// server.Trader.CollectAllCompanyFloorSheets(nepse.GetDateRange(date.StartTime, date.EndTime)...)
	// server.Trader.AggregateCompanyFloorsheets()
	go server.Trader.FetchLiveData()
	handler := http.NewServeMux()
	server.RegisterRoutes(handler)
	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nepse.PanicHandler(handler))
}
