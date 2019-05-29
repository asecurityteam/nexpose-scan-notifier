package main

import (
	"context"
	"net/http"
	"os"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/domain"
	v1 "github.com/asecurityteam/nexpose-scan-notifier/pkg/handlers/v1"
	"github.com/asecurityteam/nexpose-scan-notifier/pkg/producer"
	"github.com/asecurityteam/nexpose-scan-notifier/pkg/scanfetcher"
	"github.com/asecurityteam/serverfull"
	"github.com/asecurityteam/settings"
)

func main() {
	ctx := context.Background()
	source, err := settings.NewEnvSource(os.Environ())
	if err != nil {
		panic(err.Error())
	}

	// configure Nexpose scan fetcher
	nexposeComponent := scanfetcher.NexposeComponent{}
	nexposeClient := new(scanfetcher.NexposeClient)
	if err = settings.NewComponent(context.Background(), source, nexposeComponent, nexposeClient); err != nil {
		panic(err.Error())
	}
	nexposeClient.Client = http.DefaultClient

	// configure HTTP scan event producer
	httpProducerComponent := producer.ProducerComponent{}
	httpProducer := new(producer.HTTP)
	if err = settings.NewComponent(context.Background(), source, httpProducerComponent, httpProducer); err != nil {
		panic(err.Error())
	}
	httpProducer.Client = http.DefaultClient

	notificationHandler := &v1.NotificationHandler{
		// TODO: implement domain.TimestampFetcher interface
		// TODO: implement domain.TimestampStorer interface
		ScanFetcher: nexposeClient,
		Producer:    httpProducer,
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
