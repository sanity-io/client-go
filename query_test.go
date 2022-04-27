package sanity_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sanity "github.com/sanity-io/client-go"
	"github.com/sanity-io/client-go/api"
)

func TestQuery_basic(t *testing.T) {
	groq := "*[0]"

	now := time.Date(2020, 1, 2, 23, 01, 44, 0, time.UTC)

	testDoc := &testDocument{
		ID:        "123",
		Type:      "doc",
		CreatedAt: now,
		UpdatedAt: now,
		Value:     "hello world",
	}

	withSuite(t, func(s *Suite) {
		s.mux.Get("/v1/data/query/myDataset", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, groq, r.URL.Query().Get("query"))
			for k := range r.URL.Query() {
				assert.False(t, strings.HasPrefix("$", k))
			}

			w.WriteHeader(http.StatusOK)
			_, err := w.Write(mustJSONBytes(&api.QueryResponse{
				Ms:     12,
				Result: mustJSONMsg(testDoc),
			}))
			assert.NoError(t, err)
		})

		result, err := s.client.Query(groq).Do(context.Background())
		require.NoError(t, err)

		assert.Equal(t, 12*time.Millisecond, result.Time)

		b, err := json.Marshal(testDoc)
		require.NoError(t, err)
		assert.Equal(t, string(b), string(*result.Result))
	})
}

func TestQuery_params(t *testing.T) {
	groq := "*[0]"

	for _, tc := range []struct {
		desc string
		val  interface{}
	}{
		{"integer", 1},
		{"float", 1.23},
		{"string", "hello"},
		{"bool", true},

		{"integer array", []int{1}},
		{"float array", []float64{1.23}},
		{"string array", []string{"hello", "world"}},
		{"bool array", []bool{true, false}},

		{"custom", valueWithCustomJSONMarshaler(123)},
	} {
		t := t
		t.Run(tc.desc, func(t *testing.T) {
			withSuite(t, func(s *Suite) {
				s.mux.Get("/v1/data/query/myDataset", func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, groq, r.URL.Query().Get("query"))

					b, err := json.Marshal(tc.val)
					require.NoError(t, err)
					assert.Equal(t, string(b), r.URL.Query().Get("$val"))

					w.WriteHeader(http.StatusOK)
					_, err = w.Write(mustJSONBytes(&api.QueryResponse{
						Ms:     12,
						Result: mustJSONMsg(nil),
					}))
					assert.NoError(t, err)
				})

				_, err := s.client.Query(groq).Param("val", tc.val).Do(context.Background())
				require.NoError(t, err)
			})
		})
	}
}

func TestQuery_large(t *testing.T) {
	groq := "*[foo=='" + strings.Repeat("foo", 1000) + "']"

	withSuite(t, func(s *Suite) {
		s.mux.Post("/v1/data/query/myDataset", func(w http.ResponseWriter, r *http.Request) {
			var req api.QueryRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			assert.Equal(t, groq, req.Query)
			assert.Equal(t, "1.23", string(*req.Params["val"]))
			assert.Empty(t, r.URL.Query())

			w.WriteHeader(http.StatusOK)
			_, err := w.Write(mustJSONBytes(&api.QueryResponse{
				Ms:     12,
				Result: mustJSONMsg(nil),
			}))
			assert.NoError(t, err)
		})

		builder := s.client.Query(groq).Param("val", 1.23)

		_, err := builder.Do(context.Background())
		require.NoError(t, err)
	})
}

func TestQuery_tag(t *testing.T) {
	t.Run("small queries with default tag", func(t *testing.T) {
		groq := "*[foo=='" + strings.Repeat("foo", 1) + "']"

		withSuite(t, func(s *Suite) {
			s.mux.Get("/v1/data/query/myDataset", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "default", r.URL.Query().Get("tag"))

				w.WriteHeader(http.StatusOK)
				_, err := w.Write(mustJSONBytes(&api.QueryResponse{}))
				assert.NoError(t, err)
			})
			_, err := s.client.Query(groq).Param("val", 1.23).Do(context.Background())
			require.NoError(t, err)
		}, sanity.WithTag("default"))
	})
	t.Run("small queries overwrites tag", func(t *testing.T) {
		groq := "*[foo=='" + strings.Repeat("foo", 1) + "']"

		withSuite(t, func(s *Suite) {
			s.mux.Get("/v1/data/query/myDataset", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "tag", r.URL.Query().Get("tag"))

				w.WriteHeader(http.StatusOK)
				_, err := w.Write(mustJSONBytes(&api.QueryResponse{}))
				assert.NoError(t, err)
			})
			_, err := s.client.Query(groq).Param("val", 1.23).Tag("tag").Do(context.Background())
			require.NoError(t, err)
		}, sanity.WithTag("default"))
	})
	t.Run("large queries with default tag", func(t *testing.T) {
		groq := "*[foo=='" + strings.Repeat("foo", 1000) + "']"

		withSuite(t, func(s *Suite) {
			s.mux.Post("/v1/data/query/myDataset", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "default", r.URL.Query().Get("tag"))

				w.WriteHeader(http.StatusOK)
				_, err := w.Write(mustJSONBytes(&api.QueryResponse{}))
				assert.NoError(t, err)
			})
			_, err := s.client.Query(groq).Param("val", 1.23).Do(context.Background())
			require.NoError(t, err)
		}, sanity.WithTag("default"))
	})
	t.Run("large queries overwrites tag", func(t *testing.T) {
		groq := "*[foo=='" + strings.Repeat("foo", 1000) + "']"

		withSuite(t, func(s *Suite) {
			s.mux.Post("/v1/data/query/myDataset", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "tag", r.URL.Query().Get("tag"))

				w.WriteHeader(http.StatusOK)
				_, err := w.Write(mustJSONBytes(&api.QueryResponse{}))
				assert.NoError(t, err)
			})
			_, err := s.client.Query(groq).Param("val", 1.23).Tag("tag").Do(context.Background())
			require.NoError(t, err)
		}, sanity.WithTag("default"))
	})
}
