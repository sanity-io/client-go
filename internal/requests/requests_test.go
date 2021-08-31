package requests_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sanity-io/client-go/internal/requests"
)

func TestRequest_AppendPath(t *testing.T) {
	tests := []struct {
		name  string
		elems []string
		want  string
	}{
		{
			name:  "empty string",
			elems: []string{""},
			want:  "//localhost",
		},
		{
			name:  "with path",
			elems: []string{"foo"},
			want:  "//localhost/foo",
		},
		{
			name:  "with multiple paths",
			elems: []string{"foo", "bar"},
			want:  "//localhost/foo/bar",
		},
		{
			name:  "with multiple paths and empty string",
			elems: []string{"foo", "", "bar"},
			want:  "//localhost/foo/bar",
		},
		{
			name:  "with empty string at the start",
			elems: []string{"", "foo", "bar"},
			want:  "//localhost/foo/bar",
		},
		{
			name:  "with empty string at the end",
			elems: []string{"foo", "bar", ""},
			want:  "//localhost/foo/bar",
		},
		{
			name:  "with slash in path string",
			elems: []string{"foo", "/", "bar"},
			want:  "//localhost/foo/bar",
		},
		{
			name:  "with slash in path string at the start",
			elems: []string{"/", "foo", "bar"},
			want:  "//localhost/foo/bar",
		},
		{
			name:  "with slash in path string at the end",
			elems: []string{"foo", "/", "bar"},
			want:  "//localhost/foo/bar",
		},
	}
	baseURL := url.URL{Host: "localhost"}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := requests.New(baseURL)
			got := r.AppendPath(tt.elems...)
			require.Equal(t, got.EncodeURL(), tt.want)
		})
	}
}
