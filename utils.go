package sanity

import (
	"net/http"
)

func isStatusCodeRetriable(code int) bool {
	switch code {
	case http.StatusServiceUnavailable, http.StatusGatewayTimeout, http.StatusRequestTimeout:
		return true
	default:
		return false
	}
}
