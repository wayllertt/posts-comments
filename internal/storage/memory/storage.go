package memory

import (
	"errors"
	"posts-comments-1/internal/domain"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPostNotFound           = errors.New("post not found")
	ErrParentCommentWrongPost = errors.New("parent comment belongs to another post")
	ErrCommentNotFound        = errors.New("comment not found")
	ErrCommentTooLong         = errors.New("comment content exceeds maximum length")
	ErrCommentsDisabled       = errors.New("comments aredisabled")
)

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
	m.mu.Lock()
	defer m.mu.Unlock()

	if post.ID == uuid.Nil {
		post.ID = uuid.New()
	}
	if post.CreatedAt.IsZero() {
		post.CreatedAt = time.Now()
	}

	m.posts[post.ID] = post
	return nil
}

func (m *MemoryStorage) GetPost(id uuid.UUID) (*domain.Post, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	post, ok := m.posts[id]
	if !ok {
		return nil, ErrPostNotFound
	}
	return &post, nil
}

func (m *MemoryStorage) ListPosts(limit, offset int) ([]domain.Post, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		return []domain.Post{}, nil
	}

	posts := make([]domain.Post, 0, len(m.posts))
	for _, p := range m.posts {
		posts = append(posts, p)
	}

	sort.Slice(posts, func(i, j int) bool {
		if posts[i].CreatedAt.Equal(posts[j].CreatedAt) {
			return posts[i].ID.String() < posts[j].ID.String()
		}
		return posts[i].CreatedAt.Before(posts[j].CreatedAt)
	})

	if offset >= len(posts) {
		return []domain.Post{}, nil
	}
	end := offset + limit
	if end > len(posts) {
		end = len(posts)
	}
	return posts[offset:end], nil
}

func (m *MemoryStorage) UpdatePost(post domain.Post) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.posts[post.ID]; !ok {
		return ErrPostNotFound
	}
	m.posts[post.ID] = post
	return nil
}

func (m *MemoryStorage) DeletePost(id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.posts[id]; !ok {
		return ErrPostNotFound
	}
	delete(m.posts, id)

	ids := m.commentsByPost[id]
	for _, cid := range ids {
		delete(m.commentsByID, cid)
	}
	delete(m.commentsByPost, id)

	return nil
}

func (m *MemoryStorage) CreateComment(c domain.Comment) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	post, ok := m.posts[c.PostID]
	if !ok {
		return ErrPostNotFound
	}
	if !post.CommentsAllowed {
		return ErrCommentsDisabled
	}

	if len(c.Content) > 2000 {
		return ErrCommentTooLong
	}

	if c.ParentID != nil {
		parent, ok := m.commentsByID[*c.ParentID]
		if !ok {
			return ErrCommentNotFound
		}
		if parent.PostID != c.PostID {
			return ErrParentCommentWrongPost
		}
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

func (m *MemoryStorage) GetComment(id uuid.UUID) (*domain.Comment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	comment, ok := m.commentsByID[id]
	if !ok {
		return nil, ErrCommentNotFound
	}
	return &comment, nil
}

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
