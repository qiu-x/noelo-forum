package page

import (
	"forumapp/session"
	"forumapp/storage"
	"forumapp/tmpl"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func MakeAddPostHandler(ses *session.Sessions, strg *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		AddPostHandler(w, r, ses, strg)
	}
}

func AddPostHandler(w http.ResponseWriter, r *http.Request, ses *session.Sessions, strg *storage.Storage) {
	if r.Method == http.MethodPost {
		addPostAction(ses, strg, w, r)
	} else if r.Method == http.MethodGet {
		addPostPage(ses, w, r, "", "", "")
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
		return
	}
}

func addPostPage(ses *session.Sessions, w http.ResponseWriter, r *http.Request, status string, title string, text string) {
	page := tmpl.PageBase[struct {
		AddPostError string
		Title        string
		Text         string
	}]{
		PageName: "addpost",
		Content: struct {
			AddPostError string
			Title        string
			Text         string
		}{status, title, text},
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
		addPostPage(ses, w, r, "Please log in", "", "")
		return
	}
	username, isLoggedIn := ses.CheckAuth(sessionCookie.Value)
	if !isLoggedIn {
		addPostPage(ses, w, r, "Please log in", "", "")
		return
	}

	title := r.FormValue("title")
	text := r.FormValue("text")

	if strings.TrimSpace(title) == "" || strings.TrimSpace(text) == "" {
		log.Println("Error while adding post: title or text are empty")
		addPostPage(ses, w, r, "Please make sure both the title and text include at least one letter and aren't just empty.", title, text)
		return
	}

	if len(title) > 200 {
		log.Println("Error while adding post: title it too long, max 200 chars")
		addPostPage(ses, w, r, "Please make sure the title is up to 200 letters.", title, text)
		return
	}

	err = strg.AddPost(username, title, text)

	if err != nil {
		log.Println("Error while adding post:", err)
		addPostPage(ses, w, r, "An unexpected error has occurred", title, text)
		return
	}

	http.Redirect(w, r, "/active", http.StatusSeeOther)
}
