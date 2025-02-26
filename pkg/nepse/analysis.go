package nepse

import (
	"sort"

	"github.com/oarkflow/trading/pkg/nepse/models"
)

func (n *Trader) AnalyzeTotalBuySellByBroker(floorSheet models.FloorSheet, companies map[string]models.Company) map[string]map[string]map[string]map[string]int {
	brokerTrade := make(map[string]map[string]map[string]map[string]int)
	for _, txn := range floorSheet.Content {
		symbol := txn.StockSymbol
		company, exists := companies[symbol]
		if !exists {
			continue
		}
		sector := company.SectorName
		buyer := txn.BuyerMemberId
		seller := txn.SellerMemberId
		if _, ok := brokerTrade[sector]; !ok {
			brokerTrade[sector] = make(map[string]map[string]map[string]int)
		}
		if _, ok := brokerTrade[sector][symbol]; !ok {
			brokerTrade[sector][symbol] = make(map[string]map[string]int)
		}
		if _, ok := brokerTrade[sector][symbol][buyer]; !ok {
			brokerTrade[sector][symbol][buyer] = make(map[string]int)
		}
		if _, ok := brokerTrade[sector][symbol][seller]; !ok {
			brokerTrade[sector][symbol][seller] = make(map[string]int)
		}
		brokerTrade[sector][symbol][buyer]["buy"] += txn.ContractQuantity
		brokerTrade[sector][symbol][seller]["sell"] += txn.ContractQuantity
	}
	return brokerTrade
}

func (n *Trader) AnalyzeSectorWise(floorSheet models.FloorSheet, companies map[string]models.Company) map[string]map[string]map[string]int {
	sectorWise := make(map[string]map[string]map[string]int)
	for _, txn := range floorSheet.Content {
		symbol := txn.StockSymbol
		company, exists := companies[symbol]
		if !exists {
			continue
		}
		sector := company.SectorName
		if _, ok := sectorWise[sector]; !ok {
			sectorWise[sector] = make(map[string]map[string]int)
		}
		if _, ok := sectorWise[sector][symbol]; !ok {
			sectorWise[sector][symbol] = make(map[string]int)
		}
		sectorWise[sector][symbol][txn.BuyerMemberId] += txn.ContractQuantity
		sectorWise[sector][symbol][txn.SellerMemberId] -= txn.ContractQuantity
	}
	return sectorWise
}

func (n *Trader) AnalyzeBrokerWise(floorSheet models.FloorSheet) map[string]map[string]int {
	brokerWise := make(map[string]map[string]int)
	for _, txn := range floorSheet.Content {
		if _, ok := brokerWise[txn.BuyerMemberId]; !ok {
			brokerWise[txn.BuyerMemberId] = make(map[string]int)
		}
		if _, ok := brokerWise[txn.SellerMemberId]; !ok {
			brokerWise[txn.SellerMemberId] = make(map[string]int)
		}
		brokerWise[txn.BuyerMemberId][txn.StockSymbol] += txn.ContractQuantity
		brokerWise[txn.SellerMemberId][txn.StockSymbol] -= txn.ContractQuantity
	}
	return brokerWise
}

func (n *Trader) AnalyzeBrokerTransactions(floorSheet models.FloorSheet) map[string]int {
	brokerTotals := make(map[string]int)
	for _, txn := range floorSheet.Content {
		brokerTotals[txn.BuyerMemberId] += txn.ContractQuantity
		brokerTotals[txn.SellerMemberId] -= txn.ContractQuantity
	}
	return brokerTotals
}

func (n *Trader) AnalyzeTopBrokers(floorSheet models.FloorSheet) (brokers []models.BrokerVolume) {
	brokerVolume := make(map[string]int)
	for _, txn := range floorSheet.Content {
		brokerVolume[txn.BuyerMemberId] += txn.ContractQuantity
		brokerVolume[txn.SellerMemberId] += txn.ContractQuantity
	}
	for name, volume := range brokerVolume {
		brokers = append(brokers, models.BrokerVolume{Symbol: name, Volume: volume})
	}
	sort.Slice(brokers, func(i, j int) bool {
		return brokers[i].Volume > brokers[j].Volume
	})
	return brokers
}

func (n *Trader) AnalyzeTradedStocks(floorSheet models.FloorSheet) (stocks []models.BrokerVolume) {
	stockVolume := make(map[string]int)
	for _, txn := range floorSheet.Content {
		stockVolume[txn.StockSymbol] += txn.ContractQuantity
	}
	for symbol, volume := range stockVolume {
		stocks = append(stocks, models.BrokerVolume{Symbol: symbol, Volume: volume})
	}
	return stocks
}

func (n *Trader) AnalyzeStockVolatility(floorSheet models.FloorSheet) map[string][]float64 {
	volatility := make(map[string][]float64)
	for _, txn := range floorSheet.Content {
		volatility[txn.StockSymbol] = append(volatility[txn.StockSymbol], txn.ContractRate)
	}
	return volatility
}

func (n *Trader) AnalyzeBuyerSellerMatching(floorSheet models.FloorSheet) map[string]map[string]int {
	matches := make(map[string]map[string]int)
	for _, txn := range floorSheet.Content {
		if _, ok := matches[txn.BuyerMemberId]; !ok {
			matches[txn.BuyerMemberId] = make(map[string]int)
		}
		matches[txn.BuyerMemberId][txn.SellerMemberId] += txn.ContractQuantity
	}
	return matches
}

func (n *Trader) AnalyzeMarketBuySellRatio(floorSheet models.FloorSheet) models.MarketRatio {
	buy, sell := 0, 0
	for _, txn := range floorSheet.Content {
		buy += txn.ContractQuantity
		sell -= txn.ContractQuantity
	}
	return models.MarketRatio{Buy: buy, Sell: sell}
}
