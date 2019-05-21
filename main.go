package main

import (
	"context"
	"os"

	"github.com/asecurityteam/serverfull"
	"github.com/asecurityteam/settings"
)

func main() {
	ctx := context.Background()
	source, err := settings.NewEnvSource(os.Environ())
	if err != nil {
		panic(err.Error())
	}

	handlers := map[string]serverfull.Function{
		// TODO: Register lambda functions here in the form of
		// "name_or_arn": lambda.NewHandler(myHandler.Handle)
	}

	fetcher := &serverfull.StaticFetcher{Functions: handlers}
	if err := serverfull.Start(ctx, source, fetcher); err != nil {
		panic(err.Error())
	}
}
