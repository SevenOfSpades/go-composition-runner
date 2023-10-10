package runner

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunner(t *testing.T) {
	t.Run("it should start all registered runnables and close when context is canceled", func(t *testing.T) {
		t.Parallel()

		// GIVEN
		printBuff := bytes.NewBuffer(nil)
		printBuffLock := sync.Mutex{}

		r, err := New(OptionName("test"), OptionShutdownTimeout(time.Second), OptionDebugPrinter(func(msg string) {
			printBuffLock.Lock()
			defer printBuffLock.Unlock()

			printBuff.WriteString(msg)
		}))
		require.NoError(t, err)

		testRunnable1Finished := false
		_ = r.RegisterRunnableHandler("test-runnable-1", func(ctx context.Context) error {
			select {
			case <-time.After(time.Hour):
			case <-ctx.Done():
			}

			testRunnable1Finished = true

			return nil
		})

		testRunnable2Finished := false
		_ = r.RegisterRunnableHandler("test-runnable-2", func(ctx context.Context) error {
			select {
			case <-time.After(time.Hour):
			case <-ctx.Done():
			}

			testRunnable2Finished = true

			return nil
		})

		testRunnable3Finished := false
		testRunnable3Close := make(chan struct{}, 1)
		_ = r.RegisterRunnableHandler("test-runnable-3-with-shutdown-func", func(_ context.Context) error {
			<-testRunnable3Close
			testRunnable3Finished = true

			return nil
		})
		_ = r.RegisterRunnableShutdown("test-runnable-3-with-shutdown-func", func(ctx context.Context) {
			testRunnable3Close <- struct{}{}
			close(testRunnable3Close)
		})

		// WHEN
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
		defer cancel()

		// WHEN-THEN
		require.NoError(t, <-r.Start(ctx))
		assert.True(t, testRunnable1Finished)
		assert.True(t, testRunnable2Finished)
		assert.True(t, testRunnable3Finished)

		printVal := printBuff.String()
		assert.Contains(t, printVal, "Shutdown handler for runnable 'test-runnable-3-with-shutdown-func' has been called in runner[test].")
	})
	t.Run("it should stop all runnables if one of them stops without a failure", func(t *testing.T) {
		t.Parallel()

		// GIVEN
		r, err := New(OptionShutdownTimeout(time.Second))
		require.NoError(t, err)

		_ = r.RegisterRunnableHandler("test-runnable-1", func(ctx context.Context) error {
			return nil
		})

		testRunnable2Finished := false
		_ = r.RegisterRunnableHandler("test-runnable-2", func(ctx context.Context) error {
			select {
			case <-time.After(time.Hour):
			case <-ctx.Done():
			}

			testRunnable2Finished = true

			return nil
		})

		// WHEN
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
		defer cancel()

		// THEN
		require.EqualError(t, <-r.Start(ctx), "runnable 'test-runnable-1' finished operation without any cause in runner[default]")
		assert.True(t, testRunnable2Finished)
	})
	t.Run("it should stop all runnables if one of them fails", func(t *testing.T) {
		t.Parallel()

		// GIVEN
		r, err := New(OptionShutdownTimeout(time.Second))
		require.NoError(t, err)

		_ = r.RegisterRunnableHandler("test-runnable-1", func(ctx context.Context) error {
			return errors.New("test error")
		})

		testRunnable2Finished := false
		_ = r.RegisterRunnableHandler("test-runnable-2", func(ctx context.Context) error {
			select {
			case <-time.After(time.Hour):
			case <-ctx.Done():
			}

			testRunnable2Finished = true

			return nil
		})

		// WHEN
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
		defer cancel()

		// THEN
		require.EqualError(t, <-r.Start(ctx), "runnable 'test-runnable-1' failed during operation in runner[default]: test error")
		assert.True(t, testRunnable2Finished)
	})
}
