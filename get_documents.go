package sanity

import (
	"context"
	"strings"

	"github.com/sanity-io/client-go/api"
	"github.com/sanity-io/client-go/internal/requests"
)

// GetDocuments returns a new GetDocuments builder.
func (c *Client) GetDocuments(docIDs ...string) *GetDocumentsBuilder {
	return &GetDocumentsBuilder{c: c, docIDs: docIDs}
}

// QueryBuilder is a builder for GET documents API.
type GetDocumentsBuilder struct {
	c      *Client
	docIDs []string
}

// Do performs the query.
// On validation failure, this will return an error of *InvalidRequestError.
// On API request failure, this will return an error of type *RequestError.
func (b *GetDocumentsBuilder) Do(ctx context.Context) (*api.GetDocumentsResponse, error) {
	req, err := b.buildGET()
	if err != nil {
		return nil, err
	}

	var resp api.GetDocumentsResponse
	if _, err := b.c.do(ctx, req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (b *GetDocumentsBuilder) buildGET() (*requests.Request, error) {
	if len(b.docIDs) == 0 {
		return nil, &InvalidRequestError{Description: "no document ID specified"}
	}

	req := b.c.newAPIRequest().
		AppendPath("data/doc", b.c.dataset, strings.Join(b.docIDs, ","))

	if len(req.EncodeURL()) > maxGETRequestURLLength {
		return nil, &InvalidRequestError{Description: "max URL length exceeded"}
	}
	return req, nil
}
