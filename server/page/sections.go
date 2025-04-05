package page

import (
	"forumapp/session"
	"forumapp/storage"
	"forumapp/tmpl"
	"html/template"
	"log"
	"net/http"
)

func MakeActiveHandler(ses *session.Sessions) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		ActiveSection(w, r, ses)
	}
}

func ActiveSection(w http.ResponseWriter, r *http.Request, ses *session.Sessions) {
	page := tmpl.SectionPage[tmpl.ArticleItem]{
		PageName: "active",
		Content:  storage.GetAllArticles(),
	}

	sessionCookie, err := r.Cookie(session.SessionCookie)
	if err == nil {
		page.Username, page.IsLoggedIn = ses.CheckAuth(sessionCookie.Value)
	}

	t := template.Must(template.ParseFiles("../templates/page.template", "../templates/article_list.template"))
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
