package gdocs

import (
	language "cloud.google.com/go/language/apiv1"
	"context"
	"github.com/jlewi/p22h/backend/pkg/glanguage"
	"github.com/pkg/errors"
	"google.golang.org/api/docs/v1"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
)

// GetEntities gets the entities from the document.
//
// N.B. The current implementation doesn't keep track of
func GetEntities(ctx context.Context, client *language.Client, doc *docs.Document) ([]*languagepb.Entity, error) {
	text, err := ReadText(doc)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read text from documment")
	}

	// TODO(jeremy): Retries?
	resp, err := client.AnalyzeEntities(ctx, &languagepb.AnalyzeEntitiesRequest{
		Document: &languagepb.Document{
			Source: &languagepb.Document_Content{
				Content: text,
			},
			Type: languagepb.Document_PLAIN_TEXT,
		},
		EncodingType: languagepb.EncodingType_UTF8,
	})

	if err != nil {
		return nil, errors.Wrapf(err, "Failed to call NLP API.")
	}

	return newEntityCandidates(resp.GetEntities()), nil
}

// newEntityCandidates finds those entities which could represent new entities that we aren't already aware of.
// The idea is that we want to use more stringent criterion when discovering new entities versus mentions of existing
// entities that we are aware of.
//
// As noted in https://github.com/jlewi/p22h/issues/4 this is an attempt to improve precision.
//
// TODO(jeremy): We'd really like to join the information returned by the NLP API with formatting information
// and use the joint information to render a decision. For example, we'd like to see if a mention contains one or
// more hyperlinks.
func newEntityCandidates(entities []*languagepb.Entity) []*languagepb.Entity {
	cleaned := make([]*languagepb.Entity, 0, len(entities))

	for _, e := range entities {
		// The following types of entities are not ones we consider "things" and want to track.
		switch e.GetType() {
		case languagepb.Entity_ADDRESS:
		case languagepb.Entity_DATE:
		case languagepb.Entity_NUMBER:
		case languagepb.Entity_PHONE_NUMBER:
		case languagepb.Entity_PRICE:
		case languagepb.Entity_LOCATION:
		case languagepb.Entity_WORK_OF_ART:
		// Other returns a lot of spammy organizations.
		case languagepb.Entity_OTHER:
			continue
		}

		// To be included as a possible new entity one of two things must be true.
		// 1. NL API linked it to an existing entry in its Knowledge graph
		// 2. It is a proper noun as opposed to common noun.
		wikipediaURL := glanguage.GetWikipediaURL(*e)
		mid := glanguage.GetMID(*e)

		if wikipediaURL != "" || mid != "" {
			cleaned = append(cleaned, e)
			continue
		}

		func() {
			for _, m := range e.Mentions {
				if m.GetType() == languagepb.EntityMention_PROPER {
					cleaned = append(cleaned, e)
					return
				}
			}
		}()
	}

	return cleaned
}
