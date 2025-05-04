package page

import (
	"forumapp/session"
	"forumapp/storage"
	"forumapp/tmpl"
	"html/template"
	"log"
	"net/http"
)

func MakeActiveHandler(ses *session.Sessions, strg *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ActiveSection(w, r, ses, strg)
	}
}

func ActiveSection(w http.ResponseWriter, r *http.Request, ses *session.Sessions, strg *storage.Storage) {
	article_list_tmpl := tmpl.ActiveSection{
		TextPosts: strg.GetRecentlyActive(10),
	}

	page := tmpl.SectionPage[tmpl.ActiveSection]{
		PageName: "active",
		Content:  article_list_tmpl,
	}

	sessionCookie, err := r.Cookie(session.SessionCookie)
	if err == nil {
		page.Username, page.IsLoggedIn = ses.CheckAuth(sessionCookie.Value)
	}

	t := template.Must(template.ParseFiles("../templates/page.template", "../templates/active_section.template", "../templates/article_list.template"))
	err = t.Execute(w, page)
	if err != nil {
		log.Println("\"active\" page generation failed:", err)
	}
}

func MainPageHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Bad url:", r.URL.Path)
		if r.URL.Path != "/" {
			NotFoundHandler(w, r)
			return
		}
		http.Redirect(w, r, "/active", http.StatusSeeOther)
	})
}
