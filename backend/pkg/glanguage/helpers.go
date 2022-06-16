package glanguage

import languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"

const (
	// Metadata keys in the entity entry
	// https://cloud.google.com/natural-language/docs/reference/rest/v1/Entity
	WikipediaKey = "wikipedia_url"
	MIDKey       = "mid"
)

// GetWikipediaURL gets the wikipedia url from the entity if there is one otherwise returns the empty string.
func GetWikipediaURL(e languagepb.Entity) string {
	v := e.GetMetadata()[WikipediaKey]
	return v
}

// GetMID gets the mid (knowledge graph id ) from the entity if there is one otherwise returns the empty string.
func GetMID(e languagepb.Entity) string {
	v := e.GetMetadata()[MIDKey]
	return v
}
