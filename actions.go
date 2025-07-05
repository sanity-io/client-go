package sanity

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sanity-io/client-go/api"
)

// Actions returns a new actions builder.
func (c *Client) Actions() *ActionsBuilder {
	return &ActionsBuilder{c: c}
}

type ActionsBuilder struct {
	c                                   *Client
	actions                             []interface{}
	transactionID                       string
	skipCrossDatasetReferenceValidation bool
	dryRun                              bool
	tag                                 string
}

// Action appends an action to perform.
func (b *ActionsBuilder) Action(a interface{}) *ActionsBuilder {
	b.actions = append(b.actions, a)
	return b
}

// TransactionID sets the transaction ID.
func (b *ActionsBuilder) TransactionID(id string) *ActionsBuilder {
	b.transactionID = id
	return b
}

// SkipCrossDatasetReferenceValidation disables cross dataset reference validation.
func (b *ActionsBuilder) SkipCrossDatasetReferenceValidation(v bool) *ActionsBuilder {
	b.skipCrossDatasetReferenceValidation = v
	return b
}

// DryRun performs the actions without applying them.
func (b *ActionsBuilder) DryRun(v bool) *ActionsBuilder {
	b.dryRun = v
	return b
}

// Tag sets a custom request tag.
func (b *ActionsBuilder) Tag(tag string) *ActionsBuilder {
	b.tag = tag
	return b
}

// Do performs the actions.
func (b *ActionsBuilder) Do(ctx context.Context) (*api.ActionsResponse, error) {
	reqBody := &api.ActionsRequest{
		Actions:                             b.actions,
		TransactionID:                       b.transactionID,
		SkipCrossDatasetReferenceValidation: b.skipCrossDatasetReferenceValidation,
		DryRun:                              b.dryRun,
	}

	req := b.c.newAPIRequest().
		Method(http.MethodPost).
		AppendPath("data/actions", b.c.dataset).
		MarshalBody(reqBody).
		Tag(b.tag, b.c.tag)

	var resp api.ActionsResponse
	if _, err := b.c.do(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("actions: %w", err)
	}
	return &resp, nil
}
