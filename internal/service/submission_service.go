package service

import (
	"context"
	"fmt"

	"github.com/gupta/leetcode-judge/internal/models"
	"github.com/gupta/leetcode-judge/internal/queue"
	"github.com/gupta/leetcode-judge/internal/repository"
	"github.com/gupta/leetcode-judge/internal/service/judge"
)

// SubmissionService holds business logic for the submissions domain.
// It sits between the HTTP handler and the repository + queue.
type SubmissionService struct {
	subRepo     *repository.SubmissionRepository
	problemRepo *repository.ProblemRepository
	queue       *queue.RedisQueue
}

func NewSubmissionService(
	subRepo *repository.SubmissionRepository,
	problemRepo *repository.ProblemRepository,
	q *queue.RedisQueue,
) *SubmissionService {
	return &SubmissionService{
		subRepo:     subRepo,
		problemRepo: problemRepo,
		queue:       q,
	}
}

// SubmitInput is the validated input shape for submitting code.
type SubmitInput struct {
	UserID    string
	ProblemID string
	Language  string
	Code      string
}

// Submit validates the input, persists the submission as "pending",
// and enqueues a job for the judge worker to pick up.
func (s *SubmissionService) Submit(ctx context.Context, in SubmitInput) (*models.Submission, error) {
	// Validate language is supported by the judge
	if !judge.IsSupported(in.Language) {
		return nil, fmt.Errorf("unsupported language %q: must be one of %v", in.Language, judge.SupportedLanguages())
	}

	// Verify the problem actually exists
	if _, err := s.problemRepo.GetByID(ctx, in.ProblemID); err != nil {
		return nil, fmt.Errorf("problem not found: %w", err)
	}

	// Persist the submission with initial "pending" status
	sub := &models.Submission{
		UserID:    in.UserID,
		ProblemID: in.ProblemID,
		Language:  in.Language,
		Code:      in.Code,
		Status:    models.StatusPending,
	}
	if err := s.subRepo.Create(ctx, sub); err != nil {
		return nil, fmt.Errorf("create submission: %w", err)
	}

	// Enqueue the job for the judge worker
	job := queue.Job{
		SubmissionID: sub.ID,
		ProblemID:    sub.ProblemID,
		Language:     sub.Language,
		Code:         sub.Code,
	}
	if err := s.queue.Enqueue(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to enqueue submission: %w", err)
	}

	return sub, nil
}

// GetByID returns a single submission by ID (used for polling verdict status).
func (s *SubmissionService) GetByID(ctx context.Context, id string) (*models.Submission, error) {
	return s.subRepo.GetByID(ctx, id)
}

// ListByUser returns paginated submissions for a specific user, newest first.
func (s *SubmissionService) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]models.Submission, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.subRepo.ListByUser(ctx, userID, page, pageSize)
}

// ListByProblem returns paginated submissions for a problem (admin view), newest first.
func (s *SubmissionService) ListByProblem(ctx context.Context, problemID string, page, pageSize int) ([]models.Submission, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.subRepo.ListByProblem(ctx, problemID, page, pageSize)
}
