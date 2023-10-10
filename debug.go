package runner

// DebugMessagePrinterFunc can be provided by OptionDebugPrinter as a way
// for printing detailed logs from Runner instance.
type DebugMessagePrinterFunc func(msg string)
