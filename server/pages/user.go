package pages

import (
	"html/template"
	"log"
	"net/http"
)

type User struct{}

func (p *User) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	page := struct {
		PageName string
		Content  struct {
			Title  string
			Text   string
			Author string
		}
	}{}

	page.Content = struct {
		Title  string
		Text   string
		Author string
	}{"Example Post", `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus sit amet imperdiet nisl. Nulla rutrum ornare ligula ac laoreet. Nulla dignissim diam vel finibus aliquet. Duis ullamcorper est libero, non hendrerit nunc ornare quis. Suspendisse et ultricies arcu. Integer egestas odio nec orci rutrum, eget pharetra lacus congue. Sed efficitur, nibh ac placerat faucibus, augue eros venenatis massa, eu mattis quam erat vel purus. Nunc vehicula pretium ullamcorper. Aenean et dictum elit. Suspendisse tempus turpis aliquam, dignissim urna eu, posuere ex. Donec hendrerit, tortor vel facilisis rutrum, urna diam tempor magna, vel cursus urna erat nec ipsum. 
Morbi ac augue molestie, mattis tellus eget, porta felis. Pellentesque at odio in magna tempor tempus. Quisque auctor metus vitae sapien vulputate fringilla. Aliquam ex neque, molestie vel molestie at, maximus sit amet nulla. Ut iaculis a sem nec pretium. Morbi velit eros, molestie quis dui nec, cursus ullamcorper turpis. Donec id placerat orci. Morbi sed erat arcu. Pellentesque vel lectus scelerisque, lacinia tortor id, posuere diam. Aenean consectetur laoreet nunc, quis venenatis felis ullamcorper et. Maecenas hendrerit efficitur augue, sit amet auctor dui rutrum at. Mauris suscipit venenatis urna sed malesuada. Praesent posuere quam nibh, nec dictum mi porttitor et. Sed sit amet felis tellus. Maecenas posuere non ligula sit amet maximus. 
	`, "Qiu"}

	t := template.Must(template.ParseFiles("../templates/page.template", "../templates/article.template"))
	err := t.Execute(w, page)
	if err != nil {
		log.Println("\"active\" page generation failed")
	}
}
