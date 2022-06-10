package gdocs

import (
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"google.golang.org/api/drive/v3"
	"net/http"
)

// Client is a high level client for interacting with gdrive
type Client struct {
	c   *http.Client
	log logr.Logger
}

func NewClient(c *http.Client, log logr.Logger) (*Client, error) {
	if c == nil {
		return nil, errors.New("client can't be nil")
	}

	return &Client{
		c:   c,
		log: log,
	}, nil
}

// ResultFunc is invoked by search to process each result
// A non nil error causes result processing to stop.
type ResultFunc func(file *drive.File) error

type DriveSearch interface {
	Search(query string, driveId string, corpora string, resultFunc ResultFunc) error
}

// Search runs the provided search query.
func (c *Client) Search(query string, driveId string, corpora string, resultFunc ResultFunc) error {
	log := c.log
	svc, err := drive.New(c.c)

	if err != nil {
		return errors.Wrapf(err, "Failed to create drive client")
	}

	l := svc.Files.List()

	if driveId != "" {
		l.DriveId(driveId)
		// I think this needs to be true to include items from team drives
		// TODO(jeremy) Why would we set SupportsAllDrives when searching a particular drive?
		l.IncludeItemsFromAllDrives(true)
		l.SupportsAllDrives(true)
	}

	if corpora != "" {
		l.Corpora(corpora)
	}

	l.Q(query)
	pageToken := ""
	pageSize := int64(100)
	for {
		if pageToken != "" {
			l.PageToken(pageToken)
		}
		r, err := l.PageSize(pageSize).Fields("nextPageToken, files(id, name, mimeType, md5Checksum, size)").Do()

		if err != nil {
			return errors.Wrapf(err, "Failed to fetch results from drive")
		}

		for _, f := range r.Files {
			log.V(1).Info("Got file", "Name", f.Name, "ID", f.Id)
			if rErr := resultFunc(f); rErr != nil {
				log.Error(err, "resultFunc returned error", "Name", f.Name, "ID", f.Id)
				return rErr
			}
		}

		if r.NextPageToken == "" {
			break
		}
		pageToken = r.NextPageToken
	}
	return nil
}

// QueryStats contains statistics about the results of a search query
type QueryStats struct {
	Count int64
	// Size in bytes
	Size   int64
	ByType map[string]float64
}

// NewStatsBuilder returns a ResultFunc that will aggregate statistics.
func NewStatsBuilder(s *QueryStats) (ResultFunc, error) {
	if s == nil {
		return nil, errors.New("s can't be nil")
	}

	s.Count = 0
	s.Size = 0
	s.ByType = map[string]float64{}

	return func(f *drive.File) error {
		s.Count = s.Count + 1
		s.Size = s.Size + f.Size
		if _, ok := s.ByType[f.MimeType]; !ok {
			s.ByType[f.MimeType] = 0
		}

		s.ByType[f.MimeType] = s.ByType[f.MimeType] + 1

		return nil
	}, nil
}
