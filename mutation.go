package sanity

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sanity-io/client-go/api"
)

// Mutate returns a new mutation builder.
func (c *Client) Mutate() *MutationBuilder {
	return &MutationBuilder{
		c:          c,
		returnDocs: true,
		visibility: api.MutationVisibilitySync,
	}
}

type MutateResult struct {
	TransactionID string
	Results       []*api.MutateResultItem
}

type MutationBuilder struct {
	c             *Client
	items         []*api.MutationItem
	err           error
	returnIDs     bool
	returnDocs    bool
	visibility    api.MutationVisibility
	transactionID string
}

func (mb *MutationBuilder) Visibility(v api.MutationVisibility) *MutationBuilder {
	mb.visibility = v
	return mb
}

func (mb *MutationBuilder) TransactionID(id string) *MutationBuilder {
	mb.transactionID = id
	return mb
}

func (mb *MutationBuilder) ReturnIDs(enable bool) *MutationBuilder {
	mb.returnIDs = enable
	return mb
}

func (mb *MutationBuilder) ReturnDocuments(enable bool) *MutationBuilder {
	mb.returnDocs = enable
	return mb
}

func (mb *MutationBuilder) Do(ctx context.Context) (*MutateResult, error) {
	if mb.err != nil {
		return nil, fmt.Errorf("mutation builder: %w", mb.err)
	}

	req := mb.c.newRequest().
		Method(http.MethodPost).
		AppendPath("data/mutate", mb.c.dataset).
		Param("returnIds", mb.returnIDs).
		Param("returnDocuments", mb.returnDocs).
		Param("visibility", string(mb.visibility)).
		MarshalBody(&api.MutateRequest{
			Mutations: mb.items,
		})
	if mb.transactionID != "" {
		req.Param("transactionId", mb.transactionID)
	}

	var resp api.MutateResponse
	if _, err := mb.c.do(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("mutate: %w", err)
	}

	return &MutateResult{
		Results: resp.Results,
	}, nil
}

func (mb *MutationBuilder) Create(doc interface{}) *MutationBuilder {
	b, ok := mb.marshalJSON(doc)
	if ok {
		mb.items = append(mb.items, &api.MutationItem{Create: b})
	}
	return mb
}

func (mb *MutationBuilder) CreateIfNotExists(doc interface{}) *MutationBuilder {
	b, ok := mb.marshalJSON(doc)
	if ok {
		mb.items = append(mb.items, &api.MutationItem{CreateIfNotExists: b})
	}
	return mb
}

func (mb *MutationBuilder) CreateOrReplace(doc interface{}) *MutationBuilder {
	b, ok := mb.marshalJSON(doc)
	if ok {
		mb.items = append(mb.items, &api.MutationItem{CreateOrReplace: b})
	}
	return mb
}

func (mb *MutationBuilder) Delete(id string) *MutationBuilder {
	mb.items = append(mb.items, &api.MutationItem{Delete: &api.Delete{ID: id}})
	return mb
}

func (mb *MutationBuilder) Patch(id string) *PatchBuilder {
	patch := &api.Patch{ID: id}
	mb.items = append(mb.items, &api.MutationItem{Patch: patch})
	return &PatchBuilder{mb, patch}
}

func (mb *MutationBuilder) setErr(err error) {
	if mb.err == nil {
		mb.err = err
	}
}

func (mb *MutationBuilder) marshalJSON(val interface{}) (*json.RawMessage, bool) {
	b, err := marshalJSON(val)
	if err != nil {
		mb.setErr(fmt.Errorf("marshaling document: %w", err))
		return nil, false
	}

	return b, true
}

type PatchBuilder struct {
	mb    *MutationBuilder
	patch *api.Patch
}

func (pb *PatchBuilder) IfRevisionID(id string) *PatchBuilder {
	pb.patch.IfRevisionID = id
	return pb
}

func (pb *PatchBuilder) Query(query string) *PatchBuilder {
	pb.patch.Query = query
	return pb
}

func (pb *PatchBuilder) Set(path string, val interface{}) *PatchBuilder {
	if pb.patch.Set == nil {
		pb.patch.Set = map[string]*json.RawMessage{}
	}

	b, ok := pb.mb.marshalJSON(val)
	if ok {
		pb.patch.Set[path] = b
	}

	return pb
}

func (pb *PatchBuilder) SetIfMissing(path string, val interface{}) *PatchBuilder {
	if pb.patch.SetIfMissing == nil {
		pb.patch.SetIfMissing = map[string]*json.RawMessage{}
	}

	b, ok := pb.mb.marshalJSON(val)
	if ok {
		pb.patch.SetIfMissing[path] = b
	}

	return pb
}

func (pb *PatchBuilder) Unset(paths ...string) *PatchBuilder {
	pb.patch.Unset = append(pb.patch.Unset, paths...)
	return pb
}

func (pb *PatchBuilder) Inc(path string, n float64) *PatchBuilder {
	if pb.patch.Inc == nil {
		pb.patch.Inc = map[string]float64{}
	}

	pb.patch.Inc[path] = n
	return pb
}

func (pb *PatchBuilder) Dec(path string, n float64) *PatchBuilder {
	if pb.patch.Dec == nil {
		pb.patch.Dec = map[string]float64{}
	}

	pb.patch.Dec[path] = n
	return pb
}

func (pb *PatchBuilder) InsertBefore(path string, items ...interface{}) *PatchBuilder {
	bs := make([]*json.RawMessage, len(items))
	for i, item := range items {
		b, ok := pb.mb.marshalJSON(item)
		if !ok {
			return pb
		}
		bs[i] = b
	}

	pb.patch.Insert = &api.Insert{
		Before: path,
		Items:  bs,
	}
	return pb
}

func (pb *PatchBuilder) InsertAfter(path string, items ...interface{}) *PatchBuilder {
	bs := make([]*json.RawMessage, len(items))
	for i, item := range items {
		b, ok := pb.mb.marshalJSON(item)
		if !ok {
			return pb
		}
		bs[i] = b
	}

	pb.patch.Insert = &api.Insert{
		After: path,
		Items: bs,
	}
	return pb
}

func (pb *PatchBuilder) InsertReplace(path string, items ...interface{}) *PatchBuilder {
	bs := make([]*json.RawMessage, len(items))
	for i, item := range items {
		b, ok := pb.mb.marshalJSON(item)
		if !ok {
			return pb
		}
		bs[i] = b
	}

	pb.patch.Insert = &api.Insert{
		Replace: path,
		Items:   bs,
	}
	return pb
}

func (pb *PatchBuilder) End() *MutationBuilder {
	return pb.mb
}
