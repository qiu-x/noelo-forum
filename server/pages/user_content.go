package pages

import (
	"errors"
	"forumapp/session"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Matches commnet.template.
type Comment struct {
	Author      string
	Location    string
	Text        string
	Indentation int
	Replies     []Comment
}

// Matches linkpost.template.
type LinkPost struct {
	Location string
	Title    string
	Link     string
	Author   string
	Comments []Comment
}

// Matches textpost.template.
type TextPost struct {
	Location string
	Title    string
	Text     string
	Author   string
	Comments []Comment
}

type PostType interface {
	TextPost | LinkPost
}

type PostPage[T PostType] = PageBase[T]

func UserContent(w http.ResponseWriter, r *http.Request) {
	log.Println("resource url:", r.URL.Path)

	user, resourceType, resourcePath, err := parseUserResourceURI(r.URL.Path)
	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	if resourceType == "post" {
		renderPost(resourcePath, user, w, r)
	} else if resourceType == "comment" {
		// TODO: render standalone comments with replies (direct link to comment)
		// renderComment(resourcePath, user, w)
	}
}

func parseUserResourceURI(path string) (user string, resourceType string, resourcePath string, err error) {
	urlparts := strings.Split(path, "/")
	if len(urlparts) < 2 {
		err = errors.New("invalid URI")
		return
	}
	resourceURI := strings.Split(urlparts[2], ":")
	if len(resourceURI) < 2 {
		err = errors.New("invalid URI")
		return
	}
	user = urlparts[1]
	resourceID := resourceURI[1]
	resourceType = resourceURI[0]

	resourcePath, err = url.JoinPath(user, resourceType, resourceID)
	resourcePath = "../storage/users/" + resourcePath
	return
}

func renderPost(resourcePath string, user string, w http.ResponseWriter, r *http.Request) {
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

	sessionCookie, err := r.Cookie(session.SessionCookie)
	if err == nil {
		page.Username, page.IsLoggedIn = session.CheckAuth(sessionCookie.Value)
	}

	page.Content = TextPost{
		Location: r.URL.Path,
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
		if err != nil {
			continue
		}
		comment, ok := parseComment(resourcePath, user)
		if !ok {
			continue
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

func CommentAction(w http.ResponseWriter, r *http.Request) {
	// TODO: Add mutex
	// TODO: Clean up fragile fs logic; move to other module, with rest of fs logic

	sessionCookie, err := r.Cookie(session.SessionCookie)
	if err != nil {
		w.Write([]byte("Error: auth failed"))
		return
	}
	username, isLoggedIn := session.CheckAuth(sessionCookie.Value)
	if !isLoggedIn {
		w.Write([]byte("Error: not logged in"))
		return
	}

	location := r.FormValue("location")
	text := r.FormValue("comment")

	userdir := filepath.Join("../storage/users/", username)
	if _, err := os.Stat(userdir); os.IsNotExist(err) {
		w.Write([]byte("Error: logged in as non existing user"))
		return
	}

	commentDir, id, err := getNextName(filepath.Join(userdir, "comment"))
	if err != nil {
		w.Write([]byte("Error: failed to get next comment id, " + err.Error()))
		return
	}

	err = os.Mkdir(commentDir, 0755)
	if err != nil {
		w.Write([]byte("Error: failed create comment: " + err.Error()))
		return
	}

	err = os.Mkdir(filepath.Join(commentDir, "replies"), 0755)
	if err != nil {
		w.Write([]byte("Error: failed create comment: " + err.Error()))
		return
	}

	tf, err := os.Create(filepath.Join(commentDir, "text"))
	if err != nil {
		w.Write([]byte("Error: failed create comment: " + err.Error()))
		return
	}
	_, _ = tf.WriteString(text)
	tf.Close()

	lf, err := os.Create(filepath.Join(commentDir, "location"))
	if err != nil {
		w.Write([]byte("Error: failed create comment: " + err.Error()))
		return
	}
	_, _ = lf.WriteString(location)
	lf.Close()

	_, _, resourcePath, err := parseUserResourceURI(location)

	commentRef, _, err := getNextName(filepath.Join(resourcePath, "comments"))
	cr, err := os.Create(commentRef)
	if err != nil {
		w.Write([]byte("Error: failed create comment: " + err.Error()))
		return
	}
	_, _ = cr.WriteString("/" + username + "/comment:" + strconv.Itoa(id))
	lf.Close()

	http.Redirect(w, r, r.Header.Get("Referer"), 302)
}

func getNextName(basePath string) (string, int, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return "", 0, err
	}

	maxNum := -1
	for _, entry := range entries {
		num, err := strconv.Atoi(entry.Name())
		if err == nil && num > maxNum {
			maxNum = num
		}
	}

	return filepath.Join(basePath, strconv.Itoa(maxNum+1)), maxNum+1, nil
}
