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
	"sync/atomic"

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

func (lb *ListenBuilder) IncludeResult(b bool) *ListenBuilder {
	lb.includeResult = b
	return lb
}

func (lb *ListenBuilder) Tag(tag string) *ListenBuilder {
	lb.tag = tag
	return lb
}

// Do starts the listen stream and returns a ListenStream. The listen endpoint is not cacheable, so
// the request always goes to the API host, never the CDN host. It accepts GET only, so a query
// that exceeds the maximum URL length fails with an error.
func (lb *ListenBuilder) Do(ctx context.Context) (*ListenStream, error) {
	req := lb.c.newAPIRequest().
		AppendPath("data/listen", lb.c.dataset).
		Param("query", lb.query).
		SetHeader("accept", "text/event-stream").
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

	return &ListenStream{ctx: ctx, resp: resp, r: bufio.NewReader(resp.Body)}, nil
}

// ListenStream represents a stream of listen events.
type ListenStream struct {
	ctx    context.Context
	resp   *http.Response
	r      *bufio.Reader
	closed int32
}

func (ls *ListenStream) Close() error {
	atomic.StoreInt32(&ls.closed, 1)
	if ls.resp != nil && ls.resp.Body != nil {
		return ls.resp.Body.Close()
	}
	return nil
}

// Next reads the next event from the stream. It returns io.EOF when the server ends the
// stream and after Close.
func (ls *ListenStream) Next() (*api.ListenEvent, error) {
	var (
		eventType string
		id        string
		dataBuf   bytes.Buffer
		haveData  bool
	)

	for {
		line, err := ls.r.ReadString('\n')
		if err != nil {
			return nil, ls.readError(err)
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" { // end of event
			if !haveData {
				eventType, id = "", ""
				continue
			}
			payload := json.RawMessage(dataBuf.Bytes())
			return &api.ListenEvent{Type: eventType, ID: id, Data: &payload}, nil
		}

		switch {
		case strings.HasPrefix(line, "event:"):
			eventType = strings.TrimSpace(line[len("event:"):])
		case strings.HasPrefix(line, "id:"):
			id = strings.TrimSpace(line[len("id:"):])
		case strings.HasPrefix(line, "data:"):
			if haveData {
				dataBuf.WriteByte('\n')
			}
			dataBuf.WriteString(strings.TrimSpace(line[len("data:"):]))
			haveData = true
		}
	}
}

func (ls *ListenStream) readError(err error) error {
	if atomic.LoadInt32(&ls.closed) == 1 {
		return io.EOF
	}
	if ctxErr := ls.ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	return err
}
