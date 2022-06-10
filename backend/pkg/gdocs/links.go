package gdocs

import (
	"github.com/pkg/errors"
	"google.golang.org/api/docs/v1"
)

type HyperLink struct {
	Url        string
	Text       string
	StartIndex int64
	EndIndex   int64
}

// GetAllLinks gets all the links from the document.
func GetAllLinks(doc *docs.Document) ([]*HyperLink, error) {
	if doc == nil {
		return []*HyperLink{}, errors.New("doc is a required argument")
	}

	if doc.Body == nil {
		return []*HyperLink{}, nil
	}

	links := readElementLinks(doc.Body.Content)
	return links, nil
}

// readElements extracts all the text from the elements and concatenates it together.
func readElementLinks(elements []*docs.StructuralElement) []*HyperLink {
	links := make([]*HyperLink, 0, 10)
	for _, e := range elements {
		if e.Paragraph != nil {
			links = append(links, readParagraphLinks(e.Paragraph)...)
		}

		if e.Table != nil {
			links = append(links, readTableLinks(e.Table)...)
		}
	}
	return links
}

// readParagraphLinks reads all the text in the paragraph
func readParagraphLinks(p *docs.Paragraph) []*HyperLink {
	links := make([]*HyperLink, 0, 10)
	for _, e := range p.Elements {
		if l := getLinkFromRichLink(e.RichLink); l != nil {
			l.StartIndex = e.StartIndex
			l.EndIndex = e.EndIndex
			links = append(links, l)
		}
		if l := getLinkFromTextRun(e.TextRun); l != nil {
			l.StartIndex = e.StartIndex
			l.EndIndex = e.EndIndex
			links = append(links, l)
		}
	}
	return links
}

func readTableLinks(t *docs.Table) []*HyperLink {
	links := make([]*HyperLink, 0, 10)
	for _, r := range t.TableRows {
		for _, cell := range r.TableCells {
			newLinks := readElementLinks(cell.Content)
			links = append(links, newLinks...)
		}
	}
	return links
}

func getLinkFromTextRun(t *docs.TextRun) *HyperLink {
	if t == nil {
		return nil
	}
	if t.TextStyle == nil {
		return nil
	}
	if t.TextStyle.Link == nil {
		return nil
	}
	// If URL is blank then its a link within the document which we ignore.
	if t.TextStyle.Link.Url == "" {
		return nil
	}
	return &HyperLink{
		Url:  t.TextStyle.Link.Url,
		Text: t.Content,
	}
}

func getLinkFromRichLink(r *docs.RichLink) *HyperLink {
	if r == nil {
		return nil
	}

	if r.RichLinkProperties == nil {
		return nil
	}

	return &HyperLink{
		Url:  r.RichLinkProperties.Uri,
		Text: r.RichLinkProperties.Title,
	}
}
