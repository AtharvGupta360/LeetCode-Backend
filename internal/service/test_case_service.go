package service

import (
	"context"
	"fmt"

	"github.com/gupta/leetcode-judge/internal/models"
	"github.com/gupta/leetcode-judge/internal/repository"
)

// TestCaseService holds business logic for the test-cases domain.
type TestCaseService struct {
	repo        *repository.TestCaseRepository
	problemRepo *repository.ProblemRepository // used to verify problem existence
}

func NewTestCaseService(repo *repository.TestCaseRepository, problemRepo *repository.ProblemRepository) *TestCaseService {
	return &TestCaseService{repo: repo, problemRepo: problemRepo}
}

// CreateTestCaseInput is the validated input shape for creating a test case.
type CreateTestCaseInput struct {
	ProblemID      string
	Input          string
	ExpectedOutput string
	IsHidden       bool
	OrderIndex     int
}

// UpdateTestCaseInput is the validated input shape for updating a test case.
type UpdateTestCaseInput struct {
	Input          string
	ExpectedOutput string
	IsHidden       bool
	OrderIndex     int
}

// Create validates that the problem exists, then persists the test case.
func (s *TestCaseService) Create(ctx context.Context, in CreateTestCaseInput) (*models.TestCase, error) {
	// Ensure the parent problem exists before inserting
	if _, err := s.problemRepo.GetByID(ctx, in.ProblemID); err != nil {
		return nil, fmt.Errorf("problem not found: %w", err)
	}

	tc := &models.TestCase{
		ProblemID:      in.ProblemID,
		Input:          in.Input,
		ExpectedOutput: in.ExpectedOutput,
		IsHidden:       in.IsHidden,
		OrderIndex:     in.OrderIndex,
	}

	if err := s.repo.Create(ctx, tc); err != nil {
		return nil, fmt.Errorf("create test case: %w", err)
	}
	return tc, nil
}

// List returns test cases for a problem.
// Admins receive all cases (including hidden); regular users only see visible ones.
func (s *TestCaseService) List(ctx context.Context, problemID string, isAdmin bool) ([]models.TestCase, error) {
	return s.repo.ListByProblem(ctx, problemID, isAdmin)
}

// GetByID returns a single test case. Hidden cases are accessible to admins only.
func (s *TestCaseService) GetByID(ctx context.Context, id string, isAdmin bool) (*models.TestCase, error) {
	tc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tc.IsHidden && !isAdmin {
		return nil, fmt.Errorf("test case not found")
	}
	return tc, nil
}

// Update applies changes to an existing test case.
func (s *TestCaseService) Update(ctx context.Context, id string, in UpdateTestCaseInput) (*models.TestCase, error) {
	tc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tc.Input = in.Input
	tc.ExpectedOutput = in.ExpectedOutput
	tc.IsHidden = in.IsHidden
	tc.OrderIndex = in.OrderIndex

	if err := s.repo.Update(ctx, tc); err != nil {
		return nil, fmt.Errorf("update test case: %w", err)
	}
	return tc, nil
}

// Delete removes a test case by ID.
func (s *TestCaseService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
