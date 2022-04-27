package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Request struct {
	baseURL         url.URL
	path            string
	method          string
	params          url.Values
	body            io.Reader
	headers         http.Header
	maxResponseSize int64
	err             error
}

func New(baseURL url.URL) *Request {
	return &Request{
		baseURL: baseURL,
		method:  http.MethodGet,
		headers: http.Header{
			"Accept": []string{"application/json"},
		},
	}
}

func (b *Request) HTTPRequest() (*http.Request, error) {
	if b.err != nil {
		return nil, b.err
	}

	req, err := http.NewRequest(b.method, b.EncodeURL(), b.body)
	if err != nil {
		return nil, err
	}

	for k, v := range b.headers {
		req.Header[k] = v
	}
	return req, nil
}

func (b *Request) EncodeURL() string {
	u := b.baseURL
	u.Path += b.path
	if b.params != nil {
		u.RawQuery = b.params.Encode()
	}
	return u.String()
}

func (b *Request) Method(m string) *Request {
	b.method = m
	return b
}

func (b *Request) Path(elems ...string) *Request {
	b.path = ""
	return b.AppendPath(elems...)
}

func (b *Request) AppendPath(elems ...string) *Request {
	for _, elem := range elems {
		if (b.path == "" || b.path[len(b.path)-1] != '/') &&
			(len(elem) > 0 && elem[0] != '/') {
			b.path += "/"
		}
		b.path += elem
	}
	return b
}

func (b *Request) Header(name, val string) *Request {
	if b.headers == nil {
		b.headers = make(http.Header, 10) // Small capacity
	}
	b.headers.Add(name, val)
	return b
}

func (b *Request) Param(name string, val interface{}) *Request {
	if b.params == nil {
		b.params = make(url.Values, 10) // Small capacity
	}

	switch val := val.(type) {
	case string:
		b.params.Add(name, val)
	case fmt.Stringer:
		b.params.Add(name, val.String())
	case bool:
		if val {
			b.params.Add(name, "true")
		} else {
			b.params.Add(name, "false")
		}
	default:
		panic(fmt.Sprintf("cannot add %q of type %T as parameter", name, val))
	}
	return b
}

func (b *Request) Tag(tag string, defaultTag string) *Request {
	if tag != "" {
		b.Param("tag", tag)
	} else if defaultTag != "" {
		b.Param("tag", defaultTag)
	}
	return b
}

func (b *Request) MaxResponseSize(limit int64) *Request {
	b.maxResponseSize = limit
	return b
}

func (b *Request) Body(body []byte) *Request {
	b.body = bytes.NewReader(body)
	return b
}

func (b *Request) ReadBody(r io.Reader) *Request {
	b.body = r
	return b
}

func (b *Request) MarshalBody(val interface{}) *Request {
	body, err := json.Marshal(val)
	if err != nil {
		b.err = fmt.Errorf("marshaling body value to JSON: %w", err)
		return b
	}

	b.body = bytes.NewReader(body)
	return b
}
