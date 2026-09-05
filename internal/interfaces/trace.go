package interfaces

import "context"

type ISpan interface {
	End()
}

type ITrace interface {
	Stop()

	StartSpan(
		ctx context.Context,
		operationName string,
		tags map[string]any,
	) (ISpan, context.Context)
}
