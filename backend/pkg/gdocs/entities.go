package gdocs

import (
	language "cloud.google.com/go/language/apiv1"
	"context"
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

	return resp.GetEntities(), nil
}
