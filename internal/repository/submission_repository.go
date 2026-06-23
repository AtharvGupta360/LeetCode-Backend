package repository

import (
	"context"
	"fmt"

	"github.com/gupta/leetcode-judge/internal/models"
	"github.com/jmoiron/sqlx"
)

type SubmissionRepository struct {
	db *sqlx.DB
}

func NewSubmissionRepository(db *sqlx.DB) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

// Create inserts a new submission and returns the generated id + timestamps.
func (r *SubmissionRepository) Create(ctx context.Context, s *models.Submission) error {
	query := `
		INSERT INTO submissions (user_id, problem_id, language, code, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		s.UserID, s.ProblemID, s.Language, s.Code, s.Status,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

// GetByID fetches a single submission by UUID.
func (r *SubmissionRepository) GetByID(ctx context.Context, id string) (*models.Submission, error) {
	var s models.Submission
	query := `
		SELECT id, user_id, problem_id, language, code, status,
		       runtime_ms, passed_test_cases, total_test_cases,
		       error_message, created_at, updated_at
		FROM submissions WHERE id = $1`

	if err := r.db.GetContext(ctx, &s, query, id); err != nil {
		return nil, fmt.Errorf("submission not found: %w", err)
	}
	return &s, nil
}

// UpdateStatus atomically updates only the judge-produced fields.
// This avoids race conditions between the worker and any other process.
func (r *SubmissionRepository) UpdateStatus(
	ctx context.Context,
	id string,
	status models.SubmissionStatus,
	runtimeMs *int,
	passed, total int,
	errMsg *string,
) error {
	query := `
		UPDATE submissions
		SET status=$1, runtime_ms=$2, passed_test_cases=$3,
		    total_test_cases=$4, error_message=$5, updated_at=NOW()
		WHERE id=$6`

	result, err := r.db.ExecContext(ctx, query,
		status, runtimeMs, passed, total, errMsg, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update submission status: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("submission not found")
	}
	return nil
}

// ListByUser returns paginated submissions for a specific user, newest first.
func (r *SubmissionRepository) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]models.Submission, error) {
	query := `
		SELECT id, user_id, problem_id, language, code, status,
		       runtime_ms, passed_test_cases, total_test_cases,
		       error_message, created_at, updated_at
		FROM submissions WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	var subs []models.Submission
	if err := r.db.SelectContext(ctx, &subs, query, userID, pageSize, (page-1)*pageSize); err != nil {
		return nil, fmt.Errorf("failed to list submissions: %w", err)
	}
	return subs, nil
}

// ListByProblem returns paginated submissions for a problem (admin view), newest first.
func (r *SubmissionRepository) ListByProblem(ctx context.Context, problemID string, page, pageSize int) ([]models.Submission, error) {
	query := `
		SELECT id, user_id, problem_id, language, code, status,
		       runtime_ms, passed_test_cases, total_test_cases,
		       error_message, created_at, updated_at
		FROM submissions WHERE problem_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	var subs []models.Submission
	if err := r.db.SelectContext(ctx, &subs, query, problemID, pageSize, (page-1)*pageSize); err != nil {
		return nil, fmt.Errorf("failed to list submissions: %w", err)
	}
	return subs, nil
}
