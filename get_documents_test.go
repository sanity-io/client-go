package sanity_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sanity "github.com/sanity-io/client-go"
	"github.com/sanity-io/client-go/api"
)

func TestGetDocuments(t *testing.T) {
	docIDs := []string{"doc1", "doc2"}
	now := time.Date(2020, 1, 2, 23, 01, 44, 0, time.UTC)
	testDoc1 := &testDocument{
		ID:        "doc1",
		Type:      "doc",
		CreatedAt: now,
		UpdatedAt: now,
		Value:     "hello world",
	}

	testDoc2 := &testDocument{
		ID:        "doc2",
		Type:      "doc",
		CreatedAt: now,
		UpdatedAt: now,
		Value:     "hello world",
	}

	testDocuments := []api.Document{testDoc1.toMap(), testDoc2.toMap()}

	t.Run("No document ID specified", func(t *testing.T) {
		withSuite(t, func(s *Suite) {
			resp, err := s.client.GetDocuments().Do(context.Background())
			require.NoError(t, err)
			require.Nil(t, resp.Documents)
		})
	})

	t.Run("Empty document ID specified", func(t *testing.T) {
		withSuite(t, func(s *Suite) {
			s.mux.Get("/v1/data/doc/myDataset", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			})

			_, err := s.client.GetDocuments([]string{""}...).Do(context.Background())
			require.Error(t, err)

			var reqErr *sanity.RequestError
			require.True(t, errors.As(err, &reqErr))
		})
	})

	t.Run("GET URL length exceeded", func(t *testing.T) {
		withSuite(t, func(s *Suite) {
			docID := make([]rune, 1024)
			for i := range docID {
				docID[i] = 'x'
			}
			_, err := s.client.GetDocuments(string(docID)).Do(context.Background())
			require.Error(t, err)
		})
	})

	t.Run("get 2 documents", func(t *testing.T) {
		withSuite(t, func(s *Suite) {
			s.mux.Get("/v1/data/doc/myDataset/doc1,doc2", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write(mustJSONBytes(&api.GetDocumentsResponse{
					Documents: testDocuments,
				}))
				assert.NoError(t, err)
			})

			result, err := s.client.GetDocuments(docIDs...).Do(context.Background())
			require.NoError(t, err)

			assert.Equal(t, testDocuments, result.Documents)
		})
	})
}
