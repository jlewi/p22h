package httptesting

import (
	"bytes"
	"encoding/json"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

// BuildJSONRoundTrip creates a round trip function that return
func BuildJSONRoundTrip(payload interface{}) RoundTripFunc {
	log := zapr.NewLogger(zap.L())

	return func(req *http.Request) *http.Response {
		log.Info("Got http request", "request", req.URL.String())

		b, err := json.Marshal(payload)

		if err != nil {
			log.Error(err, "Failed to marshal response")
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBufferString(err.Error())),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}
		}

		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBuffer(b)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	}
}
