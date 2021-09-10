package sanity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"time"

	"github.com/jpillora/backoff"

	"github.com/sanity-io/client-go/internal/requests"
)

const (
	// Default dataset name for sanity projects
	DefaultDataset = "production"

	// API Host for skipping CDN
	APIHost = "api.sanity.io"

	// API Host which connects through CDN
	APICDNHost = "apicdn.sanity.io"

	// VersionV1 is version 1, the initial released version
	VersionV1 = Version("1")

	// VersionExperimental is the experimental version
	VersionExperimental = Version("X")

	// Latest release
	VersionV20210325 = Version("2021-03-25")
)

// Version is an API version, generally be dates in ISO format but also
// "1" (for backwards compatibility) and "X" (for experimental features)
type Version string

// String implements fmt.Stringer.
func (version Version) String() string {
	return string(version)
}

// Validate validates a version
func (version Version) Validate() error {
	if version == "" {
		return errors.New("no version given")
	}
	regExpVersion := regexp.MustCompile(`^(1|X|\d{4}-\d{2}-\d{2})$`)
	if !regExpVersion.MatchString(string(version)) {
		return fmt.Errorf("invalid version format %q", version)
	}
	return nil
}

// Client implements a client for interacting with the Sanity API.
type Client struct {
	hc            *http.Client
	apiVersion    Version
	useCDN        bool
	baseAPIURL    url.URL
	baseQueryURL  url.URL // if useCDN=false, baseQueryURL will be same as baseAPIURL.
	customHeaders http.Header
	token         string
	projectID     string
	dataset       string
	backoff       backoff.Backoff
	callbacks     Callbacks
	setHeaders    func(r *requests.Request)
}

type Option func(c *Client)

// WithHTTPClient returns an option for setting a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) { c.hc = client }
}

// WithCallbacks returns an option that enables callbacks for common events
// such as errors.
func WithCallbacks(cbs Callbacks) Option {
	return func(c *Client) { c.callbacks = cbs }
}

// WithBackoff returns an option that configures network request backoff. For how
// backoff works, see the underlying backoff package: https://github.com/jpillora/backoff.
// By default, the client uses the backoff package's default (maximum 10 seconds wait,
// backoff factor of 2).
func WithBackoff(b backoff.Backoff) Option {
	return func(c *Client) { c.backoff = b }
}

// WithToken returns an option that sets the API token to use.
func WithToken(t string) Option {
	return func(c *Client) { c.token = t }
}

// WithCDN returns an option that enables or disables the use of the Sanity API CDN.
// It is ignored when a custom HTTP host is set.
func WithCDN(b bool) Option {
	return func(c *Client) { c.useCDN = b }
}

// WithHTTPHost returns an option that changes the API URL.
func WithHTTPHost(scheme, host string) Option {
	return func(c *Client) {
		c.baseAPIURL.Scheme = scheme
		c.baseAPIURL.Host = host
		c.baseQueryURL.Scheme = scheme
		c.baseQueryURL.Host = host
	}
}

// WithHTTPHeader returns an option for setting a custom HTTP header.
// These headers are set in addition to the ones defined in Client.setHeaders().
// If a custom header is added with the same key as one of default header, then
// custom value is appended to key, and does not replace default value.
func WithHTTPHeader(key, value string) Option {
	return func(c *Client) {
		if c.customHeaders == nil {
			c.customHeaders = make(http.Header)
		}
		c.customHeaders.Add(key, value)
	}
}

// NewClient returns a new versioned client. A project ID must be provided.
// Zero or more options can be passed. For example:
//
//     client := sanity.VersionV20210325.NewClient("projectId", sanity.DefaultDataset,
//       sanity.WithCDN(true), sanity.WithToken("mytoken"))
//
func (v Version) NewClient(projectID, dataset string, opts ...Option) (*Client, error) {
	if projectID == "" {
		return nil, errors.New("project ID cannot be empty")
	}

	if dataset == "" {
		return nil, errors.New("dataset must be set")
	}

	baseAPIURL := fmt.Sprintf("%s.%s", projectID, APIHost)
	c := Client{
		backoff:    backoff.Backoff{Jitter: true},
		hc:         http.DefaultClient,
		projectID:  projectID,
		dataset:    dataset,
		apiVersion: v,
		baseAPIURL: url.URL{
			Scheme: "https",
			Host:   baseAPIURL,
			Path:   fmt.Sprintf("/v%s", v.String()),
		},
	}

	for _, opt := range opts {
		opt(&c)
	}

	c.baseQueryURL = c.baseAPIURL
	// Only use APICDN if useCDN=true and API host has not been updated by options.
	if c.useCDN && c.baseAPIURL.Host == baseAPIURL {
		c.baseQueryURL.Host = fmt.Sprintf("%s.%s", projectID, APICDNHost)
	}

	setDefaultHeaders := func(r *requests.Request) {
		r.Header("user-agent", "Sanity Go client/"+runtime.Version())
		if c.token != "" {
			r.Header("authorization", "Bearer "+c.token)
		}
	}

	c.setHeaders = func(r *requests.Request) {
		setDefaultHeaders(r)
		for key, values := range c.customHeaders {
			for _, value := range values {
				r.Header(key, value)
			}
		}
	}

	return &c, nil
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
	body := []byte("[no response body]")

	if resp.Body != nil {
		var err error
		if body, err = ioutil.ReadAll(resp.Body); err != nil {
			body = []byte(fmt.Sprintf("[failed to read response body: %s]", err))
		}
	}

	return &RequestError{
		Request:  req,
		Response: resp,
		Body:     body,
	}
}

func (c *Client) newAPIRequest() *requests.Request {
	r := requests.New(c.baseAPIURL)
	c.setHeaders(r)
	return r
}

func (c *Client) newQueryRequest() *requests.Request {
	r := requests.New(c.baseQueryURL)
	c.setHeaders(r)
	return r
}

const maxGETRequestURLLength = 1024
