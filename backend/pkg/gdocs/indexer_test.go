package gdocs

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"github.com/jlewi/pkg/backend/pkg/datastore"
	"github.com/jlewi/pkg/backend/pkg/logging"
	"github.com/pkg/errors"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func Test_Index(t *testing.T) {
	wDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory; %v", err)
	}

	testData := filepath.Join(wDir, "test_data")

	// Load the test documents
	testDocs := map[string]docs.Document{}
	docsById := map[string]docs.Document{}
	for _, n := range []string{"test_doc.json", "test_doc2.json"} {
		p := filepath.Join(testData, n)
		b, err := ioutil.ReadFile(p)
		if err != nil {
			t.Fatalf("failed to read file; %v", p)
		}

		doc := &docs.Document{}

		if err := json.Unmarshal(b, doc); err != nil {
			t.Fatalf("failed to unmarshal Document from file; %v; error %v", p, err)
		}
		testDocs[n] = *doc
		docsById[doc.DocumentId] = *doc
	}

	client := NewTestClient(func(req *http.Request) *http.Response {
		t.Logf("Got http request: %v", req.URL.String())

		// get the serialized Document to return
		payload, err := func() ([]byte, error) {
			// We can't use ParseGoogleURI because the URL is docs.googleapis.com which is different from the one
			// used by the browser.
			path := strings.Split(req.URL.Path, "/")
			if len(path) == 0 {
				return nil, errors.Errorf("Could not parse URL; %v", req.URL.String())
			}
			id := path[len(path)-1]

			doc, ok := docsById[id]

			if !ok {
				return nil, errors.Errorf("Missing test document with id: %v", id)
			}
			b, err := json.Marshal(doc)

			if err != nil {
				return b, errors.Wrapf(err, "failed to serialize document with id: %v", id)
			}

			return b, nil
		}()

		if err != nil {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBufferString(err.Error())),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}
		}

		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBuffer(payload)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	// Example of using the fake round tripper
	docsService, err := docs.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		t.Fatalf("failed to create docs service; error %v", err)
	}

	if err != nil {
		t.Fatalf("failed to get doc; error %v", err)
	}

	log, err := logging.InitLogger("info", true)

	if err != nil {
		t.Fatalf("Failed to initialize logger; %v", err)
	}

	searcher := &FakeSearch{
		Docs: []*drive.File{},
	}

	for _, d := range testDocs {
		searcher.Docs = append(searcher.Docs, &drive.File{
			Id:       d.DocumentId,
			MimeType: DocumentMimeType,
		})
	}

	dir, err := ioutil.TempDir("", "testDatabase")
	if err != nil {
		t.Fatalf("Failed to create temporary directory; error %v", err)
	}

	t.Logf("Created temporary directory: %v", dir)

	dbFile := path.Join(dir, "database.db")

	store, err := datastore.New(dbFile, *log)

	if err != nil {
		t.Fatalf("Failted to create datastore; error %v", err)
	}

	idx, err := NewIndexer(searcher, docsService, store, *log)

	if err != nil {
		t.Fatalf("Failed to create indexer; error %v", err)
	}

	if err := idx.Index("someDrive"); err != nil {
		t.Fatalf("indexing failed; error %v", err)
	}

	aLinks, err := idx.store.ListDocLinks()

	if err != nil {
		t.Errorf("failed to list doc links; error %v", err)
	}

	eLinks := []*datastore.DocLink{
		{
			SourceID:   "gdrive.1n1hJJzqpzm8igA_gL27GaFOK4zsF08soTIy6qJXy8KE",
			DestID:     "gdrive.1xC1ORtF6imxbFyyng1ABximw-xW67j29UbGAlNlT4KY",
			URI:        "https://docs.google.com/document/d/1xC1ORtF6imxbFyyng1ABximw-xW67j29UbGAlNlT4KY/edit",
			Text:       "Link to Google Document",
			StartIndex: 51,
			EndIndex:   74,
		},
		// The second link is the chip.
		{
			SourceID:   "gdrive.1n1hJJzqpzm8igA_gL27GaFOK4zsF08soTIy6qJXy8KE",
			DestID:     "gdrive.1xC1ORtF6imxbFyyng1ABximw-xW67j29UbGAlNlT4KY",
			URI:        "https://docs.google.com/document/d/1xC1ORtF6imxbFyyng1ABximw-xW67j29UbGAlNlT4KY/edit",
			Text:       "Test Doc2",
			StartIndex: 97,
			EndIndex:   98,
		},
	}

	if d := cmp.Diff(eLinks, aLinks, datastore.GormIgnored(datastore.DocLink{})); d != "" {
		t.Errorf("Did not get expected DocLinks; diff:%v\n", d)
	}
}
