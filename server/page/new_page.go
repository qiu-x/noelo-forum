package page

import (
	"forumapp/session"
	"forumapp/storage"
	"forumapp/tmpl"
	"html/template"
	"log"
	"net/http"
)
func MakeAddPostHandler(ses *session.Sessions, strg *storage.Storage) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		AddPostHandler(w, r, ses, strg)
	}
}

func AddPostHandler(w http.ResponseWriter, r *http.Request, ses *session.Sessions, strg *storage.Storage) {
	if r.Method == http.MethodPost {
		addPostAction(ses, strg, w, r)
	} else if r.Method == http.MethodGet {
		addPostPage(ses, w, r, "")
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}
}

func addPostPage(ses *session.Sessions, w http.ResponseWriter, r *http.Request, status string) {
	page := tmpl.PageBase[struct{ AddPostError string }]{
		PageName:   "addpost",
		Content:    struct{ AddPostError string }{status},
	}

	sessionCookie, err := r.Cookie(session.SessionCookie)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	page.Username, page.IsLoggedIn = ses.CheckAuth(sessionCookie.Value)

	if !page.IsLoggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	t := template.Must(template.ParseFiles("../templates/page.template", "../templates/addpost.template"))
	err = t.Execute(w, page)
	if err != nil {
		log.Println("\"addPost\" page generation failed:", err)
	}
}

func addPostAction(ses *session.Sessions, strg *storage.Storage, w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie(session.SessionCookie)
	if err != nil {
		addPostPage(ses, w, r, "Please log in")
		return
	}
	username, isLoggedIn := ses.CheckAuth(sessionCookie.Value)
	if !isLoggedIn {
		addPostPage(ses, w, r, "Please log in")
		return
	}

	title := r.FormValue("title")
	text := r.FormValue("text")

	err = strg.AddPost(username, title, text)

	if err != nil {
		log.Println("Error while adding post:", err)
		addPostPage(ses, w, r, "An unexpected error has occurred")
		return
	}

	http.Redirect(w, r, "/active", http.StatusSeeOther)
}
