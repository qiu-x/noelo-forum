package storage

// UserCmd - internal command struct for RegisterUser options
type userCmd struct {
	displayName   string
	emailVerified bool
}

type UserOption func(*userCmd)

func WithDisplayName(name string) UserOption {
	return func(c *userCmd) { c.displayName = name }
}

func WithEmailVerified() UserOption {
	return func(c *userCmd) { c.emailVerified = true }
}

// PostCmd - internal command struct for CreatePost options
type postCmd struct {
	tags          []string
	slug          string
	trackActivity bool
}

type PostOption func(*postCmd)

func WithTags(tags []string) PostOption {
	return func(c *postCmd) { c.tags = tags }
}

func WithSlug(slug string) PostOption {
	return func(c *postCmd) { c.slug = slug }
}

func WithActivityTracking() PostOption {
	return func(c *postCmd) { c.trackActivity = true }
}

// PostEditCmd - internal command struct for EditPostContent options
type postEditCmd struct {
	trackActivity bool
}

type PostEditOption func(*postEditCmd)

func WithActivityTrackingEdit() PostEditOption {
	return func(c *postEditCmd) { c.trackActivity = true }
}

// CommentCmd - internal command struct for AddComment options
type commentCmd struct {
	parentID      *CommentID
	trackActivity bool
}

type CommentOption func(*commentCmd)

func WithParentComment(parentID CommentID) CommentOption {
	return func(c *commentCmd) { c.parentID = &parentID }
}

func WithActivityTrackingComment() CommentOption {
	return func(c *commentCmd) { c.trackActivity = true }
}

// VoteCmd - internal command struct for SetVote options
type voteCmd struct {
	trackActivity bool
	maxRetries    int
}

type VoteOption func(*voteCmd)

func WithActivityTrackingVote() VoteOption {
	return func(c *voteCmd) { c.trackActivity = true }
}

func WithConflictRetry(maxRetries int) VoteOption {
	return func(c *voteCmd) { c.maxRetries = maxRetries }
}

// PostListCmd - internal command struct for ListPosts options
type postListCmd struct {
	author     UserID
	tag        string
	limit      int
	recentOnly bool
}

type PostListOption func(*postListCmd)

func WithAuthor(author UserID) PostListOption {
	return func(c *postListCmd) { c.author = author }
}

func WithTag(tag string) PostListOption {
	return func(c *postListCmd) { c.tag = tag }
}

func WithLimit(limit int) PostListOption {
	return func(c *postListCmd) { c.limit = limit }
}

func RecentOnly() PostListOption {
	return func(c *postListCmd) { c.recentOnly = true }
}
