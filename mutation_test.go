package sanity_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sanity "github.com/sanity-io/client-go"
	"github.com/sanity-io/client-go/api"
)

func TestMutation_Builder(t *testing.T) {
	now := time.Date(2020, 1, 2, 23, 01, 44, 0, time.UTC)

	testDoc := &testDocument{
		ID:        "123",
		Type:      "doc",
		CreatedAt: now,
		UpdatedAt: now,
		Value:     "hello world",
	}

	for _, tc := range []struct {
		desc      string
		buildFunc func(b *sanity.MutationBuilder)
		expect    api.MutateRequest
	}{
		{
			"Create",
			func(b *sanity.MutationBuilder) {
				b.Create(testDoc)
			},
			api.MutateRequest{
				Mutations: []*api.MutationItem{{Create: mustJSONMsg(testDoc)}},
			},
		},
		{
			"CreateIfNotExists",
			func(b *sanity.MutationBuilder) {
				b.CreateIfNotExists(testDoc)
			},
			api.MutateRequest{
				Mutations: []*api.MutationItem{{CreateIfNotExists: mustJSONMsg(testDoc)}},
			},
		},
		{
			"CreateOrReplace",
			func(b *sanity.MutationBuilder) {
				b.CreateOrReplace(testDoc)
			},
			api.MutateRequest{
				Mutations: []*api.MutationItem{{CreateOrReplace: mustJSONMsg(testDoc)}},
			},
		},
		{
			"Delete",
			func(b *sanity.MutationBuilder) {
				b.Delete("123")
			},
			api.MutateRequest{
				Mutations: []*api.MutationItem{{Delete: &api.Delete{ID: "123"}}},
			},
		},
		{
			"empty patch",
			func(b *sanity.MutationBuilder) {
				b.Patch("123")
			},
			api.MutateRequest{
				Mutations: []*api.MutationItem{{Patch: &api.Patch{
					ID: "123",
				}}},
			},
		},
		{
			"patch with End",
			func(b *sanity.MutationBuilder) {
				b.Patch("123").End().Patch("234")
			},
			api.MutateRequest{
				Mutations: []*api.MutationItem{
					{Patch: &api.Patch{ID: "123"}},
					{Patch: &api.Patch{ID: "234"}},
				},
			},
		},
		{
			"Patch/IfRevisionID",
			func(b *sanity.MutationBuilder) {
				b.Patch("123").IfRevisionID("foo").Inc("values[0]", 1)
			},
			api.MutateRequest{
				Mutations: []*api.MutationItem{{Patch: &api.Patch{
					ID:           "123",
					IfRevisionID: "foo",
					Inc:          map[string]float64{"values[0]": 1},
				}}},
			},
		},
		{
			"Patch/Query",
			func(b *sanity.MutationBuilder) {
				b.Patch("123").Query("*").Inc("values[0]", 1)
			},
			api.MutateRequest{
				Mutations: []*api.MutationItem{{Patch: &api.Patch{
					ID:    "123",
					Query: "*",
					Inc:   map[string]float64{"values[0]": 1},
				}}},
			},
		},
		{
			"Patch/Inc",
			func(b *sanity.MutationBuilder) {
				b.Patch("123").Inc("values[0]", 1)
			},
			api.MutateRequest{
				Mutations: []*api.MutationItem{{Patch: &api.Patch{
					ID:  "123",
					Inc: map[string]float64{"values[0]": 1},
				}}},
			},
		},
		{
			"Patch/Dec",
			func(b *sanity.MutationBuilder) {
				b.Patch("123").Dec("values[0]", 1)
			},
			api.MutateRequest{
				Mutations: []*api.MutationItem{{Patch: &api.Patch{
					ID:  "123",
					Dec: map[string]float64{"values[0]": 1},
				}}},
			},
		},
		{
			"Patch/Set",
			func(b *sanity.MutationBuilder) {
				b.Patch("123").Set("a", testDoc)
			},
			api.MutateRequest{
				Mutations: []*api.MutationItem{{Patch: &api.Patch{
					ID: "123",
					Set: map[string]*json.RawMessage{
						"a": mustJSONMsg(testDoc),
					},
				}}},
			},
		},
		{
			"Patch/SetIfMissing",
			func(b *sanity.MutationBuilder) {
				b.Patch("123").SetIfMissing("a", testDoc)
			},
			api.MutateRequest{
				Mutations: []*api.MutationItem{{Patch: &api.Patch{
					ID: "123",
					SetIfMissing: map[string]*json.RawMessage{
						"a": mustJSONMsg(testDoc),
					},
				}}},
			},
		},
		{
			"Patch/Unset",
			func(b *sanity.MutationBuilder) {
				b.Patch("123").Unset("a", "b")
			},
			api.MutateRequest{
				Mutations: []*api.MutationItem{{Patch: &api.Patch{
					ID:    "123",
					Unset: []string{"a", "b"},
				}}},
			},
		},
		{
			"Patch/InsertAfter",
			func(b *sanity.MutationBuilder) {
				b.Patch("123").InsertAfter("array[3]", testDoc, "doink")
			},
			api.MutateRequest{
				Mutations: []*api.MutationItem{{Patch: &api.Patch{
					ID: "123",
					Insert: &api.Insert{
						After: "array[3]",
						Items: []*json.RawMessage{
							mustJSONMsg(testDoc),
							mustJSONMsg("doink"),
						},
					},
				}}},
			},
		},
		{
			"Patch/InsertBefore",
			func(b *sanity.MutationBuilder) {
				b.Patch("123").InsertBefore("array[3]", testDoc, "doink")
			},
			api.MutateRequest{
				Mutations: []*api.MutationItem{{Patch: &api.Patch{
					ID: "123",
					Insert: &api.Insert{
						Before: "array[3]",
						Items: []*json.RawMessage{
							mustJSONMsg(testDoc),
							mustJSONMsg("doink"),
						},
					},
				}}},
			},
		},
		{
			"Patch/InsertReplace",
			func(b *sanity.MutationBuilder) {
				b.Patch("123").InsertReplace("array[3]", testDoc, "doink")
			},
			api.MutateRequest{
				Mutations: []*api.MutationItem{{Patch: &api.Patch{
					ID: "123",
					Insert: &api.Insert{
						Replace: "array[3]",
						Items: []*json.RawMessage{
							mustJSONMsg(testDoc),
							mustJSONMsg("doink"),
						},
					},
				}}},
			},
		},
	} {
		t := t
		t.Run(tc.desc, func(t *testing.T) {
			withSuite(t, func(s *Suite) {
				s.mux.Post("/v1/mutate/myDataset", func(w http.ResponseWriter, r *http.Request) {
					var req api.MutateRequest
					require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
					assert.Equal(t, tc.expect, req)

					w.WriteHeader(http.StatusOK)
					_, err := w.Write(mustJSONBytes(&api.MutateResponse{}))
					assert.NoError(t, err)
				})

				builder := s.client.Mutate()
				tc.buildFunc(builder)

				_, err := builder.Do(context.Background())
				require.NoError(t, err)
			})
		})
	}
}

func TestMutation_Builder_returnIDs(t *testing.T) {
	t.Run("can be set to true", func(t *testing.T) {
		withSuite(t, func(s *Suite) {
			s.mux.Post("/v1/mutate/myDataset", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "true", r.URL.Query().Get("returnIds"))
				w.WriteHeader(http.StatusOK)
				_, err := w.Write(mustJSONBytes(&api.MutateResponse{}))
				assert.NoError(t, err)
			})

			_, err := s.client.Mutate().ReturnIDs(true).Do(context.Background())
			require.NoError(t, err)
		})
	})

	t.Run("defaults to false", func(t *testing.T) {
		withSuite(t, func(s *Suite) {
			s.mux.Post("/v1/mutate/myDataset", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "true", r.URL.Query().Get("returnIds"))
				w.WriteHeader(http.StatusOK)
				_, err := w.Write(mustJSONBytes(&api.MutateResponse{}))
				assert.NoError(t, err)
			})

			_, err := s.client.Mutate().ReturnIDs(true).Do(context.Background())
			require.NoError(t, err)
		})
	})
}

func TestMutation_Builder_marshalError(t *testing.T) {
	withSuite(t, func(s *Suite) {
		_, err := s.client.Mutate().Create(&testDocumentWithJSONMarshalFailure{}).Do(context.Background())
		require.Error(t, err)
		assert.True(t, errors.Is(err, errMarshalFailure))
	})
}

func TestMutation_Builder_customJSONMarshaling(t *testing.T) {
	t.Run("can be set", func(t *testing.T) {
		withSuite(t, func(s *Suite) {
			s.mux.Post("/v1/mutate/myDataset", func(w http.ResponseWriter, r *http.Request) {
				b, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				assert.Equal(t, `{"mutations":[{"create":{"x":1}}]}`, string(b))

				w.WriteHeader(http.StatusOK)

				_, err = w.Write(mustJSONBytes(&api.MutateResponse{}))
				assert.NoError(t, err)
			})

			_, err := s.client.Mutate().Create(&testDocumentWithCustomJSONMarshaler{}).Do(context.Background())
			require.NoError(t, err)
		})
	})
}

func TestMutation_Builder_unmarshalResult(t *testing.T) {
	withSuite(t, func(s *Suite) {
		s.mux.Post("/v1/mutate/myDataset", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "x", r.URL.Query().Get("transactionId"))
			w.WriteHeader(http.StatusOK)
			_, err := w.Write(mustJSONBytes(&api.MutateResponse{}))
			assert.NoError(t, err)
		})

		_, err := s.client.Mutate().TransactionID("x").Do(context.Background())
		require.NoError(t, err)
	})
}

func TestMutation_Builder_transactionID(t *testing.T) {
	t.Run("can be set", func(t *testing.T) {
		withSuite(t, func(s *Suite) {
			s.mux.Post("/v1/mutate/myDataset", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "x", r.URL.Query().Get("transactionId"))
				w.WriteHeader(http.StatusOK)
				_, err := w.Write(mustJSONBytes(&api.MutateResponse{}))
				assert.NoError(t, err)
			})

			_, err := s.client.Mutate().TransactionID("x").Do(context.Background())
			require.NoError(t, err)
		})
	})

	t.Run("not included by default", func(t *testing.T) {
		withSuite(t, func(s *Suite) {
			s.mux.Post("/v1/mutate/myDataset", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "", r.URL.Query().Get("transactionId"))
				w.WriteHeader(http.StatusOK)
				_, err := w.Write(mustJSONBytes(&api.MutateResponse{}))
				assert.NoError(t, err)
			})

			_, err := s.client.Mutate().Do(context.Background())
			require.NoError(t, err)
		})
	})
}

func TestMutation_Builder_returnDocumentsOption(t *testing.T) {
	t.Run("can be set to false", func(t *testing.T) {
		withSuite(t, func(s *Suite) {
			s.mux.Post("/v1/mutate/myDataset", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "false", r.URL.Query().Get("returnDocuments"))
				w.WriteHeader(http.StatusOK)
				_, err := w.Write(mustJSONBytes(&api.MutateResponse{}))
				assert.NoError(t, err)
			})

			_, err := s.client.Mutate().ReturnDocuments(false).Do(context.Background())
			require.NoError(t, err)
		})
	})

	t.Run("defaults to true", func(t *testing.T) {
		withSuite(t, func(s *Suite) {
			s.mux.Post("/v1/mutate/myDataset", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "true", r.URL.Query().Get("returnDocuments"))
				w.WriteHeader(http.StatusOK)
				_, err := w.Write(mustJSONBytes(&api.MutateResponse{}))
				assert.NoError(t, err)
			})

			_, err := s.client.Mutate().Do(context.Background())
			require.NoError(t, err)
		})
	})
}

func TestMutation_Builder_visibilityOption(t *testing.T) {
	for _, v := range []api.MutationVisibility{
		api.MutationVisibilityAsync,
		api.MutationVisibilityDeferred,
		api.MutationVisibilitySync,
	} {
		t.Run(fmt.Sprintf("can be set to %q", v), func(t *testing.T) {
			withSuite(t, func(s *Suite) {
				s.mux.Post("/v1/mutate/myDataset", func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, string(v), r.URL.Query().Get("visibility"))
					w.WriteHeader(http.StatusOK)
					_, err := w.Write(mustJSONBytes(&api.MutateResponse{}))
					assert.NoError(t, err)
				})

				_, err := s.client.Mutate().Visibility(v).Do(context.Background())
				require.NoError(t, err)
			})
		})
	}

	t.Run("defaults to sync", func(t *testing.T) {
		withSuite(t, func(s *Suite) {
			s.mux.Post("/v1/mutate/myDataset", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, string(api.MutationVisibilitySync), r.URL.Query().Get("visibility"))
				w.WriteHeader(http.StatusOK)
				_, err := w.Write(mustJSONBytes(&api.MutateResponse{}))
				assert.NoError(t, err)
			})

			_, err := s.client.Mutate().Do(context.Background())
			require.NoError(t, err)
		})
	})
}
