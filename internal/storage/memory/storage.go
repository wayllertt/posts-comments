package memory

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"posts_comments_1/internal/domain"
	"sort"
)

var ErrPostNotFound = errors.New("post not found")
var ErrCommentNotFound = errors.New("comment not found")

type MemoryStorage struct {
	posts    map[uuid.UUID]domain.Post
	comments map[uuid.UUID]domain.Comment
}

func NewMemoryStorage() *MemoryStorage { //	УКАЗЫВАЕМ
	return &MemoryStorage{ //берем
		posts:    make(map[uuid.UUID]domain.Post),
		comments: make(map[uuid.UUID]domain.Comment),
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

// CreateComment(comment domain.Comment) error
func (m *MemoryStorage) CreateComment(comment domain.Comment) error {
	m.comments[comment.ID] = comment
	return nil
}

// GetComment(id uuid.UUID) (*domain.Comment, error)
func (m *MemoryStorage) GetComment(id uuid.UUID) (*domain.Comment, error) {
	comment, ok := m.comments[id]
	if !ok {
		return nil, fmt.Errorf("comment not found")
	}
	return &comment, nil
}

// GetComments(postID uuid.UUID, limit, offset int) ([]domain.Comment, error)
// func (m *MemoryStorage) GetComments(postID uuid.UUID, limit, offset int) ([]domain.Comment, error) {

// }

// UpdateComment(comment domain.Comment) error
// DeleteComment(id uuid.UUID) error
