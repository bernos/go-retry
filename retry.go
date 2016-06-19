package retry

import (
	"fmt"
	"time"
)

const (
	DefaultMaxRetries = 10
	DefaultBaseDelay  = time.Millisecond
	DefaultMaxDelay   = time.Minute
	Infinity          = -1
)

// Retrier holds options for retrying
type Retrier struct {
	MaxRetries     int
	BaseDelay      time.Duration
	MaxDelay       time.Duration
	ShouldRetry    func(error) bool
	CalculateDelay BackoffFunc
	Log            func(string, ...interface{})
}

func Retry(fn func() (interface{}, error), options ...func(*Retrier)) func() (interface{}, error) {
	r := Retrier{
		BaseDelay:      DefaultBaseDelay,
		MaxDelay:       DefaultMaxDelay,
		MaxRetries:     DefaultMaxRetries,
		ShouldRetry:    func(err error) bool { return true },
		CalculateDelay: calculateDelayBinary,
		Log:            func(format string, v ...interface{}) {},
	}

	for _, o := range options {
		o(&r)
	}

	return func() (interface{}, error) {
		var count int

		for {
			if count > 0 {
				r.Log("Retrying attempt %d", count)
			}

			value, err := fn()

			if err == nil {
				return value, err
			}

			if !r.ShouldRetry(err) {
				return nil, fmt.Errorf("Retrier aborted due to user supplied ShouldRetry func. Cause: %s", err.Error())
			}

			if count == r.MaxRetries {
				return nil, fmt.Errorf("Retrier exceeded max retry count of %d. Cause: %s", r.MaxRetries, err.Error())
			}

			d := r.CalculateDelay(uint(count), r.BaseDelay, r.MaxDelay)

			r.Log("Will retry in %s", d)

			time.Sleep(d)
			count++
		}
	}
}
