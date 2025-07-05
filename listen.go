package sanity

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sanity-io/client-go/api"
)

// Listen returns a new builder for the listen API.
func (c *Client) Listen(query string) *ListenBuilder {
	return &ListenBuilder{c: c, query: query}
}

type ListenBuilder struct {
	c             *Client
	query         string
	params        map[string]interface{}
	includeResult bool
	tag           string
}

// Param adds a parameter to the listen query.
func (lb *ListenBuilder) Param(name string, val interface{}) *ListenBuilder {
	if lb.params == nil {
		lb.params = make(map[string]interface{}, 10)
	}
	lb.params[name] = val
	return lb
}

// IncludeResult controls whether the event payload contains the document.
func (lb *ListenBuilder) IncludeResult(b bool) *ListenBuilder {
	lb.includeResult = b
	return lb
}

// Tag sets a custom tag on the request.
func (lb *ListenBuilder) Tag(tag string) *ListenBuilder {
	lb.tag = tag
	return lb
}

// Do starts the listen stream and returns a ListenStream.
func (lb *ListenBuilder) Do(ctx context.Context) (*ListenStream, error) {
	req := lb.c.newQueryRequest().
		AppendPath("data/listen", lb.c.dataset).
		Param("query", lb.query).
		Header("accept", "text/event-stream").
		Tag(lb.tag, lb.c.tag)

	for p, v := range lb.params {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("marshaling parameter %q to JSON: %w", p, err)
		}
		req.Param("$"+p, string(b))
	}
	if lb.includeResult {
		req.Param("includeResult", true)
	}

	resp, err := lb.c.doRaw(ctx, req)
	if err != nil {
		return nil, err
	}

	return &ListenStream{resp: resp, r: bufio.NewReader(resp.Body)}, nil
}

// ListenStream represents a stream of listen events.
type ListenStream struct {
	resp *http.Response
	r    *bufio.Reader
}

// Close closes the underlying HTTP connection.
func (ls *ListenStream) Close() error {
	if ls.resp != nil && ls.resp.Body != nil {
		return ls.resp.Body.Close()
	}
	return nil
}

// Next reads the next event from the stream. It returns io.EOF when the
// connection is closed.
func (ls *ListenStream) Next() (*api.ListenEvent, error) {
	var (
		eventType string
		id        string
		dataBuf   bytes.Buffer
	)

	for {
		line, err := ls.r.ReadString('\n')
		if err != nil {
			if err == io.EOF && dataBuf.Len() == 0 {
				return nil, io.EOF
			}
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" { // end of event
			if dataBuf.Len() == 0 {
				// keep reading until we get a data line
				continue
			}
			data := dataBuf.Bytes()
			var payload json.RawMessage = append([]byte(nil), data...)
			return &api.ListenEvent{Type: eventType, ID: id, Data: &payload}, nil
		}
		switch {
		case strings.HasPrefix(line, "event:"):
			eventType = strings.TrimSpace(line[len("event:"):])
		case strings.HasPrefix(line, "id:"):
			id = strings.TrimSpace(line[len("id:"):])
		case strings.HasPrefix(line, "data:"):
			if dataBuf.Len() > 0 {
				dataBuf.WriteByte('\n')
			}
			dataBuf.WriteString(strings.TrimSpace(line[len("data:"):]))
		}
	}
}
