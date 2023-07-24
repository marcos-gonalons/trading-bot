package types

// Candle ...
type Candle struct {
	Open       float64
	High       float64
	Low        float64
	Close      float64
	Volume     float64
	Timestamp  int64
	Indicators Indicators
}

type MovingAverage struct {
	Description   string
	Name          string
	Value         float64
	CandlesAmount int64
}

type Indicators struct {
	MovingAverages []MovingAverage
}
