package memory

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"posts_comments_1/internal/domain"
	"sync"
	"time"
)

var ErrPostNotFound = errors.New("post not found")
var ErrCommentNotFound = errors.New("comment not found")
var ErrCommentTooLong = errors.New("comment content exceeds maximum length")

type MemoryStorage struct {
	posts          map[uuid.UUID]domain.Post
	commentsByID   map[uuid.UUID]domain.Comment
	commentsByPost map[uuid.UUID][]uuid.UUID

	mu sync.RWMutex
}

func New() *MemoryStorage {
	return &MemoryStorage{
		posts:          make(map[uuid.UUID]domain.Post),
		commentsByID:   make(map[uuid.UUID]domain.Comment),
		commentsByPost: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (m *MemoryStorage) CreatePost(post domain.Post) error {
	m.posts[post.ID] = post
	return nil
}

func (m *MemoryStorage) GetPost(id uuid.UUID) (*domain.Post, error) {
	post, ok := m.posts[id]
	if !ok {
		return nil, fmt.Errorf("post not found")
	}
	return &post, nil
}

func (m *MemoryStorage) ListPosts(limit, offset int) ([]domain.Post, error) {
	posts := make([]domain.Post, 0)
	for _, post := range m.posts {
		posts = append(posts, post)
	}
	return posts, nil
}

// UpdatePost(post domain.Post) error
func (m *MemoryStorage) UpdatePost(post domain.Post) error {
	if _, ok := m.posts[post.ID]; !ok {
		return ErrPostNotFound
	}
	m.posts[post.ID] = post
	return nil
}

// DeletePost(id uuid.UUID) error
func (m *MemoryStorage) DeletePost(id uuid.UUID) error { //+удаление коммов
	delete(m.posts, id)
	return nil
}

func (m *MemoryStorage) CreateComment(c domain.Comment) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(c.Content) > 2000 {
		return ErrCommentTooLong
	}

	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}

	m.commentsByID[c.ID] = c
	m.commentsByPost[c.PostID] = append(m.commentsByPost[c.PostID], c.ID)
	return nil
}

// GetComment(id uuid.UUID) (*domain.Comment, error)
func (m *MemoryStorage) GetComment(id uuid.UUID) (*domain.Comment, error) {
	comment, ok := m.commentsByID[id]
	if !ok {
		return nil, fmt.Errorf("comment not found")
	}
	return &comment, nil
}

// GetComments(postID uuid.UUID, limit, offset int) ([]domain.Comment, error)
func (m *MemoryStorage) GetComments(postID uuid.UUID, limit, offset int) ([]domain.Comment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		return []domain.Comment{}, nil
	}

	ids := m.commentsByPost[postID]

	if offset >= len(ids) {
		return []domain.Comment{}, nil
	}

	end := offset + limit
	if end > len(ids) {
		end = len(ids)
	}

	pageIDs := ids[offset:end]

	result := make([]domain.Comment, 0, len(pageIDs))
	for _, id := range pageIDs {
		c, ok := m.commentsByID[id]
		if ok {
			result = append(result, c)
		}
	}
	return result, nil

}

// UpdateComment(comment domain.Comment) error
func (m *MemoryStorage) UpdateComment(c domain.Comment) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	old, ok := m.commentsByID[c.ID]
	if !ok {
		return ErrCommentNotFound
	}

	c.PostID = old.PostID
	c.ParentID = old.ParentID
	c.CreatedAt = old.CreatedAt

	if len(c.Content) > 2000 {
		return ErrCommentTooLong
	}

	m.commentsByID[c.ID] = c
	return nil
}

// DeleteComment(id uuid.UUID) error
func (m *MemoryStorage) DeleteComment(id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.commentsByID[id]
	if !ok {
		return ErrCommentNotFound
	}

	delete(m.commentsByID, id)

	ids := m.commentsByPost[c.PostID]
	for i := 0; i < len(ids); i++ {
		if ids[i] == id {
			ids = append(ids[:i], ids[i+1:]...)
			break
		}
	}

	if len(ids) == 0 {
		delete(m.commentsByPost, c.PostID)
	} else {
		m.commentsByPost[c.PostID] = ids
	}
	return nil
}

//+вложенные комментарии (индекс по parentID)
