package storage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/kivik/v4/couchdb"
)

type couchBackend struct {
	db     *kivik.DB
	logger *slog.Logger
}

func NewCouchBackend(db *kivik.DB, logger *slog.Logger) *couchBackend {
	return &couchBackend{
		db:     db,
		logger: logger,
	}
}

func ConnectCouchDB(ctx context.Context, dsn, dbName string) (*kivik.DB, error) {
	client, err := kivik.New("couch", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to CouchDB: %w", err)
	}

	// Ping to verify connection
	_, err = client.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping CouchDB: %w", err)
	}

	db := client.DB(dbName)
	return db, nil
}

func (cb *couchBackend) EnsureIndexes(ctx context.Context) error {
	cb.logger.Info("Indexes should be created manually or via CouchDB admin API")
	return nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	// Check error type instead of status code
	if kivik.HTTPStatus(err) == 404 {
		return ErrNotFound
	}
	if kivik.HTTPStatus(err) == 409 {
		return ErrConflict
	}
	if kivik.HTTPStatus(err) == 400 {
		return ErrInvalidInput
	}
	return err
}

func (cb *couchBackend) CreateUser(ctx context.Context, doc *userDoc) error {
	_, err := cb.db.Put(ctx, doc.ID, doc)
	return mapError(err)
}

func (cb *couchBackend) GetUserByID(ctx context.Context, id UserID) (*userDoc, error) {
	row := cb.db.Get(ctx, string(id))
	doc := &userDoc{}
	if err := row.ScanDoc(doc); err != nil {
		return nil, mapError(err)
	}
	return doc, nil
}

func (cb *couchBackend) GetUserByEmail(ctx context.Context, email string) (*userDoc, error) {
	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"docType": "user",
			"email":   email,
		},
		"limit": 1,
	}
	rows := cb.db.Find(ctx, query)
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		return nil, ErrNotFound
	}

	doc := &userDoc{}
	if err := rows.ScanDoc(doc); err != nil {
		return nil, mapError(err)
	}
	return doc, nil
}

func (cb *couchBackend) CreatePost(ctx context.Context, doc *postDoc) error {
	_, err := cb.db.Put(ctx, doc.ID, doc)
	return mapError(err)
}

func (cb *couchBackend) GetPost(ctx context.Context, id PostID) (*postDoc, error) {
	row := cb.db.Get(ctx, string(id))
	doc := &postDoc{}
	if err := row.ScanDoc(doc); err != nil {
		return nil, mapError(err)
	}
	return doc, nil
}

func (cb *couchBackend) MutatePost(
	ctx context.Context,
	id PostID,
	mutate func(*postDoc) error,
	maxRetries int,
) (*postDoc, error) {
	var doc *postDoc
	var err error

	for attempt := range maxRetries {
		// Load current doc
		row := cb.db.Get(ctx, string(id))
		doc = &postDoc{}
		if err := row.ScanDoc(doc); err != nil {
			return nil, mapError(err)
		}

		// Apply mutation
		if err := mutate(doc); err != nil {
			return nil, err
		}

		// Try to save
		_, err = cb.db.Put(ctx, doc.ID, doc)
		if err == nil {
			return doc, nil
		}

		// Check if it's a conflict
		if kivik.HTTPStatus(err) != 409 {
			return nil, mapError(err)
		}

		// Conflict - retry
		cb.logger.Debug(
			"conflict on post mutation, retrying",
			slog.String("postID", string(id)),
			slog.Int("attempt", attempt+1),
		)
	}

	return nil, ErrConflict
}

func (cb *couchBackend) QueryPosts(ctx context.Context, q PostListQuery) ([]*postDoc, error) {
	selector := map[string]interface{}{
		"docType": "post",
	}

	if q.Author != "" {
		selector["authorId"] = string(q.Author)
	}

	// Mango query
	// curl -X POST http://localhost:5984/forum/_index -H "Content-Type: application/json" \
	//   -d '{"index": {"fields": ["docType", "createdAt"]}, "type": "json"}'
	query := map[string]interface{}{
		"selector": selector,
		"limit":    q.Limit,
	}

	rows := cb.db.Find(ctx, query)
	defer func() { _ = rows.Close() }()

	var docs []*postDoc
	for rows.Next() {
		doc := &postDoc{}
		if err := rows.ScanDoc(doc); err != nil {
			cb.logger.Warn("failed to scan post doc", slog.Any("error", err))
			continue
		}
		docs = append(docs, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, mapError(err)
	}

	return docs, nil
}

func (cb *couchBackend) CreateComment(ctx context.Context, doc *commentDoc) error {
	_, err := cb.db.Put(ctx, doc.ID, doc)
	return mapError(err)
}

func (cb *couchBackend) ListComments(ctx context.Context, postID PostID) ([]*commentDoc, error) {
	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"docType": "comment",
			"postId":  string(postID),
		},
		// Note: Sorting requires explicit indexes in CouchDB. Results returned unsorted.
	}

	rows := cb.db.Find(ctx, query)
	defer func() { _ = rows.Close() }()

	var docs []*commentDoc
	for rows.Next() {
		doc := &commentDoc{}
		if err := rows.ScanDoc(doc); err != nil {
			cb.logger.Warn("failed to scan comment doc", slog.Any("error", err))
			continue
		}
		// Filter for top-level comments only (omitempty means ParentID is nil)
		if doc.ParentID == nil {
			docs = append(docs, doc)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, mapError(err)
	}

	return docs, nil
}

func (cb *couchBackend) GetVote(ctx context.Context, userID UserID, ref ResourceRef) (*voteDoc, error) {
	voteID := NewVoteID(userID, ref)
	row := cb.db.Get(ctx, voteID)
	doc := &voteDoc{}
	if err := row.ScanDoc(doc); err != nil {
		return nil, mapError(err)
	}
	return doc, nil
}

func (cb *couchBackend) PutVote(ctx context.Context, doc *voteDoc) error {
	_, err := cb.db.Put(ctx, doc.ID, doc)
	return mapError(err)
}

func (cb *couchBackend) AddActivity(ctx context.Context, doc *activityDoc) error {
	_, _ = cb.db.Put(ctx, doc.ID, doc)
	// Ignore activity logging errors (not critical)
	return nil
}
