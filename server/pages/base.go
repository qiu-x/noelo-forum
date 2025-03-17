package pages

type PageBase[T any] struct {
	PageName   string
	Username   string
	IsLoggedIn bool
	Content T
}
