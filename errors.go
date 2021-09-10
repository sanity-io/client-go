package sanity

import (
	"fmt"
	"net/http"
)

// RequestError is returned for API requests that fail with a non-successful HTTP status code.
type RequestError struct {
	// Request is the attempted HTTP request that failed.
	Request *http.Request

	// Response is the HTTP response. Note that the body will no longer be valid.
	Response *http.Response

	// Body is the body of the response.
	Body []byte
}

// Error implements the error interface.
func (e *RequestError) Error() string {
	maxBody := 500
	body := string(e.Body)
	if len(body) > maxBody {
		body = fmt.Sprintf("%s [... and %d more bytes]", body[0:maxBody], len(body)-maxBody)
	}

	msg := fmt.Sprintf("HTTP request [%s %s] failed with status %d",
		e.Request.Method, e.Request.URL.String(), e.Response.StatusCode)
	if body != "" {
		msg += ": " + body
	}
	return msg
}
