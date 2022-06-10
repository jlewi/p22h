package datastore

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jlewi/pkg/backend/pkg/logging"
	"io/ioutil"
	"path"
	"sort"
	"testing"
)

func Test_datastore(t *testing.T) {
	type testCase struct {
		name     string
		rows     []*DocReference
		expected []*DocReference
	}

	cases := []testCase{
		{
			name: "basic",
			rows: []*DocReference{
				{
					DriveId:  "someId",
					Name:     "someRow",
					MimeType: "text",
				},
			},
			expected: []*DocReference{
				{
					DriveId:  "someId",
					Name:     "someRow",
					MimeType: "text",
				},
			},
		},
		{
			name: "update",
			rows: []*DocReference{
				{
					DriveId:  "someId",
					Name:     "someRow",
					MimeType: "text",
				},
				{
					DriveId:  "someId",
					Name:     "newName",
					MimeType: "newText",
				},
			},
			expected: []*DocReference{
				{
					DriveId:  "someId",
					Name:     "newName",
					MimeType: "newText",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "testDatabase")
			if err != nil {
				t.Fatalf("Failed to create temporary directory; error %v", err)
			}

			t.Logf("Created temporary directory: %v", dir)

			dbFile := path.Join(dir, "database.db")

			log, _ := logging.InitLogger("debug", true)
			db, err := New(dbFile, *log)

			if err != nil {
				t.Fatalf("Failed to create database; error %v", err)
			}

			for _, r := range c.rows {
				err = db.UpdateDocReference(r)
				if err != nil {
					t.Errorf("Failed to add row %+v; %+v", r, err)
				}
			}
			rows, err := db.ListDocReferences()

			if err != nil {
				t.Errorf("Failed to read rows")
			}

			// TODO(jeremy): Really need to sort the rows for comparison when there are multiple rows
			if d := cmp.Diff(c.expected, rows, GormIgnored(DocReference{})); d != "" {
				t.Errorf("Read rows didn't match; diff:\n%v", d)
			}

			if err := db.Close(); err != nil {
				t.Errorf("Failed to close database; error %+v", err)
			}

		})
	}
}

func Test_ToBeIndexed(t *testing.T) {
	type testCase struct {
		name     string
		rows     []*DocReference
		expected []string
	}

	cases := []testCase{
		{
			name: "basic",
			rows: []*DocReference{
				{
					DriveId:                "doesntNeedIndexing",
					Name:                   "doesntNeedIndexing",
					MimeType:               "text",
					Md5Checksum:            "1234",
					LastIndexedMd5Checksum: "1234",
				},
				{
					DriveId:                "neverindexed",
					Name:                   "neverindexed",
					MimeType:               "text",
					Md5Checksum:            "",
					LastIndexedMd5Checksum: "",
				},
				{
					DriveId:                "outdated",
					Name:                   "outdated",
					MimeType:               "text",
					Md5Checksum:            "new",
					LastIndexedMd5Checksum: "old",
				},
			},
			expected: []string{
				"neverindexed", "outdated",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "testDatabase")
			if err != nil {
				t.Fatalf("Failed to create temporary directory; error %v", err)
			}

			t.Logf("Created temporary directory: %v", dir)

			dbFile := path.Join(dir, "database.db")

			log, _ := logging.InitLogger("debug", true)
			db, err := New(dbFile, *log)

			if err != nil {
				t.Fatalf("Failed to create database; error %v", err)
			}

			for _, r := range c.rows {
				err = db.UpdateDocReference(r)
				if err != nil {
					t.Errorf("Failed to add row %+v; %+v", r, err)
				}
			}
			rows, err := db.ToBeIndexed()

			if err != nil {
				t.Errorf("Failed to read rows")
			}

			actual := make([]string, 0, len(rows))
			for _, r := range rows {
				actual = append(actual, r.DriveId)
			}

			sort.Strings(actual)
			sort.Strings(c.expected)
			if d := cmp.Diff(c.expected, actual); d != "" {
				t.Errorf("Read rows didn't match; diff:\n%v", d)
			}

			if err := db.Close(); err != nil {
				t.Errorf("Failed to close database; error %+v", err)
			}

		})
	}
}

func Test_doclinks(t *testing.T) {
	// Some docs to populate the database with
	docs := []*DocReference{
		{
			DriveId:  "doc1",
			Name:     "someRow",
			MimeType: "text",
		},
		{
			DriveId:  "doc2",
			Name:     "someRow",
			MimeType: "text",
		},
		{
			DriveId:  "doc3",
			Name:     "someRow",
			MimeType: "text",
		},
	}

	type testCase struct {
		name     string
		destId   string
		links    []*DocLink
		expected []*DocLink
	}

	cases := []testCase{
		{
			name: "basic",
			links: []*DocLink{
				{
					SourceID: "doc1",
					DestID:   "doc2",
				},
			},
			expected: []*DocLink{
				{
					SourceID: "doc1",
					DestID:   "doc2",
				},
			},
		},
		{
			name: "empty-dest",
			links: []*DocLink{
				{
					SourceID: "doc1",
					DestID:   "",
				},
			},
			expected: []*DocLink{
				{
					SourceID: "doc1",
					DestID:   "",
				},
			},
		},
		{
			// Verify inserting the same link twice doesn't give duplicates
			name: "update-no-duplicates",
			links: []*DocLink{
				{
					SourceID: "doc1",
					DestID:   "doc2",
				},
				{
					SourceID: "doc1",
					DestID:   "doc2",
				},
			},
			expected: []*DocLink{
				{
					SourceID: "doc1",
					DestID:   "doc2",
				},
			},
		},
		{
			// Verify start and end index are taken into account so we can have multiple links between the same
			// set of documents
			name: "multiple-links",
			links: []*DocLink{
				{
					SourceID:   "doc1",
					DestID:     "doc2",
					StartIndex: 10,
					EndIndex:   20,
				},
				{
					SourceID:   "doc1",
					DestID:     "doc2",
					StartIndex: 30,
					EndIndex:   40,
				},
			},
			expected: []*DocLink{
				{
					SourceID:   "doc1",
					DestID:     "doc2",
					StartIndex: 10,
					EndIndex:   20,
				},
				{
					SourceID:   "doc1",
					DestID:     "doc2",
					StartIndex: 30,
					EndIndex:   40,
				},
			},
		},
		{
			name:   "query-by-dest-id",
			destId: "doc2",
			links: []*DocLink{
				{
					SourceID: "doc1",
					DestID:   "doc2",
				},
				{
					SourceID: "doc1",
					DestID:   "doc3",
				},
			},
			expected: []*DocLink{
				{
					SourceID: "doc1",
					DestID:   "doc2",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "testDatabase")
			if err != nil {
				t.Fatalf("Failed to create temporary directory; error %v", err)
			}

			t.Logf("Created temporary directory: %v", dir)

			dbFile := path.Join(dir, "database.db")

			log, _ := logging.InitLogger("debug", true)
			db, err := New(dbFile, *log)

			if err != nil {
				t.Fatalf("Failed to create database; error %v", err)
			}

			// Preload the docs
			for _, r := range docs {
				// Make a copy of the doc because UpdateDocReference can modify it.
				dCopy := *r
				err = db.UpdateDocReference(&dCopy)
				if err != nil {
					t.Fatalf("Failed to add row %+v; %+v", r, err)
				}
			}

			// Update the doc links
			for _, l := range c.links {
				err = db.UpdateDocLink(l)
				if err != nil {
					t.Fatalf("Failed to add link %+v; %+v", l, err)
				}
			}
			links, err := db.ListDocLinks(c.destId)

			if err != nil {
				t.Errorf("Failed to read links")
			}

			// TODO(jeremy): Really need to sort the rows for comparison when there are multiple rows
			opts := cmpopts.IgnoreFields(DocLink{}, "ID", "CreatedAt", "DeletedAt", "UpdatedAt")
			if d := cmp.Diff(c.expected, links, opts); d != "" {
				t.Errorf("Read links didn't match; diff:\n%v", d)
			}

			if err := db.Close(); err != nil {
				t.Errorf("Failed to close database; error %+v", err)
			}
		})
	}
}
