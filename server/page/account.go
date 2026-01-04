package page

import (
	"errors"
	"forumapp/session"
	"forumapp/storage"
	"forumapp/tmpl"
	"html/template"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(ses *session.Sessions, store *storage.Store) http.HandlerFunc {
	// Precompute template
	loginTemplate := template.Must(template.ParseFiles(
		"../templates/page.template",
		"../templates/login.template",
	))

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			loginAction(loginTemplate, ses, store, w, r)
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
		rawUsername, ok := ses.CheckAuth(sessionCookie.Value)
		page.IsLoggedIn = ok
		page.Username = storage.UserNameFromID(storage.UserID(rawUsername))
	}

	err = t.Execute(w, page)
	if err != nil {
		log.Println("\"login\" page generation failed:", err)
	}
}

func loginAction(
	t *template.Template,
	ses *session.Sessions,
	store *storage.Store,
	w http.ResponseWriter,
	r *http.Request,
) {
	username := r.FormValue("uname")
	pass := r.FormValue("psw")

	if username == "" || pass == "" {
		loginPage(t, ses, w, r, "Invalid username or password")
		return
	}

	ctx := r.Context()
	tx := store.With(ctx)

	// User IDs are in the format "user:username"
	user, err := tx.GetUserByID(storage.NewUserID(username))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			loginPage(t, ses, w, r, "Invalid username or password")
			return
		}
		log.Println("Login error:", err)
		loginPage(t, ses, w, r, "An unexpected error has occurred")
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash), []byte(pass),
	); err != nil {
		loginPage(t, ses, w, r, "Invalid username or password")
		return
	}

	// Create session and set cookie
	sessionToken := ses.CreateSession(string(user.ID))
	http.SetCookie(w, &http.Cookie{
		Name:     session.SessionCookie,
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 7, // 7 days
	})

	http.Redirect(w, r, "/active", http.StatusSeeOther)
}

func RegisterHandler(ses *session.Sessions, store *storage.Store) http.HandlerFunc {
	// Precompute template
	t := template.Must(template.ParseFiles(
		"../templates/page.template",
		"../templates/register.template",
	))

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			registerAction(t, ses, store, w, r)
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
		rawUsername, ok := ses.CheckAuth(sessionCookie.Value)
		page.IsLoggedIn = ok
		page.Username = storage.UserNameFromID(storage.UserID(rawUsername))
	}

	err = t.Execute(w, page)
	if err != nil {
		log.Println("\"register\" page generation failed:", err)
	}
}

func registerAction(
	t *template.Template,
	ses *session.Sessions,
	store *storage.Store,
	w http.ResponseWriter,
	r *http.Request,
) {
	email := r.FormValue("email")
	username := r.FormValue("uname")
	pass := r.FormValue("psw")

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Password hashing error:", err)
		registerPage(t, ses, w, r, "An unexpected error has occurred")
		return
	}

	// Register user via Store
	tx := store.With(r.Context())
	_, err = tx.RegisterUser(email, username, string(passwordHash))

	switch {
	case errors.Is(err, storage.ErrInvalidInput):
		registerPage(t, ses, w, r, "Invalid registration request")
		return
	case errors.Is(err, storage.ErrUserExists):
		registerPage(t, ses, w, r, "Account already exists")
		return
	case err != nil:
		log.Println("Account creation error:", err)
		registerPage(t, ses, w, r, "An unexpected error has occurred")
		return
	}

	registerPage(t, ses, w, r, "success")
}
