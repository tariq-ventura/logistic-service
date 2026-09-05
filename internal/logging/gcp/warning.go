package logging_gcp

import "cloud.google.com/go/logging"

func (gc *GCPClient) LogWarning(msg string, fields map[string]any) {
	gc.logs.Log(logging.Entry{
		Payload:  fields,
		Severity: logging.Warning,
	})
}
