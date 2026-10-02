package server

import (
	"crypto/rand"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"software-maintenance-agent/internal/job"
)

type ConcurrentMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

type Server struct {
	mux  *http.ServeMux
	jobs *ConcurrentMap[string, *job.Job]
}

func NewConcurrentMap[K comparable, V any]() *ConcurrentMap[K, V] {
	return &ConcurrentMap[K, V]{
		m: make(map[K]V),
	}
}

func (m *ConcurrentMap[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.m[key]
	return value, ok
}

func (m *ConcurrentMap[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.m[key] = value
}

func New() *Server {
	s := &Server{
		mux:  http.NewServeMux(),
		jobs: NewConcurrentMap[string, *job.Job](),
	}

	s.mux.HandleFunc("GET /health", s.healthHandler)
	s.mux.HandleFunc("POST /jobs", s.createJobHandler)
	s.mux.HandleFunc("GET /jobs/{id}", s.getJobHandler)

	return s
}

func (s *Server) healthHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func (s *Server) createJobHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req struct {
		BugReport      string `json:"bug_report"`
		RepositoryPath string `json:"repository_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.BugReport) == "" {
		http.Error(w, "bug_report is required", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.RepositoryPath) == "" {
		http.Error(w, "repository_path is required", http.StatusBadRequest)
		return
	}

	id := rand.Text()
	j := job.New(id, req.BugReport, req.RepositoryPath)

	s.jobs.Set(j.ID, j)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(j); err != nil {
		// The response may already have been partially written.
		// Log this error in a later milestone.
		return
	}
}

func (s *Server) getJobHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := r.PathValue("id")

	j, exists := s.jobs.Get(id)
	if !exists || j == nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(j); err != nil {
		// The response may already have been partially written.
		// Log this error in a later milestone.
		return
	}
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}
