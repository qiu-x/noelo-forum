package page

import (
	"forumapp/session"
	"forumapp/storage"
	"forumapp/tmpl"
	"html/template"
	"log"
	"net/http"
)

func MakeUserContent(ses *session.Sessions) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		UserContent(w, r, ses)
	}
}

func UserContent(w http.ResponseWriter, r *http.Request, ses *session.Sessions) {
	log.Println("resource url:", r.URL.Path)

	user, resourceType, resourcePath, err := storage.ParseUserResourceURI(r.URL.Path)
	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	if resourceType == "post" {
		renderPost(ses, resourcePath, user, w, r)
	} else if resourceType == "comment" {
		// TODO: render standalone comments with replies (direct link to comment)
		// renderComment(resourcePath, user, w)
	}
}


func renderPost(ses *session.Sessions, resourcePath string, user string, w http.ResponseWriter, r *http.Request) {
	var page tmpl.PostPage[tmpl.TextPost]

	sessionCookie, err := r.Cookie(session.SessionCookie)
	if err == nil {
		page.Username, page.IsLoggedIn = ses.CheckAuth(sessionCookie.Value)
	}

	content, err := storage.GetPost(resourcePath, r.URL.Path, user)
	if err != nil {
		log.Println("\"post\" page generation failed:", err)
	}

    page.Content = content

	fns := template.FuncMap{
		"indent": func(comments []tmpl.Comment) []tmpl.Comment {
			for i := range comments {
				comments[i].Indentation += 20
			}
			return comments
		},
	}

	t := template.Must(template.New("").Funcs(fns).ParseFiles(
		"../templates/page.template",
		"../templates/textpost.template",
		"../templates/comments.template",
		"../templates/comment.template",
	))

	err = t.ExecuteTemplate(w, "page.template", page)
	if err != nil {
		log.Println("\"post\" page generation failed:", err)
	}
}

func MakeCommentAction(ses *session.Sessions) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		CommentAction(w, r, ses)
	}
}

func CommentAction(w http.ResponseWriter, r *http.Request, ses *session.Sessions) {
	// TODO: Add mutex
	// TODO: Clean up fragile fs logic; move to other module, with rest of fs logic

	sessionCookie, err := r.Cookie(session.SessionCookie)
	if err != nil {
		w.Write([]byte("Error: auth failed"))
		return
	}
	username, isLoggedIn := ses.CheckAuth(sessionCookie.Value)
	if !isLoggedIn {
		w.Write([]byte("Error: not logged in"))
		return
	}

	location := r.FormValue("location")
	text := r.FormValue("comment")

	err = storage.AddComment(username, text, location)
	if err != nil {
		w.Write([]byte("error: " + err.Error()))
		return
	}

	http.Redirect(w, r, r.Header.Get("Referer"), 302)
}
