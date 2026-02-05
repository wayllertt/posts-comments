package memory

import (
	"errors"
	"testing"
	"time"

	"posts-comments-1/internal/domain"

	"github.com/google/uuid"
)

func newPost() domain.Post {
	return domain.Post{
		ID:              uuid.New(),
		Title:           "t",
		AuthorID:        uuid.New(),
		Content:         "c",
		CreatedAt:       time.Now(),
		CommentsAllowed: true,
	}
}

func newComment(postID uuid.UUID) domain.Comment {
	return domain.Comment{
		ID:        uuid.New(),
		PostID:    postID,
		AuthorID:  uuid.New(),
		ParentID:  nil,
		Content:   "hello",
		CreatedAt: time.Now(),
	}
}

func TestMemoryStorage_CreateAndGetPost(t *testing.T) {
	s := New()

	p := newPost()
	p.ID = uuid.Nil
	p.CreatedAt = time.Time{}

	if err := s.CreatePost(p); err != nil {
		t.Fatalf("CreatePost error: %v", err)
	}

	posts, err := s.ListPosts(10, 0)
	if err != nil {
		t.Fatalf("ListPosts error: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}

	got, err := s.GetPost(posts[0].ID)
	if err != nil {
		t.Fatalf("GetPost error: %v", err)
	}
	if got.Title != p.Title {
		t.Fatalf("expected title %q, got %q", p.Title, got.Title)
	}
}

func TestMemoryStorage_GetPost_NotFound(t *testing.T) {
	s := New()
	_, err := s.GetPost(uuid.New())
	if !errors.Is(err, ErrPostNotFound) {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
}

func TestMemoryStorage_ListPosts_Pagination(t *testing.T) {
	s := New()

	for i := 0; i < 5; i++ {
		p := newPost()
		p.CreatedAt = time.Now().Add(time.Duration(i) * time.Second)
		if err := s.CreatePost(p); err != nil {
			t.Fatalf("CreatePost error: %v", err)
		}
	}

	page1, err := s.ListPosts(2, 0)
	if err != nil {
		t.Fatalf("ListPosts error: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2, got %d", len(page1))
	}

	page2, _ := s.ListPosts(2, 2)
	if len(page2) != 2 {
		t.Fatalf("expected 2, got %d", len(page2))
	}

	page3, _ := s.ListPosts(2, 4)
	if len(page3) != 1 {
		t.Fatalf("expected 1, got %d", len(page3))
	}
}

func TestMemoryStorage_CreateComment_TooLong(t *testing.T) {
	s := New()
	p := newPost()
	if err := s.CreatePost(p); err != nil {
		t.Fatalf("CreatePost error: %v", err)
	}

	c := newComment(p.ID)
	c.Content = makeString(2001)

	err := s.CreateComment(c)
	if !errors.Is(err, ErrCommentTooLong) {
		t.Fatalf("expected ErrCommentTooLong, got %v", err)
	}
}

func makeString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

func TestMemoryStorage_CreateComment_ParentNotFound(t *testing.T) {
	s := New()
	p := newPost()
	if err := s.CreatePost(p); err != nil {
		t.Fatalf("CreatePost error: %v", err)
	}

	parentID := uuid.New()
	c := newComment(p.ID)
	c.ParentID = &parentID

	err := s.CreateComment(c)
	if !errors.Is(err, ErrCommentNotFound) {
		t.Fatalf("expected ErrCommentNotFound, got %v", err)
	}
}

func TestMemoryStorage_CreateComment_ParentWrongPost(t *testing.T) {
	s := New()

	p1 := newPost()
	p2 := newPost()
	if err := s.CreatePost(p1); err != nil {
		t.Fatal(err)
	}
	if err := s.CreatePost(p2); err != nil {
		t.Fatal(err)
	}

	parent := newComment(p1.ID)
	if err := s.CreateComment(parent); err != nil {
		t.Fatal(err)
	}

	child := newComment(p2.ID)
	child.ParentID = &parent.ID

	err := s.CreateComment(child)
	if !errors.Is(err, ErrParentCommentWrongPost) {
		t.Fatalf("expected ErrParentCommentWrongPost, got %v", err)
	}
}

func TestMemoryStorage_CreateComment_Disabled(t *testing.T) {
	s := New()

	p := newPost()
	p.CommentsAllowed = false
	if err := s.CreatePost(p); err != nil {
		t.Fatalf("CreatePost error: %v", err)
	}

	c := newComment(p.ID)
	err := s.CreateComment(c)
	if !errors.Is(err, ErrCommentsDisabled) {
		t.Fatalf("expected ErrCommentsDisabled, got %v", err)
	}
}

func TestMemoryStorage_GetComments_Pagination(t *testing.T) {
	s := New()
	p := newPost()
	if err := s.CreatePost(p); err != nil {
		t.Fatalf("CreatePost error: %v", err)
	}

	for i := 0; i < 5; i++ {
		c := newComment(p.ID)
		c.CreatedAt = time.Now().Add(time.Duration(i) * time.Second)
		if err := s.CreateComment(c); err != nil {
			t.Fatalf("CreateComment error: %v", err)
		}
	}

	page1, err := s.GetComments(p.ID, 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2, got %d", len(page1))
	}

	page2, _ := s.GetComments(p.ID, 2, 2)
	if len(page2) != 2 {
		t.Fatalf("expected 2, got %d", len(page2))
	}

	page3, _ := s.GetComments(p.ID, 2, 4)
	if len(page3) != 1 {
		t.Fatalf("expected 1, got %d", len(page3))
	}
}

func TestMemoryStorage_UpdatePost_NotFound(t *testing.T) {
	s := New()

	p := domain.Post{ID: uuid.New()}
	err := s.UpdatePost(p)
	if !errors.Is(err, ErrPostNotFound) {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
}

func TestMemoryStorage_DeletePost_NotFound(t *testing.T) {
	s := New()

	err := s.DeletePost(uuid.New())
	if !errors.Is(err, ErrPostNotFound) {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
}

func TestMemoryStorage_DeletePost_RemovesComments(t *testing.T) {
	s := New()

	p := newPost()
	if err := s.CreatePost(p); err != nil {
		t.Fatal(err)
	}

	c1 := newComment(p.ID)
	c2 := newComment(p.ID)
	if err := s.CreateComment(c1); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateComment(c2); err != nil {
		t.Fatal(err)
	}

	if err := s.DeletePost(p.ID); err != nil {
		t.Fatal(err)
	}

	_, err := s.GetComment(c1.ID)
	if !errors.Is(err, ErrCommentNotFound) {
		t.Fatalf("expected ErrCommentNotFound, got %v", err)
	}
}

func TestMemoryStorage_UpdateComment_NotFound(t *testing.T) {
	s := New()
	c := domain.Comment{ID: uuid.New(), Content: "x"}
	err := s.UpdateComment(c)
	if !errors.Is(err, ErrCommentNotFound) {
		t.Fatalf("expected ErrCommentNotFound, got %v", err)
	}
}

func TestMemoryStorage_DeleteComment_NotFound(t *testing.T) {
	s := New()
	err := s.DeleteComment(uuid.New())
	if !errors.Is(err, ErrCommentNotFound) {
		t.Fatalf("expected ErrCommentNotFound, got %v", err)
	}
}
