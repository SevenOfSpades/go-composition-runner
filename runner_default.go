package runner

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type (
	finishedDTO struct {
		Name string
		Err  error
	}
	defaultRunner struct {
		debugPrinterFunc   DebugMessagePrinterFunc
		runnables          map[RunnableName]RunnableFunc
		runnablesShutdowns map[RunnableName]ShutdownFunc
		done               chan error
		shutdownTimeout    time.Duration
		name               string
		started            bool
	}
)

func newRunner(shutdownTimeout time.Duration, debugPrinterFunc DebugMessagePrinterFunc, name string) Runner {
	return &defaultRunner{
		shutdownTimeout:    shutdownTimeout,
		debugPrinterFunc:   debugPrinterFunc,
		runnables:          make(map[RunnableName]RunnableFunc),
		runnablesShutdowns: make(map[RunnableName]ShutdownFunc),
		done:               make(chan error, 1),
		name:               name,
	}
}

func (r *defaultRunner) RegisterRunnableHandler(name RunnableName, runnableFunc RunnableFunc) error {
	if r.started {
		return fmt.Errorf("failed to register shutdown handler for runnable '%s': %w", name.String(), ErrAlreadyStarted)
	}
	if _, ok := r.runnables[name]; ok {
		return fmt.Errorf("failed to register runnable '%s': %w", name.String(), ErrRunnableAlreadyDefined)
	}
	r.runnables[name] = runnableFunc
	return nil
}

func (r *defaultRunner) RegisterRunnableShutdown(name RunnableName, shutdownFunc ShutdownFunc) error {
	if r.started {
		return fmt.Errorf("failed to register shutdown handler for runnable '%s': %w", name.String(), ErrAlreadyStarted)
	}
	if _, ok := r.runnables[name]; !ok {
		return fmt.Errorf("failed to register shutdown handler for runnable '%s': %w", name.String(), ErrRunnableNotFound)
	}
	if _, ok := r.runnablesShutdowns[name]; ok {
		return fmt.Errorf("failed to register shutdown handler for runnable '%s': %w", name.String(), ErrShutdownAlreadyDefined)
	}
	r.runnablesShutdowns[name] = shutdownFunc
	return nil
}

func (r *defaultRunner) Start(ctx context.Context) <-chan error {
	defer func() {
		r.started = true
	}()

	ready := make(chan struct{})

	go func() {
		iCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if r.started {
			ready <- struct{}{}
			close(ready)

			r.finishWithTimeout(context.Background(), nil, fmt.Errorf("runner[%s] start cannot be completed: %w", r.name, ErrAlreadyStarted))

			return
		}

		if len(r.runnables) == 0 {
			ready <- struct{}{}
			close(ready)

			r.finishWithTimeout(context.Background(), nil, fmt.Errorf("runner[%s] start cannot be completed: %w", r.name, ErrNoRunnables))

			return
		}

		r.debugPrinterFunc(fmt.Sprintf("Runner[%s] initialized.", r.name))

		wg := sync.WaitGroup{}
		anyDone := make(chan finishedDTO, 1)
		isAnyDone := false
		runLock := sync.Mutex{}

		for x, y := range r.runnables {
			r.debugPrinterFunc(fmt.Sprintf("Runnable '%s' run has been taken from pool in runner[%s].", x.String(), r.name))

			wg.Add(1)
			go func(n string, m RunnableFunc) {
				r.debugPrinterFunc(fmt.Sprintf("Subroutine for '%s' has been initialized in runner[%s].", n, r.name))

				defer func() {
					r.debugPrinterFunc(fmt.Sprintf("Subroutine for '%s' has been closed in runner[%s].", n, r.name))

					runLock.Lock()
					if !isAnyDone {
						anyDone <- finishedDTO{
							Name: n,
							Err:  nil,
						}
						close(anyDone)
						isAnyDone = true
					}
					runLock.Unlock()

					wg.Done()
				}()

				if err := m(iCtx); err != nil {
					runLock.Lock()
					if !isAnyDone {
						anyDone <- finishedDTO{
							Name: n,
							Err:  fmt.Errorf("runnable '%s' failed during operation in runner[%s]: %w", n, r.name, err),
						}
						close(anyDone)
						isAnyDone = true
					}
					runLock.Unlock()
				}
			}(x.String(), y)
		}

		done := make(chan struct{}, 1)
		go func() {
			wg.Wait()
			done <- struct{}{}
			close(done)
		}()

		ready <- struct{}{}
		close(ready)

		select {
		case finDTO := <-anyDone:
			sCtx, sCancel := context.WithTimeout(context.Background(), r.shutdownTimeout)
			defer sCancel()

			if finDTO.Err != nil {
				r.debugPrinterFunc(fmt.Sprintf("Runner[%s] started shutdown procedure due to an error in '%s' runnable.", r.name, finDTO.Name))
			} else {
				r.debugPrinterFunc(fmt.Sprintf("Runner[%s] started shutdown procedure due to '%s' runnable stopping its operation.", r.name, finDTO.Name))
				finDTO.Err = fmt.Errorf("runnable '%s' finished operation without any cause in runner[%s]", finDTO.Name, r.name)
			}

			cancel()
			r.triggerRunnableShutdownHandlers(sCtx)

			r.finishWithTimeout(sCtx, done, finDTO.Err)
		case <-ctx.Done():
			sCtx, sCancel := context.WithTimeout(context.Background(), r.shutdownTimeout)
			defer sCancel()

			r.debugPrinterFunc(fmt.Sprintf("Runner[%s] received shutdown signal from context.", r.name))

			cancel()
			r.triggerRunnableShutdownHandlers(sCtx)

			r.finishWithTimeout(sCtx, done, nil)
		}
	}()

	<-ready

	return r.done
}

func (r *defaultRunner) triggerRunnableShutdownHandlers(ctx context.Context) {
	for m, x := range r.runnablesShutdowns {
		go func(n string, y ShutdownFunc) {
			r.debugPrinterFunc(fmt.Sprintf("Shutdown handler for runnable '%s' has been called in runner[%s].", n, r.name))
			y(ctx)
		}(m.String(), x)
	}
}

func (r *defaultRunner) finishWithTimeout(ctx context.Context, done <-chan struct{}, err error) {
	defer func() {
		r.done <- err
		close(r.done)
	}()

	if done == nil {
		return
	}

	select {
	case <-done:
	case <-ctx.Done():
		if err == nil {
			err = fmt.Errorf("runner[%s] shutdown procedure finished with an error: %w", r.name, ErrShutdownTimeout)
		}
		if err != nil {
			err = fmt.Errorf("%w [%s]", err, ErrShutdownTimeout.Error())
		}
	}
}
