package pages

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type User struct{}

// Matches linkpost.template
type LinkPost struct{
	Title  string
	Link   string
	Author string
}

// Matches textpost.template
type TextPost struct{
	Title  string
	Text   string
	Author string
}

type PostType interface {
	TextPost | LinkPost
}

type PostPage[T PostType] struct{
	PageName string
	Content T
}

func (p *User) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("resource url:", r.URL.Path)
	urlparts := strings.Split(r.URL.Path, "/")

	user := urlparts[1]
	resourceURI := strings.Split(urlparts[2], ":")
	if len(resourceURI) < 2 {
		NotFoundHandler(w, r)
		return
	}
	resourceId := resourceURI[1]
	resourceType := resourceURI[0]

	resourcePath, err :=  url.JoinPath(user, resourceType, resourceId)
	resourcePath = "../storage/users/" + resourcePath

	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	resourceData, err := readFiles(resourcePath)
	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	if resourceType == "post" {
		var page PostPage[TextPost]
		page.Content = TextPost{
			Title:  resourceData["title"],
			Text:   resourceData["text"],
			Author: user,
		}

		t := template.Must(template.ParseFiles("../templates/page.template", "../templates/textpost.template"))
		err := t.Execute(w, page)
		if err != nil {
			log.Println("\"active\" page generation failed")
		}
	}
}

func readFiles(dirPath string) (map[string]string, error) {
    filesMap := make(map[string]string)

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dirPath, file.Name()))
		if err != nil {
			return nil, err
		}
		filesMap[file.Name()] = string(data)
	}

	return filesMap, nil
}
