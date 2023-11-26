package runner

import (
	"time"

	"github.com/SevenOfSpades/go-just-options"
)

const (
	optionShutdownTimeout options.OptionKey = `shutdown_timeout`
	optionDebugPrinter    options.OptionKey = `debug_printer_func`
	optionName            options.OptionKey = `name`
)

// OptionShutdownTimeout allows to overwrite default timeout value for shutdown procedure (graceful and failure).
func OptionShutdownTimeout(shutdownTimeout time.Duration) options.Option {
	return func(r options.Resolver) {
		options.WriteOrPanic[time.Duration](r, optionShutdownTimeout, shutdownTimeout)
	}
}

// OptionDebugPrinter allows to set DebugMessagePrinterFunc which can be used for displaying detailed logs
// during Runner operation.
func OptionDebugPrinter(debugPrinterFunc DebugMessagePrinterFunc) options.Option {
	return func(r options.Resolver) {
		options.WriteOrPanic[DebugMessagePrinterFunc](r, optionDebugPrinter, debugPrinterFunc)
	}
}

// OptionName sets name for Runner which makes easier to identify logs from specific Runner in OptionDebugPrinter
// when multiple runners are active in the same project.
func OptionName(name string) options.Option {
	return func(r options.Resolver) {
		options.WriteOrPanic[string](r, optionName, name)
	}
}
