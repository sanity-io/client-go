package sanity

import (
	"encoding/json"
	"fmt"
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

func isMethodRetriable(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodDelete, http.MethodOptions:
		return true
	default:
		return false
	}
}

func marshalJSON(val interface{}) (*json.RawMessage, error) {
	switch val := val.(type) {
	case *json.RawMessage:
		return val, nil
	case []byte:
		return (*json.RawMessage)(&val), nil
	default:
		b, err := json.Marshal(val)
		if err != nil {
			return nil, fmt.Errorf("marshaling value of type %T to JSON: %w", val, err)
		}

		return (*json.RawMessage)(&b), nil
	}
}
