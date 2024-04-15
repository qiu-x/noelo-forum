package pages

import (
	"html/template"
	"log"
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		loginAction(w, r)
	} else if r.Method == "GET" {
		loginPage(w)
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}
}

func loginAction(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

func loginPage(w http.ResponseWriter) {
	page := struct {
		PageName string
		Content  struct{}
	}{}
	t := template.Must(template.ParseFiles("../templates/page.template", "../templates/login.template"))
	err := t.Execute(w, page)
	if err != nil {
		log.Println("\"login\" page generation failed")
	}
}
