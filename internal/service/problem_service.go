package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/gupta/leetcode-judge/internal/models"
	"github.com/gupta/leetcode-judge/internal/repository"
)

// ProblemService holds business logic for the problems domain.
// It sits between the HTTP handler and the repository.
type ProblemService struct {
	repo *repository.ProblemRepository
}

func NewProblemService(repo *repository.ProblemRepository) *ProblemService {
	return &ProblemService{repo: repo}
}

// CreateProblemInput is the validated input shape for creating a problem.
type CreateProblemInput struct {
	Title       string
	Description string
	Difficulty  models.Difficulty
	Tags        []string
	Examples    string
	Constraints string
	IsPublished bool
	CreatedBy   string // userID from JWT
}

// UpdateProblemInput is the validated input shape for updating a problem.
type UpdateProblemInput struct {
	Title       string
	Description string
	Difficulty  models.Difficulty
	Tags        []string
	Examples    string
	Constraints string
	IsPublished bool
}

// Create validates input, generates a slug, and persists the problem.
func (s *ProblemService) Create(ctx context.Context, in CreateProblemInput) (*models.Problem, error) {
	if err := validateDifficulty(in.Difficulty); err != nil {
		return nil, err
	}

	p := &models.Problem{
		Title:       in.Title,
		Slug:        slugify(in.Title),
		Description: in.Description,
		Difficulty:  in.Difficulty,
		Tags:        in.Tags,
		Examples:    in.Examples,
		Constraints: in.Constraints,
		IsPublished: in.IsPublished,
		CreatedBy:   in.CreatedBy,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("create problem: %w", err)
	}
	return p, nil
}

// GetByID returns a problem by UUID.
func (s *ProblemService) GetByID(ctx context.Context, id string) (*models.Problem, error) {
	return s.repo.GetByID(ctx, id)
}

// GetBySlug returns a problem by its URL slug.
func (s *ProblemService) GetBySlug(ctx context.Context, slug string) (*models.Problem, error) {
	return s.repo.GetBySlug(ctx, slug)
}

// List returns paginated problems with optional difficulty filter.
// Only admins see unpublished problems; regular users only see published ones.
func (s *ProblemService) List(ctx context.Context, difficulty string, isAdmin bool, page, pageSize int) ([]models.Problem, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := repository.ListFilter{
		Difficulty: difficulty,
		Limit:      pageSize,
		Offset:     (page - 1) * pageSize,
	}

	// Non-admins only see published problems
	if !isAdmin {
		published := true
		filter.IsPublished = &published
	}

	return s.repo.List(ctx, filter)
}

// Update applies changes to an existing problem.
func (s *ProblemService) Update(ctx context.Context, id string, in UpdateProblemInput) (*models.Problem, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := validateDifficulty(in.Difficulty); err != nil {
		return nil, err
	}

	p.Title = in.Title
	p.Slug = slugify(in.Title)
	p.Description = in.Description
	p.Difficulty = in.Difficulty
	p.Tags = in.Tags
	p.Examples = in.Examples
	p.Constraints = in.Constraints
	p.IsPublished = in.IsPublished

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("update problem: %w", err)
	}
	return p, nil
}

// Delete removes a problem by UUID.
func (s *ProblemService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// --- helpers ---

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

// slugify converts "Two Sum" → "two-sum"
func slugify(title string) string {
	s := strings.ToLower(title)
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func validateDifficulty(d models.Difficulty) error {
	switch d {
	case models.DifficultyEasy, models.DifficultyMedium, models.DifficultyHard:
		return nil
	default:
		return fmt.Errorf("invalid difficulty %q: must be easy, medium, or hard", d)
	}
}
