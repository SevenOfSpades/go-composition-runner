package runner

import "context"

type (
	// RunnableFunc expects the content of handler to run indefinitely unless the provided context is cancelled.
	RunnableFunc func(context.Context) error

	// ShutdownFunc is designed to stop runnable that uses function to close its operation loop
	// rather than relying on context to be cancelled.
	ShutdownFunc func(context.Context)
)
