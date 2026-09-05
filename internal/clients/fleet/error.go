package clients_fleet

import "fmt"

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf(
		"fleet-service returned HTTP %d: %s",
		e.StatusCode,
		e.Body,
	)
}
