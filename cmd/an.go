package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	"github.com/oarkflow/trading/pkg/nepse"
	"github.com/oarkflow/trading/pkg/nepse/models"
)

func main() {
	analysis, err := nepse.GetAnalysis()
	if err != nil {
		log.Fatalf("Error during analysis: %v", err)
	}
	// writeJSON("data-analysis.json", analysis)
	writeHTML("data-analysis.html", analysis)
}

func writeJSON(file string, analysis models.AnalysisResult) {
	// For demonstration, marshal the analysis result as JSON and print.
	data, err := json.MarshalIndent(analysis, "", "\t")
	if err != nil {
		log.Fatalf("Error marshaling analysis result: %v", err)
	}
	os.WriteFile(file, data, 0644)
}

// signalClass returns a CSS class based on the signal value.
func signalClass(signal string) string {
	switch strings.ToUpper(signal) {
	case "BUY":
		return "buy"
	case "SELL":
		return "sell"
	default:
		return "hold"
	}
}

// formatFloat rounds a float64 to 3 decimals.
func formatFloat(f float64) string {
	return fmt.Sprintf("%.3f", f)
}

func writeHTML(file string, data models.AnalysisResult) {
	// Template function map.
	funcMap := template.FuncMap{
		"signalClass": signalClass,
		"formatFloat": formatFloat,
	}

	// Create and parse the template.
	tmpl, err := template.New("analysis").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	// Create (or overwrite) the output HTML file.
	outFile, err := os.Create(file)
	if err != nil {
		log.Fatalf("Error creating HTML file: %v", err)
	}
	defer outFile.Close()

	// Execute the template with our analysis result.
	if err := tmpl.Execute(outFile, data); err != nil {
		log.Fatalf("Error executing template: %v", err)
	}

	log.Println(file + " generated successfully.")

}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>NEPSE Analysis Report</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <!-- Bootstrap CSS from CDN -->
  <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css">
  <style>
    body {
      padding-top: 10px;
      font-size: 0.95em;
    }
    p {
      margin-bottom: 5px;
    }
    .buy {
      color: green;
      font-weight: bold;
    }
    .sell {
      color: red;
      font-weight: bold;
    }
    .hold {
      color: #555;
      font-weight: bold;
    }
    .signal-reason {
      font-size: 0.85em;
      color: #666;
    }
    .card {
      margin-bottom: 10px;
    }
    .broker-table th,
    .broker-table td {
      font-size: 0.8em;
      padding: 2px 5px;
    }
    .filter-container {
      position: sticky;
      top: 0;
      z-index: 1000;
      background-color: #fff;
      padding: 10px 0;
      margin-bottom: 10px;
      border-bottom: 1px solid #ddd;
    }
    .stock-card {
      margin-bottom: 10px;
    }
    .card-header h4 {
      font-size: 1.1em;
      margin: 0;
    }
    .badge-action {
      font-size: 0.8em;
      margin-left: 10px;
    }
  </style>
</head>
<body>
<div class="container">
  <h1 class="mb-3">NEPSE Stock Analysis Report</h1>
  <p><strong>Date Range:</strong> {{.StartDate}} to {{.EndDate}}</p>
  
  <!-- Tabs for Analysis Types -->
  <ul class="nav nav-tabs" id="analysisTabs" role="tablist">
    <li class="nav-item">
      <a class="nav-link" id="daily-tab" data-toggle="tab" href="#daily" role="tab" aria-controls="daily" aria-selected="false">Daily Analysis</a>
    </li>
  </ul>
  
  <!-- Tab Content -->
  <div class="tab-content mt-3" id="analysisTabsContent">
    <!-- Daily Analysis Tab -->
    <div class="tab-pane fade" id="daily" role="tabpanel" aria-labelledby="daily-tab">
      {{range .Daily}}
        {{template "stockAnalysis" .}}
      {{else}}
        <p>No Daily analysis data available.</p>
      {{end}}
    </div>
  </div>
  
  <!-- Hidden block to store the raw JSON analysis data in case needed for client-side caching -->
  <script type="application/json" id="analysisData">
    {{- /* You can output the JSON analysis result here if desired.
            For example, using a custom template function to output JSON:
            {{ toJSON . }}
        */ -}}
  </script>
  
</div>

<!-- Stock Analysis Card Template -->
{{define "stockAnalysis"}}
<div class="card stock-card" data-symbol="{{.Symbol}}" data-action="{{.Action}}">
  <div class="card-header">
    <h4>{{.Symbol}} <span class="badge badge-info badge-action">{{.Action}}</span></h4>
  </div>
  <div class="card-body">
    <div class="row">
      <!-- Signals Column -->
      <div class="col-md-6">
        <p>
          <strong>Short Term:</strong>
          <span class="{{signalClass .ShortTermSignal}}">{{.ShortTermSignal}}</span>
          <br><small class="signal-reason">{{.ShortTermReason}}</small>
        </p>
        <p>
          <strong>MACD:</strong>
          <span class="{{signalClass .MACDSignal}}">{{.MACDSignal}}</span>
          <br><small class="signal-reason">{{.MACDReason}}</small>
        </p>
        <p>
          <strong>ATR:</strong>
          <span class="{{signalClass .ATRSignal}}">{{.ATRSignal}}</span>
          <br><small class="signal-reason">{{.ATRReason}}</small>
        </p>
        <p>
          <strong>EMA Crossover:</strong>
          <span class="{{signalClass .EMACrossoverSignal}}">{{.EMACrossoverSignal}}</span>
          <br><small class="signal-reason">{{.EMACrossoverReason}}</small>
        </p>
        <p>
          <strong>Bollinger:</strong>
          <span class="{{signalClass .BollingerSignal}}">{{.BollingerSignal}}</span>
          <br><small class="signal-reason">{{.BollingerReason}}</small>
        </p>
        <p>
          <strong>SMC:</strong>
          <span class="{{signalClass .SMCSignal}}">{{.SMCSignal}}</span>
          <br><small class="signal-reason">{{.SMCReason}}</small>
        </p>
        <!-- New Signals -->
        <p>
          <strong>OBV:</strong>
          <span class="{{signalClass .OBVSignal}}">{{.OBVSignal}}</span>
          <br><small class="signal-reason">{{.OBVReason}}</small>
        </p>
        <p>
          <strong>A/D:</strong>
          <span class="{{signalClass .ADSignal}}">{{.ADSignal}}</span>
          <br><small class="signal-reason">{{.ADReason}}</small>
        </p>
        <p>
          <strong>CMF:</strong>
          <span class="{{signalClass .CMFSignal}}">{{.CMFSignal}}</span>
          <br><small class="signal-reason">{{.CMFReason}}</small>
        </p>
        <p>
          <strong>Order Imbalance:</strong>
          <span class="{{signalClass .OrderImbalanceSignal}}">{{.OrderImbalanceSignal}}</span>
          <br><small class="signal-reason">{{.OrderImbalanceReason}}</small>
        </p>
        <p>
          <strong>Candlestick:</strong>
          <span class="{{signalClass .CandlestickSignal}}">{{.CandlestickSignal}}</span>
          <br><small class="signal-reason">{{.CandlestickReason}}</small>
        </p>
        <p>
          <strong>Divergence:</strong>
          <span class="{{signalClass .DivergenceSignal}}">{{.DivergenceSignal}}</span>
          <br><small class="signal-reason">{{.DivergenceReason}}</small>
        </p>
        <p>
          <strong>Market Profile:</strong>
          <span class="{{signalClass .MarketProfileSignal}}">{{.MarketProfileSignal}}</span>
          <br><small class="signal-reason">{{.MarketProfileReason}}</small>
        </p>
      </div>
      <!-- Metrics Column -->
      <div class="col-md-6">
        <p><strong>Trades:</strong> {{.Trades}}</p>
        <p><strong>Avg Risk Reward:</strong> {{formatFloat .AvgRiskReward}}</p>
        <p><strong>Max Drawdown:</strong> {{formatFloat .MaxDrawdown}}</p>
        <p><strong>Net P/L:</strong> {{formatFloat .NetPL}}</p>
        <p><strong>Avg P/L:</strong> {{formatFloat .AvgPL}}</p>
        <p><strong>Win Rate:</strong> {{.WinRate}}%</p>
        <p><strong>Boom:</strong> {{.Boom}}</p>
      </div>
    </div>
    <hr>
    <div class="row">
      <!-- Price & Indicator Data -->
      <div class="col-md-6">
        <p><strong>VWAP:</strong> {{formatFloat .VWAP}}</p>
        <p>
          <strong>Bollinger Bands:</strong>
          Upper: {{formatFloat .BollingerUpper}},
          Middle: {{formatFloat .BollingerMiddle}},
          Lower: {{formatFloat .BollingerLower}}
        </p>
        <p>
          <strong>Stochastic:</strong>
          Index: {{formatFloat .Stochastic.Index}},
          Signal: {{.Stochastic.Signal}}
        </p>
      </div>
      <!-- Price Range & Volume -->
      <div class="col-md-6">
        <p><strong>Holding Info:</strong> {{.HoldingInfo}}</p>
        <p><strong>Days Held:</strong> {{.Days}}</p>
        <p>
          <strong>Price Range:</strong>
          High: {{formatFloat .HighPrice}},
          Low: {{formatFloat .LowPrice}},
          Avg: {{formatFloat .AvgPrice}}
        </p>
        <p><strong>Total Volume:</strong> {{formatFloat .TotalVolume}}</p>
      </div>
    </div>
    <hr>
    <div class="row">
      <!-- Broker Statistics -->
      <div class="col-md-12">
        <h6>Top Brokers</h6>
        <table class="table table-sm table-bordered broker-table">
          <thead>
            <tr>
              <th>Broker</th>
              <th>Total Buy</th>
              <th>Total Sell</th>
              <th>Net</th>
            </tr>
          </thead>
          <tbody>
            {{range $broker, $stats := .TopBrokers}}
            <tr>
              <td>{{$broker}}</td>
              <td>{{formatFloat $stats.TotalBuy}}</td>
              <td>{{formatFloat $stats.TotalSell}}</td>
              <td>{{formatFloat $stats.Net}}</td>
            </tr>
            {{end}}
          </tbody>
        </table>
      </div>
    </div>
    <hr>
    <div class="row">
      <!-- Advanced Risk and External Signals -->
      <div class="col-md-6">
        <p><strong>Optimized SMA Period:</strong> {{.OptimizedSMAPeriod}}</p>
        <p>
          <strong>Risk Metrics:</strong>
          <span class="{{signalClass .RiskMetricsSignal}}">{{.RiskMetricsSignal}}</span>
          <br><small class="signal-reason">{{.RiskMetricsReason}}</small>
        </p>
        <p>
          <strong>Sentiment:</strong>
          <span class="{{signalClass .SentimentSignal}}">{{.SentimentSignal}}</span>
          <br><small class="signal-reason">{{.SentimentReason}}</small>
        </p>
      </div>
      <div class="col-md-6">
        <p>
          <strong>Macro Data:</strong>
          <span class="{{signalClass .MacroSignal}}">{{.MacroSignal}}</span>
          <br><small class="signal-reason">{{.MacroReason}}</small>
        </p>
        <p>
          <strong>Prediction:</strong>
          <span class="{{signalClass .PredictionSignal}}">{{.PredictionSignal}}</span>
          <br><small class="signal-reason">{{.PredictionReason}}</small>
        </p>
        <p>
          <strong>Anomaly:</strong>
          <span class="{{signalClass .AnomalySignal}}">{{.AnomalySignal}}</span>
          <br><small class="signal-reason">{{.AnomalyReason}}</small>
        </p>
      </div>
    </div>
  </div>
</div>
{{end}}

<!-- jQuery and Bootstrap JS (from CDN) -->
<script src="https://code.jquery.com/jquery-3.5.1.min.js"></script>
<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/js/bootstrap.min.js"></script>
<script>
// ----- LocalStorage and Intraday Period Control -----

// On page load, check localStorage for the last chosen intraday period.
$(document).ready(function(){
  var savedPeriod = localStorage.getItem("intradayPeriod");
  if(savedPeriod) {
    $('#intradayPeriod').val(savedPeriod);
  } else {
    // Save default (15 minutes) if not set.
    localStorage.setItem("intradayPeriod", $('#intradayPeriod').val());
  }
  
  // When the user changes the intraday period, store the new value.
  $('#intradayPeriod').on('change', function(){
    localStorage.setItem("intradayPeriod", $(this).val());
    // In a real application, you might now re-fetch the intraday data
    // using an API call and then re-render the intraday tab.
    alert("Intraday period set to " + $(this).val() + " minutes. Reload page to see updated data.");
  });
  
  // Activate Live tab if live analysis data exists; otherwise default to Intraday.
  var liveCount = $('#live .stock-card').length;
  $('#daily-tab').addClass('active');
    $('#daily').addClass('active show');
  
  // Filtering function (if needed)
  function applyFilter() {
    var symbolFilter = $('#filterSymbol').val().toLowerCase();
    var actionFilter = $('#filterAction').val().toUpperCase();
    $('.stock-card').each(function() {
      var symbol = $(this).data('symbol').toLowerCase();
      var action = $(this).data('action').toUpperCase();
      var show = true;
      if (symbolFilter && symbol.indexOf(symbolFilter) === -1) {
        show = false;
      }
      if (actionFilter && action !== actionFilter) {
        show = false;
      }
      $(this).toggle(show);
    });
  }
  
  $('#applyFilter').click(applyFilter);
  $('#filterSymbol').keyup(applyFilter);
  $('#filterAction').change(applyFilter);
});
</script>
</body>
</html>
`
