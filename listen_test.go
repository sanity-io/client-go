package sanity_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sanity "github.com/sanity-io/client-go"
)

var errRecordedRequest = errors.New("recorded")

type recordingTransport struct {
	req *http.Request
}

func (t *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.req = req
	return nil, errRecordedRequest
}

func TestListen_basic(t *testing.T) {
	withSuite(t, func(s *Suite) {
		s.mux.Get("/v1/data/listen/myDataset", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "*[_type == \"test\"]", r.URL.Query().Get("query"))
			assert.Equal(t, []string{"text/event-stream"}, r.Header["Accept"])
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

func TestListen_eventWithoutData(t *testing.T) {
	withSuite(t, func(s *Suite) {
		s.mux.Get("/v1/data/listen/myDataset", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			flusher := w.(http.Flusher)
			fmt.Fprint(w, "event: welcome\nid:0\n\n")
			fmt.Fprint(w, "event: mutation\ndata:{\"foo\":\"bar\"}\n\n")
			flusher.Flush()
		})
		stream, err := s.client.Listen("*").Do(context.Background())
		require.NoError(t, err)
		defer stream.Close()

		event, err := stream.Next()
		require.NoError(t, err)
		assert.Equal(t, "mutation", event.Type)
		assert.Equal(t, "", event.ID)
	})
}

func TestListen_multilineData(t *testing.T) {
	withSuite(t, func(s *Suite) {
		s.mux.Get("/v1/data/listen/myDataset", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			flusher := w.(http.Flusher)
			fmt.Fprint(w, "event: mutation\ndata:{\"foo\":\ndata:\"bar\"}\n\n")
			fmt.Fprint(w, "event: mutation\nid:2\ndata:{}\n\n")
			flusher.Flush()
		})
		stream, err := s.client.Listen("*").Do(context.Background())
		require.NoError(t, err)
		defer stream.Close()

		first, err := stream.Next()
		require.NoError(t, err)
		var m map[string]string
		require.NoError(t, json.Unmarshal(*first.Data, &m))
		assert.Equal(t, "bar", m["foo"])

		second, err := stream.Next()
		require.NoError(t, err)
		assert.Equal(t, "2", second.ID)
		assert.Equal(t, "{}", string(*second.Data))
	})
}

func TestListen_error(t *testing.T) {
	withSuite(t, func(s *Suite) {
		s.mux.Get("/v1/data/listen/myDataset", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":"bad query"}`)
		})
		stream, err := s.client.Listen("*").Do(context.Background())
		require.Error(t, err)
		assert.Nil(t, stream)

		var reqErr *sanity.RequestError
		require.True(t, errors.As(err, &reqErr))
		assert.Equal(t, http.StatusBadRequest, reqErr.Response.StatusCode)
		assert.Equal(t, `{"error":"bad query"}`, string(reqErr.Body))
	})
}

func TestListen_closeYieldsEOF(t *testing.T) {
	done := make(chan struct{})
	defer close(done)

	withSuite(t, func(s *Suite) {
		s.mux.Get("/v1/data/listen/myDataset", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			flusher := w.(http.Flusher)
			fmt.Fprint(w, "event: mutation\nid:1\ndata:{}\n\n")
			flusher.Flush()
			<-done // hold the stream open, so the client sees a close and not an EOF
		})
		stream, err := s.client.Listen("*").Do(context.Background())
		require.NoError(t, err)

		_, err = stream.Next()
		require.NoError(t, err)
		require.NoError(t, stream.Close())

		_, err = stream.Next()
		assert.Equal(t, io.EOF, err)
	})
}

func TestListen_contextCancellation(t *testing.T) {
	done := make(chan struct{})
	defer close(done)

	withSuite(t, func(s *Suite) {
		s.mux.Get("/v1/data/listen/myDataset", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			flusher := w.(http.Flusher)
			fmt.Fprint(w, "event: mutation\nid:1\ndata:{}\n\n")
			flusher.Flush()
			<-done
		})
		ctx, cancel := context.WithCancel(context.Background())
		stream, err := s.client.Listen("*").Do(ctx)
		require.NoError(t, err)
		defer stream.Close()

		_, err = stream.Next()
		require.NoError(t, err)

		cancel()
		_, err = stream.Next()
		assert.Equal(t, context.Canceled, err)
	})
}

func TestListen_skipsCDN(t *testing.T) {
	tr := &recordingTransport{}
	c, err := sanity.VersionV1.NewClient("myProject", "myDataset",
		sanity.WithCDN(true),
		sanity.WithHTTPClient(&http.Client{Transport: tr}))
	require.NoError(t, err)

	_, err = c.Listen("*").Do(context.Background())
	require.Error(t, err)

	require.NotNil(t, tr.req)
	assert.Equal(t, "myProject.api.sanity.io", tr.req.URL.Host)
	assert.Equal(t, []string{"text/event-stream"}, tr.req.Header["Accept"])
}
