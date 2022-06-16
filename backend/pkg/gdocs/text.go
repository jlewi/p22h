package gdocs

import (
	"github.com/pkg/errors"
	"google.golang.org/api/docs/v1"
)

// ReadText reads all the text from the provided document.
// It is based on https://developers.google.com/docs/api/samples/extract-text#python.
//
// TODO(https://github.com/jlewi/p22h/issues/1): Linearize text so as to preserve positioning.
func ReadText(doc *docs.Document) (string, error) {
	if doc == nil {
		return "", errors.New("doc is a required argument")
	}

	if doc.Body == nil {
		return "", nil
	}

	return readElements(doc.Body.Content), nil
}

// readElements extracts all the text from the elements and concatenates it together.
func readElements(elements []*docs.StructuralElement) string {
	c := ""
	for _, e := range elements {
		if e.Paragraph != nil {
			c = c + readParagraph(e.Paragraph)
		}

		if e.Table != nil {
			c = c + readTable(e.Table)
		}

		if e.TableOfContents != nil {
			// Recursively read the table of contents text
			c = c + readElements(e.TableOfContents.Content)
		}
	}
	return c
}

// readParagraph reads all the text in the paragraph
func readParagraph(p *docs.Paragraph) string {
	c := ""
	for _, e := range p.Elements {
		if e.TextRun == nil {
			continue
		}
		c = c + e.TextRun.Content
	}
	return c
}

func readTable(t *docs.Table) string {
	c := ""
	for _, r := range t.TableRows {
		for _, cell := range r.TableCells {
			c = c + readElements(cell.Content)
		}
	}
	return c
}
