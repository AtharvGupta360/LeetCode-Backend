package models

import "time"

// SubmissionStatus represents the current state of a submission in the judge pipeline.
type SubmissionStatus string

const (
	StatusPending           SubmissionStatus = "pending"
	StatusRunning           SubmissionStatus = "running"
	StatusAccepted          SubmissionStatus = "accepted"
	StatusWrongAnswer       SubmissionStatus = "wrong_answer"
	StatusTimeLimitExceeded SubmissionStatus = "time_limit_exceeded"
	StatusRuntimeError      SubmissionStatus = "runtime_error"
	StatusCompileError      SubmissionStatus = "compile_error"
)

// Submission represents a user's code submission for a problem.
type Submission struct {
	ID              string           `json:"id"              db:"id"`
	UserID          string           `json:"userId"          db:"user_id"`
	ProblemID       string           `json:"problemId"       db:"problem_id"`
	Language        string           `json:"language"        db:"language"`
	Code            string           `json:"code"            db:"code"`
	Status          SubmissionStatus `json:"status"          db:"status"`
	RuntimeMs       *int             `json:"runtimeMs"       db:"runtime_ms"`
	PassedTestCases int              `json:"passedTestCases" db:"passed_test_cases"`
	TotalTestCases  int              `json:"totalTestCases"  db:"total_test_cases"`
	ErrorMessage    *string          `json:"errorMessage"    db:"error_message"`
	CreatedAt       time.Time        `json:"createdAt"       db:"created_at"`
	UpdatedAt       time.Time        `json:"updatedAt"       db:"updated_at"`
}
