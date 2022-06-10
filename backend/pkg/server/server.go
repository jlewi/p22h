package server

import (
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/jlewi/p22h/backend/api"
	"github.com/jlewi/p22h/backend/pkg/datastore"
	"github.com/jlewi/p22h/backend/pkg/debug"
	"github.com/pkg/errors"
	"net"
	"net/http"
	"path/filepath"
)

const (
	// What if we want to get all the links to some reference which is not a Document? e.g. all the
	// links pointing at www.kubernetes.com
	backLinksPath = "/documents/{name}:backLinks"
)

type Server struct {
	log        logr.Logger
	staticPath string
	listener   net.Listener
	store      *datastore.Datastore
}

func NewServer(staticPath string, listener net.Listener, store *datastore.Datastore, log logr.Logger) (*Server, error) {
	if staticPath == "" {
		return nil, errors.Errorf("staticPath must be set")
	}

	if listener == nil {
		return nil, errors.Errorf("listener must be set")
	}
	if store == nil {
		return nil, errors.Errorf("store must be set")
	}
	resolved, err := filepath.Abs(staticPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get absolute path; path: %v", staticPath)
	}
	log.Info("resolved static path", "input", staticPath, "output", resolved)
	return &Server{
		log:        log,
		staticPath: resolved,
		listener:   listener,
		store:      store,
	}, nil
}

func (s *Server) Address() string {
	return fmt.Sprintf("http://localhost:%v", s.listener.Addr().(*net.TCPAddr).Port)
}

func (s *Server) writeStatus(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	resp := api.RequestStatus{
		Kind:    "RequestStatus",
		Message: message,
		Code:    code,
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		s.log.Error(err, "Failed to marshal RequestStatus", "RequestStatus", resp, "code", code)
	}

	if code != http.StatusOK {
		caller := debug.ThisCaller()
		s.log.Info("HTTP error", "RequestStatus", resp, "code", code, "caller", caller)
	}
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	s.writeStatus(w, "Feed backend is running", http.StatusOK)
}

// BackLinks returns backLinks for a given document.
func (s *Server) BackLinks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		s.writeStatus(w, "Missing document name", http.StatusBadRequest)
		return
	}

	links, err := s.store.ListDocLinks(name)

	if err != nil {
		s.writeStatus(w, fmt.Sprintf("Failed to get backlinks for doc: %v; error %v", name, err), http.StatusInternalServerError)
		return
	}

	linkList := &api.BackLinkList{
		Items: make([]api.BackLink, len(links)),
	}

	for i, l := range links {
		linkList.Items[i] = api.BackLink{
			Text:  l.Text,
			DocId: l.SourceID,
		}
	}
	payload, err := json.Marshal(linkList)
	if err != nil {
		s.writeStatus(w, fmt.Sprintf("Failed to encode BackLinkList; error %v", err), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(payload); err != nil {
		s.log.Error(err, "Failed to write response")
	}
}

func (s *Server) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	s.writeStatus(w, fmt.Sprintf("feed backend server doesn't handle the path; url: %v", r.URL), http.StatusNotFound)
}

// StartAndBlock starts the server and blocks.
func (s *Server) StartAndBlock() error {
	log := s.log

	router := mux.NewRouter().StrictSlash(true)

	// This will serve files under http://localhost:8000/ui/<filename>
	// This should match the --base-href argument used to compile the flutter web app
	log.Info("Configuring /ui/", "staticPath", s.staticPath)
	router.PathPrefix("/ui/").Handler(http.StripPrefix("/ui/", http.FileServer(http.Dir(s.staticPath))))

	router.HandleFunc("/healthz", s.HealthCheck)
	router.HandleFunc(backLinksPath, s.BackLinks)
	router.NotFoundHandler = http.HandlerFunc(s.NotFoundHandler)

	log.Info("Gateway is running", "address", s.Address())
	err := http.Serve(s.listener, router)

	if err != nil {
		log.Error(err, "Server returned error")
	}
	return err
}
