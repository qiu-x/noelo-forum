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

func LoginHandler(ses *session.Sessions) http.HandlerFunc {
	// Precompute template
	loginTemplate := template.Must(template.ParseFiles(
		"../templates/page.template",
		"../templates/login.template",
	))

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			loginAction(loginTemplate, ses, w, r)
		case http.MethodGet:
			loginPage(loginTemplate, ses, w, r, "")
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
				http.StatusMethodNotAllowed)
		}
	}
}

func loginPage(
	t *template.Template,
	ses *session.Sessions,
	w http.ResponseWriter,
	r *http.Request,
	status string,
) {
	page := tmpl.PageBase[struct{ LoginStatus string }]{
		Content: struct{ LoginStatus string }{status},
	}

	sessionCookie, err := r.Cookie(session.SessionCookie)
	if err == nil {
		page.Username, page.IsLoggedIn = ses.CheckAuth(sessionCookie.Value)
	}

	err = t.Execute(w, page)
	if err != nil {
		log.Println("\"login\" page generation failed:", err)
	}
}

func loginAction(
	t *template.Template,
	ses *session.Sessions,
	w http.ResponseWriter,
	r *http.Request,
) {
	username := r.FormValue("uname")
	pass := r.FormValue("psw")

	sessionToken, err := ses.Auth(username, pass)
	if err != nil {
		loginPage(t, ses, w, r, "Invalid credentials")

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

func RegisterHandler(ses *session.Sessions, strg *storage.Storage) http.HandlerFunc {
	// Precompute template
	t := template.Must(template.ParseFiles(
		"../templates/page.template",
		"../templates/register.template",
	))

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			registerAction(t, ses, strg, w, r)
		case http.MethodGet:
			registerPage(t, ses, w, r, "")
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
				http.StatusMethodNotAllowed)
		}
	}
}

func registerPage(
	t *template.Template,
	ses *session.Sessions,
	w http.ResponseWriter,
	r *http.Request,
	status string,
) {
	page := tmpl.PageBase[struct{ RegisterStatus string }]{
		Content: struct{ RegisterStatus string }{status},
	}

	sessionCookie, err := r.Cookie(session.SessionCookie)
	if err == nil {
		page.Username, page.IsLoggedIn = ses.CheckAuth(sessionCookie.Value)
	}

	err = t.Execute(w, page)
	if err != nil {
		log.Println("\"register\" page generation failed:", err)
	}
}

func registerAction(
	t *template.Template,
	ses *session.Sessions,
	strg *storage.Storage,
	w http.ResponseWriter,
	r *http.Request,
) {
	email := r.FormValue("email")
	username := r.FormValue("uname")
	pass := r.FormValue("psw")

	err := strg.AddUser(email, username, pass)

	if errors.Is(err, storage.ErrInvalidUserData) {
		registerPage(t, ses, w, r, "Invalid registration request")
		return
	} else if errors.Is(err, storage.ErrUserExists) {
		registerPage(t, ses, w, r, "Account already exists")
		return
	} else if err != nil {
		log.Println("Account creation error:", err)
		registerPage(t, ses, w, r, "An unexpected error has occurred") // should never happen
		return
	}

	registerPage(t, ses, w, r, "success")
}
