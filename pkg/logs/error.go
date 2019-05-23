package logs

// ScanFetcherFailure is logged when there is a failure with the scan fetcher dependency.
type ScanFetcherFailure struct {
	Message string `logevent:"message,default=scan-fetcher-error"`
	Reason  string `logevent:"reason"`
}

// ProducerFailure is logged when the producer fails to put a scan on the queue.
type ProducerFailure struct {
	Message string `logevent:"message,default=producer-failure"`
	Reason  string `logevent:"reason"`
}

// StorageFailure is logged when there is a failure with the storage layer.
type StorageFailure struct {
	Message string `logevent:"message,default=storage-failure"`
	Reason  string `logevent:"reason"`
}
