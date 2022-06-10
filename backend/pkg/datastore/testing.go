package datastore

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// GormIgnored returns IgnoreFields for the given type
// e.g GormIgnored(DocReference{})
func GormIgnored(typ interface{}) cmp.Option {
	return cmpopts.IgnoreFields(typ, "ID", "CreatedAt", "DeletedAt", "UpdatedAt")
}
