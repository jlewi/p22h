package gdocs

import (
	language "cloud.google.com/go/language/apiv1"
	"context"
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"github.com/jlewi/p22h/backend/pkg/datastore"
	"github.com/jlewi/p22h/backend/pkg/logging"
	"google.golang.org/api/docs/v1"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
)

type testDocs struct {
	docsbyName map[string]docs.Document
	docsById   map[string]docs.Document
	refsByName map[string]*datastore.DocReference
}

func TestIndexer_ProcessLinks(t *testing.T) {
	dir, err := ioutil.TempDir("", "testDatabase")
	if err != nil {
		t.Fatalf("Failed to create temporary directory; error %v", err)
	}

	log, err := logging.InitLogger("info", true)

	if err != nil {
		t.Fatalf("Failed to initialize logger; %v", err)
	}

	t.Logf("Created temporary directory: %v", dir)

	dbFile := path.Join(dir, "database.db")

	store, err := datastore.New(dbFile, *log)

	if err != nil {
		t.Fatalf("Failted to create datastore; error %v", err)
	}

	// Load the test documents

	data := loadTestDocs(t)
	for _, r := range data.refsByName {
		if err := store.UpdateDocReference(r); err != nil {
			t.Fatalf("Failed to load document reference into database; error %v", err)
		}
	}

	idx := &Indexer{
		log:         *log,
		store:       store,
		searcher:    nil,
		docsService: nil,
	}

	if err != nil {
		t.Fatalf("Failed to create indexer; error %v", err)
	}

	doc := data.docsbyName["test_doc.json"]
	if err := idx.ProcessDocLinks(data.refsByName["test_doc.json"], &doc); err != nil {
		t.Fatalf("indexing failed; error %v", err)
	}

	aLinks, err := idx.store.ListDocLinks("")

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

func TestIndexer_ProcessEntities(t *testing.T) {
	dir, err := ioutil.TempDir("", "testDatabase")
	if err != nil {
		t.Fatalf("Failed to create temporary directory; error %v", err)
	}

	log, err := logging.InitLogger("info", true)

	if err != nil {
		t.Fatalf("Failed to initialize logger; %v", err)
	}

	t.Logf("Created temporary directory: %v", dir)

	dbFile := path.Join(dir, "database.db")

	store, err := datastore.New(dbFile, *log)

	if err != nil {
		t.Fatalf("Failted to create datastore; error %v", err)
	}

	data := loadTestDocs(t)

	nlpClient, err := language.NewClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	idx := &Indexer{
		log:         *log,
		store:       store,
		searcher:    nil,
		docsService: nil,
		nlpClient:   nlpClient,
	}

	mockLanguage.Resps = []proto.Message{
		&languagepb.AnalyzeEntitiesResponse{
			Entities: []*languagepb.Entity{
				{
					Name: "john",
					Type: languagepb.Entity_PERSON,
					Mentions: []*languagepb.EntityMention{
						{
							Text: &languagepb.TextSpan{
								Content:     "john",
								BeginOffset: 10,
							},
						},
					},
				},
			},
		},
	}

	if err != nil {
		t.Fatalf("Failed to create indexer; error %v", err)
	}

	doc := data.docsbyName["test_doc.json"]
	if err := idx.ProcessEntities(data.refsByName["test_doc.json"], &doc); err != nil {
		t.Fatalf("indexing failed; error %v", err)
	}

	aEntities, err := idx.store.ListEntities()

	if err != nil {
		t.Errorf("failed to list doc links; error %v", err)
	}

	aMentions, err := idx.store.ListEntityMentions("")

	if err != nil {
		t.Errorf("failed to list entity mentions; error %v", err)
	}

	eEntities := []*datastore.Entity{
		{
			Name: "john",
			Type: languagepb.Entity_PERSON.String(),
		},
	}

	eMentions := []*datastore.EntityMention{
		{
			EntityID:   aEntities[0].ID,
			DocID:      data.refsByName["test_doc.json"].ID,
			Text:       "john",
			StartIndex: 10,
			EndIndex:   14,
		},
	}

	if d := cmp.Diff(eEntities, aEntities, datastore.GormIgnored(datastore.Entity{})); d != "" {
		t.Errorf("Did not get expected Entities; diff:%v\n", d)
	}

	if d := cmp.Diff(eMentions, aMentions, datastore.GormIgnored(datastore.EntityMention{})); d != "" {
		t.Errorf("Did not get expected EntityMentions; diff:%v\n", d)
	}
}

func loadTestDocs(t *testing.T) *testDocs {
	wDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory; %v", err)
	}

	testData := filepath.Join(wDir, "test_data")

	data := &testDocs{
		docsbyName: map[string]docs.Document{},
		docsById:   map[string]docs.Document{},
		refsByName: map[string]*datastore.DocReference{},
	}

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
		data.docsbyName[n] = *doc
		data.docsById[doc.DocumentId] = *doc

		r := &datastore.DocReference{
			ID:                     datastore.DriveKey(doc.DocumentId),
			DriveId:                doc.DocumentId,
			Name:                   doc.Title,
			MimeType:               DocumentMimeType,
			Md5Checksum:            "",
			LastIndexedMd5Checksum: "",
		}
		data.refsByName[n] = r
	}
	return data
}
