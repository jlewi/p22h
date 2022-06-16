package datastore

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/jlewi/p22h/backend/pkg/logging"
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

// EntityMentionKey generates the primary key for the given EntityMention.
// WARNING: Changing this code will break existing databases since existing data will not have
// compatible keys.
func EntityMentionKey(m EntityMention) string {
	return fmt.Sprintf("%v-%v-%v-%v", m.DocID, m.EntityID, m.StartIndex, m.EndIndex)
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
	if err := d.db.AutoMigrate(&Entity{}); err != nil {
		return errors.Wrapf(err, "Failed to automigrate the schema for Entity")
	}
	if err := d.db.AutoMigrate(&EntityMention{}); err != nil {
		return errors.Wrapf(err, "Failed to automigrate the schema for EntityMention")
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

// UpdateEntity updates or creates the Entity
//
// TODO(jeremy): The semantics for dealing with multiple entities with the same name are ill defined. Right now
// it is the caller's job to do entity linking before calling UpdateEntity. If an entity with a given name already
// exists in the database but m represents a different entity with the same name then caller should assign a unique
// id to it.
func (d *Datastore) UpdateEntity(m *Entity) error {
	if m.Name == "" {
		return fmt.Errorf("Name is required")
	}

	// Default ID to Name
	if m.ID == "" {
		m.ID = m.Name
	}

	log := d.log.WithValues("id", m.ID)
	db := d.db

	current := &Entity{
		ID: m.ID,
	}
	result := db.First(current)

	if result.RowsAffected == 0 {
		log.V(logging.Debug).Info("Record not found; it will be created")
	} else {
		log.V(logging.Debug).Info("Record found", "id", current.ID)
		m.ID = current.ID
	}

	log.V(logging.Debug).Info("Updating record")
	if result := db.Save(m); result.Error != nil {
		return errors.Wrapf(result.Error, "Failed to update Entity ID: %v", m.ID)
	}

	return nil
}

// ListEntities lists all the entities.
func (d *Datastore) ListEntities() ([]*Entity, error) {
	db := d.db
	// TODO(jeremy): Should we introduce some form of pagination? See
	// https://gorm.io/docs/scopes.html#pagination
	entities := make([]*Entity, 0, 0)

	if result := db.Find(&entities); result.Error != nil {
		return nil, errors.Wrapf(result.Error, "Failed to find all entities")
	}

	return entities, nil
}

type EntityQuery struct {
	Name         string
	WikipediaURL string
	MID          string
}

// FindEntities is a primitive form of entity linking.
//
// TODO(jeremy): How should we handle the case where we could potentially have multiple entries in the database
// that would match?
func (d *Datastore) FindEntity(q EntityQuery) ([]*Entity, error) {
	db := d.db

	entities := make([]*Entity, 0, 0)

	if q.Name != "" {
		db = db.Or("name = ?", q.Name)
	}

	if q.WikipediaURL != "" {
		db = db.Or("wikipedia_url = ?", q.WikipediaURL)
	}

	if q.MID != "" {
		db = db.Or("mid = ?", q.MID)
	}

	if result := db.Find(&entities); result.Error != nil {
		return nil, errors.Wrapf(result.Error, "Failed to find all entitymentions")
	}

	return entities, nil
}

// UpdateEntityMention updates or creates the EntityMention
//
func (d *Datastore) UpdateEntityMention(m *EntityMention) error {
	if m.DocID == "" {
		return errors.New("DocID must be set")
	}

	// TODO(jeremy): Should we be using hooks https://gorm.io/docs/hooks.html to achieve this
	expectedId := EntityMentionKey(*m)
	if m.ID != "" && m.ID != expectedId {
		return errors.Errorf("ID and EntityMention are inconsistent; ID should be empty or %v", expectedId)
	}

	m.ID = expectedId

	log := d.log.WithValues("docId", m.DocID, "id", m.ID)
	db := d.db

	current := &EntityMention{
		ID: m.ID,
	}
	result := db.First(current)

	if result.RowsAffected == 0 {
		log.V(logging.Debug).Info("Record not found; it will be created")
	} else {
		log.V(logging.Debug).Info("Record found", "id", current.ID)
		m.ID = current.ID
	}

	log.V(logging.Debug).Info("Updating record")
	if result := db.Save(m); result.Error != nil {
		return errors.Wrapf(result.Error, "Failed to update EntityMention ID: %v", m.ID)
	}

	return nil
}

// ListEntityMentions lists all the entity mentions.
// docId is optional if supplied list all the mentions for the provided doc.
func (d *Datastore) ListEntityMentions(docId string) ([]*EntityMention, error) {
	db := d.db
	// TODO(jeremy): Should we introduce some form of pagination? See
	// https://gorm.io/docs/scopes.html#pagination
	links := make([]*EntityMention, 0, 0)

	if docId != "" {
		db = db.Where("doc_id = ? ", docId)
	}

	if result := db.Find(&links); result.Error != nil {
		return nil, errors.Wrapf(result.Error, "Failed to find all entitymentions")
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
