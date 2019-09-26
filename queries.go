package sanity

import (
	"context"
	"encoding/json"
	"net/url"
	"reflect"
)

// QueryResult holds the result of a query API call.
type QueryResult struct {
	// Ms is the time taken, in milliseconds.
	Ms float64 `json:"ms"`

	// Query is the GROQ query.
	Query string `json:"query"`

	// Result is the raw JSON of the query result.
	Result *json.RawMessage `json:"result"`
}

func (q *QueryResult) unmarshal(out interface{}) error {
	if q.Result == nil {
		v := reflect.ValueOf(&out)
		if v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
			i := reflect.Indirect(v)
			i.Set(reflect.Zero(i.Type()))
		}
		return nil
	} else {
		return json.Unmarshal([]byte(*q.Result), out)
	}
}

// Query performs a query. The out parameter must point to a value or struct that
// will receive the result. json.Unmarshal is used to deserialize the JSON; if the type
// supports json.Unmarshaller, then it can override the unmarshalling.
//
// Zero or more parameters can be passed. For simplicity, use the Param() function
// to construct each parameter.
//
// On API failure, this will return an error of type *RequestError.
func (c *Client) Query(ctx context.Context, query string, out interface{}, params ...Parameter) error {
	uv := url.Values{
		"query": []string{query},
	}
	for _, p := range params {
		if err := p.build(uv); err != nil {
			return err
		}
	}

	var resp QueryResult
	if _, err := c.performGET(ctx, "data/query/"+c.dataset, uv, &resp); err != nil {
		return err
	}

	if c.callbacks.OnQueryResult != nil {
		c.callbacks.OnQueryResult(&resp)
	}

	return resp.unmarshal(out)
}
