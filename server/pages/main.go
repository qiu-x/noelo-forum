package pages

import (
	"forumapp/templates"
	"net/http"
)

type MainPage struct{}

func (p *MainPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var posts []struct {
		Title  string
		Author string
	}

	posts = append(posts, struct {
		Title  string
		Author string
	}{"Example Post", "Qiu"})

	posts = append(posts, struct {
		Title  string
		Author string
	}{"Great Example Post", "Qik"})

	templates.CreateTemplate("../templates/main_page.template").Execute(w, posts)
}
