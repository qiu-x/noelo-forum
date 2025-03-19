package pages

import (
	"errors"
	"forumapp/session"
	"html/template"
	"log"
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		loginAction(w, r)
	} else if r.Method == http.MethodGet {
		loginPage(w, r, "")
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}
}

func loginPage(w http.ResponseWriter, r *http.Request, status string) {
	page := PageBase[struct{ LoginStatus string }]{
		Content: struct{ LoginStatus string }{status},
	}

	sessionCookie, err := r.Cookie(session.SessionCookie)
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
	log.Println("psw:", pass)

	sessionToken, err := session.Auth(username, pass)
	if err != nil {
		loginPage(w, r, "Invalid credentials")

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

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		registerAction(w, r)
	} else if r.Method == http.MethodGet {
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

	sessionCookie, err := r.Cookie(session.SessionCookie)
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
		registerPage(w, r, "Invalid registration request")
		return
	} else if errors.Is(err, session.ErrUserExists) {
		registerPage(w, r, "Account already exists")
		return
	} else if err != nil {
		log.Println("Account creation error:", err)
		registerPage(w, r, "An unexpected error has occurred") // should never happen
		return
	}

	registerPage(w, r, "")
}
