package graph

import (
	"sync"

	"github.com/google/uuid"
	"posts-comments-1/internal/domain"
	"posts-comments-1/internal/storage"
)

type Resolver struct {
	Storage storage.Storage

	mu          sync.Mutex
	subscribers map[uuid.UUID]map[int]chan *domain.Comment
	nextSubID   int
}
