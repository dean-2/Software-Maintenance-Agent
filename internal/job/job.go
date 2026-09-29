package job

import (
	"time"
	"fmt"
)

type Status string

const(
	StatusQueued              Status = "QUEUED"
    StatusPreparing           Status = "PREPARING"
    StatusAnalyzing           Status = "ANALYZING"
    StatusReproducing         Status = "REPRODUCING"
    StatusPatching            Status = "PATCHING"
    StatusValidating          Status = "VALIDATING"
    StatusRepairing           Status = "REPAIRING"
    StatusAwaitingReview      Status = "AWAITING_REVIEW"
    StatusApproved            Status = "APPROVED"
    StatusRejected            Status = "REJECTED"

	StatusFailed              Status = "FAILED"
	StatusCancelled           Status = "CANCELLED"
	StatusTimedOut            Status = "TIMED_OUT"
	StatusNotReproduced	      Status = "NOT_REPRODUCED"
	StatusRepairLimitReached  Status = "REPAIR_LIMIT_REACHED"
)

type Job struct {
	ID             string
    BugReport      string
    RepositoryPath string

    Status         Status

    CreatedAt      time.Time
    UpdatedAt      time.Time

    Error          string
    RepairAttempts int
}

func New(id, bugReport, repositoryPath string) *Job {
	now := time.Now()
	return &Job{
		ID:             id,
		BugReport:      bugReport,
		RepositoryPath: repositoryPath,
		Status:         StatusQueued,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func CanTransition(from, to Status) bool {
	switch from {
	case StatusQueued:
		return to == StatusPreparing || to == StatusFailed || to == StatusCancelled || to == StatusTimedOut
	case StatusPreparing:
		return to == StatusAnalyzing || to == StatusFailed || to == StatusCancelled || to == StatusTimedOut
	case StatusAnalyzing:
		return to == StatusReproducing || to == StatusFailed || to == StatusCancelled || to == StatusTimedOut
	case StatusReproducing:
		return to == StatusPatching || to == StatusNotReproduced || to == StatusFailed || to == StatusCancelled || to == StatusTimedOut
	case StatusPatching:
		return to == StatusValidating || to == StatusFailed || to == StatusCancelled || to == StatusTimedOut
	case StatusValidating:
		return to == StatusAwaitingReview || to == StatusRepairing || to == StatusFailed || to == StatusCancelled || to == StatusTimedOut || to == StatusRepairLimitReached
	case StatusRepairing:
		return to == StatusPatching || to == StatusFailed || to == StatusCancelled || to == StatusTimedOut || to == StatusRepairLimitReached
	case StatusAwaitingReview:
		return to == StatusApproved || to == StatusRejected
	}
	return false
}

func (j * Job) Transition(to Status) error {
	if !CanTransition(j.Status, to) {
		return fmt.Errorf("invalid transition from %s to %s", j.Status, to)
	}
	j.Status = to
	j.UpdatedAt = time.Now()
	return nil
}