package login

// APIResponse ...
type APIResponse struct {
	Status   string `json:"s"`
	ErrorMsg string `json:"errmsg"`
	Data     struct {
		AccessToken string  `json:"access_token"`
		Expiration  float64 `json:"expiration"`
	} `json:"d"`
}
