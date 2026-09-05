package logging_gcp

import "cloud.google.com/go/logging"

func (gc *GCPClient) LogInfo(msg string, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}

	fields["message"] = msg

	gc.logs.Log(logging.Entry{
		Payload:  fields,
		Severity: logging.Info,
	})
}
