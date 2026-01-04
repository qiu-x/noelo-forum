package storage

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Type aliases for convenience in tests

// MockBackend is a mock implementation of Backend for testing
type MockBackend struct {
	users         map[string]*userDoc
	posts         map[string]*postDoc
	comments      map[string]*commentDoc
	votes         map[string]*voteDoc
	createUserErr error
	getPostErr    error
}

func NewMockBackend() *MockBackend {
	return &MockBackend{
		users:    make(map[string]*userDoc),
		posts:    make(map[string]*postDoc),
		comments: make(map[string]*commentDoc),
		votes:    make(map[string]*voteDoc),
	}
}

func (m *MockBackend) CreateUser(ctx context.Context, doc *userDoc) error {
	if m.createUserErr != nil {
		return m.createUserErr
	}
	if _, exists := m.users[doc.ID]; exists {
		return ErrConflict
	}
	m.users[doc.ID] = doc
	return nil
}

func (m *MockBackend) GetUserByID(ctx context.Context, id UserID) (*userDoc, error) {
	doc, ok := m.users[string(id)]
	if !ok {
		return nil, ErrNotFound
	}
	return doc, nil
}

func (m *MockBackend) GetUserByEmail(ctx context.Context, email string) (*userDoc, error) {
	for _, doc := range m.users {
		if doc.Email == email {
			return doc, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockBackend) CreatePost(ctx context.Context, doc *postDoc) error {
	if _, exists := m.posts[doc.ID]; exists {
		return ErrConflict
	}
	m.posts[doc.ID] = doc
	return nil
}

func (m *MockBackend) GetPost(ctx context.Context, id PostID) (*postDoc, error) {
	if m.getPostErr != nil {
		return nil, m.getPostErr
	}
	doc, ok := m.posts[string(id)]
	if !ok {
		return nil, ErrNotFound
	}
	return doc, nil
}

func (m *MockBackend) MutatePost(
	ctx context.Context,
	id PostID,
	mutate func(*postDoc) error,
	maxRetries int,
) (*postDoc, error) {
	doc, ok := m.posts[string(id)]
	if !ok {
		return nil, ErrNotFound
	}
	if err := mutate(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (m *MockBackend) QueryPosts(ctx context.Context, q PostListQuery) ([]*postDoc, error) {
	var result []*postDoc
	for _, doc := range m.posts {
		result = append(result, doc)
	}
	return result, nil
}

func (m *MockBackend) CreateComment(ctx context.Context, doc *commentDoc) error {
	if _, exists := m.comments[doc.ID]; exists {
		return ErrConflict
	}
	m.comments[doc.ID] = doc
	return nil
}

func (m *MockBackend) ListComments(ctx context.Context, postID PostID) ([]*commentDoc, error) {
	var result []*commentDoc
	for _, doc := range m.comments {
		if doc.PostID == postID && doc.ParentID == nil {
			result = append(result, doc)
		}
	}
	return result, nil
}

func (m *MockBackend) GetVote(ctx context.Context, userID UserID, ref ResourceRef) (*voteDoc, error) {
	voteID := generateVoteID(userID, ref)
	doc, ok := m.votes[voteID]
	if !ok {
		return nil, ErrNotFound
	}
	return doc, nil
}

func (m *MockBackend) PutVote(ctx context.Context, doc *voteDoc) error {
	m.votes[doc.ID] = doc
	return nil
}

func (m *MockBackend) AddActivity(ctx context.Context, doc *activityDoc) error {
	return nil
}

func generateVoteID(userID UserID, ref ResourceRef) string {
	return NewVoteID(userID, ref)
}

// Tests

func TestVoteDocIDFormat(t *testing.T) {
	t.Parallel()
	ref := ResourceRef{Kind: "post", ID: "post:123"}
	got := NewVoteID("user:user1", ref)
	want := "vote:user:user1:post:123"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRegisterUser(t *testing.T) {
	t.Parallel()
	backend := NewMockBackend()
	clock := func() time.Time { return time.Date(2025, 1, 4, 12, 0, 0, 0, time.UTC) }
	store := NewStore(WithBackend(backend), WithClock(clock))

	ctx := t.Context()
	tx := store.With(ctx)

	// Test successful registration
	user, err := tx.RegisterUser("alice@example.com", "alice", "hash123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != "alice@example.com" {
		t.Fatalf("expected email 'alice@example.com', got %q", user.Email)
	}
	if user.ID != "user:alice" {
		t.Fatalf("expected ID 'user:alice', got %q", user.ID)
	}

	// Test duplicate user
	_, err = tx.RegisterUser("bob@example.com", "alice", "hash456")
	if !errors.Is(err, ErrUserExists) {
		t.Fatalf("expected ErrUserExists, got %v", err)
	}

	// Test invalid input
	_, err = tx.RegisterUser("", "charlie", "hash789")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for empty email, got %v", err)
	}
}

func TestGetUserByID(t *testing.T) {
	t.Parallel()
	backend := NewMockBackend()
	store := NewStore(WithBackend(backend))
	ctx := t.Context()
	tx := store.With(ctx)

	// Create a user first
	_, _ = tx.RegisterUser("alice@example.com", "alice", "hash123")

	// Retrieve it
	user, err := tx.GetUserByID("user:alice")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != "user:alice" {
		t.Fatalf("expected ID 'user:alice', got %q", user.ID)
	}

	// Try to get non-existent user
	_, err = tx.GetUserByID("nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCreatePost(t *testing.T) {
	t.Parallel()
	backend := NewMockBackend()
	store := NewStore(WithBackend(backend))
	ctx := t.Context()
	tx := store.With(ctx)

	// Create a post
	post, err := tx.CreatePost("user:author1", "Test Title", "Test Body")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if post.Title != "Test Title" {
		t.Fatalf("expected title 'Test Title', got %q", post.Title)
	}
	if post.Body != "Test Body" {
		t.Fatalf("expected body 'Test Body', got %q", post.Body)
	}
	if post.AuthorID != "user:author1" {
		t.Fatalf("expected author 'user:author1', got %q", post.AuthorID)
	}

	// Test invalid input
	_, err = tx.CreatePost("", "Title", "Body")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestSetVote(t *testing.T) {
	t.Parallel()
	backend := NewMockBackend()
	store := NewStore(WithBackend(backend))
	ctx := t.Context()
	tx := store.With(ctx)

	// Create a post
	post, _ := tx.CreatePost("user:author1", "Test", "Body")

	// Set an upvote
	ref := ResourceRef{Kind: "post", ID: string(post.ID)}
	err := tx.SetVote("user:voter1", ref, VoteUp)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check vote
	vote, err := tx.GetUserVote("user:voter1", ref)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if vote != VoteUp {
		t.Fatalf("expected VoteUp, got %v", vote)
	}

	// Change vote to downvote
	err = tx.SetVote("user:voter1", ref, VoteDown)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify change
	vote, _ = tx.GetUserVote("user:voter1", ref)
	if vote != VoteDown {
		t.Fatalf("expected VoteDown, got %v", vote)
	}

	// Test invalid input
	err = tx.SetVote("", ref, VoteUp)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestAddComment(t *testing.T) {
	t.Parallel()
	backend := NewMockBackend()
	store := NewStore(WithBackend(backend))
	ctx := t.Context()
	tx := store.With(ctx)

	// Create a post
	post, _ := tx.CreatePost("user:author1", "Test", "Body")

	// Add a comment
	comment, err := tx.AddComment(post.ID, "user:commenter1", "Great post!")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if comment.Body != "Great post!" {
		t.Fatalf("expected body 'Great post!', got %q", comment.Body)
	}
	if comment.AuthorID != "user:commenter1" {
		t.Fatalf("expected author 'user:commenter1', got %q", comment.AuthorID)
	}

	// Test invalid input
	_, err = tx.AddComment(post.ID, "", "Body")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestListPosts(t *testing.T) {
	t.Parallel()
	backend := NewMockBackend()
	store := NewStore(WithBackend(backend))
	ctx := t.Context()
	tx := store.With(ctx)

	// Create some posts
	_, _ = tx.CreatePost("user:author1", "Post 1", "Body 1")
	_, _ = tx.CreatePost("user:author1", "Post 2", "Body 2")
	_, _ = tx.CreatePost("user:author2", "Post 3", "Body 3")

	// List all posts
	posts, err := tx.ListPosts()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(posts) != 3 {
		t.Fatalf("expected 3 posts, got %d", len(posts))
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()
	backend := NewMockBackend()
	store := NewStore(WithBackend(backend))
	ctx := t.Context()
	tx := store.With(ctx)

	// Test PostOption
	post, err := tx.CreatePost("user:author1", "Title", "Body", WithTags([]string{"go", "testing"}), WithSlug("my-post"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(post.Tags) != 2 || post.Tags[0] != "go" {
		t.Fatalf("expected tags [go testing], got %v", post.Tags)
	}
	if post.Slug != "my-post" {
		t.Fatalf("expected slug 'my-post', got %q", post.Slug)
	}

	// Test UserOption
	user, _ := tx.RegisterUser("bob@example.com", "bob", "hash", WithDisplayName("Bob Smith"))
	if user.DisplayName != "Bob Smith" {
		t.Fatalf("expected display name 'Bob Smith', got %q", user.DisplayName)
	}
}
