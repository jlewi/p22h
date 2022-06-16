package gdocs

import (
	language "cloud.google.com/go/language/apiv1"
	"context"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/jlewi/p22h/backend/pkg/datastore"
	"github.com/jlewi/p22h/backend/pkg/glanguage"
	"github.com/jlewi/p22h/backend/pkg/logging"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"net/http"
)

// Indexer indexes Google Drive.
type Indexer struct {
	log        logr.Logger
	store      *datastore.Datastore
	httpClient *http.Client
	searcher   DriveSearch

	docsService *docs.Service
	nlpClient   *language.Client
}

// NewIndexer creates a new indexer
func NewIndexer(searcher DriveSearch, docsService *docs.Service, store *datastore.Datastore, nlpClient *language.Client, log logr.Logger, opts ...IndexerOption) (*Indexer, error) {
	if searcher == nil {
		return nil, errors.New("client is required")
	}

	if store == nil {
		return nil, errors.New("store is required")
	}

	if docsService == nil {
		return nil, errors.New("docsService is required")
	}

	if nlpClient == nil {
		return nil, errors.New("nlpClient is required")
	}

	idx := &Indexer{
		log:         log,
		store:       store,
		searcher:    searcher,
		docsService: docsService,
		nlpClient:   nlpClient,
	}

	for _, o := range opts {
		o(idx)
	}

	return idx, nil
}

type IndexerOption func(*Indexer)

func IndexerWithLogger(log logr.Logger) IndexerOption {
	return func(idx *Indexer) {
		idx.log = log
	}
}

func IndexerWithHTTPClient(c *http.Client) IndexerOption {
	return func(idx *Indexer) {
		idx.httpClient = c
	}
}

// newDbInserter returns a ResultFunc that will insert documents into a datastore.
func newDbInserter(store *datastore.Datastore) (ResultFunc, error) {
	if store == nil {
		return nil, errors.New("store can't be nil")
	}

	return func(f *drive.File) error {
		r := datastore.DocReference{
			DriveId:     f.Id,
			Name:        f.Name,
			MimeType:    f.MimeType,
			Md5Checksum: f.Md5Checksum,
		}

		return store.UpdateDocReference(&r)
	}, nil
}

// TODO(jeremy): Should rename this IndexFolder or IndexDrive
func (idx *Indexer) Index(driveId string) error {
	log := idx.log
	log.Info("Indexing drive", "driveId", driveId)

	query := ""
	corpora := "drive"

	f, err := newDbInserter(idx.store)
	if err != nil {
		return errors.Wrapf(err, "Failed to create newDbInserter")
	}
	if err := idx.searcher.Search(query, driveId, corpora, f); err != nil {
		return errors.Wrapf(err, "Failed to search: driveId %v", driveId)
	}

	docReferences, err := idx.store.ToBeIndexed()

	if err != nil {
		return errors.Wrapf(err, "Failed to get docs needing indexing")
	}

	for _, r := range docReferences {
		idx.ProcessDoc(r)
	}

	return nil
}

// IndexDocument indexes a specific document
func (idx *Indexer) IndexDocument(docId string) error {
	log := idx.log
	log.Info("Indexing doc", "driveId", docId)

	svc, err := drive.NewService(context.Background(), option.WithHTTPClient(idx.httpClient))

	if err != nil {
		return errors.Wrapf(err, "Failed to create drive client")
	}

	f, err := svc.Files.Get(docId).Do()

	if err != nil {
		return errors.Wrapf(err, "Failed to get Drive document: %v", docId)
	}

	r := &datastore.DocReference{
		DriveId:     f.Id,
		Name:        f.Name,
		MimeType:    f.MimeType,
		Md5Checksum: f.Md5Checksum,
	}

	if err := idx.store.UpdateDocReference(r); err != nil {
		return errors.Wrapf(err, "Failed to UpdateDocReference; DocId: %v", docId)
	}

	idx.ProcessDoc(r)
	return nil
}

// ProcessDoc processes the referenced doc.
//
// TODO(jeremy): This function should really return an error. Originally it wasn't returning an error because it
// was only being called from Index which just continued but now its being called from IndexDocument and we should
// propogate the error to that.
func (idx *Indexer) ProcessDoc(r *datastore.DocReference) {
	log := idx.log.WithValues("driveId", r.ID, "name", r.Name)

	// TODO(jeremy): This is a bit weird. We are relying on the data in DocReference to be up to date. This is
	// implicitly relying on the fact that when we scan GoogleDrive to find files we create the DocReference. Arguably,
	// we should be updating DocReference as part of ProcessDoc. We should change this so that we update the
	// DocReference as part of ProcessDoc. Basically look at the code in IndexDocument.
	if r.MimeType != DocumentMimeType {
		log.V(logging.Debug).Info("Skipping document; not a Google Document", "driveId", r.DriveId)
		return
	}
	d, err := idx.docsService.Documents.Get(r.DriveId).Do()
	// TODO(jeremy): We probably need to handle the case where a document was deleted or we lost access to it.
	if err != nil {
		log.Error(err, "Failed to get document; Was it deleted?", "driveId", r.ID, "name", r.Name)
		return
	}

	if err := idx.ProcessDocLinks(r, d); err != nil {
		// Keep going to try to degrade gracefully
		log.Error(err, "Failed to process links")
	}

	// If there is an error try to keep going even though this means some data might end up being missed.
	if err := idx.ProcessEntities(r, d); err != nil {
		log.Error(err, "failed to get entities for document", "driveId", r.ID)
	}

	r.Md5Checksum = d.RevisionId
	r.LastIndexedMd5Checksum = d.RevisionId

	if err := idx.store.UpdateDocReference(r); err != nil {
		log.Error(err, "failed to update doc reference", "doc", r)
	}
}

// ProcessDocLinks processes all the docs for the doc referenced by r and represented by d.
func (idx *Indexer) ProcessDocLinks(r *datastore.DocReference, d *docs.Document) error {
	log := idx.log.WithValues("driveId", r.ID, "name", r.Name)
	links, err := GetAllLinks(d)
	if err != nil {
		log.Error(err, "Failed to get document links")
		return nil
	}

	// TODO(jeremy): This will lead to duplicates because each time we reindex the document we will re-add the
	// same links. To fix this I think we want to add some sort of version number to the links. We can then
	// overwrite any existing links updating the version number. We can then delete any links whose version
	// number doesn't equal the new version number. This avoids the need for transactions and ensures we would
	// never be in a state where the is no data for a document that had been previously indexed. There would
	// be a state where we would potentially return links which had since been deleted.
	for _, l := range links {
		g, err := ParseGoogleDocUri(l.Url)

		if err != nil {
			// In the event of an error log it but try to keep going.
			log.Error(err, "Failed to parse GoogleDocURI", "url", l.Url)
		}

		destId := ""

		// TODO(jeremy): Should we verify the ID? Should we do something with the heading?
		if g != nil {
			destId = datastore.DriveKey(g.ID)
		}

		docLink := &datastore.DocLink{
			SourceID:   r.ID,
			DestID:     destId,
			URI:        l.Url,
			Text:       l.Text,
			StartIndex: l.StartIndex,
			EndIndex:   l.EndIndex,
		}

		// If there is an error try to keep going even though this means some data might end up being missed.
		if err := idx.store.UpdateDocLink(docLink); err != nil {
			log.Error(err, "failed to update doclink", "docLink", docLink)
		}
	}
	return nil
}

// ProcessEntities gets all the entities in the document
func (idx *Indexer) ProcessEntities(r *datastore.DocReference, d *docs.Document) error {
	log := idx.log.WithValues("driveId", r.ID, "name", r.Name)
	// TODO(jeremy): This will lead to duplicates because each time we reindex the document we will re-add the
	// same entity mentions. To fix this I think we want to add some sort of version number to the mentions. We can then
	// overwrite any existing mentions updating the version number. We can then delete any mentions whose version
	// number doesn't equal the new version number. This avoids the need for transactions and ensures we would
	// never be in a state where the is no data for a document that had been previously indexed. There would
	// be a state where we would potentially return mentions which had since been deleted.
	//
	// Get the entities in the document

	entities, err := GetEntities(context.Background(), idx.nlpClient, d)
	if err != nil {
		return errors.Wrapf(err, "Failed to get entities")
	}

	// For each entity found in the doc try to resolve it to an entity already in the database.
	// If there isn't one then create a new entry.
	for _, e := range entities {
		var dEntity *datastore.Entity

		q := datastore.EntityQuery{
			Name:         e.GetName(),
			WikipediaURL: glanguage.GetWikipediaURL(*e),
			MID:          glanguage.GetMID(*e),
		}
		entities, err := idx.store.FindEntity(q)

		if err != nil {
			return errors.Wrapf(err, "Failed to find matching entities")
		}

		if len(entities) > 1 {
			log.Info("Found more than one matching entity", "query", q, "numMatched", len(entities))
		}

		if len(entities) == 0 {
			uid, err := uuid.NewUUID()
			if err != nil {
				log.Error(err, "Failed to create UID for entity", "query", q)
				// Try to degrade gracefully and continue processing the other entities
				continue
			}
			dEntity = &datastore.Entity{
				ID:           uid.String(),
				Name:         q.Name,
				Type:         e.Type.String(),
				WikipediaUrl: q.WikipediaURL,
				MID:          q.MID,
			}
			log.Info("Creating Entity", "name", q.Name)

			if err := idx.store.UpdateEntity(dEntity); err != nil {
				log.Error(err, "Failed to add entity to database", "id", dEntity.ID, "name", dEntity.Name)
				// Try to degrade gracefully and continue processing the other entities
				continue
			}
		} else {
			dEntity = entities[0]
		}

		// Now add all the entity mentions to the doc.
		for _, m := range e.Mentions {
			content := m.Text.GetContent()
			dMention := &datastore.EntityMention{
				DocID:      r.ID,
				EntityID:   dEntity.ID,
				Text:       content,
				StartIndex: int64(m.Text.GetBeginOffset()),
				EndIndex:   int64(m.Text.GetBeginOffset()) + int64(len(content)),
			}

			if err := idx.store.UpdateEntityMention(dMention); err != nil {
				log.Error(err, "Failed to add entity mention to database", "id", dEntity.ID, "name", dEntity.Name)
				// Try to degrade gracefully and continue processing the other entities
				continue
			}
		}
	}

	return nil
}
