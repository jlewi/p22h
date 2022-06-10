package gdocs

import "google.golang.org/api/drive/v3"

// FakeSearch implements the search interface for an in memory set of drive documents.
// FakeSearch is intended for testing.
type FakeSearch struct {
	Docs []*drive.File
}

func (f *FakeSearch) Search(query string, driveId string, corpora string, resultFunc ResultFunc) error {
	for _, d := range f.Docs {
		if err := resultFunc(d); err != nil {
			return err
		}
	}
	return nil
}
