package token

import "fmt"

// HTTPClientError reports errors reaching the remote service
type HTTPClientError struct {
	code    int
	message string
}

func (e *HTTPClientError) Error() string {
	return fmt.Sprintf("status: %d message: %s", e.code, e.message)
}
