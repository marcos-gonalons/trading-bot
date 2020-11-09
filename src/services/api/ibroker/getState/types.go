package getstate

// APIResponse ...
type APIResponse struct {
	Status   string        `json:"s"`
	ErrorMsg string        `json:"errmsg"`
	Data     ResponseState `json:"d"`
}

// ResponseState ...
type ResponseState struct {
	Balance      float64     `json:"balance"`
	UnrealizedPL float64     `json:"unrealizedPl"`
	Equity       float64     `json:"equity"`
	AmData       interface{} `json:"amData"`
}
