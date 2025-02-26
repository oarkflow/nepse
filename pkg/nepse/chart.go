package nepse

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/oarkflow/trading/pkg/techan"
)

func CreateInteractiveChart(ts *techan.TimeSeries, chartTitle, outputPath string) error {
	var dates []string
	var klineData []opts.KlineData
	for _, candle := range ts.Candles {
		dates = append(dates, candle.Period.Start.Format("2006-01-02"))
		klineData = append(klineData, opts.KlineData{
			Value: []float64{
				candle.OpenPrice.Float(),
				candle.ClosePrice.Float(),
				candle.MinPrice.Float(),
				candle.MaxPrice.Float(),
			},
		})
	}
	kline := charts.NewKLine()
	kline.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: chartTitle}),
		charts.WithXAxisOpts(opts.XAxis{Type: "category", Data: dates}),
	)
	kline.SetXAxis(dates).AddSeries("Price", klineData)
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return kline.Render(f)
}

// CreateSMCSignalsBarChart generates a bar chart (HTML) for the given SMC signals.
func CreateSMCSignalsBarChart(signals map[string]SMCDetailedSignal, outputPath string) error {
	var signalNames []string
	var barData []opts.BarData
	for name, detail := range signals {
		signalNames = append(signalNames, name)
		barData = append(barData, opts.BarData{
			Value: detail.Signal,
			Name:  detail.Reason,
		})
	}
	sort.Strings(signalNames)
	bar := charts.NewBar()
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "SMC Signals Overview"}))
	bar.SetXAxis(signalNames).AddSeries("SMC Signals", barData)
	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return err
	}
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return bar.Render(f)
}
