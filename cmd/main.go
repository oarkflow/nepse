package main

import (
	"fmt"
	"net/http"
	
	"github.com/oarkflow/trading/pkg/nepse"
)

func main() {
	server := nepse.NewTraderHandler()
	// server.Trader.CollectAllCompanyFloorSheets("2025-02-25")
	// server.Trader.AggregateCompanyFloorsheets()
	go server.Trader.FetchLiveData()
	handler := http.NewServeMux()
	server.RegisterRoutes(handler)
	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nepse.PanicHandler(handler))
}
