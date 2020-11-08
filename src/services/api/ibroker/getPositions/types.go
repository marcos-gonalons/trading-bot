package getpositions

// APIResponse ...
type APIResponse struct {
	Status   string             `json:"s"`
	ErrorMsg string             `json:"errmsg"`
	Data     []ResponsePosition `json:"d"`
}

// ResponsePosition ...
type ResponsePosition struct {
	ID           string  `json:"id"` // The ID of a position is the symbol's name, for example "GER30"
	Instrument   string  `json:"instrument"`
	Qty          float32 `json:"qty"`
	Side         string  `json:"side"`
	AvgPrice     float32 `json:"avgPrice"`
	UnrealizedPL float64 `json:"unrealizedPl"`
}
