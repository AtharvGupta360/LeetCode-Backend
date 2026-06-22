package models

import (
    "time"

    "github.com/lib/pq"
)

type Difficulty string

const (
    DifficultyEasy   Difficulty = "easy"
    DifficultyMedium Difficulty = "medium"
    DifficultyHard   Difficulty = "hard"
)

type Problem struct {
    ID          string         `json:"id"          db:"id"`
    Title       string         `json:"title"        db:"title"`
    Slug        string         `json:"slug"         db:"slug"`
    Description string         `json:"description"  db:"description"`
    Difficulty  Difficulty     `json:"difficulty"   db:"difficulty"`
    Tags        pq.StringArray `json:"tags"         db:"tags"`
    Examples    string         `json:"examples"     db:"examples"`
    Constraints string         `json:"constraints"  db:"constraints"`
    IsPublished bool           `json:"isPublished"  db:"is_published"`
    CreatedBy   string         `json:"createdBy"    db:"created_by"`
    CreatedAt   time.Time      `json:"createdAt"    db:"created_at"`
    UpdatedAt   time.Time      `json:"updatedAt"    db:"updated_at"`
}
