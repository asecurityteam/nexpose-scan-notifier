package main

import (
	"context"
	"os"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/domain"
	v1 "github.com/asecurityteam/nexpose-scan-notifier/pkg/handlers/v1"
	"github.com/asecurityteam/serverfull"
	"github.com/asecurityteam/settings"
)

func main() {
	ctx := context.Background()
	source, err := settings.NewEnvSource(os.Environ())
	if err != nil {
		panic(err.Error())
	}

	notificationHandler := &v1.NotificationHandler{
		// TODO: implement domain.ScanFetcher interface
		// TODO: implement domain.TimestampFetcher interface
		// TODO: implement domain.TimestampStorer interface
		// TODO: implement domain.Producer interface
		LogFn: domain.LoggerFromContext,
	}
	handlers := map[string]serverfull.Function{
		"notification": serverfull.NewFunction(notificationHandler),
	}

	fetcher := &serverfull.StaticFetcher{Functions: handlers}
	if err := serverfull.Start(ctx, source, fetcher); err != nil {
		panic(err.Error())
	}
}
