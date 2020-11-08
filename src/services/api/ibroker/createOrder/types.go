package createorder

// APIResponse ...
type APIResponse struct {
	Status   string `json:"s"`
	ErrorMsg string `json:"errmsg"`
	Data     struct {
		OrderID int64 `json:"orderId"`
	} `json:"d"`
}
