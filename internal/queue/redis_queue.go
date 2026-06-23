package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const queueKey = "judge:submissions"

// Job represents a submission waiting to be judged.
type Job struct {
	SubmissionID string `json:"submissionId"`
	ProblemID    string `json:"problemId"`
	Language     string `json:"language"`
	Code         string `json:"code"`
}

// RedisQueue is a FIFO job queue backed by a Redis List.
type RedisQueue struct {
	client *redis.Client
}

// NewRedisQueue creates a new queue instance.
func NewRedisQueue(client *redis.Client) *RedisQueue {
	return &RedisQueue{client: client}
}

// Enqueue pushes a job to the queue (LPUSH = add to left).
func (q *RedisQueue) Enqueue(ctx context.Context, job Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("queue marshal error: %w", err)
	}
	return q.client.LPush(ctx, queueKey, data).Err()
}

// Dequeue blocks until a job is available or timeout is reached (BRPOP = pop from right).
// Returns nil, nil on timeout (caller should loop and retry).
func (q *RedisQueue) Dequeue(ctx context.Context, timeout time.Duration) (*Job, error) {
	result, err := q.client.BRPop(ctx, timeout, queueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // timeout, no job available
		}
		return nil, fmt.Errorf("queue dequeue error: %w", err)
	}

	// BRPop returns [key, value] — we want index 1
	var job Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, fmt.Errorf("queue unmarshal error: %w", err)
	}
	return &job, nil
}

// Len returns the current number of jobs waiting in the queue.
func (q *RedisQueue) Len(ctx context.Context) (int64, error) {
	return q.client.LLen(ctx, queueKey).Result()
}
