package data

import (
	"context"
	"time"
)

// Duration to use for SQL operation timeouts.
const QueryTimeout = 3 * time.Second

// createTimeoutContext accepts a time duration and returns a context and cancel
// function with a timeout of that duration.
//
// The caller should defer calling the cancel() function.
func CreateTimeoutContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return ctx, cancel
}
