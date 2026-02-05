package domain

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID              uuid.UUID
	Title           string
	AuthorID          uuid.UUID
	Content         string
	CreatedAt       time.Time
	CommentsAllowed bool
}
