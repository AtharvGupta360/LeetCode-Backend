package models

import "time"

// TestCase represents a single input/output pair used to judge a submission.
type TestCase struct {
	ID             string    `json:"id"             db:"id"`
	ProblemID      string    `json:"problemId"      db:"problem_id"`
	Input          string    `json:"input"          db:"input"`
	ExpectedOutput string    `json:"expectedOutput" db:"expected_output"`
	IsHidden       bool      `json:"isHidden"       db:"is_hidden"`
	OrderIndex     int       `json:"orderIndex"     db:"order_index"`
	CreatedAt      time.Time `json:"createdAt"      db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt"      db:"updated_at"`
}
