package getquote

// APIResponse ...
type APIResponse struct {
	Status   string `json:"s"`
	ErrorMsg string `json:"errmsg"`
	Data     []struct {
		Status string     `json:"s"`
		Name   string     `json:"n"`
		Value  QuoteValue `json:"v"`
	} `json:"d"`
}

// QuoteValue ...
type QuoteValue struct {
	Ch                 float64 `json:"ch"`
	Chp                float64 `json:"chp"`
	CurrentPrice       float64 `json:"lp"`
	Ask                float64 `json:"ask"`
	Bid                float64 `json:"bid"`
	OpenPrice          float64 `json:"open_price"`
	HighPrice          float64 `json:"high_price"`
	LowPrice           float64 `json:"low_price"`
	PreviousClosePrice float64 `json:"prev_close_price"`
	Volume             float64 `json:"volume"`
}
