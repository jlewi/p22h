package server

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	"github.com/jlewi/p22h/backend/pkg/datastore"
	"github.com/jlewi/p22h/backend/pkg/logging"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
)

func createDatastore(t *testing.T, logger logr.Logger, docLinks []*datastore.DocLink) *datastore.Datastore {
	dir, err := ioutil.TempDir("", "testDatabase")
	if err != nil {
		t.Fatalf("Failed to create temporary directory; error %v", err)
	}

	t.Logf("Created temporary directory: %v", dir)

	dbFile := path.Join(dir, "database.db")
	db, err := datastore.New(dbFile, logger)

	if err != nil {
		t.Fatalf("Failed to create database; error %v", err)
	}

	// Update the doc links
	for _, l := range docLinks {
		err = db.UpdateDocLink(l)
		if err != nil {
			t.Fatalf("Failed to add link %+v; %+v", l, err)
		}
	}
	return db
}

func TestServer_Backlinks(t *testing.T) {
	type testCase struct {
		name     string
		docName  string
		code     int
		docLinks []*datastore.DocLink
		body     string
	}
	cases := []testCase{
		{
			name:    "basic",
			docName: "doc2",
			code:    http.StatusOK,
			docLinks: []*datastore.DocLink{
				{
					Text:     "sometext",
					SourceID: "doc1",
					DestID:   "doc2",
				},
				{
					Text:     "otherText",
					SourceID: "doc4",
					DestID:   "doc2",
				},
				{
					Text:     "otherText",
					SourceID: "doc1",
					DestID:   "doc3",
				},
			},
			body: `{"items":[{"text":"sometext","docId":"doc1"},{"text":"otherText","docId":"doc4"}]}`,
		},
	}
	log, err := logging.InitLogger("info", true)
	if err != nil {
		t.Fatalf("Failed to initialize the logger")
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			store := createDatastore(t, *log, c.docLinks)
			s := Server{
				log:   *log,
				store: store,
			}
			path := fmt.Sprintf("/documents/%v:backLinks", c.docName)
			req := httptest.NewRequest(http.MethodGet, path, nil)
			resp := httptest.NewRecorder()

			// Need to create a router that we can pass the request through so that the vars will be added to the context
			router := mux.NewRouter()
			router.HandleFunc(backLinksPath, s.BackLinks)
			router.ServeHTTP(resp, req)

			result := resp.Result()
			if result.StatusCode != c.code {
				t.Fatalf("Got Code %v; want %v", result.StatusCode, c.code)
			}

			read, err := ioutil.ReadAll(result.Body)
			if err != nil {
				t.Fatalf("failed to read the response; error: %v", err)
			}
			if d := cmp.Diff(c.body, string(read)); d != "" {
				t.Errorf("Unexpected diff for body; Got:\n%v", d)
			}
		})
	}
}
