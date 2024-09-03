package techan

// A strategy is a holder struct which bundles the raw timeseries data, indicators, and rules used to make
// trading decisions for a given security. The strategy can be used to calculate allocations and map algo
// trading triggers over time.
type Strategy struct {
	Security   string
	Timeseries TimeSeries
	Indicators map[string]Indicator
	Rule       Rule
}

// Helper function to get the last index of the strategy's data.
func (s *Strategy) LastIndex() int {
	return s.Timeseries.LastIndex()
}
