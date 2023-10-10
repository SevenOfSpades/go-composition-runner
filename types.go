package runner

import (
	"context"
	"errors"
)

var (
	ErrAlreadyStarted         = errors.New("runner has been already started")
	ErrRunnableAlreadyDefined = errors.New("runnable already defined")
	ErrShutdownAlreadyDefined = errors.New("shutdown for runnable already defined")
	ErrRunnableNotFound       = errors.New("runnable not found")
	ErrNoRunnables            = errors.New("nothing to run")
	ErrShutdownTimeout        = errors.New("shutdown deadline reached")
)

type Runner interface {
	// RegisterRunnableHandler adds handler to pool.
	// If existing RunnableName will be provided then function will fail with ErrRunnableAlreadyDefined.
	RegisterRunnableHandler(RunnableName, RunnableFunc) error
	// RegisterRunnableShutdown adds separate function for shutdown procedure to handler.
	// Runnable must be registered with RegisterRunnableHandler before this.
	// If there isn't any runnable with provided name then function will fail with ErrRunnableNotFound.
	// If there is already defined shutdown procedure for runnable then function will fail with ErrShutdownAlreadyDefined.
	RegisterRunnableShutdown(RunnableName, ShutdownFunc) error
	// Start initiates all handlers in pool.
	// It will wait for all registered handlers to be set up but does not guarantee their
	// start upon returning value.
	// Returned channel will receive `nil` if all handlers gracefully shut down or an error
	// if one of handlers failed during execution (it will report first problematic handler).
	// Provided context can be used to send shutdown signal by cancelling it.
	Start(context.Context) <-chan error
}
