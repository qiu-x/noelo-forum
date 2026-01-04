package storage

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// ID types
type UserID string
type PostID string
type CommentID string

// represents the direction of a vote
type VoteDirection int

const (
	VoteNone VoteDirection = iota
	VoteUp
	VoteDown
)

// ResourceRef identifies a post or comment
type ResourceRef struct {
	Kind string // "post" or "comment"
	ID   string // PostID or CommentID, including any type prefix (e.g., "post:1234", "comment:5678")
}

type User struct {
	ID            UserID
	Email         string
	DisplayName   string
	PasswordHash  string
	CreatedAt     time.Time
	EmailVerified bool
}

type Post struct {
	ID           PostID
	AuthorID     UserID
	Title        string
	Body         string
	Tags         []string
	Slug         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Score        int
	Upvotes      int
	Downvotes    int
	CommentCount int
}

type Comment struct {
	ID        CommentID
	PostID    PostID
	AuthorID  UserID
	Body      string
	CreatedAt time.Time
	ParentID  *CommentID // nil for top-level
}

type PostListQuery struct {
	Author     UserID
	Tag        string
	Limit      int
	RecentOnly bool
}

// ID helpers for CouchDB documents
// Format: type:identifier (e.g., user:alice, post:1735992000000000000, comment:1735992000000000001)

func NewUserID(username string) UserID {
	return UserID("user:" + username)
}

func UserNameFromID(id UserID) string {
	return strings.TrimPrefix(string(id), "user:")
}

func NewPostID(ts time.Time) PostID {
	return PostID("post:" + strconv.FormatInt(ts.UnixNano(), 10))
}

func NewCommentID(ts time.Time) CommentID {
	return CommentID("comment:" + strconv.FormatInt(ts.UnixNano(), 10))
}

func NewActivityID(ts time.Time, resource ResourceRef) string {
	return "activity:" + ts.Format(time.RFC3339Nano) + ":" + resource.ID
}

func NewVoteID(userID UserID, ref ResourceRef) string {
	return "vote:" + string(userID) + ":" + ref.ID
}

var (
	ErrUserExists   = errors.New("user already exists")
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrInvalidInput = errors.New("invalid input")
)
