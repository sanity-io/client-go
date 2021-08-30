package sanity_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sanity "github.com/sanity-io/client-go"
	"github.com/sanity-io/client-go/api"
)

func TestAuth(t *testing.T) {
	withSuite(t, func(s *Suite) {
		s.mux.Get("/v1/data/query/myDataset", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer bork", r.Header.Get("Authorization"))

			w.WriteHeader(http.StatusOK)
			_, err := w.Write(mustJSONBytes(&api.QueryResponse{
				Ms:     12,
				Result: mustJSONMsg(nil),
			}))
			assert.NoError(t, err)
		})

		_, err := s.client.Query("*").Do(context.Background())
		require.NoError(t, err)
	}, sanity.WithToken("bork"))
}
