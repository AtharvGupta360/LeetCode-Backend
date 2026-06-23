package repository

import (
	"context"
	"fmt"

	"github.com/gupta/leetcode-judge/internal/models"
	"github.com/jmoiron/sqlx"
)

type TestCaseRepository struct {
	db *sqlx.DB
}

func NewTestCaseRepository(db *sqlx.DB) *TestCaseRepository {
	return &TestCaseRepository{db: db}
}

// Create inserts a new test case and returns the generated id + timestamps.
func (r *TestCaseRepository) Create(ctx context.Context, tc *models.TestCase) error {
	query := `
		INSERT INTO test_cases (problem_id, input, expected_output, is_hidden, order_index)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		tc.ProblemID, tc.Input, tc.ExpectedOutput, tc.IsHidden, tc.OrderIndex,
	).Scan(&tc.ID, &tc.CreatedAt, &tc.UpdatedAt)
}

// GetByID fetches a single test case by UUID.
func (r *TestCaseRepository) GetByID(ctx context.Context, id string) (*models.TestCase, error) {
	var tc models.TestCase
	query := `
		SELECT id, problem_id, input, expected_output, is_hidden, order_index, created_at, updated_at
		FROM test_cases WHERE id = $1`

	if err := r.db.GetContext(ctx, &tc, query, id); err != nil {
		return nil, fmt.Errorf("test case not found: %w", err)
	}
	return &tc, nil
}

// ListByProblem returns all test cases for a problem ordered by order_index.
// If includeHidden is false, hidden cases are excluded (for non-admin users).
func (r *TestCaseRepository) ListByProblem(ctx context.Context, problemID string, includeHidden bool) ([]models.TestCase, error) {
	var (
		cases []models.TestCase
		query string
		args  []interface{}
	)

	if includeHidden {
		query = `
			SELECT id, problem_id, input, expected_output, is_hidden, order_index, created_at, updated_at
			FROM test_cases WHERE problem_id = $1
			ORDER BY order_index ASC`
		args = []interface{}{problemID}
	} else {
		query = `
			SELECT id, problem_id, input, expected_output, is_hidden, order_index, created_at, updated_at
			FROM test_cases WHERE problem_id = $1 AND is_hidden = FALSE
			ORDER BY order_index ASC`
		args = []interface{}{problemID}
	}

	if err := r.db.SelectContext(ctx, &cases, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list test cases: %w", err)
	}
	return cases, nil
}

// Update saves changes to an existing test case.
func (r *TestCaseRepository) Update(ctx context.Context, tc *models.TestCase) error {
	query := `
		UPDATE test_cases
		SET input=$1, expected_output=$2, is_hidden=$3, order_index=$4, updated_at=NOW()
		WHERE id=$5
		RETURNING updated_at`

	return r.db.QueryRowContext(ctx, query,
		tc.Input, tc.ExpectedOutput, tc.IsHidden, tc.OrderIndex, tc.ID,
	).Scan(&tc.UpdatedAt)
}

// Delete removes a test case by ID.
func (r *TestCaseRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM test_cases WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete test case: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("test case not found")
	}
	return nil
}
