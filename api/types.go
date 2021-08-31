package api

import (
	"encoding/json"
)

type MutateRequest struct {
	Mutations []*MutationItem `json:"mutations"`
}

type MutateResponse struct {
	TransactionID string              `json:"transactionId"`
	Results       []*MutateResultItem `json:"results"`
}

type MutationItem struct {
	Create            *json.RawMessage `json:"create,omitempty"`
	CreateIfNotExists *json.RawMessage `json:"createIfNotExists,omitempty"`
	CreateOrReplace   *json.RawMessage `json:"createOrReplace,omitempty"`
	Delete            *Delete          `json:"delete,omitempty"`
	Patch             *Patch           `json:"patch,omitempty"`
}

type Delete struct {
	ID string `json:"id"`
}

type Patch struct {
	ID             string                      `json:"id"`
	IfRevisionID   string                      `json:"ifRevisionID,omitempty"`
	Query          string                      `json:"query,omitempty"`
	Set            map[string]*json.RawMessage `json:"set,omitempty"`
	SetIfMissing   map[string]*json.RawMessage `json:"setIfMissing,omitempty"`
	DiffMatchPatch map[string]string           `json:"diffMatchPatch,omitempty"`
	Unset          []string                    `json:"unset,omitempty"`
	Insert         *Insert                     `json:"insert,omitempty"`
	Inc            map[string]float64          `json:"inc,omitempty"`
	Dec            map[string]float64          `json:"dec,omitempty"`
}

type Insert struct {
	Before  string             `json:"before,omitempty"`
	After   string             `json:"after,omitempty"`
	Replace string             `json:"replace,omitempty"`
	Items   []*json.RawMessage `json:"items"`
}

type MutateResultItem struct {
	Document *json.RawMessage `json:"document"`
}

// Unmarshal unmarshals the document into the passed-in struct.
func (i *MutateResultItem) Unmarshal(dest interface{}) error {
	return json.Unmarshal(*i.Document, dest)
}

type MutationVisibility string

const (
	MutationVisibilitySync     MutationVisibility = "sync"
	MutationVisibilityAsync    MutationVisibility = "async"
	MutationVisibilityDeferred MutationVisibility = "deferred"
)

type QueryRequest struct {
	Query  string                      `json:"query"`
	Params map[string]*json.RawMessage `json:"params"`
}

// QueryResponse holds the result of a query API call.
type QueryResponse struct {
	// Ms is the time taken, in milliseconds.
	Ms float64 `json:"ms"`

	// Query is the GROQ query.
	Query string `json:"query"`

	// Result is the raw JSON of the query result.
	Result *json.RawMessage `json:"result"`
}

// GetDocumentsResponse holds result of GET documents API call.
type GetDocumentsResponse struct {
	// Documents is slice of documents
	Documents []Document `json:"documents"`
}

// Document is a map of document attributes
type Document map[string]interface{}
