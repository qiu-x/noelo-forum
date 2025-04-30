package tmpl

// Base for filling out page.template
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
		Title        string
		Author       string
		CreationDate string
		PostLink     string
	}

	ItemType interface {
		ArticleItem
	}
)

// User content related templates
type (
	// Matches user_page.template
	UserPage struct {
		Username           string
		LogoutButtonActive bool
	}

	// Matches comment.template
	Comment struct {
		Author       string
		CreationDate string
		Location     string
		Text         string
		Indentation  int
		Replies      []Comment
	}

	// Matches linkpost.template
	LinkPost struct {
		Location string
		Title    string
		Link     string
		Author   string
		Comments []Comment
	}

	// Matches textpost.template
	TextPost struct {
		Location      string
		Title         string
		Text          string
		Author        string
		CreationDate  string
		TextPostError string
		Comments      []Comment
	}

	PageType interface {
		UserPage | TextPost | LinkPost
	}

	UserContentPage[T PageType] = PageBase[T]
)
