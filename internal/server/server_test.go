package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"software-maintenance-agent/internal/job"
)

func TestNewConcurrentMap(t *testing.T) {
	m := NewConcurrentMap[string, int]()
	if m == nil {
		t.Fatal("NewConcurrentMap returned nil")
	}
}

func TestConcurrentMapSetGet(t *testing.T) {
	m := NewConcurrentMap[string, int]()
	m.Set("key", 42)
	value, exists := m.Get("key")
	if !exists {
		t.Fatal("Expected key not found")
	}
	if value != 42 {
		t.Fatalf("Expected value 42, got %d", value)
	}
}

func TestHealthHandler(t *testing.T) {
	s := New()
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}
	var response struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode health response: %v", err)
	}
	if response.Status != "ok" {
		t.Fatalf("Expected health status %q, got %q", "ok", response.Status)
	}
}

func TestCreateJobHandler(t *testing.T) {
	s := New()
	reqBody := `{"bug_report": "test bug", "repository_path": "./repo"}`
	req, err := http.NewRequest("POST", "/jobs", strings.NewReader(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("Expected status code %d, got %d", http.StatusCreated, rr.Code)
	}
	var createdJob job.Job
	if err := json.NewDecoder(rr.Body).Decode(&createdJob); err != nil {
		t.Fatalf("Failed to decode created job: %v", err)
	}
	if createdJob.ID == "" {
		t.Fatal("Expected created job to have a nonempty ID")
	}
	if createdJob.Status != job.StatusQueued {
		t.Fatalf("Expected job status %q, got %q", job.StatusQueued, createdJob.Status)
	}
	storedJob, exists := s.jobs.Get(createdJob.ID)
	if !exists || storedJob == nil {
		t.Fatalf("Expected job with ID %q to be stored", createdJob.ID)
	}
}

func TestCreateJobHandlerInvalidJSON(t *testing.T) {
	s := New()
	reqBody := `{"bug_report": "test bug", "repository_path": "./repo"`
	req, err := http.NewRequest("POST", "/jobs", strings.NewReader(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestCreateJobHandlerMissingFields(t *testing.T) {
	s := New()
	reqBody := `{"bug_report": "", "repository_path": ""}`
	req, err := http.NewRequest("POST", "/jobs", strings.NewReader(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestGetJobHandler(t *testing.T) {
	s := New()
	// First, create a job to retrieve
	reqBody := `{"bug_report": "test bug", "repository_path": "./repo"}`
	createReq, err := http.NewRequest("POST", "/jobs", strings.NewReader(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	s.mux.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("Expected status code %d, got %d", http.StatusCreated, createRR.Code)
	}
	var createdJob job.Job
	if err := json.NewDecoder(createRR.Body).Decode(&createdJob); err != nil {
		t.Fatalf("Failed to decode created job: %v", err)
	}

	// Now, retrieve the job using its ID
	getReq, err := http.NewRequest("GET", "/jobs/"+createdJob.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	getRR := httptest.NewRecorder()
	s.mux.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, getRR.Code)
	}
	var retrievedJob job.Job
	if err := json.NewDecoder(getRR.Body).Decode(&retrievedJob); err != nil {
		t.Fatalf("Failed to decode retrieved job: %v", err)
	}
	if retrievedJob.ID != createdJob.ID {
		t.Fatalf("Expected retrieved job ID %q, got %q", createdJob.ID, retrievedJob.ID)
	}
	if retrievedJob.BugReport != createdJob.BugReport {
		t.Fatalf("Expected retrieved job bug report %q, got %q", createdJob.BugReport, retrievedJob.BugReport)
	}
	if retrievedJob.RepositoryPath != createdJob.RepositoryPath {
		t.Fatalf("Expected retrieved job repository path %q, got %q", createdJob.RepositoryPath, retrievedJob.RepositoryPath)
	}
	if retrievedJob.Status != createdJob.Status {
		t.Fatalf("Expected retrieved job status %q, got %q", createdJob.Status, retrievedJob.Status)
	}
}

func TestGetJobHandlerNotFound(t *testing.T) {
	s := New()
	req, err := http.NewRequest("GET", "/jobs/nonexistent-id", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	s.mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("Expected status code %d, got %d", http.StatusNotFound, rr.Code)
	}
}