package sanity

import (
	"fmt"
	"net/http"
)

type RequestError struct {
	Request  *http.Request
	Response *http.Response
	Body     []byte
}

func (e *RequestError) Error() string {
	maxBody := 500
	body := string(e.Body)
	if len(body) > maxBody {
		body = fmt.Sprintf("%s [... and %d more bytes]", body[0:maxBody], len(body)-maxBody)
	}
	return fmt.Sprintf("HTTP request [%s %s] failed with status %d: %s",
		e.Request.Method, e.Request.URL.String(), e.Response.StatusCode, body)
}
