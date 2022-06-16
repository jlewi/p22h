package datastore

import (
	"gorm.io/gorm"
	"time"
)

// DocReference is a reference to a document stored in some system such as Google Drive.
type DocReference struct {
	// The unique id follows the convention $namespace.id where namespace identifies a namespace with respect
	// to which id is unique. Typically namespace is the source of the file e.g. Google Drive.
	ID        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// The ID of the file in Google Drive. We create a unique index named uid to ensure there is
	// one row for each doc. This could eventually become a composite index because we want to allow for
	// documents in different systems (e.g. Drive and GitHub). In which case the uid index would be a composite
	// key on DriveId and GitHub and only one will be set.
	DriveId  string `gorm:"index:uid,unique"`
	Name     string
	MimeType string

	// TODO(jeremy): We should rename the checksum fields. To be opaque version numbers. They won't always be
	// checksums.
	// Md5Checksum is current checksum
	Md5Checksum string

	// LastIndexedMd5Checksum is the checksum at which it was last indexed
	LastIndexedMd5Checksum string
}

// DocLink is a directional link between two docs.
// We do not rely on GormAssociations for a couple reasons
// 1. We want to stick with a CRUD API to allow for more flexible backends
// 2. Using associations adds complexity in terms of how it gets used
//    I believe in order to populate associations its doing joins
//    There's also confusion on which fields user should set to update when using a BelongsTo association
//    as there are separate fields for the foreign key and the reference.
// 3. GraphQL might be a better API for joins.
//
// A link can appear more than once between two documents; i.e. a given doc can have multiple hyperlinks to another
// document.
// Not all links will have destId set.
type DocLink struct {
	// The unique id follows the convention sourceId-destId-startIndex-endindex.
	// This is arguably not space efficient but we can optimize later.
	ID        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	// SourceID is the id of the destination doc
	SourceID string `gorm:"index"`
	// DestID is the destination doc
	DestID string `gorm:"index"`
	// URI is the URI the link is pointing to
	URI string
	// Text is the text associated with the link.
	Text string
	// StartIndex of the text for the link.
	StartIndex int64
	// EndIndex of the text for the link.
	EndIndex int64
}

// EntityMention is the mention of some entity in a doc.
//
// We do not rely on GormAssociations for a couple reasons
// 1. We want to stick with a CRUD API to allow for more flexible backends
// 2. Using associations adds complexity in terms of how it gets used
//    I believe in order to populate associations its doing joins
//    There's also confusion on which fields user should set to update when using a BelongsTo association
//    as there are separate fields for the foreign key and the reference.
// 3. GraphQL might be a better API for joins.
//
// A specific entity can appear more than once in a given doc.
//
// TODO(jeremy): We also need an Entity table and should attempt to do some entity linking.
type EntityMention struct {
	// The unique id follows the convention docId-startIndex-endindex.
	// Assumption is a given range can only be a single entity.
	ID        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	// DocID is the id of the doc
	DocID    string `gorm:"index"`
	EntityID string
	// Text associated with the entity
	Text string
	// StartIndex of the text for the link.
	StartIndex int64
	// EndIndex of the text for the link.
	EndIndex int64
}

// Entity is a unique entity.
type Entity struct {
	ID        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Name is the canonical name of the entity
	Name string

	// Type of entity
	Type string

	// WikipediaURL associated with this entity if there is one.
	WikipediaUrl string

	// MID is the Google Knowledge Graph MID if there is one
	MID string `gorm:"column:mid"`
}
