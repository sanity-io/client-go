package sanity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/jpillora/backoff"

	"github.com/sanity-io/client-go/internal/requests"
)

const DefaultDataset = "production"

// Client implements a client for interacting with the Sanity API.
type Client struct {
	hc        *http.Client
	useCDN    bool
	baseURL   url.URL
	token     string
	projectID string
	dataset   string
	backoff   backoff.Backoff
	callbacks Callbacks
}

type Option func(c *Client) error

// WithHTTPClient returns an option for setting a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) error {
		c.hc = client
		return nil
	}
}

// WithCallbacks returns an option that enables callbacks for common events
// such as errors.
func WithCallbacks(cbs Callbacks) Option {
	return func(c *Client) error {
		c.callbacks = cbs
		return nil
	}
}

// WithBackoff returns an option that configures network request backoff. For how
// backoff works, see the underlying backoff package: https://github.com/jpillora/backoff.
// By default, the client uses the backoff package's default (maximum 10 seconds wait,
// backoff factor of 2).
func WithBackoff(b backoff.Backoff) Option {
	return func(c *Client) error {
		c.backoff = b
		return nil
	}
}

// WithToken returns an option that sets the API token to use.
func WithToken(t string) Option {
	return func(c *Client) error {
		c.token = t
		return nil
	}
}

// WithDataset returns an option that sets the dataset name.
func WithDataset(id string) Option {
	return func(c *Client) error {
		c.dataset = id
		return nil
	}
}

// WithCDN returns an option that enables or disables the use of the Sanity API CDN.
func WithCDN(b bool) Option {
	return func(c *Client) error { c.useCDN = b; return nil }
}

// WithBaseURL returns an option that changes the API URL.
func WithBaseURL(url url.URL) Option {
	return func(c *Client) error { c.baseURL = url; return nil }
}

// New returns a new client. A project ID must be provided. Zero or more options can
// be passed. For example:
//
//     client := sanity.New("foobar123",
//       sanity.WithCDN(true), sanity.WithToken("mytoken"))
//
func New(projectID string, opts ...Option) (*Client, error) {
	if projectID == "" {
		return nil, errors.New("project ID cannot be empty")
	}

	c := Client{
		backoff:   backoff.Backoff{Jitter: true},
		hc:        http.DefaultClient,
		projectID: projectID,
		dataset:   DefaultDataset,
		baseURL: url.URL{
			Scheme: "https",
			Path:   "/v1",
		},
	}

	for _, opt := range opts {
		if err := opt(&c); err != nil {
			return nil, err
		}
	}

	if c.dataset == "" {
		return nil, errors.New("dataset must be set")
	}

	if c.baseURL.Host == "" {
		if c.useCDN {
			c.baseURL.Host = fmt.Sprintf("%s.apicdn.sanity.io", c.projectID)
		} else {
			c.baseURL.Host = fmt.Sprintf("%s.api.sanity.io", c.projectID)
		}
	}

	return &c, nil
}

// WithOptions returns a new client instance with options modified.
func (c *Client) WithOptions(opts ...Option) (*Client, error) {
	copy := *c
	for _, opt := range opts {
		if err := opt(&copy); err != nil {
			return nil, err
		}
	}
	return &copy, nil
}

func (c *Client) do(ctx context.Context, r *requests.Request, dest interface{}) (*http.Response, error) {
	bckoff := c.backoff
	for {
		req, err := r.HTTPRequest()
		if err != nil {
			return nil, err
		}

		req = req.WithContext(ctx)

		resp, err := c.hc.Do(req)
		if err != nil {
			return nil, fmt.Errorf("[%s %s] failed: %w", req.Method, req.URL.String(), err)
		}

		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			return resp, json.NewDecoder(resp.Body).Decode(dest)
		}

		if !isMethodRetriable(req.Method) || !isStatusCodeRetriable(resp.StatusCode) {
			return nil, c.handleErrorResponse(req, resp)
		}

		_ = resp.Body.Close()

		if c.callbacks.OnErrorWillRetry != nil {
			c.callbacks.OnErrorWillRetry(err)
		}

		time.Sleep(bckoff.Duration())
	}
}

func (c *Client) handleErrorResponse(req *http.Request, resp *http.Response) error {
	var body []byte
	if req.Body != nil {
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			b = []byte(fmt.Sprintf("[failed to read response body: %s]", err))
		}
		body = b
	}

	return &RequestError{
		Request:  req,
		Response: resp,
		Body:     body,
	}
}

func (c *Client) newRequest() *requests.Request {
	r := requests.New(c.baseURL)
	r.Header("Accept", "application/json")
	if c.token != "" {
		r.Header("Authorization", "Bearer "+c.token)
	}
	return r
}
