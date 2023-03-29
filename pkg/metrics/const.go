package metrics

type Instrument string

const (
	InstrumentDeleted   Instrument = "deleted"
	InstrumentGauge     Instrument = "gauge"
	InstrumentAdditive  Instrument = "additive"
	InstrumentHistogram Instrument = "histogram"
	InstrumentCounter   Instrument = "counter"
)
