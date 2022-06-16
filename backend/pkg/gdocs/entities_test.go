package gdocs

import (
	language "cloud.google.com/go/language/apiv1"
	"context"
	"encoding/json"
	"flag"
	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jlewi/p22h/backend/pkg/glanguage"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func Test_GetEntities(t *testing.T) {
	wDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory; %v", err)
	}

	testData := filepath.Join(wDir, "test_data")

	type testCase struct {
		fileName string
		response *languagepb.AnalyzeEntitiesResponse
		expected []*languagepb.Entity
	}

	cases := []testCase{
		{
			fileName: "test_doc.json",
			response: &languagepb.AnalyzeEntitiesResponse{
				Entities: []*languagepb.Entity{
					{
						Name: "john",
						Type: languagepb.Entity_PERSON,
					},
				},
			},
			expected: []*languagepb.Entity{
				{
					Name: "john",
					Type: languagepb.Entity_PERSON,
				},
			},
		},
	}

	lClient, err := language.NewClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
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

			mockLanguage.Resps = []proto.Message{c.response}

			links, err := GetEntities(context.Background(), lClient, doc)
			if err != nil {
				t.Fatalf("failed to get links; error %v", err)
			}

			if d := cmp.Diff(c.expected, links, EntityIgnored); d != "" {
				t.Errorf("Actual links didn't match; diff:\n%v", d)
			}
		})
	}
}

var EntityIgnored = cmpopts.IgnoreFields(languagepb.Entity{}, "state", "sizeCache", "unknownFields")

// clientOpt is the option tests should use to connect to the test server.
// It is initialized by TestMain.
var clientOpt option.ClientOption

var (
	mockLanguage glanguage.MockLanguageServer
)

// TestMain runs once and initializes the server.
func TestMain(m *testing.M) {
	flag.Parse()

	serv := grpc.NewServer()
	languagepb.RegisterLanguageServiceServer(serv, &mockLanguage)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatal(err)
	}
	go serv.Serve(lis)

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	clientOpt = option.WithGRPCConn(conn)

	os.Exit(m.Run())
}
