package job

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	before := time.Now()
	j := New("job-1", "bug-report", "./repo")

	after := time.Now()
	if j.ID != "job-1" {
		t.Errorf("Expected job ID 'job-1', got '%s'", j.ID)
	}

	if j.BugReport != "bug-report" {
		t.Errorf("Expected bug report 'bug-report', got '%s'", j.BugReport)
	}

	if j.RepositoryPath != "./repo" {
		t.Errorf("Expected repository path './repo', got '%s'", j.RepositoryPath)
	}

	if j.Status != StatusQueued {
		t.Errorf("Expected status 'QUEUED', got '%s'", j.Status)
	}

	if j.CreatedAt.Before(before) || j.CreatedAt.After(after) {
		t.Errorf("Expected created at time between %v and %v, got %v", before, after, j.CreatedAt)
	}

	if j.UpdatedAt.Before(before) || j.UpdatedAt.After(after) {
		t.Errorf("Expected updated at time between %v and %v, got %v", before, after, j.UpdatedAt)
	}
}

func TestCanTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     Status
		to       Status
		want     bool
	}{
		{
			name: "Queued to Preparing",
			from: StatusQueued,
			to:   StatusPreparing,
			want: true,
		},
		{
			name: "Queued to Analyzing",
			from: StatusQueued,
			to:   StatusAnalyzing,
			want: false,
		},
		{
			name: "Preparing to Analyzing",
			from: StatusPreparing,
			to:   StatusAnalyzing,
			want: true,
		},
		{
			name: "Analyzing to Reproducing",
			from: StatusAnalyzing,
			to:   StatusReproducing,
			want: true,
		},
		{
			name: "Reproducing to Patching",
			from: StatusReproducing,
			to:   StatusPatching,
			want: true,
		},
		{
			name: "Patching to Validating",
			from: StatusPatching,
			to:   StatusValidating,
			want: true,
		},
		{
			name: "Approved to Patching",
			from: StatusApproved,
			to:   StatusPatching,
			want: false,
		},
		{
			name: "Rejected to Queued",
			from: StatusRejected,
			to:   StatusQueued,
			want: false,
		},
		{
			name: "Failed to Preparing",
			from: StatusFailed,
			to:   StatusPreparing,
			want: false,
		},
		{
			name: "Cancelled to Analyzing",
			from: StatusCancelled,
			to:   StatusAnalyzing,
			want: false,
		},
		{
			name: "Queued to Analyzing",
			from: StatusQueued,
			to:   StatusAnalyzing,
			want: false,
		},
		{
			name: "Preparing to Patching",
			from: StatusPreparing,
			to:   StatusPatching,
			want: false,
		},
		{
			name: "Analyzing to Validating",
			from: StatusAnalyzing,
			to:   StatusValidating,
			want: false,
		},
		{
			name: "Awaiting Review to Patching",
			from: StatusAwaitingReview,
			to:   StatusPatching,
			want: false,
		},
		{
			name: "Unknown state to Preparing",
			from: Status("UNKNOWN"),
			to:   StatusPreparing,
			want: false,
		},
		{
			name: "Repairing to Patching",
			from: StatusRepairing,
			to:   StatusPatching,
			want: true,
		},
	}
	for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := CanTransition(tt.from, tt.to)

            if got != tt.want {
                t.Errorf(
                    "CanTransition(%s, %s) = %v; want %v",
                    tt.from, tt.to, got, tt.want,
                )
            }
        })
    }
}

func TestTransition(t *testing.T) {
	j := New("job-1", "bug-report", "./repo")
	err := j.Transition(StatusPreparing)
	if err != nil {
		t.Errorf("Expected transition to succeed, got error: %v", err)
	}

	if j.Status != StatusPreparing {
		t.Errorf("Expected status 'PREPARING', got '%s'", j.Status)
	}

	if(j.UpdatedAt.Before(j.CreatedAt)) {
		t.Errorf("Expected updated at time to be after created at time, got %v", j.UpdatedAt)
	}
}

func TestTransitionSuccess(t *testing.T) {
    j := New("job-1", "bug-report", "./repo")
    originalTime := j.UpdatedAt

    err := j.Transition(StatusPreparing)

    if err != nil {
        t.Fatalf("Expected transition to succeed, got: %v", err)
    }

    if j.Status != StatusPreparing {
        t.Errorf("Expected status PREPARING, got %s", j.Status)
    }

    if j.UpdatedAt.Before(originalTime) {
        t.Errorf("UpdatedAt should not be before the original timestamp")
    }
}

func TestTransitionInvalid(t *testing.T) {
    j := New("job-1", "bug-report", "./repo")
    originalTime := j.UpdatedAt

    err := j.Transition(StatusApproved)

    if err == nil {
        t.Fatal("Expected an error for invalid transition")
    }

    if j.Status != StatusQueued {
        t.Errorf("Expected status QUEUED, got %s", j.Status)
    }

    if !j.UpdatedAt.Equal(originalTime) {
        t.Errorf("UpdatedAt changed after invalid transition")
    }
}