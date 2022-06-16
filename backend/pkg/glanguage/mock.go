package glanguage

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
	"google.golang.org/grpc/metadata"
	"strings"
)

// MockLanguageServer is a mock implementation to be used in tests.
//
// Based on https://github.com/googleapis/google-cloud-go/blob/507a2be8e4fda152d517dcb972be6353a6da2914/language/apiv1/mock_test.go#L47
type MockLanguageServer struct {
	// Embed for forward compatibility.
	// Tests will keep working if more methods are added
	// in the future.
	languagepb.LanguageServiceServer

	Reqs []proto.Message

	// If set, all calls return this error.
	Err error

	// responses to return if Err == nil
	Resps []proto.Message
}

func (s *MockLanguageServer) AnalyzeSentiment(ctx context.Context, req *languagepb.AnalyzeSentimentRequest) (*languagepb.AnalyzeSentimentResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.Reqs = append(s.Reqs, req)
	if s.Err != nil {
		return nil, s.Err
	}
	return s.Resps[0].(*languagepb.AnalyzeSentimentResponse), nil
}

func (s *MockLanguageServer) AnalyzeEntities(ctx context.Context, req *languagepb.AnalyzeEntitiesRequest) (*languagepb.AnalyzeEntitiesResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.Reqs = append(s.Reqs, req)
	if s.Err != nil {
		return nil, s.Err
	}
	return s.Resps[0].(*languagepb.AnalyzeEntitiesResponse), nil
}

func (s *MockLanguageServer) AnalyzeEntitySentiment(ctx context.Context, req *languagepb.AnalyzeEntitySentimentRequest) (*languagepb.AnalyzeEntitySentimentResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.Reqs = append(s.Reqs, req)
	if s.Err != nil {
		return nil, s.Err
	}
	return s.Resps[0].(*languagepb.AnalyzeEntitySentimentResponse), nil
}

func (s *MockLanguageServer) AnalyzeSyntax(ctx context.Context, req *languagepb.AnalyzeSyntaxRequest) (*languagepb.AnalyzeSyntaxResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.Reqs = append(s.Reqs, req)
	if s.Err != nil {
		return nil, s.Err
	}
	return s.Resps[0].(*languagepb.AnalyzeSyntaxResponse), nil
}

func (s *MockLanguageServer) ClassifyText(ctx context.Context, req *languagepb.ClassifyTextRequest) (*languagepb.ClassifyTextResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.Reqs = append(s.Reqs, req)
	if s.Err != nil {
		return nil, s.Err
	}
	return s.Resps[0].(*languagepb.ClassifyTextResponse), nil
}

func (s *MockLanguageServer) AnnotateText(ctx context.Context, req *languagepb.AnnotateTextRequest) (*languagepb.AnnotateTextResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.Reqs = append(s.Reqs, req)
	if s.Err != nil {
		return nil, s.Err
	}
	return s.Resps[0].(*languagepb.AnnotateTextResponse), nil
}
