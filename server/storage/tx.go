package storage

// Tx is the public transaction interface for domain operations.
// Methods take context implicitly from the Tx (created via Store.With(ctx)).
type Tx interface {
	// Users
	RegisterUser(email, username, passwordHash string, opts ...UserOption) (*User, error)
	GetUserByID(id UserID) (*User, error)
	GetUserByEmail(email string) (*User, error)

	// Posts
	CreatePost(authorID UserID, title, body string, opts ...PostOption) (*Post, error)
	EditPostContent(postID PostID, title, body string, opts ...PostEditOption) (*Post, error)
	GetPost(postID PostID) (*Post, error)
	ListPosts(opts ...PostListOption) ([]Post, error)

	// Comments
	AddComment(postID PostID, authorID UserID, body string, opts ...CommentOption) (*Comment, error)
	GetComment(commentID CommentID) (*Comment, error)
	ListComments(postID PostID) ([]Comment, error)

	// Votes
	SetVote(userID UserID, ref ResourceRef, dir VoteDirection, opts ...VoteOption) error
	GetUserVote(userID UserID, ref ResourceRef) (VoteDirection, error)

	// Resources
	// GetResource fetches a resource by raw ID and dispatches to the appropriate handler.
	// Handlers are type-safe closures that receive the fully-typed resource.
	GetResource(id string, handlers ResourceHandlers) error
}

// groups type-safe callbacks for each resource type.
type ResourceHandlers struct {
	Post    func(*Post) error
	Comment func(*Comment) error
	User    func(*User) error

	Unknown func(kind, id string) error
}
