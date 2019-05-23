package domain

import "context"

// The Producer interface is used to produce completed scans onto a queue.
type Producer interface {
	Produce(ctx context.Context, scan CompletedScan) error
}
