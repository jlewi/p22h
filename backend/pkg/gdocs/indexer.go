package gdocs

import (
	"github.com/go-logr/logr"
	"github.com/jlewi/pkg/backend/pkg/datastore"
	"github.com/jlewi/pkg/backend/pkg/logging"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
)

// Indexer indexes Google Drive.
type Indexer struct {
	log      logr.Logger
	store    *datastore.Datastore
	searcher DriveSearch

	docsService *docs.Service
}

// NewIndexer creates a new indexer
func NewIndexer(searcher DriveSearch, docsService *docs.Service, store *datastore.Datastore, log logr.Logger) (*Indexer, error) {
	if searcher == nil {
		return nil, errors.New("client is required")
	}

	if store == nil {
		return nil, errors.New("store is required")
	}

	if docsService == nil {
		return nil, errors.New("docsService is required")
	}

	return &Indexer{
		log:         log,
		store:       store,
		searcher:    searcher,
		docsService: docsService,
	}, nil
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
		if r.MimeType != DocumentMimeType {
			log.V(logging.Debug).Info("Skipping document; not a Google Document", "driveId", r.DriveId)
			continue
		}
		d, err := idx.docsService.Documents.Get(r.DriveId).Do()
		// TODO(jeremy): We probably need to handle the case where a document was deleted or we lost access to it.
		if err != nil {
			log.Error(err, "Failed to get document; Was it deleted?", "driveId", r.ID, "name", r.Name)
			continue
		}

		links, err := GetAllLinks(d)
		if err != nil {
			log.Error(err, "Failed to get document links", "driveId", r.ID, "name", r.Name)
			continue
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

		r.Md5Checksum = d.RevisionId
		r.LastIndexedMd5Checksum = d.RevisionId

		if err := idx.store.UpdateDocReference(r); err != nil {
			log.Error(err, "failed to update doc reference", "doc", r)
		}

	}

	return nil
}
