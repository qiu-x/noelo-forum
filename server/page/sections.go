package page

import (
	"forumapp/session"
	"forumapp/tmpl"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func MakeActiveHandler(ses *session.Sessions) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		ActiveSection(w, r, ses)
	}
}

func ActiveSection(w http.ResponseWriter, r *http.Request, ses *session.Sessions) {
	page := tmpl.SectionPage[tmpl.ArticleItem]{
		PageName: "active",
		Content:  getAllArticles(),
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

// Temporary hack to get all articles listed on the "active" page.
func getAllArticles() []tmpl.ArticleItem {
	var articles []tmpl.ArticleItem
	dirPath := "../storage/users"
	users, err := os.ReadDir(dirPath)
	if err != nil {
		return articles
	}

	for _, userData := range users {
		if !userData.IsDir() {
			continue
		}
		userPosts := filepath.Join(dirPath, userData.Name(), "post")
		userPostsDir, err := os.ReadDir(userPosts)
		if err != nil {
			continue
		}
		for _, v := range userPostsDir {
			title, err := os.ReadFile(filepath.Join(userPosts, v.Name(), "title"))
			if err != nil {
				continue
			}
			articles = append(articles, tmpl.ArticleItem{
				Title: string(title),
				Author: userData.Name(),
				PostLink: "/u/" + userData.Name() + "/post:" + v.Name(),
			})
		}
	}

	return articles
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
