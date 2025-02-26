package models

import (
	"fmt"
	"time"
)

type Transaction struct {
	BusinessDate     string     `json:"businessDate"`
	BuyerBrokerName  string     `json:"buyerBrokerName"`
	BuyerMemberId    string     `json:"buyerMemberId"`
	ContractAmount   float64    `json:"contractAmount"`
	ContractId       int        `json:"contractId"`
	ContractQuantity int        `json:"contractQuantity"`
	ContractRate     float64    `json:"contractRate"`
	SecurityName     string     `json:"securityName"`
	SellerBrokerName string     `json:"sellerBrokerName"`
	SellerMemberId   string     `json:"sellerMemberId"`
	StockId          int        `json:"stockId"`
	StockSymbol      string     `json:"stockSymbol"`
	TradeBookId      int        `json:"tradeBookId"`
	TradeTime        CustomTime `json:"tradeTime"`
}

type CustomTime struct {
	time.Time
}

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	if len(s) < 2 {
		return fmt.Errorf("invalid time string")
	}
	s = s[1 : len(s)-1]
	parsedTime, err := time.Parse("2006-01-02T15:04:05.999999", s)
	if err != nil {
		return err
	}
	ct.Time = parsedTime
	return nil
}

type FloorSheet struct {
	Date        string        `json:"date"`
	Content     []Transaction `json:"content"`
	TotalAmount float64       `json:"totalAmount"`
	TotalQty    int           `json:"totalQty"`
	TotalTrades int           `json:"totalTrades"`
}
