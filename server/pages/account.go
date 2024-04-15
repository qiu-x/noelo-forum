package pages

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
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
	// username := r.FormValue("uname")
	// pass := r.FormValue("psw")

	// // Sanitize username
	// username = strings.Replace(username, "/", "∕", -1)
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

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		registerAction(w, r)
	} else if r.Method == "GET" {
		registerPage(w, "none")
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}
}

func registerAction(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	username := r.FormValue("uname")
	pass := r.FormValue("psw")

	// Sanitize username
	username = strings.Replace(username, "/", "∕", -1)
	username = strings.TrimSpace(username)

	if !strings.Contains(email, "@") {
		registerPage(w, "invalid")
		return
	}

	if email == "" || pass == "" || username == "" {
		registerPage(w, "invalid")
		return
	}

	if _, err := os.Stat("../storage/users/" + username); !os.IsNotExist(err) {
		registerPage(w, "exists")
		return
	}
	registerPage(w, "success")
}

func registerPage(w http.ResponseWriter, status string) {
	page := struct {
		PageName string
		Content  struct {
			RegisterStatus string
		}
	}{}
	page.Content = struct {
		RegisterStatus string
	}{status}

	t := template.Must(template.ParseFiles("../templates/page.template", "../templates/register.template"))
	err := t.Execute(w, page)
	if err != nil {
		log.Println("\"register\" page generation failed")
	}
}
