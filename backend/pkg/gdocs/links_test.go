package gdocs

import (
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/docs/v1"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func Test_GetAllLinks(t *testing.T) {
	wDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory; %v", err)
	}

	testData := filepath.Join(wDir, "test_data")

	type testCase struct {
		fileName string
		expected []*HyperLink
	}

	cases := []testCase{
		{
			fileName: "test_doc.json",
			expected: []*HyperLink{
				{

					Url:        "https://docs.google.com/document/d/1xC1ORtF6imxbFyyng1ABximw-xW67j29UbGAlNlT4KY/edit",
					Text:       "Link to Google Document",
					StartIndex: 51,
					EndIndex:   74,
				},
				{

					Url:        "https://docs.google.com/document/d/1xC1ORtF6imxbFyyng1ABximw-xW67j29UbGAlNlT4KY/edit",
					Text:       "Test Doc2",
					StartIndex: 97,
					EndIndex:   98,
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.fileName, func(t *testing.T) {
			p := filepath.Join(testData, c.fileName)
			b, err := ioutil.ReadFile(p)
			if err != nil {
				t.Fatalf("failed to read file; %v", p)
			}

			doc := &docs.Document{}

			if err := json.Unmarshal(b, doc); err != nil {
				t.Fatalf("failed to unmarshal Document from file; %v; error %v", p, err)
			}

			links, err := GetAllLinks(doc)
			if err != nil {
				t.Fatalf("failed to get links; error %v", err)
			}

			if d := cmp.Diff(c.expected, links); d != "" {
				t.Errorf("Actual links didn't match; diff:\n%v", d)
			}
		})
	}
}
