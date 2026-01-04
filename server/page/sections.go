package page

import (
	"forumapp/session"
	"forumapp/storage"
	"forumapp/tmpl"
	"html/template"
	"log"
	"net/http"
)

func ActiveHandler(ses *session.Sessions, store *storage.Store) http.Handler {
	// Precompute template
	t := template.Must(template.ParseFiles(
		"../templates/page.template",
		"../templates/active_section.template",
		"../templates/article_list.template",
	))

	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		tx := store.With(r.Context())
		posts, err := tx.ListPosts(storage.RecentOnly(), storage.WithLimit(10))
		if err != nil {
			log.Println("Failed to fetch recent posts:", err)
			posts = []storage.Post{}
		}

		article_list_tmpl := tmpl.ActiveSection{
			TextPosts: convertPostsToTextPosts(posts),
		}

		page := tmpl.SectionPage[tmpl.ActiveSection]{
			PageName: "active",
			Content:  article_list_tmpl,
		}

		sessionCookie, err := r.Cookie(session.SessionCookie)
		if err == nil {
			rawUsername, ok := ses.CheckAuth(sessionCookie.Value)
			page.IsLoggedIn = ok
			page.Username = storage.UserNameFromID(storage.UserID(rawUsername))
		}

		err = t.Execute(w, page)
		if err != nil {
			log.Println("\"active\" page generation failed:", err)
		}
	})
}

func MainPageHandler() http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
				http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/" {
			log.Println("Bad url:", r.URL.Path)
			NotFoundHandler(w, r)
			return
		}
		http.Redirect(w, r, "/active", http.StatusSeeOther)
	})
}
