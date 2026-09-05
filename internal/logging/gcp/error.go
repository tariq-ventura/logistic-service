package logging_gcp

import (
	"fmt"

	"cloud.google.com/go/errorreporting"
	"cloud.google.com/go/logging"
)

func (gc *GCPClient) LogError(msg string, fields map[string]any) {
	clientError := fmt.Errorf("Error: %s, Details: %v", msg, fields)

	gc.logs.Log(logging.Entry{
		Payload:  fields,
		Severity: logging.Error,
	})

	gc.errors.Report(errorreporting.Entry{
		Error: clientError,
	})
}
