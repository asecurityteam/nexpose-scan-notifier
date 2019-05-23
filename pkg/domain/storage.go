package domain

import (
	"context"
	"fmt"
	"time"
)

// TimestampStorer provides a method to persist a timestamp to storage.
type TimestampStorer interface {
	StoreTimestamp(context.Context, time.Time) error
}

// TimestampFetcher provides methods to retrieve a timestamp from storage.
type TimestampFetcher interface {
	FetchTimestamp(context.Context) (time.Time, error)
}

// TimestampNotFound is used to indicate that no timestamp value exists in storage.
type TimestampNotFound struct{}

func (e TimestampNotFound) Error() string {
	return fmt.Sprintf("no timestamp found in storage")
}
