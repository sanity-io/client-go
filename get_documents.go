package sanity

import (
	"context"
	"strings"

	"github.com/sanity-io/client-go/api"
)

// GetDocuments returns a new GetDocuments builder.
func (c *Client) GetDocuments(docIDs ...string) *GetDocumentsBuilder {
	return &GetDocumentsBuilder{c: c, docIDs: docIDs}
}

// QueryBuilder is a builder for GET documents API.
type GetDocumentsBuilder struct {
	c      *Client
	docIDs []string
	tag    string
}

func (b *GetDocumentsBuilder) Tag(tag string) *GetDocumentsBuilder {
	b.tag = tag
	return b
}

// Do performs the query.
// On API request failure, this will return an error of type *RequestError.
func (b *GetDocumentsBuilder) Do(ctx context.Context) (*api.GetDocumentsResponse, error) {
	if len(b.docIDs) == 0 {
		return &api.GetDocumentsResponse{}, nil
	}

	req := b.c.newAPIRequest().
		AppendPath("data/doc", b.c.dataset, strings.Join(b.docIDs, ",")).
		Tag(b.tag, b.c.tag)

	var resp api.GetDocumentsResponse
	if _, err := b.c.do(ctx, req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
