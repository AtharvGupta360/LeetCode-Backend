-- 000004_create_submissions.up.sql
-- Submissions table for the LeetCode Judge platform

CREATE TYPE submission_status AS ENUM (
    'pending',
    'running',
    'accepted',
    'wrong_answer',
    'time_limit_exceeded',
    'runtime_error',
    'compile_error'
);

CREATE TABLE IF NOT EXISTS submissions (
    id                UUID              PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id           UUID              NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_id        UUID              NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    language          VARCHAR(20)       NOT NULL,
    code              TEXT              NOT NULL,
    status            submission_status NOT NULL DEFAULT 'pending',
    runtime_ms        INTEGER,
    passed_test_cases INTEGER           NOT NULL DEFAULT 0,
    total_test_cases  INTEGER           NOT NULL DEFAULT 0,
    error_message     TEXT,
    created_at        TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ       NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_submissions_user_id    ON submissions(user_id);
CREATE INDEX IF NOT EXISTS idx_submissions_problem_id ON submissions(problem_id);
CREATE INDEX IF NOT EXISTS idx_submissions_status     ON submissions(status);
CREATE INDEX IF NOT EXISTS idx_submissions_created_at ON submissions(created_at DESC);
