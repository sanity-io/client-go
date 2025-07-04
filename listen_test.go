package sanity_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListen_basic(t *testing.T) {
	withSuite(t, func(s *Suite) {
		s.mux.Get("/v1/data/listen/myDataset", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "*[_type == \"test\"]", r.URL.Query().Get("query"))
			w.Header().Set("Content-Type", "text/event-stream")
			flusher, ok := w.(http.Flusher)
			require.True(t, ok)
			fmt.Fprint(w, "event: mutation\nid:1\ndata:{\"foo\":\"bar\"}\n\n")
			flusher.Flush()
		})
		stream, err := s.client.Listen("*[_type == \"test\"]").Do(context.Background())
		require.NoError(t, err)
		event, err := stream.Next()
		require.NoError(t, err)
		require.NoError(t, stream.Close())
		assert.Equal(t, "mutation", event.Type)
		assert.Equal(t, "1", event.ID)
		var m map[string]string
		require.NoError(t, json.Unmarshal(*event.Data, &m))
		assert.Equal(t, "bar", m["foo"])
	})
}
