package runner

// RunnableName identifies runnable within single Runner instance.
// Proper naming also will make easier to identify specific runnable in logs provided by DebugMessagePrinterFunc.
type RunnableName string

func (n RunnableName) String() string {
	return string(n)
}
