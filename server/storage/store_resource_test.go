package storage

import (
	"errors"
	"testing"
	"time"
)

func TestGetResourceDispatchesPost(t *testing.T) {
	t.Parallel()
	backend := NewMockBackend()
	clock := func() time.Time { return time.Date(2025, 1, 4, 12, 0, 0, 0, time.UTC) }
	store := NewStore(WithBackend(backend), WithClock(clock))
	ctx := t.Context()
	tx := store.With(ctx)

	// Dispatch to a post
	postID := PostID("post:123")
	backend.posts[string(postID)] = &postDoc{
		baseDoc:  baseDoc{ID: string(postID), Type: "post"},
		Title:    "Test Title",
		Body:     "Test Body",
		AuthorID: "user:author1",
	}

	handlerCalled := false
	err := tx.GetResource(string(postID), ResourceHandlers{
		Post: func(p *Post) error {
			handlerCalled = true
			if p.Title != "Test Title" {
				t.Fatalf("expected title 'Test Title', got %q", p.Title)
			}
			return nil
		},
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !handlerCalled {
		t.Fatalf("expected Post handler to be called")
	}
}

func TestGetResourceDispatchesUser(t *testing.T) {
	t.Parallel()
	backend := NewMockBackend()
	store := NewStore(WithBackend(backend))
	ctx := t.Context()
	tx := store.With(ctx)

	// Create a user
	_, _ = tx.RegisterUser("alice@example.com", "alice", "hash123")

	// Dispatch to it
	handlerCalled := false
	err := tx.GetResource("user:alice", ResourceHandlers{
		User: func(u *User) error {
			handlerCalled = true
			if u.Email != "alice@example.com" {
				t.Fatalf("expected email 'alice@example.com', got %q", u.Email)
			}
			return nil
		},
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !handlerCalled {
		t.Fatalf("expected User handler to be called")
	}
}

func TestGetResourceCommentUnsupported(t *testing.T) {
	t.Parallel()
	backend := NewMockBackend()
	store := NewStore(WithBackend(backend))
	ctx := t.Context()
	tx := store.With(ctx)

	// Try to dispatch to a comment (should not call Comment handler, should call Unknown or fail)
	unknownCalled := false
	err := tx.GetResource("comment:456", ResourceHandlers{
		Comment: func(c *Comment) error {
			t.Fatalf("Comment handler should not be called for unsupported comments")
			return nil
		},
		Unknown: func(kind, id string) error {
			unknownCalled = true
			if kind != "comment" {
				t.Fatalf("expected kind 'comment', got %q", kind)
			}
			return nil
		},
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !unknownCalled {
		t.Fatalf("expected Unknown handler to be called for comments")
	}
}

func TestGetResourceUnknownType(t *testing.T) {
	t.Parallel()
	backend := NewMockBackend()
	store := NewStore(WithBackend(backend))
	ctx := t.Context()
	tx := store.With(ctx)

	// Try to dispatch to unknown resource type
	unknownCalled := false
	err := tx.GetResource("vote:foo", ResourceHandlers{
		Unknown: func(kind, id string) error {
			unknownCalled = true
			if kind != "unknown" {
				t.Fatalf("expected kind 'unknown', got %q", kind)
			}
			return nil
		},
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !unknownCalled {
		t.Fatalf("expected Unknown handler to be called")
	}
}

func TestGetResourceHandlerError(t *testing.T) {
	t.Parallel()
	backend := NewMockBackend()
	clock := func() time.Time { return time.Date(2025, 1, 4, 12, 0, 0, 0, time.UTC) }
	store := NewStore(WithBackend(backend), WithClock(clock))
	ctx := t.Context()
	tx := store.With(ctx)

	// Create a post
	post, _ := tx.CreatePost("user:author1", "Test Title", "Test Body")
	postID := string(post.ID)

	// Dispatch and have the handler return an error
	testErr := ErrConflict
	err := tx.GetResource(postID, ResourceHandlers{
		Post: func(p *Post) error {
			return testErr
		},
	})

	if !errors.Is(err, testErr) {
		t.Fatalf("expected test error, got %v", err)
	}
}

func TestGetResourcerror(t *testing.T) {
	t.Parallel()
	backend := NewMockBackend()
	backend.getPostErr = ErrNotFound
	store := NewStore(WithBackend(backend))
	ctx := t.Context()
	tx := store.With(ctx)

	// Try to get a post that doesn't exist
	err := tx.GetResource("post:nonexistent", ResourceHandlers{
		Post: func(p *Post) error {
			return nil
		},
	})

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetResourceNilHandler(t *testing.T) {
	t.Parallel()
	backend := NewMockBackend()
	clock := func() time.Time { return time.Date(2025, 1, 4, 12, 0, 0, 0, time.UTC) }
	store := NewStore(WithBackend(backend), WithClock(clock))
	ctx := t.Context()
	tx := store.With(ctx)

	// Create a post
	post, _ := tx.CreatePost("user:author1", "Test Title", "Test Body")
	postID := string(post.ID)

	// Dispatch with nil Post handler and nil Unknown handler
	err := tx.GetResource(postID, ResourceHandlers{
		Post:    nil,
		Unknown: nil,
	})

	if err != nil {
		t.Fatalf("expected no error when handlers are nil, got %v", err)
	}
}
