package runner

import (
	"fmt"
	"time"

	"github.com/SevenOfSpades/go-just-options"
)

// New will create new Runner instance.
// It can be configured with options:
// * OptionShutdownTimeout - set custom shutdown timeout (default is 10 seconds).
// * OptionDebugPrinter - set handler for debug logs (default handler is an empty function).
// * OptionName - set custom name for Runner ("default" is a default name for Runner).
func New(opts ...options.Option) (Runner, error) {
	opt := options.Resolve(opts)

	optShutdownTimeout, err := options.ReadOrDefault[time.Duration](opt, optionShutdownTimeout, time.Second*10)
	if err != nil {
		return nil, fmt.Errorf("runner initialization failed: %w", err)
	}
	optDebugPrinter, err := options.ReadOrDefault[DebugMessagePrinterFunc](opt, optionDebugPrinter, func(_ string) {})
	if err != nil {
		return nil, fmt.Errorf("runner initialization failed: %w", err)
	}
	optRunnerName, err := options.ReadOrDefault[string](opt, optionName, "default")
	if err != nil {
		return nil, fmt.Errorf("runner initialization failed: %w", err)
	}

	return newRunner(optShutdownTimeout, optDebugPrinter, optRunnerName), nil
}
