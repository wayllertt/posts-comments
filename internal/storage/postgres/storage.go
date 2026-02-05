package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"posts-comments-1/internal/domain"
)

var (
	ErrPostNotFound           = errors.New("post not found")
	ErrCommentNotFound        = errors.New("comment not found")
	ErrCommentTooLong         = errors.New("comment content exceeds maximum length")
	ErrCommentsDisabled       = errors.New("comments are disabled")
	ErrParentCommentWrongPost = errors.New("parent comment belongs to another post")
	ErrParentCommentNotFound  = errors.New("parent comment not found")
)

type Storage struct {
	db *pgxpool.Pool
}

func New() (*Storage, error) {
	dsn := dsnFromEnv()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Storage{db: pool}, nil
}

func (s *Storage) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

func dsnFromEnv() string {
	host := getenv("POSTGRES_HOST", "localhost")
	port := getenv("POSTGRES_PORT", "5432")
	user := getenv("POSTGRES_USER", "postgres")
	pass := getenv("POSTGRES_PASSWORD", "pass")
	db := getenv("POSTGRES_DB", "postgres")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func withTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)
}

func (s *Storage) CreatePost(p domain.Post) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}

	ctx, cancel := withTimeout()
	defer cancel()

	const q = `
INSERT INTO posts (id, title, content, comments_allowed, created_at)
VALUES ($1, $2, $3, $4, $5);
`
	_, err := s.db.Exec(ctx, q, p.ID, p.Title, p.Content, p.CommentsAllowed, p.CreatedAt)
	if err != nil {
		return fmt.Errorf("create post: %w", err)
	}
	return nil
}

func (s *Storage) GetPost(id uuid.UUID) (*domain.Post, error) {
	ctx, cancel := withTimeout()
	defer cancel()

	const q = `
SELECT id, title, content, comments_allowed, created_at
FROM posts
WHERE id = $1;
`
	var p domain.Post
	err := s.db.QueryRow(ctx, q, id).Scan(&p.ID, &p.Title, &p.Content, &p.CommentsAllowed, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("get post: %w", err)
	}
	return &p, nil
}

func (s *Storage) ListPosts(limit, offset int) ([]domain.Post, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		return []domain.Post{}, nil
	}

	ctx, cancel := withTimeout()
	defer cancel()

	const q = `
SELECT id, title, content, comments_allowed, created_at
FROM posts
ORDER BY created_at, id
LIMIT $1 OFFSET $2;
`
	rows, err := s.db.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list posts query: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Post, 0, limit)
	for rows.Next() {
		var p domain.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.CommentsAllowed, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("list posts scan: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list posts rows: %w", err)
	}

	return out, nil
}

func (s *Storage) UpdatePost(p domain.Post) error {
	ctx, cancel := withTimeout()
	defer cancel()

	const q = `
UPDATE posts
SET title = $2,
    content = $3,
    comments_allowed = $4
WHERE id = $1;
`
	tag, err := s.db.Exec(ctx, q, p.ID, p.Title, p.Content, p.CommentsAllowed)
	if err != nil {
		return fmt.Errorf("update post: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrPostNotFound
	}
	return nil
}

func (s *Storage) DeletePost(id uuid.UUID) error {
	ctx, cancel := withTimeout()
	defer cancel()

	const q = `DELETE FROM posts WHERE id = $1;`
	tag, err := s.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrPostNotFound
	}
	return nil
}

func (s *Storage) CreateComment(c domain.Comment) error {
	if len(c.Content) > 2000 {
		return ErrCommentTooLong
	}

	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}

	ctx, cancel := withTimeout()
	defer cancel()

	const qPost = `
SELECT comments_allowed
FROM posts
WHERE id = $1;
`
	var allowed bool
	if err := s.db.QueryRow(ctx, qPost, c.PostID).Scan(&allowed); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPostNotFound
		}
		return fmt.Errorf("check post for comment: %w", err)
	}
	if !allowed {
		return ErrCommentsDisabled
	}

	if c.ParentID != nil {
		const qParent = `
SELECT post_id
FROM comments
WHERE id = $1;
`
		var parentPostID uuid.UUID
		err := s.db.QueryRow(ctx, qParent, *c.ParentID).Scan(&parentPostID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrParentCommentNotFound
			}
			return fmt.Errorf("check parent comment: %w", err)
		}
		if parentPostID != c.PostID {
			return ErrParentCommentWrongPost
		}
	}

	const qInsert = `
INSERT INTO comments (id, post_id, parent_id, content, created_at)
VALUES ($1, $2, $3, $4, $5);
`
	_, err := s.db.Exec(ctx, qInsert, c.ID, c.PostID, c.ParentID, c.Content, c.CreatedAt)
	if err != nil {
		return fmt.Errorf("create comment: %w", err)
	}

	return nil
}

func (s *Storage) GetComment(id uuid.UUID) (*domain.Comment, error) {
	ctx, cancel := withTimeout()
	defer cancel()

	const q = `
SELECT id, post_id, parent_id, content, created_at
FROM comments
WHERE id = $1;
`
	var c domain.Comment
	err := s.db.QueryRow(ctx, q, id).Scan(&c.ID, &c.PostID, &c.ParentID, &c.Content, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCommentNotFound
		}
		return nil, fmt.Errorf("get comment: %w", err)
	}
	return &c, nil
}

func (s *Storage) GetComments(postID uuid.UUID, limit, offset int) ([]domain.Comment, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		return []domain.Comment{}, nil
	}

	ctx, cancel := withTimeout()
	defer cancel()

	const q = `
SELECT id, post_id, parent_id, content, created_at
FROM comments
WHERE post_id = $1
ORDER BY created_at, id
LIMIT $2 OFFSET $3;
`
	rows, err := s.db.Query(ctx, q, postID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get comments query: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Comment, 0, limit)
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.PostID, &c.ParentID, &c.Content, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("get comments scan: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get comments rows: %w", err)
	}

	return out, nil
}

func (s *Storage) UpdateComment(c domain.Comment) error {
	if len(c.Content) > 2000 {
		return ErrCommentTooLong
	}

	ctx, cancel := withTimeout()
	defer cancel()

	const q = `
UPDATE comments
SET content = $2
WHERE id = $1;
`
	tag, err := s.db.Exec(ctx, q, c.ID, c.Content)
	if err != nil {
		return fmt.Errorf("update comment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrCommentNotFound
	}
	return nil
}

func (s *Storage) DeleteComment(id uuid.UUID) error {
	ctx, cancel := withTimeout()
	defer cancel()

	const q = `DELETE FROM comments WHERE id = $1;`
	tag, err := s.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrCommentNotFound
	}
	return nil
}
