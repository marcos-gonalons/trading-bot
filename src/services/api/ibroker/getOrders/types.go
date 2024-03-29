package getorders

// APIResponse ...
type APIResponse struct {
	Status   string `json:"s"`
	ErrorMsg string `json:"errmsg"`
	Data     []struct {
		ID         string  `json:"id"`
		Instrument string  `json:"instrument"`
		Qty        float64 `json:"qty"`
		Side       string  `json:"side"`
		Type       string  `json:"type"`
		FilledQty  float64 `json:"filledQty"`
		AvgPrice   float64 `json:"avgPrice"`
		LimitPrice float64 `json:"limitPrice"`
		StopPrice  float64 `json:"stopPrice"`
		Duration   struct {
			Type     string  `json:"type"`
			Datetime float64 `json:"datetime"`
		} `json:"duration"`
		Status     string  `json:"status"`
		ParentID   *string `json:"parentId"`
		ParentType *string `json:"parentType"`
	} `json:"d"`
}
