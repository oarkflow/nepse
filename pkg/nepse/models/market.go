package models

type MarketOpen struct {
	AsOf   string `json:"asOf"`
	ID     int32  `json:"id"`
	IsOpen string `json:"isOpen"`
}

type Holiday struct {
	Date  string `json:"date"`
	Event string `json:"event"`
}
