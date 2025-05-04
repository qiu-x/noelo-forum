package page

import (
	"fmt"
	"forumapp/session"
	"forumapp/storage"
	"forumapp/tmpl"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func LogoutHandler(ses *session.Sessions) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		sessionCookie, err := r.Cookie(session.SessionCookie)
		if err == nil {
			ses.Logout(sessionCookie.Value)
		}

		http.Redirect(w, r, "/active", http.StatusSeeOther)
	})
}

func UserContent(ses *session.Sessions, strg *storage.Storage) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		resourceType, err := storage.TypeFromURI(r.URL.Path)
		if err != nil {
			NotFoundHandler(w, r)
			return
		}

		switch resourceType {
		case storage.USER_RESOURCE:
			username := strings.FieldsFunc(r.URL.Path, func(c rune) bool {
				return c == '/'
			})[0]
			renderUserPage(ses, strg, username, w, r)
		case storage.POST_RESOURCE:
			renderPost(ses, strg, r.URL.Path, "", w, r)
		case storage.COMMENT_RESOURCE:
			// TODO: render standalone comments with replies (direct link to comment)
			// renderComment(resourcePath, user, w)
			panic("unimplemented!")
		default:
			NotFoundHandler(w, r)
			return
		}
	})
}

func renderUserPage(
	ses *session.Sessions,
	strg *storage.Storage,
	username string,
	w http.ResponseWriter,
	r *http.Request,
) {
	var page tmpl.UserContentPage[tmpl.UserPage]

	sessionCookie, err := r.Cookie(session.SessionCookie)
	if err == nil {
		page.Username, page.IsLoggedIn = ses.CheckAuth(sessionCookie.Value)
	}

	page.Content = tmpl.UserPage{
		Username:           username,
		LogoutButtonActive: page.Username == username,
		TextPosts:          strg.GetUserArticles(username),
	}

	t := template.Must(template.New("").ParseFiles(
		"../templates/user_page.template",
		"../templates/article_list.template",
		"../templates/page.template",
	))

	err = t.ExecuteTemplate(w, "page.template", page)
	if err != nil {
		log.Println("\"user\" page generation failed:", err)
	}
}

func renderPost(
	ses *session.Sessions,
	strg *storage.Storage,
	uri string,
	status string,
	w http.ResponseWriter,
	r *http.Request,
) {
	var page tmpl.UserContentPage[tmpl.TextPost]

	sessionCookie, err := r.Cookie(session.SessionCookie)
	if err == nil {
		page.Username, page.IsLoggedIn = ses.CheckAuth(sessionCookie.Value)
	}

	content, err := strg.GetPost(uri)
	if err != nil {
		log.Println("\"post\" page generation failed:", err)
	}
	content.TextPostError = status
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

func CommentAction(ses *session.Sessions, strg *storage.Storage) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		location := r.FormValue("location")
		text := r.FormValue("comment")

		sessionCookie, err := r.Cookie(session.SessionCookie)
		if err != nil {
			log.Println("Error: auth failed")
			renderPost(ses, strg, location, "auth failed", w, r)
			return
		}
		username, isLoggedIn := ses.CheckAuth(sessionCookie.Value)
		if !isLoggedIn {
			log.Println("Error: not logged in")
			renderPost(ses, strg, location, "not logged in", w, r)
			return
		}

		if strings.TrimSpace(text) == "" {
			log.Println("Error while adding comment: text is empty")
			renderPost(ses, strg, location, "Please make sure the text include at least one letter and isn't just empty.", w, r)
			return
		}

		err = strg.AddComment(username, text, location)
		if err != nil {
			log.Println("Error: ", err)
			renderPost(ses, strg, location, fmt.Sprint("Error: ", err), w, r)
			return
		}

		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusFound)
	})
}
