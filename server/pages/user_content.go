package pages

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type UserContent struct{}

// Matches commnet.template
type Comment struct {
	Author      string
	Location    string
	Text        string
	Indentation int
	Replies     []Comment
}

// Matches linkpost.template
type LinkPost struct {
	Title    string
	Link     string
	Author   string
	Comments []Comment
}

// Matches textpost.template
type TextPost struct {
	Title    string
	Text     string
	Author   string
	Comments []Comment
}

type PostType interface {
	TextPost | LinkPost
}

type PostPage[T PostType] struct {
	PageName string
	Content  T
}

func (p *UserContent) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("resource url:", r.URL.Path)

	user, resourceType, resourcePath, err := parseUserResourceURI(r.URL.Path)
	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	if resourceType == "post" {
		renderPost(resourcePath, user, w)
	} else if resourceType == "comment" {
		// TODO: render standalone comments (direct link to comment)
		// renderComments(resourcePath, user, w)
	}
}

func parseUserResourceURI(path string) (user string, resourceType string, resourcePath string, err error) {
	urlparts := strings.Split(path, "/")
	if len(urlparts) < 2 {
		err = errors.New("Invalid URI")
		return
	}
	resourceURI := strings.Split(urlparts[2], ":")
	if len(resourceURI) < 2 {
		err = errors.New("Invalid URI")
		return
	}
	user = urlparts[1]
	resourceId := resourceURI[1]
	resourceType = resourceURI[0]

	resourcePath, err = url.JoinPath(user, resourceType, resourceId)
	resourcePath = "../storage/users/" + resourcePath
	return
}

func renderPost(resourcePath string, user string, w http.ResponseWriter) {
	title, err := os.ReadFile(filepath.Join(resourcePath, "title"))
	if err != nil {
		log.Println("failed to get post title", err)
		return
	}
	text, err := os.ReadFile(filepath.Join(resourcePath, "text"))
	if err != nil {
		log.Println("failed to get post text", err)
		return
	}
	comments, err := getReplies(resourcePath, "comments")
	if err != nil && !os.IsNotExist(err) {
		log.Println("failed to get post comments", err)
		return
	}

	var page PostPage[TextPost]
	page.Content = TextPost{
		Title:    string(title),
		Text:     string(text),
		Author:   user,
		Comments: comments,
	}

	fns := template.FuncMap{
		"indent": func(comments []Comment) []Comment {
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

func parseComment(resourcePath string, user string) (Comment, bool) {
	location, err := os.ReadFile(filepath.Join(resourcePath, "location"))
	if err != nil {
		return Comment{}, false
	}
	text, err := os.ReadFile(filepath.Join(resourcePath, "text"))
	if err != nil {
		return Comment{}, false
	}
	replies, err := getReplies(resourcePath, "replies")
	if err != nil {
		return Comment{}, false
	}
	return Comment{
		Author:   user,
		Location: string(location),
		Text:     string(text),
		Replies:  replies,
	}, true
}

func getReplies(postPath, commentDirName string) ([]Comment, error) {
	var comments []Comment
	commentDir := filepath.Join(postPath, commentDirName)
	files, err := os.ReadDir(commentDir)
	if err != nil {
		return []Comment{}, err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		commentURI, err := os.ReadFile(filepath.Join(commentDir, file.Name()))
		if err != nil {
			continue
		}
		user, _, resourcePath, err := parseUserResourceURI(strings.TrimSpace(string(commentURI)))
		comment, ok := parseComment(resourcePath, user)
		if !ok {
			continue
		}
		comments = append(comments, comment)
	}
	return comments, nil
}
