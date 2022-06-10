package gdocs

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func Test_ParseGoogleDocUri(t *testing.T) {
	type testCase struct {
		Name     string
		In       string
		Expected *GoogleDocUri
	}

	cases := []testCase{
		{
			Name: "basic",
			In:   "https://docs.google.com/document/d/1qPd2W0jgD/edit",
			Expected: &GoogleDocUri{
				ID:      "1qPd2W0jgD",
				Heading: "",
			},
		},
		{
			Name: "heading",
			In:   "https://docs.google.com/document/d/1qPd2W0jgD/edit#heading=h.75b5l",
			Expected: &GoogleDocUri{
				ID:      "1qPd2W0jgD",
				Heading: "h.75b5l",
			},
		},

		{
			Name: "headingQuery",
			In:   "https://docs.google.com/document/d/1qPd2W0jgD/edit#heading=h.75b5l?arg1=2",
			Expected: &GoogleDocUri{
				ID:      "1qPd2W0jgD",
				Heading: "h.75b5l",
			},
		},
		{
			Name:     "notadoc",
			In:       "https://some/other/url",
			Expected: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			a, err := ParseGoogleDocUri(c.In)

			if err != nil {
				t.Fatalf("Failed to parse: %v; error %v", c.In, err)
			}

			if d := cmp.Diff(c.Expected, a); d != "" {
				t.Errorf("Didn't get expected result; diff:\n%v", d)
			}
		})
	}
}
