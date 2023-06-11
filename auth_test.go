package sanity

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestIsValidRequest(t *testing.T) {
	t.Parallel()

	type args struct {
		r             *http.Request
		webhookSecret string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "returns true when valid signature",
			args: args{
				r: &http.Request{
					Header: func() http.Header {
						header := make(http.Header)
						header.Set(SignatureHeaderName, "t=1633519811129,v1=tLa470fx7qkLLEcMOcEUFuBbRSkGujyskxrNXcoh0N0")
						return header
					}(),
					Body: func() io.ReadCloser {
						payload, _ := json.Marshal(map[string]string{"_id": "resume"})
						return io.NopCloser(bytes.NewReader(payload))
					}(),
				},
				webhookSecret: "test",
			},
			want: true,
		},
		{
			name: "returns false when invalid signature",
			args: args{
				r: &http.Request{
					Header: func() http.Header {
						header := make(http.Header)
						header.Set(SignatureHeaderName, "t=1633519811129,v1=tLa470fx7qkLLEcMOcEUFuBbRSkGujyskxrNXcoh0N0")
						return header
					}(),
					Body: func() io.ReadCloser {
						payload, _ := json.Marshal(map[string]string{"_id": "invalid"})
						return io.NopCloser(bytes.NewReader(payload))
					}(),
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsValidRequest(tt.args.r, tt.args.webhookSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValidRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsValidRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidSignature(t *testing.T) {
	t.Parallel()

	type args struct {
		payload       string
		signature     string
		webhookSecret string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "returns true when valid signature",
			args: args{
				payload: func() string {
					payload, _ := json.Marshal(map[string]string{"_id": "resume"})
					return string(payload)
				}(),
				signature:     "t=1633519811129,v1=tLa470fx7qkLLEcMOcEUFuBbRSkGujyskxrNXcoh0N0",
				webhookSecret: "test",
			},
			want: true,
		},
		{
			name: "returns false when invalid signature",
			args: args{
				payload: func() string {
					payload, _ := json.Marshal(map[string]string{"_id": "invalid"})
					return string(payload)
				}(),
				signature:     "t=1633519811129,v1=tLa470fx7qkLLEcMOcEUFuBbRSkGujyskxrNXcoh0N0",
				webhookSecret: "test",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsValidSignature(tt.args.payload, tt.args.signature, tt.args.webhookSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValidSignature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsValidSignature() got = %v, want %v", got, tt.want)
			}
		})
	}
}
