package pages

import (
	"html/template"
	"log"
	"net/http"
)

type Active struct{}

func (p *Active) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	page := struct {
		PageName string
		Content  []struct {
			Title  string
			Author string
		}
	}{PageName: "active"}

	page.Content = append(page.Content, struct {
		Title  string
		Author string
	}{"Example Post", "Qiu"})

	page.Content = append(page.Content, struct {
		Title  string
		Author string
	}{"Great Example Post", "Qik"})

	t := template.Must(template.ParseFiles("../templates/page.template", "../templates/article_list.template"))
	err := t.Execute(w, page)
	if err != nil {
		log.Println("\"active\" page generation failed")
	}
}
