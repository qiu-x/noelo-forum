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
			Title  string
			Author string
		}
	}{PageName: "active"}

	page.Content = getAllArticles()

	t := template.Must(template.ParseFiles("../templates/page.template", "../templates/article_list.template"))
	err := t.Execute(w, page)
	if err != nil {
		log.Println("\"active\" page generation failed")
	}
}

// Temporary hack to get all articles listed on the main page
func getAllArticles() []struct {
	Title  string
	Author string
} {
	var articles []struct {
		Title  string
		Author string
	}
	dirPath := "../storage/users"
	users, err := os.ReadDir(dirPath)
	if err != nil {
		return articles
	}

	for _, useraData := range users {
		if !useraData.IsDir() {
			continue
		}
		userDataDir := filepath.Join(dirPath, useraData.Name(), "post")
		userArticles, err := os.ReadDir(userDataDir)
		if err != nil {
			continue
		}
		for _, v := range userArticles {
			title, err := os.ReadFile(filepath.Join(userDataDir, v.Name(), "title"))
			log.Println("title:", filepath.Join(userDataDir, v.Name(), "title"))
			if err != nil {
				continue
			}
			articles = append(articles, struct {
				Title  string
				Author string
			}{string(title), useraData.Name()})
		}
	}

	return articles
}
