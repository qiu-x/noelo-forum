package storage

import (
	"context"
	"time"
)

type Backend interface {
	CreateUser(ctx context.Context, doc *userDoc) error
	GetUserByID(ctx context.Context, id UserID) (*userDoc, error)
	GetUserByEmail(ctx context.Context, email string) (*userDoc, error)

	CreatePost(ctx context.Context, doc *postDoc) error
	GetPost(ctx context.Context, id PostID) (*postDoc, error)
	MutatePost(ctx context.Context, id PostID, mutate func(*postDoc) error, maxRetries int) (*postDoc, error)
	QueryPosts(ctx context.Context, q PostListQuery) ([]*postDoc, error)

	CreateComment(ctx context.Context, doc *commentDoc) error
	ListComments(ctx context.Context, postID PostID) ([]*commentDoc, error)

	GetVote(ctx context.Context, userID UserID, ref ResourceRef) (*voteDoc, error)
	PutVote(ctx context.Context, doc *voteDoc) error

	AddActivity(ctx context.Context, doc *activityDoc) error
}

// Internal document types (not exposed to handlers)

type baseDoc struct {
	ID   string `json:"id"`
	Rev  string `json:"rev,omitempty"`
	Type string `json:"docType"` // "user", "post", "comment", "vote", "activity"
}

type userDoc struct {
	baseDoc
	Email         string    `json:"email"`
	DisplayName   string    `json:"displayName"`
	PasswordHash  string    `json:"passwordHash"`
	CreatedAt     time.Time `json:"createdAt"`
	EmailVerified bool      `json:"emailVerified"`
}

type postDoc struct {
	baseDoc
	AuthorID     UserID    `json:"authorId"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	Tags         []string  `json:"tags,omitempty"`
	Slug         string    `json:"slug,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Score        int       `json:"score"`
	Upvotes      int       `json:"upvotes"`
	Downvotes    int       `json:"downvotes"`
	CommentCount int       `json:"commentCount"`
	LastActivity time.Time `json:"lastActivity"`
}

type commentDoc struct {
	baseDoc
	PostID    PostID     `json:"postId"`
	AuthorID  UserID     `json:"authorId"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"createdAt"`
	ParentID  *CommentID `json:"parentId,omitempty"`
}

type voteDoc struct {
	baseDoc
	UserID    UserID        `json:"userId"`
	Resource  ResourceRef   `json:"resource"`
	Direction VoteDirection `json:"direction"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

type activityDoc struct {
	baseDoc
	Resource  ResourceRef `json:"resource"`
	ActorID   UserID      `json:"actorId"`
	Timestamp time.Time   `json:"timestamp"`
}
