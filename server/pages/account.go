package pages

import (
	"errors"
	"forumapp/session"
	"html/template"
	"log"
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		loginAction(w, r)
	} else if r.Method == "GET" {
		loginPage(w, r, "")
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}
}

func loginPage(w http.ResponseWriter, r *http.Request, status string) {
	page := PageBase[struct{ LoginStatus string }]{
		Content:    struct{ LoginStatus string }{status},
	}

	sessionCookie, err := r.Cookie(session.SESSION_COOKIE)
	if err == nil {
		page.Username, page.IsLoggedIn = session.CheckAuth(sessionCookie.Value)
	}

	t := template.Must(template.ParseFiles("../templates/page.template", "../templates/login.template"))
	err = t.Execute(w, page)
	if err != nil {
		log.Println("\"login\" page generation failed:", err)
	}
}

func loginAction(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("uname")
	pass := r.FormValue("psw")
	log.Println("psw:", string(pass))

	sessionToken, err := session.Auth(username, pass)
	if err != nil {
		loginPage(w, r, "Invalid credentials")

		http.SetCookie(w, &http.Cookie{
			Name:  session.SESSION_COOKIE,
			Value: "",
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  session.SESSION_COOKIE,
		Value: sessionToken,
	})

	http.Redirect(w, r, "/active", http.StatusSeeOther)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		registerAction(w, r)
	} else if r.Method == "GET" {
		registerPage(w, r, "")
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}
}

func registerPage(w http.ResponseWriter, r *http.Request, status string) {
	page := PageBase[struct{ RegisterStatus string }]{
		Content: struct{ RegisterStatus string }{status},
	}

	sessionCookie, err := r.Cookie(session.SESSION_COOKIE)
	if err == nil {
		page.Username, page.IsLoggedIn = session.CheckAuth(sessionCookie.Value)
	}

	t := template.Must(template.ParseFiles("../templates/page.template", "../templates/register.template"))
	err = t.Execute(w, page)
	if err != nil {
		log.Println("\"register\" page generation failed:", err)
	}
}

func registerAction(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	username := r.FormValue("uname")
	pass := r.FormValue("psw")

	err := session.AddUser(email, username, pass)

	if errors.Is(err, session.ErrInvalidUserData) {
		registerPage(w, r, "Invalid registartion request")
		return
	} else if errors.Is(err, session.ErrUserExists) {
		registerPage(w, r, "Account already exists")
		return
	} else if err != nil {
		log.Println("Account creation error:", err)
		registerPage(w, r, "An unexpected error has occured") // should never happen
		return
	}

	registerPage(w, r, "success")
}
