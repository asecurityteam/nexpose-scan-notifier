package main

import (
	"context"
	"net/http"
	"os"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/scanfetcher"

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

	nexposeConfigComponent := scanfetcher.NexposeConfigComponent{}
	nexposeClient := new(scanfetcher.NexposeClient)
	if err = settings.NewComponent(context.Background(), source, nexposeConfigComponent, nexposeClient); err != nil {
		panic(err.Error())
	}
	nexposeClient.Client = http.DefaultClient

	notificationHandler := &v1.NotificationHandler{
		// TODO: implement domain.TimestampFetcher interface
		// TODO: implement domain.TimestampStorer interface
		// TODO: implement domain.Producer interface
		ScanFetcher: nexposeClient,
		LogFn:       domain.LoggerFromContext,
	}
	handlers := map[string]serverfull.Function{
		"notification": serverfull.NewFunction(notificationHandler),
	}

	fetcher := &serverfull.StaticFetcher{Functions: handlers}
	if err := serverfull.Start(ctx, source, fetcher); err != nil {
		panic(err.Error())
	}
}
