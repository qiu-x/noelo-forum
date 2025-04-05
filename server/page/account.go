package page

import (
	"errors"
	"forumapp/session"
	"forumapp/storage"
	"forumapp/tmpl"
	"html/template"
	"log"
	"net/http"
)

func MakeLoginHandler(ses *session.Sessions) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		LoginHandler(w, r, ses)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request, ses *session.Sessions) {
	if r.Method == http.MethodPost {
		loginAction(ses, w, r)
	} else if r.Method == http.MethodGet {
		loginPage(ses, w, r, "")
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}
}

func loginPage(ses *session.Sessions, w http.ResponseWriter, r *http.Request, status string) {
	page := tmpl.PageBase[struct{ LoginStatus string }]{
		Content: struct{ LoginStatus string }{status},
	}

	sessionCookie, err := r.Cookie(session.SessionCookie)
	if err == nil {
		page.Username, page.IsLoggedIn = ses.CheckAuth(sessionCookie.Value)
	}

	t := template.Must(template.ParseFiles("../templates/page.template", "../templates/login.template"))
	err = t.Execute(w, page)
	if err != nil {
		log.Println("\"login\" page generation failed:", err)
	}
}

func loginAction(ses *session.Sessions, w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("uname")
	pass := r.FormValue("psw")
	log.Println("psw:", pass)

	sessionToken, err := ses.Auth(username, pass)
	if err != nil {
		loginPage(ses, w, r, "Invalid credentials")

		http.SetCookie(w, &http.Cookie{
			Name:  session.SessionCookie,
			Value: "",
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  session.SessionCookie,
		Value: sessionToken,
	})

	http.Redirect(w, r, "/active", http.StatusSeeOther)
}

func MakeRegisterHandler(ses *session.Sessions, strg *storage.Storage) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		RegisterHandler(w, r, ses, strg)
	}
}

func RegisterHandler(w http.ResponseWriter, r *http.Request, ses *session.Sessions, strg *storage.Storage) {
	if r.Method == http.MethodPost {
		registerAction(ses, strg, w, r)
	} else if r.Method == http.MethodGet {
		registerPage(ses, w, r, "")
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}
}

func registerPage(ses *session.Sessions, w http.ResponseWriter, r *http.Request, status string) {
	page := tmpl.PageBase[struct{ RegisterStatus string }]{
		Content: struct{ RegisterStatus string }{status},
	}

	sessionCookie, err := r.Cookie(session.SessionCookie)
	if err == nil {
		page.Username, page.IsLoggedIn = ses.CheckAuth(sessionCookie.Value)
	}

	t := template.Must(template.ParseFiles("../templates/page.template", "../templates/register.template"))
	err = t.Execute(w, page)
	if err != nil {
		log.Println("\"register\" page generation failed:", err)
	}
}

func registerAction(ses *session.Sessions, strg *storage.Storage, w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	username := r.FormValue("uname")
	pass := r.FormValue("psw")

	err := strg.AddUser(email, username, pass)

	if errors.Is(err, storage.ErrInvalidUserData) {
		registerPage(ses, w, r, "Invalid registration request")
		return
	} else if errors.Is(err, storage.ErrUserExists) {
		registerPage(ses, w, r, "Account already exists")
		return
	} else if err != nil {
		log.Println("Account creation error:", err)
		registerPage(ses, w, r, "An unexpected error has occurred") // should never happen
		return
	}

	registerPage(ses, w, r, "")
}
