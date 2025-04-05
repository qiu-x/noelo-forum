package tmpl

// base for filling out page.template
type PageBase[T any] struct {
	PageName   string
	Username   string
	IsLoggedIn bool
	Content    T
}

// Main view related templates
type (
    SectionPage[T ItemType] = PageBase[[]T]

    ArticleItem struct {
        Title    string
        Author   string
        PostLink string
    }

    ItemType interface {
        ArticleItem
    }
)

// User content related templates
type (
    // Matches commnet.template.
    Comment struct {
        Author      string
        Location    string
        Text        string
        Indentation int
        Replies     []Comment
    }

    // Matches linkpost.template.
    LinkPost struct {
        Location string
        Title    string
        Link     string
        Author   string
        Comments []Comment
    }

    // Matches textpost.template.
    TextPost struct {
        Location string
        Title    string
        Text     string
        Author   string
        Comments []Comment
    }

    PostType interface {
        TextPost | LinkPost
    }

    PostPage[T PostType] = PageBase[T]
)
