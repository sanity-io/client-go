package sanity

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/sanity-io/client-go/api"
	"github.com/sanity-io/client-go/internal/requests"
)

// Query returns a new query builder.
func (c *Client) Query(query string) *QueryBuilder {
	return &QueryBuilder{c: c, query: query}
}

// QueryResult holds the result of a query API call.
type QueryResult struct {
	// Time is the time taken.
	Time time.Duration

	// Result is the raw JSON of the query result.
	Result *json.RawMessage
}

// Unmarshal unmarshals the result into a Go value or struct. If there were no results, the
// destination value is set to the zero value.
func (q *QueryResult) Unmarshal(dest interface{}) error {
	if q.Result == nil {
		v := reflect.ValueOf(&dest)
		if v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
			i := reflect.Indirect(v)
			i.Set(reflect.Zero(i.Type()))
		}
		return nil
	}

	return json.Unmarshal([]byte(*q.Result), dest)
}

// QueryBuilder is a builder for queries.
type QueryBuilder struct {
	c      *Client
	query  string
	params map[string]interface{}
}

// Param adds a query parameter. For example, Param("foo", "bar") makes $foo usable inside the
// query. The passed-in value must be serializable to a JSON primitive.
func (qb *QueryBuilder) Param(name string, val interface{}) *QueryBuilder {
	if qb.params == nil {
		qb.params = make(map[string]interface{}, 10) // Small size
	}

	qb.params[name] = val
	return qb
}

// Query performs the query. On API failure, this will return an error of type *RequestError.
func (qb *QueryBuilder) Do(ctx context.Context) (*QueryResult, error) {
	req, err := qb.buildGET()
	if err != nil {
		return nil, err
	}

	if len(req.EncodeURL()) > maxGETRequestURLLength {
		req, err = qb.buildPOST()
		if err != nil {
			return nil, err
		}
	}

	var resp api.QueryResponse
	if _, err := qb.c.do(ctx, req, &resp); err != nil {
		return nil, err
	}

	result := &QueryResult{
		Time:   time.Duration(resp.Ms) * time.Millisecond,
		Result: resp.Result,
	}

	if qb.c.callbacks.OnQueryResult != nil {
		qb.c.callbacks.OnQueryResult(result)
	}

	return result, nil
}

func (qb *QueryBuilder) buildGET() (*requests.Request, error) {
	req := qb.c.newRequest().
		AppendPath("data/query", qb.c.dataset).
		Param("query", qb.query)
	for p, v := range qb.params {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("marshaling parameter %q to JSON: %w", p, err)
		}
		req.Param("$"+p, string(b))
	}
	return req, nil
}

func (qb *QueryBuilder) buildPOST() (*requests.Request, error) {
	request := &api.QueryRequest{
		Query:  qb.query,
		Params: make(map[string]*json.RawMessage, len(qb.params)),
	}

	for p, v := range qb.params {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("marshaling parameter %q to JSON: %w", p, err)
		}
		request.Params[p] = (*json.RawMessage)(&b)
	}

	return qb.c.newRequest().
		Method(http.MethodPost).
		AppendPath("data/query", qb.c.dataset).
		MarshalBody(request), nil
}

const maxGETRequestURLLength = 1024
