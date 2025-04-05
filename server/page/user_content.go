package page

import (
	"forumapp/session"
	"forumapp/storage"
	"forumapp/tmpl"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func MakeUserContent(ses *session.Sessions) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		UserContent(w, r, ses)
	}
}

func UserContent(w http.ResponseWriter, r *http.Request, ses *session.Sessions) {
	log.Println("resource url:", r.URL.Path)

	user, resourceType, resourcePath, err := storage.ParseUserResourceURI(r.URL.Path)
	if err != nil {
		NotFoundHandler(w, r)
		return
	}

	if resourceType == "post" {
		renderPost(ses, resourcePath, user, w, r)
	} else if resourceType == "comment" {
		// TODO: render standalone comments with replies (direct link to comment)
		// renderComment(resourcePath, user, w)
	}
}


func renderPost(ses *session.Sessions, resourcePath string, user string, w http.ResponseWriter, r *http.Request) {
	var page tmpl.PostPage[tmpl.TextPost]

	sessionCookie, err := r.Cookie(session.SessionCookie)
	if err == nil {
		page.Username, page.IsLoggedIn = ses.CheckAuth(sessionCookie.Value)
	}

	content, err := storage.GetPost(resourcePath, r.URL.Path, user)
	if err != nil {
		log.Println("\"post\" page generation failed:", err)
	}

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

func MakeCommentAction(ses *session.Sessions) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		CommentAction(w, r, ses)
	}
}

func CommentAction(w http.ResponseWriter, r *http.Request, ses *session.Sessions) {
	// TODO: Add mutex
	// TODO: Clean up fragile fs logic; move to other module, with rest of fs logic

	sessionCookie, err := r.Cookie(session.SessionCookie)
	if err != nil {
		w.Write([]byte("Error: auth failed"))
		return
	}
	username, isLoggedIn := ses.CheckAuth(sessionCookie.Value)
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

	_, _, resourcePath, err := storage.ParseUserResourceURI(location)

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
