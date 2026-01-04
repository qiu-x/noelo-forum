package storage

import (
	"context"
	"errors"
	"strings"
	"time"
)

// Store is the root interface for transactions
type Store struct {
	backend Backend
	clock   func() time.Time
}

type StoreOption func(*Store)

func WithBackend(backend Backend) StoreOption {
	return func(s *Store) { s.backend = backend }
}

func WithClock(clock func() time.Time) StoreOption {
	return func(s *Store) { s.clock = clock }
}

func NewStore(opts ...StoreOption) *Store {
	s := &Store{
		clock: time.Now,
	}
	for _, opt := range opts {
		opt(s)
	}
	if s.backend == nil {
		panic("storage.NewStore: backend is required")
	}
	return s
}

func (s *Store) With(ctx context.Context) Tx {
	return &txImpl{
		ctx:   ctx,
		store: s,
	}
}

type txImpl struct {
	ctx   context.Context
	store *Store
}

func (tx *txImpl) RegisterUser(email, username, passwordHash string, opts ...UserOption) (*User, error) {
	if email == "" || username == "" || passwordHash == "" {
		return nil, ErrInvalidInput
	}

	cmd := &userCmd{}
	for _, opt := range opts {
		opt(cmd)
	}

	doc := &userDoc{
		baseDoc: baseDoc{
			ID:   string(NewUserID(username)),
			Type: "user",
		},
		Email:         email,
		DisplayName:   cmd.displayName,
		PasswordHash:  passwordHash,
		CreatedAt:     tx.store.clock(),
		EmailVerified: cmd.emailVerified,
	}

	if err := tx.store.backend.CreateUser(tx.ctx, doc); err != nil {
		if errors.Is(err, ErrConflict) {
			return nil, ErrUserExists
		}
		return nil, err
	}

	return &User{
		ID:            UserID(doc.ID),
		Email:         doc.Email,
		DisplayName:   doc.DisplayName,
		PasswordHash:  doc.PasswordHash,
		CreatedAt:     doc.CreatedAt,
		EmailVerified: doc.EmailVerified,
	}, nil
}

func (tx *txImpl) GetUserByID(id UserID) (*User, error) {
	doc, err := tx.store.backend.GetUserByID(tx.ctx, id)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:            UserID(doc.ID),
		Email:         doc.Email,
		DisplayName:   doc.DisplayName,
		PasswordHash:  doc.PasswordHash,
		CreatedAt:     doc.CreatedAt,
		EmailVerified: doc.EmailVerified,
	}, nil
}

func (tx *txImpl) GetUserByEmail(email string) (*User, error) {
	doc, err := tx.store.backend.GetUserByEmail(tx.ctx, email)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:            UserID(doc.ID),
		Email:         doc.Email,
		DisplayName:   doc.DisplayName,
		PasswordHash:  doc.PasswordHash,
		CreatedAt:     doc.CreatedAt,
		EmailVerified: doc.EmailVerified,
	}, nil
}

func (tx *txImpl) CreatePost(authorID UserID, title, body string, opts ...PostOption) (*Post, error) {
	if authorID == "" || title == "" || body == "" {
		return nil, ErrInvalidInput
	}

	cmd := &postCmd{}
	for _, opt := range opts {
		opt(cmd)
	}

	now := tx.store.clock()
	postID := NewPostID(now)
	doc := &postDoc{
		baseDoc: baseDoc{
			ID:   string(postID),
			Type: "post",
		},
		AuthorID:     authorID,
		Title:        title,
		Body:         body,
		Tags:         cmd.tags,
		Slug:         cmd.slug,
		CreatedAt:    now,
		UpdatedAt:    now,
		Score:        0,
		Upvotes:      0,
		Downvotes:    0,
		CommentCount: 0,
		LastActivity: now,
	}

	if err := tx.store.backend.CreatePost(tx.ctx, doc); err != nil {
		return nil, err
	}

	returnDoc := &Post{
		ID:           postID,
		AuthorID:     doc.AuthorID,
		Title:        doc.Title,
		Body:         doc.Body,
		Tags:         doc.Tags,
		Slug:         doc.Slug,
		CreatedAt:    doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
		Score:        doc.Score,
		Upvotes:      doc.Upvotes,
		Downvotes:    doc.Downvotes,
		CommentCount: doc.CommentCount,
	}

	if cmd.trackActivity {
		tx.trackActivityHelper(now, ResourceRef{Kind: "post", ID: string(postID)}, authorID)
	}

	return returnDoc, nil
}

func (tx *txImpl) EditPostContent(postID PostID, title, body string, opts ...PostEditOption) (*Post, error) {
	if postID == "" || title == "" || body == "" {
		return nil, ErrInvalidInput
	}

	cmd := &postEditCmd{}
	for _, opt := range opts {
		opt(cmd)
	}

	now := tx.store.clock()
	doc, err := tx.store.backend.MutatePost(tx.ctx, postID, func(pd *postDoc) error {
		pd.Title = title
		pd.Body = body
		pd.UpdatedAt = now
		return nil
	}, 3)

	if err != nil {
		return nil, err
	}

	if cmd.trackActivity {
		tx.trackActivityHelper(now, ResourceRef{Kind: "post", ID: string(postID)}, doc.AuthorID)
	}

	return &Post{
		ID:           postID,
		AuthorID:     doc.AuthorID,
		Title:        doc.Title,
		Body:         doc.Body,
		Tags:         doc.Tags,
		Slug:         doc.Slug,
		CreatedAt:    doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
		Score:        doc.Score,
		Upvotes:      doc.Upvotes,
		Downvotes:    doc.Downvotes,
		CommentCount: doc.CommentCount,
	}, nil
}

func (tx *txImpl) GetPost(postID PostID) (*Post, error) {
	doc, err := tx.store.backend.GetPost(tx.ctx, postID)
	if err != nil {
		return nil, err
	}
	return &Post{
		ID:           postID,
		AuthorID:     doc.AuthorID,
		Title:        doc.Title,
		Body:         doc.Body,
		Tags:         doc.Tags,
		Slug:         doc.Slug,
		CreatedAt:    doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
		Score:        doc.Score,
		Upvotes:      doc.Upvotes,
		Downvotes:    doc.Downvotes,
		CommentCount: doc.CommentCount,
	}, nil
}

func (tx *txImpl) ListPosts(opts ...PostListOption) ([]Post, error) {
	cmd := &postListCmd{limit: 100} // default limit
	for _, opt := range opts {
		opt(cmd)
	}

	query := PostListQuery{
		Author:     cmd.author,
		Tag:        cmd.tag,
		Limit:      cmd.limit,
		RecentOnly: cmd.recentOnly,
	}

	docs, err := tx.store.backend.QueryPosts(tx.ctx, query)
	if err != nil {
		return nil, err
	}

	posts := make([]Post, len(docs))
	for i, doc := range docs {
		posts[i] = Post{
			ID:           PostID(doc.ID),
			AuthorID:     doc.AuthorID,
			Title:        doc.Title,
			Body:         doc.Body,
			Tags:         doc.Tags,
			Slug:         doc.Slug,
			CreatedAt:    doc.CreatedAt,
			UpdatedAt:    doc.UpdatedAt,
			Score:        doc.Score,
			Upvotes:      doc.Upvotes,
			Downvotes:    doc.Downvotes,
			CommentCount: doc.CommentCount,
		}
	}
	return posts, nil
}

func (tx *txImpl) AddComment(postID PostID, authorID UserID, body string, opts ...CommentOption) (*Comment, error) {
	if postID == "" || authorID == "" || body == "" {
		return nil, ErrInvalidInput
	}

	cmd := &commentCmd{}
	for _, opt := range opts {
		opt(cmd)
	}

	now := tx.store.clock()
	commentID := NewCommentID(now)

	doc := &commentDoc{
		baseDoc: baseDoc{
			ID:   string(commentID),
			Type: "comment",
		},
		PostID:    postID,
		AuthorID:  authorID,
		Body:      body,
		CreatedAt: now,
		ParentID:  cmd.parentID,
	}

	if err := tx.store.backend.CreateComment(tx.ctx, doc); err != nil {
		return nil, err
	}

	comment := &Comment{
		ID:        commentID,
		PostID:    postID,
		AuthorID:  authorID,
		Body:      body,
		CreatedAt: now,
		ParentID:  cmd.parentID,
	}

	// Update post comment count and last activity
	_, err := tx.store.backend.MutatePost(tx.ctx, postID, func(pd *postDoc) error {
		pd.CommentCount++
		pd.LastActivity = now
		return nil
	}, 3)
	if err != nil {
		// Accept partial failure - comment created but post not updated
		// Add logs.
		_ = err
	}

	if cmd.trackActivity {
		tx.trackActivityHelper(now, ResourceRef{Kind: "comment", ID: string(commentID)}, authorID)
	}

	return comment, nil
}

func (tx *txImpl) GetComment(commentID CommentID) (*Comment, error) {
	// For now, we don't have a backend method to get a comment by ID directly.
	// This is a limitation of the current backend design (focused on listing by post).
	// We would need to add GetCommentByID to Backend, or query all comments and filter.
	// For simplicity, return not found for now.
	return nil, ErrNotFound
}

func (tx *txImpl) ListComments(postID PostID) ([]Comment, error) {
	docs, err := tx.store.backend.ListComments(tx.ctx, postID)
	if err != nil {
		return nil, err
	}

	comments := make([]Comment, len(docs))
	for i, doc := range docs {
		comments[i] = Comment{
			ID:        CommentID(doc.ID),
			PostID:    doc.PostID,
			AuthorID:  doc.AuthorID,
			Body:      doc.Body,
			CreatedAt: doc.CreatedAt,
			ParentID:  doc.ParentID,
		}
	}
	return comments, nil
}

func (tx *txImpl) SetVote(userID UserID, ref ResourceRef, dir VoteDirection, opts ...VoteOption) error {
	if userID == "" || ref.Kind == "" || ref.ID == "" {
		return ErrInvalidInput
	}

	cmd := &voteCmd{maxRetries: 3}
	for _, opt := range opts {
		opt(cmd)
	}

	now := tx.store.clock()

	// Get existing vote
	oldVote, err := tx.store.backend.GetVote(tx.ctx, userID, ref)
	oldDir := VoteNone
	var existingRev string
	if err == nil && oldVote != nil {
		oldDir = oldVote.Direction
		existingRev = oldVote.Rev
	} else if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}

	// If no change, return early
	if oldDir == dir {
		return nil
	}

	// Create or update vote doc (include _rev for updates)
	voteID := NewVoteID(userID, ref)
	voteDoc := &voteDoc{
		baseDoc: baseDoc{
			ID:   voteID,
			Rev:  existingRev,
			Type: "vote",
		},
		UserID:    userID,
		Resource:  ref,
		Direction: dir,
		UpdatedAt: now,
	}

	if err := tx.store.backend.PutVote(tx.ctx, voteDoc); err != nil {
		return err
	}

	// Update post counters (only for posts, not comments yet)
	if ref.Kind == "post" {
		// Calculate delta
		// Accept partial failure - vote recorded but post not updated
		err := tx.updatePostCounters(ref, oldDir, dir, now, cmd)
		if err != nil {
			// ignore, add logs
			_ = err
		}
	}

	if cmd.trackActivity {
		tx.trackActivityHelper(now, ref, userID)
	}

	return nil
}

func (tx *txImpl) GetUserVote(userID UserID, ref ResourceRef) (VoteDirection, error) {
	doc, err := tx.store.backend.GetVote(tx.ctx, userID, ref)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return VoteNone, nil
		}
		return VoteNone, err
	}
	return doc.Direction, nil
}

// GetResource fetches a resource by raw ID and dispatches to the appropriate type-safe handler.
// The ID should be in the canonical form: post:123, user:alice, comment:456, etc.
// Comments are currently not directly addressable and will be treated as unsupported.
func (tx *txImpl) GetResource(id string, handlers ResourceHandlers) error {
	// Determine resource type from ID prefix
	if strings.HasPrefix(id, "post:") {
		if handlers.Post == nil {
			if handlers.Unknown != nil {
				return handlers.Unknown("post", id)
			}
			return nil
		}
		post, err := tx.GetPost(PostID(id))
		if err != nil {
			return err
		}
		return handlers.Post(post)
	}

	if strings.HasPrefix(id, "comment:") {
		// Comments are not directly addressable yet; always treat as unsupported
		if handlers.Unknown != nil {
			return handlers.Unknown("comment", id)
		}
		return ErrNotFound
	}

	if strings.HasPrefix(id, "user:") {
		if handlers.User == nil {
			if handlers.Unknown != nil {
				return handlers.Unknown("user", id)
			}
			return nil
		}
		user, err := tx.GetUserByID(UserID(id))
		if err != nil {
			return err
		}
		return handlers.User(user)
	}

	// Unknown resource type
	if handlers.Unknown != nil {
		return handlers.Unknown("unknown", id)
	}
	return ErrNotFound
}

func (tx *txImpl) updatePostCounters(
	ref ResourceRef,
	oldDir VoteDirection,
	dir VoteDirection,
	now time.Time,
	cmd *voteCmd,
) error {
	postID := PostID(ref.ID)
	_, err := tx.store.backend.MutatePost(tx.ctx, postID, func(pd *postDoc) error {
		switch oldDir {
		case VoteUp:
			pd.Upvotes--
			pd.Score--
		case VoteDown:
			pd.Downvotes--
			pd.Score++
		case VoteNone:
		default:
		}

		switch dir {
		case VoteUp:
			pd.Upvotes++
			pd.Score++
		case VoteDown:
			pd.Downvotes++
			pd.Score--
		case VoteNone:
		default:
		}

		pd.LastActivity = now
		return nil
	}, cmd.maxRetries)
	return err
}

func (tx *txImpl) trackActivityHelper(now time.Time, ref ResourceRef, userID UserID) {
	_ = tx.store.backend.AddActivity(tx.ctx, &activityDoc{
		baseDoc: baseDoc{
			ID:   NewActivityID(now, ref),
			Type: "activity",
		},
		Resource:  ref,
		ActorID:   userID,
		Timestamp: now,
	})
}
