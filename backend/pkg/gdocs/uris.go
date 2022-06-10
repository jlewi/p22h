package gdocs

import (
	"github.com/pkg/errors"
	"net/url"
	"regexp"
	"strings"
)

const (
	GoogleDocsHost = "docs.google.com"
)

var (
	headingRe *regexp.Regexp
)

type GoogleDocUri struct {
	ID      string
	Heading string
}

// ParseGoogleDocUri parses a google document URI
// Return nil if not a googledocument.
func ParseGoogleDocUri(u string) (*GoogleDocUri, error) {
	p, err := url.Parse(u)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse URL; %v", u)
	}

	if p.Host != GoogleDocsHost {
		return nil, nil
	}

	pieces := strings.Split(p.Path, "/")

	if len(pieces) < 3 {
		return nil, errors.Wrapf(err, "u doesn't have path documents/d/{driveId}...")
	}

	if pieces[1] != "documents" && pieces[2] != "d" {
		return nil, errors.Wrapf(err, "u doesn't have path documents/d/{driveId}...")
	}

	gUri := &GoogleDocUri{
		ID: pieces[3],
	}

	if p.Fragment != "" {
		if headingRe == nil {
			headingRe, err = regexp.Compile(`.*heading=(?P<heading>h\.[0-9a-zA-Z]+)`)
			//headingRe, err = regexp.Compile(`.*`)
			if err != nil {
				return gUri, errors.Wrapf(err, "Failed to compile regex ")
			}
		}

		matches := headingRe.FindStringSubmatch(p.Fragment)
		if len(matches) > 0 {
			gUri.Heading = matches[1]
		}
	}
	return gUri, nil
}
