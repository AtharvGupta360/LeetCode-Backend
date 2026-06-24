package judge

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gupta/leetcode-judge/internal/common"
	"github.com/gupta/leetcode-judge/internal/models"
	"github.com/gupta/leetcode-judge/internal/queue"
	"github.com/gupta/leetcode-judge/internal/repository"
)

// Service is the judge orchestrator.
// It dequeues jobs, runs them against all test cases, and writes the verdict back to the DB.
type Service struct {
	runner      *Runner
	subRepo     *repository.SubmissionRepository
	testCaseRepo *repository.TestCaseRepository
	queue       *queue.RedisQueue
}

func NewService(
	runner *Runner,
	subRepo *repository.SubmissionRepository,
	testCaseRepo *repository.TestCaseRepository,
	q *queue.RedisQueue,
) *Service {
	return &Service{
		runner:       runner,
		subRepo:      subRepo,
		testCaseRepo: testCaseRepo,
		queue:        q,
	}
}

// Start launches the worker loop. It blocks until ctx is cancelled (graceful shutdown).
// Call this in a goroutine from main.
func (s *Service) Start(ctx context.Context) {
	common.Logger.Info("judge worker started")
	for {
		select {
		case <-ctx.Done():
			common.Logger.Info("judge worker stopped")
			return
		default:
			// Block for up to 2 seconds waiting for a job, then loop again.
			// This allows ctx.Done() to be checked between polls.
			job, err := s.queue.Dequeue(ctx, 2*time.Second)
			if err != nil {
				common.Logger.Errorf("judge dequeue error: %v", err)
				continue
			}
			if job == nil {
				continue // timeout, no job — loop again
			}

			common.Logger.Infof("judge processing submission %s (language: %s)", job.SubmissionID, job.Language)
			s.process(ctx, job)
		}
	}
}

// process runs a single job through the full judge pipeline.
func (s *Service) process(ctx context.Context, job *queue.Job) {
	// Mark as running immediately so the user sees progress
	if err := s.subRepo.UpdateStatus(ctx, job.SubmissionID, models.StatusRunning, nil, 0, 0, nil); err != nil {
		common.Logger.Errorf("judge: failed to set running status for %s: %v", job.SubmissionID, err)
		return
	}

	verdict, runtimeMs, passed, total, errMsg := s.evaluate(ctx, job)

	var runtimePtr *int
	if runtimeMs > 0 {
		ms := int(runtimeMs)
		runtimePtr = &ms
	}

	var errMsgPtr *string
	if errMsg != "" {
		errMsgPtr = &errMsg
	}

	if err := s.subRepo.UpdateStatus(ctx, job.SubmissionID, verdict, runtimePtr, passed, total, errMsgPtr); err != nil {
		common.Logger.Errorf("judge: failed to write verdict for %s: %v", job.SubmissionID, err)
	}

	common.Logger.Infof("judge: submission %s → %s (%d/%d test cases, %dms)",
		job.SubmissionID, verdict, passed, total, runtimeMs)
}

// evaluate runs the code against all test cases and returns the final verdict.
func (s *Service) evaluate(ctx context.Context, job *queue.Job) (
	verdict models.SubmissionStatus,
	totalRuntimeMs int64,
	passed, total int,
	errMsg string,
) {
	// Step 1: Static sanitization — catch obvious attacks before spinning up Docker
	if err := Sanitize(job.Language, job.Code); err != nil {
		return models.StatusCompileError, 0, 0, 0, fmt.Sprintf("code sanitization failed: %v", err)
	}

	// Step 2: Fetch all test cases for this problem (including hidden ones — judge sees all)
	testCases, err := s.testCaseRepo.ListByProblem(ctx, job.ProblemID, true)
	if err != nil {
		return models.StatusRuntimeError, 0, 0, 0, fmt.Sprintf("failed to load test cases: %v", err)
	}
	if len(testCases) == 0 {
		return models.StatusRuntimeError, 0, 0, 0, "no test cases found for this problem"
	}

	total = len(testCases)

	// Step 3: Run the code against each test case in order.
	// Stop early on the first failure (same as real LeetCode behaviour).
	for _, tc := range testCases {
		result, err := s.runner.Run(ctx, job.Language, job.Code, tc.Input)
		if err != nil {
			return models.StatusRuntimeError, totalRuntimeMs, passed, total,
				fmt.Sprintf("runner error on test case %s: %v", tc.ID, err)
		}

		totalRuntimeMs += result.RuntimeMs

		// Check for special failure conditions first
		if result.TimedOut {
			return models.StatusTimeLimitExceeded, totalRuntimeMs, passed, total,
				fmt.Sprintf("time limit exceeded on test case %d", passed+1)
		}
		if result.OOMKilled {
			return models.StatusRuntimeError, totalRuntimeMs, passed, total,
				fmt.Sprintf("memory limit exceeded on test case %d", passed+1)
		}
		if result.Stderr != "" {
			return models.StatusRuntimeError, totalRuntimeMs, passed, total,
				truncate(result.Stderr, 500)
		}

		// Compare actual output vs expected output (normalized)
		if !outputMatches(result.Stdout, tc.ExpectedOutput) {
			return models.StatusWrongAnswer, totalRuntimeMs, passed, total,
				fmt.Sprintf("wrong answer on test case %d", passed+1)
		}

		passed++
	}

	// All test cases passed
	return models.StatusAccepted, totalRuntimeMs, passed, total, ""
}

// outputMatches compares two outputs with whitespace normalization.
// Trims trailing spaces and normalizes line endings — same approach real judges use.
func outputMatches(actual, expected string) bool {
	return normalizeOutput(actual) == normalizeOutput(expected)
}

func normalizeOutput(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n") // Windows line endings → Unix
	lines := strings.Split(s, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		result = append(result, trimmed)
	}
	// Remove trailing empty lines
	for len(result) > 0 && result[len(result)-1] == "" {
		result = result[:len(result)-1]
	}
	return strings.Join(result, "\n")
}

// truncate limits an error message length to avoid storing huge stack traces.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "... (truncated)"
}
