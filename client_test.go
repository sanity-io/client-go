package sanity_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sanity "github.com/sanity-io/client-go"
)

func TestAuthorization(t *testing.T) {
	withSuite(t, func(s *Suite) {
		s.mux.Get("/v1/data/query/myDataset", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer bork", r.Header.Get("Authorization"))

			_, err := w.Write([]byte("{}"))
			assert.NoError(t, err)
		})

		_, err := s.client.Query("*").Do(context.Background())
		require.NoError(t, err)
	},
		sanity.WithToken("bork"),
	)
}

func TestCustomHeaders(t *testing.T) {
	withSuite(t, func(s *Suite) {
		s.mux.Get("/v1/data/query/myDataset", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "bar", r.Header.Get("foo"))
			assert.Equal(t, []string{"application/json", "text/xml"}, r.Header.Values("accept"))

			_, err := w.Write([]byte("{}"))
			assert.NoError(t, err)
		})

		_, err := s.client.Query("*").Do(context.Background())
		require.NoError(t, err)
	},
		sanity.WithHTTPHeader("foo", "bar"),
		sanity.WithHTTPHeader("foo", "baz"), // Should be ignored
		sanity.WithHTTPHeader("accept", "text/xml"),
	)
}

func TestVersion_Validate(t *testing.T) {
	tests := []struct {
		name    string
		version sanity.Version
		wantErr bool
	}{
		{
			name:    "empty string",
			version: sanity.Version(""),
			wantErr: true,
		},
		{
			name:    "invalid date format",
			version: sanity.Version("2021-01"),
			wantErr: true,
		},
		{
			name:    "invalid date format",
			version: sanity.Version("20210101"),
			wantErr: true,
		},
		{
			name:    "valid date version",
			version: sanity.Version("2021-01-01"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.version.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
