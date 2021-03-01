package retryFacade

import "time"

// RetryParams ...
type RetryParams struct {
	DelayBetweenRetries time.Duration
	MaxRetries          uint
	SuccessCallback     func()
	ErrorCallback       func(err error)
}
