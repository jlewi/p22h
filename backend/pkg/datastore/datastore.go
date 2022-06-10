package datastore

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/jlewi/pkg/backend/pkg/logging"
	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	driveNamespace = "gdrive"
)

type Datastore struct {
	log    logr.Logger
	dbFile string
	db     *gorm.DB
}

// DriveKey generates the primary key for the given Google Drive file
// WARNING: Changing this code will break existing databases since existing data will not have
// compatible keys.
func DriveKey(id string) string {
	return driveNamespace + "." + id
}

// DocLinkKey generates the primary key for the given DocLink.
// WARNING: Changing this code will break existing databases since existing data will not have
// compatible keys.
func DocLinkKey(link DocLink) string {
	return fmt.Sprintf("%v-%v-%v-%v", link.SourceID, link.DestID, link.StartIndex, link.EndIndex)
}

// New creates a new datastore.
func New(dbFile string, logger logr.Logger) (*Datastore, error) {
	if dbFile == "" {
		return nil, errors.New("dbFile is required")
	}

	db := &Datastore{
		log:    logger,
		dbFile: dbFile,
	}

	logger.Info("Opening database", "dbFile", dbFile)
	sqlDb, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open database: %v", dbFile)
	}

	db.db = sqlDb

	err = db.updateSchema()
	if err != nil {
		return nil, errors.Wrapf(err, "CreateTables failed")
	}

	return db, nil
}

// updateSchema automatically updates the schema
func (d *Datastore) updateSchema() error {
	log := d.log
	log.Info("Automigrating the schema")
	// Migrate the schema
	if err := d.db.AutoMigrate(&DocReference{}); err != nil {
		return errors.Wrapf(err, "Failed to automigrate the schema for DocReference")
	}
	if err := d.db.AutoMigrate(&DocLink{}); err != nil {
		return errors.Wrapf(err, "Failed to automigrate the schema for DocLink")
	}
	return nil
}

// UpdateDocReference updates or creates the DocReference
func (d *Datastore) UpdateDocReference(r *DocReference) error {
	if r.ID != "" && r.ID != DriveKey(r.DriveId) {
		return errors.Errorf("ID and DriveID are inconsistent ID should be empty or %v", DriveKey(r.DriveId))
	}

	r.ID = DriveKey(r.DriveId)

	log := d.log.WithValues("DriveId", r.DriveId)
	db := d.db
	current := &DocReference{
		ID: DriveKey(r.DriveId),
	}
	result := db.First(current)

	if result.RowsAffected == 0 {
		log.V(logging.Debug).Info("Record not found; it will be created")
	} else {
		log.V(logging.Debug).Info("Record found", "id", current.ID)
		r.ID = current.ID
	}

	log.V(logging.Debug).Info("Updating record")
	if result := db.Save(r); result.Error != nil {
		return errors.Wrapf(result.Error, "Failed to update DocReference DriveId: %v", r.DriveId)
	}

	return nil
}

// DocReferenceIter is an iterator over DocReferences
type DocReferenceIter func(r *DocReference) error

// ListDocReferences lists all the docreferences.
func (d *Datastore) ListDocReferences() ([]*DocReference, error) {
	db := d.db
	// TODO(jeremy): Should we introduce some form of pagination? See
	// https://gorm.io/docs/scopes.html#pagination
	references := make([]*DocReference, 0, 0)
	if result := db.Find(&references); result.Error != nil {
		return nil, errors.Wrapf(result.Error, "Failed to find all doc references")
	}

	return references, nil
}

// UpdateDocLink updates or creates the DocLink
//
// TODO(jeremy): These function needs to be updated to allow for a given link to appear multiple times in a doc.
// In that case we want to have multiple entries in the doc.
func (d *Datastore) UpdateDocLink(l *DocLink) error {
	// DestID isn't required because not all links point ot Google Docs.
	if l.SourceID == "" {
		return errors.New("SourceID must be set")
	}

	// TODO(jeremy): Should we be using hooks https://gorm.io/docs/hooks.html to achieve this
	expectedId := DocLinkKey(*l)
	if l.ID != "" && l.ID != expectedId {
		return errors.Errorf("ID and DockLink are inconsistent ID should be empty or %v", expectedId)
	}

	l.ID = expectedId

	log := d.log.WithValues("source", l.SourceID, "dest", l.DestID, "id", l.ID)
	db := d.db

	current := &DocLink{
		ID: l.ID,
	}
	result := db.First(current)

	if result.RowsAffected == 0 {
		log.V(logging.Debug).Info("Record not found; it will be created")
	} else {
		log.V(logging.Debug).Info("Record found", "id", current.ID)
		l.ID = current.ID
	}

	log.V(logging.Debug).Info("Updating record")
	if result := db.Save(l); result.Error != nil {
		return errors.Wrapf(result.Error, "Failed to update Doclink Source: %v, Dest: %v", l.SourceID, l.DestID)
	}

	return nil
}

// ListDocLinks lists all the doc links.
// destId optional if supplied list all the links pointing at this destination id
func (d *Datastore) ListDocLinks(destId string) ([]*DocLink, error) {
	db := d.db
	// TODO(jeremy): Should we introduce some form of pagination? See
	// https://gorm.io/docs/scopes.html#pagination
	links := make([]*DocLink, 0, 0)

	if destId != "" {
		db = db.Where("dest_id = ? ", destId)
	}

	if result := db.Find(&links); result.Error != nil {
		return nil, errors.Wrapf(result.Error, "Failed to find all doc links")
	}

	return links, nil
}

// ToBeIndexed returns a list of DocReferences that need to be indexed.
func (d *Datastore) ToBeIndexed() ([]*DocReference, error) {
	// Find all documents for which the current sha and last indexed sha don't match; and/or
	// last indexed sha is empty
	db := d.db
	references := make([]*DocReference, 0, 0)

	if result := db.Where("last_indexed_md5_checksum = '' or md5_checksum != last_indexed_md5_checksum").Find(&references); result.Error != nil {
		return nil, errors.Wrapf(result.Error, "Failed to find docs needing update")
	}

	return references, nil
}

func (d *Datastore) Close() error {
	// Currently a null op.
	return nil
}
