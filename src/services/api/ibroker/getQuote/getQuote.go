package getquote

// Response ...
type Response struct {
	Status   string `json:"s"`
	ErrorMsg string `json:"errmsg"`
	Data     []struct {
		Status string `json:"s"`
		Name   string `json:"n"`
		Value  struct {
			Ch                 float32 `json:"ch"`
			Chp                float64 `json:"chp"`
			CurrentPrice       float32 `json:"lp"`
			Ask                float32 `json:"ask"`
			Bid                float32 `json:"bid"`
			OpenPrice          float32 `json:"open_price"`
			HighPrice          float32 `json:"high_price"`
			LowPrice           float32 `json:"low_price"`
			PreviousClosePrice float32 `json:"prev_close_price"`
			Volume             float32 `json:"volume"`
		} `json:"v"`
	} `json:"d"`
}
