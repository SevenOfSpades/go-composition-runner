# go-composition-runner
Composition runner is designed to control multiple subroutines with continuous processes as single entity.
The basic rule is "if all works - everything works; if one fails/stops - everything stops".

## Usage
```go
package main

import (
	"context"
	"log"
	"time"

	runner "github.com/SevenOfSpades/go-composition-runner"
)

func main() {
	// Create new runner instance (with custom shutdown timeout).
	r, err := runner.New(runner.OptionShutdownTimeout(time.Second * 5))
	if err != nil {
		log.Fatalln(err)
	}

	// Register first handler with ticker.
	// We can skip checking for errors if there is no automation for registering handlers.
	_ = r.RegisterRunnableHandler("ticker", func(ctx context.Context) error {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// ...do something in here...
			case <-ctx.Done():
				// When context is cancelled runner expects all registered handlers to close
				// within defined shutdown time range.
				return nil
			}
		}
	})

	httpServer := NewHTTPServer()

	// Register second handler with HTTP server which will be closed using ShutdownFunc.
	_ = r.RegisterRunnableHandler("http-server", func(_ context.Context) error {
		// "Start" will keep process active until "Close" is called.
		if err := httpServer.Start(":80"); err != nil {
			return err
		}
		return nil
	})
	// Because HTTP server uses function to shut down itself it needs to be registered
	// as shutdown handler.
	// Context in this handler is set with "shutdown timeout" deadline.
	_ = r.RegisterRunnableShutdown("http-server", func(_ context.Context) {
		// "Close" is called separately.
		if err := httpServer.Close(); err != nil {
			log.Fatalln(err)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// "Start" will keep main function active until all registered handlers
	// are operational or provided context is canceled.
	// Channel returned by function will provide cause of failure (an error) or
	// `nil` in case of graceful shutdown.
	if err := <-r.Start(ctx); err != nil {
		log.Fatalln(err)
	}
}
```