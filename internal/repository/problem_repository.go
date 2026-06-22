package repository

import (
	"context"
	"fmt"

	"github.com/gupta/leetcode-judge/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type ProblemRepository struct {
	db *sqlx.DB
}

func NewProblemRepository(db *sqlx.DB) *ProblemRepository {
	return &ProblemRepository{db: db}
}

// Create inserts a new problem and returns the generated id + timestamps.
func (r *ProblemRepository) Create(ctx context.Context, p *models.Problem) error {
	query := `
		INSERT INTO problems (title, slug, description, difficulty, tags, examples, constraints, is_published, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		p.Title, p.Slug, p.Description, p.Difficulty,
		pq.Array(p.Tags), p.Examples, p.Constraints,
		p.IsPublished, p.CreatedBy,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

// GetByID fetches a single problem by its UUID.
func (r *ProblemRepository) GetByID(ctx context.Context, id string) (*models.Problem, error) {
	var p models.Problem
	query := `
		SELECT id, title, slug, description, difficulty, tags, examples, constraints, is_published, created_by, created_at, updated_at
		FROM problems WHERE id = $1`

	if err := r.db.GetContext(ctx, &p, query, id); err != nil {
		return nil, fmt.Errorf("problem not found: %w", err)
	}
	return &p, nil
}

// GetBySlug fetches a single problem by its slug (used in public-facing URLs).
func (r *ProblemRepository) GetBySlug(ctx context.Context, slug string) (*models.Problem, error) {
	var p models.Problem
	query := `
		SELECT id, title, slug, description, difficulty, tags, examples, constraints, is_published, created_by, created_at, updated_at
		FROM problems WHERE slug = $1`

	if err := r.db.GetContext(ctx, &p, query, slug); err != nil {
		return nil, fmt.Errorf("problem not found: %w", err)
	}
	return &p, nil
}

// ListFilter holds optional filter/pagination params for listing problems.
type ListFilter struct {
	Difficulty  string // "easy" | "medium" | "hard" | "" (all)
	IsPublished *bool  // nil = all, true = published only, false = drafts only
	Limit       int
	Offset      int
}

// List returns a paginated list of problems with optional filters.
func (r *ProblemRepository) List(ctx context.Context, f ListFilter) ([]models.Problem, error) {
	// Build the query dynamically based on filters
	query := `
		SELECT id, title, slug, description, difficulty, tags, examples, constraints, is_published, created_by, created_at, updated_at
		FROM problems WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if f.Difficulty != "" {
		query += fmt.Sprintf(" AND difficulty = $%d", argIdx)
		args = append(args, f.Difficulty)
		argIdx++
	}
	if f.IsPublished != nil {
		query += fmt.Sprintf(" AND is_published = $%d", argIdx)
		args = append(args, *f.IsPublished)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, f.Limit, f.Offset)

	var problems []models.Problem
	if err := r.db.SelectContext(ctx, &problems, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list problems: %w", err)
	}
	return problems, nil
}

// Update saves changes to an existing problem. Only updates mutable fields.
func (r *ProblemRepository) Update(ctx context.Context, p *models.Problem) error {
	query := `
		UPDATE problems
		SET title=$1, slug=$2, description=$3, difficulty=$4, tags=$5,
		    examples=$6, constraints=$7, is_published=$8, updated_at=NOW()
		WHERE id=$9
		RETURNING updated_at`

	return r.db.QueryRowContext(ctx, query,
		p.Title, p.Slug, p.Description, p.Difficulty,
		pq.Array(p.Tags), p.Examples, p.Constraints,
		p.IsPublished, p.ID,
	).Scan(&p.UpdatedAt)
}

// Delete removes a problem by ID.
func (r *ProblemRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM problems WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete problem: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("problem not found")
	}
	return nil
}
