-- 000002_create_problems.up.sql
-- Problems table for the LeetCode Judge platform

CREATE TYPE difficulty_level AS ENUM ('easy', 'medium', 'hard');

CREATE TABLE IF NOT EXISTS problems (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title        VARCHAR(255) UNIQUE NOT NULL,
    slug         VARCHAR(255) UNIQUE NOT NULL,
    description  TEXT NOT NULL,
    difficulty   difficulty_level NOT NULL,
    tags         TEXT[] NOT NULL DEFAULT '{}',
    examples     TEXT NOT NULL DEFAULT '[]',
    constraints  TEXT NOT NULL DEFAULT '',
    is_published BOOLEAN NOT NULL DEFAULT FALSE,
    created_by   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_problems_difficulty   ON problems(difficulty);
CREATE INDEX IF NOT EXISTS idx_problems_slug         ON problems(slug);
CREATE INDEX IF NOT EXISTS idx_problems_is_published ON problems(is_published);
