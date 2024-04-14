package pages

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Active struct{}

func (p *Active) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	page := struct {
		PageName string
		Content  []struct {
			Title    string
			Author   string
			PostLink string
		}
	}{PageName: "active"}

	page.Content = getAllArticles()

	t := template.Must(template.ParseFiles("../templates/page.template", "../templates/article_list.template"))
	err := t.Execute(w, page)
	if err != nil {
		log.Println("\"active\" page generation failed")
	}
}

// Temporary hack to get all articles listed on the "active" page
func getAllArticles() []struct {
	Title    string
	Author   string
	PostLink string
} {
	var articles []struct {
		Title    string
		Author   string
		PostLink string
	}
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
			articles = append(articles, struct {
				Title    string
				Author   string
				PostLink string
			}{
				string(title),
				userData.Name(),
				"/u/" + userData.Name() + "/post:" + v.Name(),
			})
		}
	}

	return articles
}
