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

// ListenEvent represents a single event from the listen API.
type ListenEvent struct {
	Type string
	ID   string
	Data *json.RawMessage
}

// ActionsRequest represents a request to the actions API.
type ActionsRequest struct {
	Actions                             []interface{} `json:"actions"`
	TransactionID                       string        `json:"transactionId,omitempty"`
	SkipCrossDatasetReferenceValidation bool          `json:"skipCrossDatasetReferenceValidation,omitempty"`
	DryRun                              bool          `json:"dryRun,omitempty"`
}

// ActionsResponse holds the result of an actions API call.
type ActionsResponse struct {
	TransactionID string `json:"transactionId"`
}

// DocumentCreateAction represents a document.create action.
type DocumentCreateAction struct {
	ActionType  string           `json:"actionType"`
	PublishedID string           `json:"publishedId"`
	IfExists    string           `json:"ifExists,omitempty"`
	Document    *json.RawMessage `json:"document"`
}

// DocumentDeleteAction represents a document.delete action.
type DocumentDeleteAction struct {
	ActionType    string   `json:"actionType"`
	PublishedID   string   `json:"publishedId"`
	IncludeDrafts []string `json:"includeDrafts,omitempty"`
	Purge         bool     `json:"purge,omitempty"`
}

// DocumentDiscardAction represents a document.discard action (deprecated).
type DocumentDiscardAction struct {
	ActionType string `json:"actionType"`
	VersionID  string `json:"versionId"`
	Purge      bool   `json:"purge,omitempty"`
}

// DocumentEditAction represents a document.edit action.
type DocumentEditAction struct {
	ActionType  string           `json:"actionType"`
	PublishedID string           `json:"publishedId"`
	VersionID   string           `json:"versionId"`
	Patch       *json.RawMessage `json:"patch"`
}

// DocumentPublishAction represents a document.publish action.
type DocumentPublishAction struct {
	ActionType            string `json:"actionType"`
	PublishedID           string `json:"publishedId"`
	VersionID             string `json:"versionId"`
	IfDraftRevisionID     string `json:"ifDraftRevisionId,omitempty"`
	IfPublishedRevisionID string `json:"ifPublishedRevisionId,omitempty"`
}

// DocumentUnpublishAction represents a document.unpublish action.
type DocumentUnpublishAction struct {
	ActionType  string `json:"actionType"`
	PublishedID string `json:"publishedId"`
	VersionID   string `json:"versionId"`
}

// DocumentReplaceDraftAction represents a document.replaceDraft action (deprecated).
type DocumentReplaceDraftAction struct {
	ActionType  string           `json:"actionType"`
	PublishedID string           `json:"publishedId"`
	Purge       bool             `json:"purge,omitempty"`
	Attributes  *json.RawMessage `json:"attributes"`
}

// DocumentVersionCreateAction represents a document.version.create action.
type DocumentVersionCreateAction struct {
	ActionType       string           `json:"actionType"`
	PublishedID      string           `json:"publishedId"`
	Document         *json.RawMessage `json:"document,omitempty"`
	VersionID        string           `json:"versionId,omitempty"`
	BaseID           string           `json:"baseId,omitempty"`
	IfBaseRevisionID string           `json:"ifBaseRevisionId,omitempty"`
}

// DocumentVersionDiscardAction represents a document.version.discard action.
type DocumentVersionDiscardAction struct {
	ActionType string `json:"actionType"`
	VersionID  string `json:"versionId"`
	Purge      bool   `json:"purge,omitempty"`
}

// DocumentVersionReplaceAction represents a document.version.replace action.
type DocumentVersionReplaceAction struct {
	ActionType string           `json:"actionType"`
	VersionID  string           `json:"versionId"`
	Attributes *json.RawMessage `json:"attributes"`
	Purge      bool             `json:"purge,omitempty"`
}

// DocumentVersionUnpublishAction represents a document.version.unpublish action.
type DocumentVersionUnpublishAction struct {
	ActionType string `json:"actionType"`
	VersionID  string `json:"versionId"`
}

// ReleaseMetadata contains release metadata fields.
type ReleaseMetadata struct {
	Title             string `json:"title,omitempty"`
	Description       string `json:"description,omitempty"`
	ReleaseType       string `json:"releaseType,omitempty"`
	IntendedPublishAt string `json:"intendedPublishAt,omitempty"`
}

// ReleaseCreateAction represents a release.create action.
type ReleaseCreateAction struct {
	ActionType string           `json:"actionType"`
	ReleaseID  string           `json:"releaseId"`
	Metadata   *ReleaseMetadata `json:"metadata,omitempty"`
}

// ReleaseEditAction represents a release.edit action.
type ReleaseEditAction struct {
	ActionType string           `json:"actionType"`
	ReleaseID  string           `json:"releaseId"`
	Metadata   *ReleaseMetadata `json:"metadata,omitempty"`
}

// ReleaseDeleteAction represents a release.delete action.
type ReleaseDeleteAction struct {
	ActionType string `json:"actionType"`
	ReleaseID  string `json:"releaseId"`
	Purge      bool   `json:"purge,omitempty"`
}

// ReleasePublishAction represents a release.publish action.
type ReleasePublishAction struct {
	ActionType string `json:"actionType"`
	ReleaseID  string `json:"releaseId"`
}

// ReleaseScheduleAction represents a release.schedule action.
type ReleaseScheduleAction struct {
	ActionType string `json:"actionType"`
	ReleaseID  string `json:"releaseId"`
	PublishAt  string `json:"publishAt"`
}

// ReleaseArchiveAction represents a release.archive action.
type ReleaseArchiveAction struct {
	ActionType string `json:"actionType"`
	ReleaseID  string `json:"releaseId"`
}

// ReleaseUnarchiveAction represents a release.unarchive action.
type ReleaseUnarchiveAction struct {
	ActionType string `json:"actionType"`
	ReleaseID  string `json:"releaseId"`
}

// ReleaseUnscheduleAction represents a release.unschedule action.
type ReleaseUnscheduleAction struct {
	ActionType string `json:"actionType"`
	ReleaseID  string `json:"releaseId"`
}
