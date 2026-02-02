package postscomments1

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID        uuid.UUID
	Title     string
	Author    uuid.UUID
	Content   string
	CreatedAt time.Time
}
