package sanity_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sanity-io/client-go/api"
)

func TestActions_basic(t *testing.T) {
	withSuite(t, func(s *Suite) {
		s.mux.Post("/v1/data/actions/myDataset", func(w http.ResponseWriter, r *http.Request) {
			var body api.ActionsRequest
			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)
			assert.Equal(t, "test-tx", body.TransactionID)
			assert.Equal(t, true, body.DryRun)
			if assert.Len(t, body.Actions, 1) {
				m, ok := body.Actions[0].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "sanity.action.document.create", m["actionType"])
			}
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(mustJSONBytes(&api.ActionsResponse{}))
			assert.NoError(t, err)
		})

		action := map[string]interface{}{"actionType": "sanity.action.document.create"}
		_, err := s.client.Actions().Action(action).TransactionID("test-tx").DryRun(true).Do(context.Background())
		require.NoError(t, err)
	})
}

func TestActions_defaults(t *testing.T) {
	withSuite(t, func(s *Suite) {
		s.mux.Post("/v1/data/actions/myDataset", func(w http.ResponseWriter, r *http.Request) {
			var body api.ActionsRequest
			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)
			assert.Equal(t, "", body.TransactionID)
			assert.Equal(t, false, body.DryRun)
			assert.False(t, body.SkipCrossDatasetReferenceValidation)
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(mustJSONBytes(&api.ActionsResponse{}))
			assert.NoError(t, err)
		})

		action := map[string]interface{}{"actionType": "x"}
		_, err := s.client.Actions().Action(action).Do(context.Background())
		require.NoError(t, err)
	})
}
