-- 000003_create_test_cases.up.sql
-- Test cases table for the LeetCode Judge platform

CREATE TABLE IF NOT EXISTS test_cases (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    problem_id      UUID        NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    input           TEXT        NOT NULL,
    expected_output TEXT        NOT NULL,
    is_hidden       BOOLEAN     NOT NULL DEFAULT FALSE,
    order_index     INTEGER     NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_test_cases_problem_id ON test_cases(problem_id);
CREATE INDEX IF NOT EXISTS idx_test_cases_order      ON test_cases(problem_id, order_index);
