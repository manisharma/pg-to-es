package business

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"pg-to-es/internal/contract"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Server struct {
	srv     *http.Server
	es      contract.Elastic
	esIndex string
}

func NewServer(es contract.Elastic, port int, esIndex string) *Server {
	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler:      nil,
	}
	return &Server{s, es, esIndex}
}

func (s *Server) InitRoutes() {
	r := mux.NewRouter()
	r.HandleFunc("/", s.Root).Methods("GET")
	r.HandleFunc("/search/user/{userID}", s.SearchProjectsByUser).Methods("GET")
	r.HandleFunc("/search/hashtags/{hashtag}", s.SearchProjectsByHashtag).Methods("GET")
	r.HandleFunc("/search/fuzzy/{query}", s.FuzzySearchProjects).Methods("GET")
	s.srv.Handler = r
}

func (s *Server) Start() error {
	if s.srv.Handler == nil {
		return fmt.Errorf("can not start server, routes not initialized, use InitRoutes()")
	}
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *Server) Root(w http.ResponseWriter, r *http.Request) {
	res := map[string]string{
		"To search for projects created by a particular user visit": "/search/user/{userID}",
		"To search for projects that use specific hashtags visit":   "/search/hashtags/{hashtag}",
		"To do full-text fuzzy search for projects visit":           "/search/fuzzy/{query}",
	}
	encode(w, res)
}

func (s *Server) SearchProjectsByUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["userID"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, err := s.es.SearchByUser(r.Context(), s.esIndex, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	encode(w, res)
}

func (s *Server) SearchProjectsByHashtag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hashtag := vars["hashtag"]
	res, err := s.es.SearchByHashtags(r.Context(), s.esIndex, hashtag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	encode(w, res)
}

func (s *Server) FuzzySearchProjects(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query := vars["query"]
	res, err := s.es.FuzzySearchProjects(r.Context(), s.esIndex, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	encode(w, res)
}

func encode(w http.ResponseWriter, res interface{}) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", " ")
	encoder.Encode(res)
}
