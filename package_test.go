package sanity_test

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"

	sanity "github.com/sanity-io/client-go"
)

type Suite struct {
	server *httptest.Server
	mux    chi.Router
	client *sanity.Client
}

func withSuite(t *testing.T, f func(*Suite), opts ...sanity.Option) {
	t.Helper()

	mux := chi.NewRouter()

	suite := &Suite{
		mux:    mux,
		server: httptest.NewServer(mux),
	}

	url, err := url.Parse(suite.server.URL)
	require.NoError(t, err)
	url.Path = "/v1"

	opts = append(opts, sanity.WithHTTPHost(url.Scheme, url.Host))

	c, err := sanity.VersionV1.NewClient("myProject", "myDataset", opts...)
	require.NoError(t, err)

	suite.client = c

	f(suite)
}

type testDocument struct {
	ID        string    `json:"_id"`
	Type      string    `json:"_type"`
	CreatedAt time.Time `json:"_createdAt"`
	UpdatedAt time.Time `json:"_updatedAt"`
	Value     string    `json:"value"`
}

func (d testDocument) toMap() map[string]interface{} {
	m, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}

	var doc map[string]interface{}
	if err := json.Unmarshal(m, &doc); err != nil {
		panic(err)
	}
	return doc
}

type testDocumentWithCustomJSONMarshaler struct{}

func (testDocumentWithCustomJSONMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(`{"x":1}`), nil
}

type testDocumentWithJSONMarshalFailure struct{}

func (testDocumentWithJSONMarshalFailure) MarshalJSON() ([]byte, error) {
	return nil, errMarshalFailure
}

type valueWithCustomJSONMarshaler float64

func (*valueWithCustomJSONMarshaler) MarshalJSON() ([]byte, error) {
	return []byte("x"), nil
}

var errMarshalFailure = errors.New("failure")

func mustJSONMsg(val interface{}) *json.RawMessage {
	b, err := json.Marshal(val)
	if err != nil {
		panic(err)
	}
	return (*json.RawMessage)(&b)
}

func mustJSONBytes(val interface{}) []byte {
	b, err := json.Marshal(val)
	if err != nil {
		panic(err)
	}
	return b
}
