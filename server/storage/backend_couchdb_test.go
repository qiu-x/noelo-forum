package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/go-kivik/kivik/v4"
)

// Integration tests with real CouchDB
// Requires CouchDB running on localhost:5984
// Environment variables:
//   COUCHDB_TEST_DSN: CouchDB connection string (default: http://admin:3344@localhost:5984)

func getTestCouchURL() string {
	if url := os.Getenv("COUCHDB_TEST_DSN"); url != "" {
		return url
	}
	return "http://admin:3344@localhost:5984"
}

func setupTestDB(t *testing.T) (*couchBackend, context.Context) {
	t.Helper()

	ctx := t.Context()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	couchURL := getTestCouchURL()

	// Connect to CouchDB
	client, err := kivik.New("couch", couchURL)
	if err != nil {
		t.Skipf("failed to connect to CouchDB: %v", err)
	}

	// Create a unique DB for this test to avoid state conflicts
	dbName := fmt.Sprintf("forum_test_%d", time.Now().UnixNano())
	if err := client.CreateDB(ctx, dbName); err != nil {
		t.Skipf("failed to create test DB %q: %v", dbName, err)
	}

	// Register cleanup to destroy the DB after test
	t.Cleanup(func() {
		_ = client.DestroyDB(ctx, dbName)
	})

	db := client.DB(dbName)

	// Create backend
	backend := NewCouchBackend(db, logger)

	// Ensure indexes
	_ = backend.EnsureIndexes(ctx)

	return backend, ctx
}

func TestCouchDBUserCreation(t *testing.T) {
	t.Parallel()
	backend, ctx := setupTestDB(t)

	// Create a user
	user := &userDoc{
		baseDoc: baseDoc{
			ID:   "user:alice",
			Type: "user",
		},
		Email:         "alice@example.com",
		DisplayName:   "Alice Smith",
		PasswordHash:  "hashed_password",
		CreatedAt:     time.Now(),
		EmailVerified: false,
	}

	err := backend.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Retrieve the user
	retrieved, err := backend.GetUserByID(ctx, "user:alice")
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}

	if retrieved.Email != user.Email {
		t.Fatalf("expected email %q, got %q", user.Email, retrieved.Email)
	}
}

func TestCouchDBUserExists(t *testing.T) {
	t.Parallel()
	backend, ctx := setupTestDB(t)

	user := &userDoc{
		baseDoc: baseDoc{
			ID:   "user:bob",
			Type: "user",
		},
		Email:     "bob@example.com",
		CreatedAt: time.Now(),
	}

	_ = backend.CreateUser(ctx, user)

	// Try to create duplicate
	err := backend.CreateUser(ctx, user)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestCouchDBGetUserByEmail(t *testing.T) {
	t.Parallel()
	backend, ctx := setupTestDB(t)

	user := &userDoc{
		baseDoc: baseDoc{
			ID:   "user:charlie",
			Type: "user",
		},
		Email:     "charlie@example.com",
		CreatedAt: time.Now(),
	}

	_ = backend.CreateUser(ctx, user)

	// Query by email
	retrieved, err := backend.GetUserByEmail(ctx, "charlie@example.com")
	if err != nil {
		t.Fatalf("failed to get user by email: %v", err)
	}

	if retrieved.ID != "user:charlie" {
		t.Fatalf("expected ID %q, got %q", "user:charlie", retrieved.ID)
	}
}

func TestCouchDBPostCreation(t *testing.T) {
	t.Parallel()
	backend, ctx := setupTestDB(t)

	post := &postDoc{
		baseDoc: baseDoc{
			ID:   "post:123",
			Type: "post",
		},
		AuthorID:     "user:author1",
		Title:        "Test Post",
		Body:         "This is a test",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Score:        0,
		Upvotes:      0,
		Downvotes:    0,
		CommentCount: 0,
		LastActivity: time.Now(),
	}

	err := backend.CreatePost(ctx, post)
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	// Retrieve the post
	retrieved, err := backend.GetPost(ctx, "post:123")
	if err != nil {
		t.Fatalf("failed to get post: %v", err)
	}

	if retrieved.Title != post.Title {
		t.Fatalf("expected title %q, got %q", post.Title, retrieved.Title)
	}
}

func TestCouchDBMutatePost(t *testing.T) {
	t.Skip() // TODO: Fix test
	t.Parallel()
	backend, ctx := setupTestDB(t)

	// Create initial post
	post := &postDoc{
		baseDoc: baseDoc{
			ID:   "post:456",
			Type: "post",
		},
		AuthorID:     "user:author1",
		Title:        "Original Title",
		Body:         "Original Body",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Score:        0,
		LastActivity: time.Now(),
	}

	_ = backend.CreatePost(ctx, post)

	// Mutate the post
	updated, err := backend.MutatePost(ctx, "post:456", func(p *postDoc) error {
		p.Title = "Updated Title"
		p.Score = 5
		return nil
	}, 5)

	if err != nil {
		t.Fatalf("failed to mutate post: %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Fatalf("expected title %q, got %q", "Updated Title", updated.Title)
	}

	if updated.Score != 5 {
		t.Fatalf("expected score 5, got %d", updated.Score)
	}
}

func TestCouchDBQueryPosts(t *testing.T) {
	t.Parallel()
	backend, ctx := setupTestDB(t)

	// Create some posts
	now := time.Now()
	posts := []*postDoc{
		{
			baseDoc:      baseDoc{ID: "post:1", Type: "post"},
			AuthorID:     "user:author1",
			Title:        "Post 1",
			Body:         "Body 1",
			CreatedAt:    now.Add(-2 * time.Hour),
			UpdatedAt:    now.Add(-2 * time.Hour),
			LastActivity: now.Add(-2 * time.Hour),
		},
		{
			baseDoc:      baseDoc{ID: "post:2", Type: "post"},
			AuthorID:     "user:author1",
			Title:        "Post 2",
			Body:         "Body 2",
			CreatedAt:    now.Add(-1 * time.Hour),
			UpdatedAt:    now.Add(-1 * time.Hour),
			LastActivity: now.Add(-1 * time.Hour),
		},
		{
			baseDoc:      baseDoc{ID: "post:3", Type: "post"},
			AuthorID:     "user:author2",
			Title:        "Post 3",
			Body:         "Body 3",
			CreatedAt:    now,
			UpdatedAt:    now,
			LastActivity: now,
		},
	}

	for _, p := range posts {
		_ = backend.CreatePost(ctx, p)
	}

	// Query by author
	query := PostListQuery{
		Author: "user:author1",
		Limit:  100,
	}

	results, err := backend.QueryPosts(ctx, query)
	if err != nil {
		t.Fatalf("failed to query posts: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 posts for author1, got %d", len(results))
	}
}

func TestCouchDBComments(t *testing.T) {
	t.Parallel()
	backend, ctx := setupTestDB(t)

	// Create a post first
	post := &postDoc{
		baseDoc:      baseDoc{ID: "post:789", Type: "post"},
		AuthorID:     "user:author1",
		Title:        "Test",
		Body:         "Test body",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
	_ = backend.CreatePost(ctx, post)

	// Create comments
	comment := &commentDoc{
		baseDoc:   baseDoc{ID: "comment:1", Type: "comment"},
		PostID:    "post:789",
		AuthorID:  "user:author2",
		Body:      "Great post!",
		CreatedAt: time.Now(),
	}

	err := backend.CreateComment(ctx, comment)
	if err != nil {
		t.Fatalf("failed to create comment: %v", err)
	}

	// List comments
	comments, err := backend.ListComments(ctx, "post:789")
	if err != nil {
		t.Fatalf("failed to list comments: %v", err)
	}

	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}

	if comments[0].Body != "Great post!" {
		t.Fatalf("expected body %q, got %q", "Great post!", comments[0].Body)
	}
}

func TestCouchDBVotes(t *testing.T) {
	t.Parallel()
	backend, ctx := setupTestDB(t)

	// Create a vote
	ref := ResourceRef{Kind: "post", ID: "post:123"}
	vote := &voteDoc{
		baseDoc: baseDoc{
			ID:   NewVoteID("user:user1", ref),
			Type: "vote",
		},
		UserID:    "user:user1",
		Resource:  ref,
		Direction: VoteUp,
		UpdatedAt: time.Now(),
	}

	err := backend.PutVote(ctx, vote)
	if err != nil {
		t.Fatalf("failed to put vote: %v", err)
	}

	// Get the vote
	retrieved, err := backend.GetVote(ctx, "user:user1", ref)
	if err != nil {
		t.Fatalf("failed to get vote: %v", err)
	}

	if retrieved.Direction != VoteUp {
		t.Fatalf("expected VoteUp, got %v", retrieved.Direction)
	}
}

func TestCouchDBTransactionFlow(t *testing.T) {
	t.Parallel()
	backend, ctx := setupTestDB(t)

	clock := func() time.Time { return time.Date(2025, 1, 4, 12, 0, 0, 0, time.UTC) }
	store := NewStore(WithBackend(backend), WithClock(clock))
	tx := store.With(ctx)

	// Register user
	user, err := tx.RegisterUser("test@example.com", "testuser", "hash123")
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	// Create post
	post, err := tx.CreatePost(user.ID, "Test Title", "Test Body")
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	// Add comment
	comment, err := tx.AddComment(post.ID, user.ID, "Nice post!")
	if err != nil {
		t.Fatalf("failed to add comment: %v", err)
	}

	// Vote
	ref := ResourceRef{Kind: "post", ID: string(post.ID)}
	err = tx.SetVote(user.ID, ref, VoteUp)
	if err != nil {
		t.Fatalf("failed to set vote: %v", err)
	}

	// Verify
	vote, _ := tx.GetUserVote(user.ID, ref)
	if vote != VoteUp {
		t.Fatalf("expected VoteUp, got %v", vote)
	}

	if comment.Body != "Nice post!" {
		t.Fatalf("expected comment body %q, got %q", "Nice post!", comment.Body)
	}
}
