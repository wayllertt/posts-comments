package storage

import (
	"github.com/google/uuid"
	"posts_comments_1/internal/domain"
	"time"
)

type Storage interface {
	CreatePost(post domain.Post) error
	GetPost(id uuid.UUID) (*domain.Post, error)
	ListPosts(limit, offset int) ([]domain.Post, error)
	UpdatePost(post domain.Post) error
	DeletePost(id uuid.UUID) error

	CreateComment(comment domain.Comment) error
	GetComment(id uuid.UUID) (*domain.Comment, error)
	GetComments(postID uuid.UUID, limit, offset int) ([]domain.Comment, error)
	UpdateComment(comment domain.Comment) error
	DeleteComment(id uuid.UUID) error
}
