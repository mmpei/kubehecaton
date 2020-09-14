package utils

import (
	"fmt"
	"time"
)

// RetryError ...
type RetryError struct {
	n int
}

// Error will return the msg of this error
func (e *RetryError) Error() string {
	return fmt.Sprintf("still failing after %d retries", e.n)
}

// IsRetryFailure will judge the err
func IsRetryFailure(err error) bool {
	_, ok := err.(*RetryError)
	return ok
}

// ConditionFunc ...
type ConditionFunc func() (bool, error)

// Retry retries f every interval until after maxRetries.
// The interval won't be affected by how long f takes.
// For example, if interval is 3s, f takes 1s, another f will be called 2s later.
// However, if f takes longer than interval, it will be delayed.
func Retry(interval time.Duration, maxRetries int, f ConditionFunc) error {
	if maxRetries <= 0 {
		return fmt.Errorf("maxRetries (%d) should be > 0", maxRetries)
	}
	tick := time.NewTicker(interval)
	defer tick.Stop()

	for i := 0; ; i++ {
		ok, err := f()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		if i >= maxRetries {
			break
		}
		<-tick.C
	}
	return &RetryError{maxRetries}
}
