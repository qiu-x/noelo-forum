package page

import (
	"fmt"
	"forumapp/session"
	"forumapp/storage"
	"forumapp/tmpl"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// displayUsername extracts the username from a full user ID for display purposes.
func displayUsername(userID string) string {
	return storage.UserNameFromID(storage.UserID(userID))
}

func LogoutHandler(ses *session.Sessions) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		sessionCookie, err := r.Cookie(session.SessionCookie)
		if err == nil {
			ses.Logout(sessionCookie.Value)
		}

		http.Redirect(w, r, "/active", http.StatusSeeOther)
	})
}

func UserContentPost(ses *session.Sessions, store *storage.Store) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		switch r.FormValue("type") {
		case "comment":
			CommentAction(ses, store, w, r)
		case "vote":
			VoteAction(ses, store, w, r)
		default:
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("400 bad request"))
			if err != nil {
				log.Println("Failed to write 400 response")
			}
			log.Println("400: Bad Request: Wrong `type` field of incoming request")
			return
		}
	})
}

func NormalizeResourceIDFromPath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimPrefix(path, "u/")
	path = strings.Trim(path, "/")
	return path
}

func UserContentGet(ses *session.Sessions, store *storage.Store) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		// Extract resource ID from path (normalize to canonical form)
		// Paths like: /u/posts/post:123 → posts/post:123
		resourceID := NormalizeResourceIDFromPath(r.URL.Path)

		if resourceID == "" {
			NotFoundHandler(w, r)
			return
		}

		tx := store.With(r.Context())

		// Dispatch to appropriate handler based on resource type
		err := tx.GetResource(resourceID, storage.ResourceHandlers{
			Post: func(p *storage.Post) error {
				renderPost(ses, store, string(p.ID), "", w, r)
				return nil
			},
			Comment: func(c *storage.Comment) error {
				// Comments are not directly addressable yet
				NotFoundHandler(w, r)
				return nil
			},
			User: func(u *storage.User) error {
				username := storage.UserNameFromID(u.ID)
				renderUserPage(ses, store, username, w, r)
				return nil
			},
			Unknown: func(kind, id string) error {
				// If not a recognized resource type, treat as a bare username for backward compat
				renderUserPage(ses, store, resourceID, w, r)
				return nil
			},
		})

		if err != nil {
			log.Println("GetResource failed:", err)
			NotFoundHandler(w, r)
		}
	})
}

func renderUserPage(
	ses *session.Sessions,
	store *storage.Store,
	username string,
	w http.ResponseWriter,
	r *http.Request,
) {
	var page tmpl.UserContentPage[tmpl.UserPage]

	sessionCookie, err := r.Cookie(session.SessionCookie)
	if err == nil {
		rawUsername, ok := ses.CheckAuth(sessionCookie.Value)
		page.IsLoggedIn = ok
		page.Username = displayUsername(rawUsername)
	}

	tx := store.With(r.Context())
	// Username passed as bare username, convert to proper UserID format
	userID := storage.NewUserID(username)
	posts, err := tx.ListPosts(storage.WithAuthor(userID))
	if err != nil {
		log.Println("Failed to fetch user articles:", err)
		posts = []storage.Post{}
	}

	page.Content = tmpl.UserPage{
		Username:           username,
		LogoutButtonActive: page.Username == username,
		TextPosts:          convertPostsToTextPosts(posts),
	}

	t := template.Must(template.New("").ParseFiles(
		"../templates/user_page.template",
		"../templates/article_list.template",
		"../templates/page.template",
	))

	err = t.ExecuteTemplate(w, "page.template", page)
	if err != nil {
		log.Println("\"user\" page generation failed:", err)
	}
}

func renderPost(
	ses *session.Sessions,
	store *storage.Store,
	uri string,
	status string,
	w http.ResponseWriter,
	r *http.Request,
) {
	var page tmpl.UserContentPage[tmpl.TextPost]

	sessionCookie, err := r.Cookie(session.SessionCookie)
	if err == nil {
		rawUsername, ok := ses.CheckAuth(sessionCookie.Value)
		page.IsLoggedIn = ok
		page.Username = displayUsername(rawUsername)
	}

	tx := store.With(r.Context())
	postID := storage.PostID(ParsePostID(uri))
	post, err := tx.GetPost(postID)
	if err != nil {
		log.Println("\"post\" page generation failed:", err)
		post = &storage.Post{}
	}

	content := convertStoragePostToTextPost(post)
	content.TextPostError = status
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

func VoteAction(ses *session.Sessions, store *storage.Store, w http.ResponseWriter, r *http.Request) {
	location := r.FormValue("location")
	vote_type := r.FormValue("vote_type")

	sessionCookie, err := r.Cookie(session.SessionCookie)
	if err != nil {
		log.Println("Error: auth failed")
		renderPost(ses, store, location, "auth failed", w, r)
		return
	}
	username, isLoggedIn := ses.CheckAuth(sessionCookie.Value)
	if !isLoggedIn {
		log.Println("Error: not logged in")
		renderPost(ses, store, location, "not logged in", w, r)
		return
	}

	if vote_type != "+" && vote_type != "-" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("400 bad request"))
		log.Println("400: Bad Request: Wrong `vote_type` of incoming request")
		return
	}

	tx := store.With(r.Context())
	userID := storage.UserID(username)
	postID := storage.PostID(ParsePostID(location))
	ref := storage.ResourceRef{Kind: "post", ID: string(postID)}

	// Get existing vote
	existingVote, err := tx.GetUserVote(userID, ref)
	if err != nil {
		log.Println("Error: ", err)
		renderPost(ses, store, location, "Error checking vote", w, r)
		return
	}

	// Determine vote direction
	var voteDir storage.VoteDirection
	if vote_type == "+" {
		voteDir = storage.VoteUp
	} else {
		voteDir = storage.VoteDown
	}

	// If same vote already exists, nothing to do
	if existingVote == voteDir {
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusFound)
		return
	}

	// Set the new vote
	err = tx.SetVote(userID, ref, voteDir)
	if err != nil {
		log.Println("Error setting vote:", err)
		renderPost(ses, store, location, "Error updating vote", w, r)
		return
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusFound)
}

func CommentAction(ses *session.Sessions, store *storage.Store, w http.ResponseWriter, r *http.Request) {
	location := r.FormValue("location")
	text := r.FormValue("comment")
	sessionCookie, err := r.Cookie(session.SessionCookie)

	if err != nil {
		log.Println("Error: auth failed")
		renderPost(ses, store, location, "auth failed", w, r)
		return
	}
	username, isLoggedIn := ses.CheckAuth(sessionCookie.Value)
	if !isLoggedIn {
		log.Println("Error: not logged in")
		renderPost(ses, store, location, "not logged in", w, r)
		return
	}

	if strings.TrimSpace(text) == "" {
		log.Println("Error while adding comment: text is empty")
		renderPost(ses, store, location, "Please make sure the text include at least one letter and isn't just empty.", w, r)
		return
	}

	tx := store.With(r.Context())
	postID := storage.PostID(ParsePostID(location))
	authorID := storage.UserID(username)

	comment, err := tx.AddComment(postID, authorID, text)
	if err != nil {
		log.Println("Error: ", err)
		renderPost(ses, store, location, fmt.Sprint("Error: ", err), w, r)
		return
	}

	// Log successful comment addition
	log.Println("Comment added:", comment.ID)

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusFound)
}

func ReplyAction(ses *session.Sessions, store *storage.Store) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		location := r.FormValue("location")
		user_location := r.FormValue("user_location")
		text := r.FormValue("comment")

		sessionCookie, err := r.Cookie(session.SessionCookie)
		if err != nil {
			log.Println("Error: auth failed")
			renderPost(ses, store, location, "auth failed", w, r)
			return
		}
		username, isLoggedIn := ses.CheckAuth(sessionCookie.Value)
		if !isLoggedIn {
			log.Println("Error: not logged in")
			renderPost(ses, store, location, "not logged in", w, r)
			return
		}

		if strings.TrimSpace(text) == "" {
			log.Println("Error while adding comment: text is empty")
			renderPost(ses, store, location, "Please make sure the text include at least one letter and isn't just empty.", w, r)
			return
		}

		tx := store.With(r.Context())
		postID := storage.PostID(ParsePostID(location))
		parentCommentID := storage.CommentID(ParseCommentID(user_location))
		authorID := storage.UserID(username)

		comment, err := tx.AddComment(postID, authorID, text, storage.WithParentComment(parentCommentID))
		if err != nil {
			log.Println("Error: ", err)
			renderPost(ses, store, location, fmt.Sprint("Error: ", err), w, r)
			return
		}

		// Log successful reply addition
		log.Println("Reply added:", comment.ID)

		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusFound)
	})
}

// Helper functions for type conversion

// ParsePostID normalizes a URI or path to a canonical post ID.
// Handles paths like "/u/posts/post:123" or bare IDs like "posts/post:123".
// Returns canonical form: "posts/post:123"
func ParsePostID(uri string) string {
	return NormalizeResourceIDFromPath(uri)
}

// ParseCommentID normalizes a URI or path to a canonical comment ID.
// Handles paths like "/u/comments/comment:456" or bare IDs like "comments/comment:456".
// Returns canonical form: "comments/comment:456"
func ParseCommentID(uri string) string {
	return NormalizeResourceIDFromPath(uri)
}

// convertPostsToTextPosts converts storage.Post slice to tmpl.ArticleItem slice
func convertPostsToTextPosts(posts []storage.Post) []tmpl.ArticleItem {
	articles := make([]tmpl.ArticleItem, len(posts))
	for i, post := range posts {
		articles[i] = convertStoragePostToArticleItem(&post)
	}
	return articles
}

// PostURLFromID generates a URL for a post ID
func PostURLFromID(id storage.PostID) string {
	return "/u/" + string(id)
}

// convertStoragePostToArticleItem converts a storage.Post to a tmpl.ArticleItem
func convertStoragePostToArticleItem(post *storage.Post) tmpl.ArticleItem {
	authorName := storage.UserNameFromID(post.AuthorID)
	return tmpl.ArticleItem{
		Title:        post.Title,
		Author:       authorName,
		CreationDate: post.CreatedAt.Format("2006-01-02"),
		PostLink:     PostURLFromID(post.ID),
	}
}

// convertStoragePostToTextPost converts a storage.Post to a tmpl.TextPost
func convertStoragePostToTextPost(post *storage.Post) tmpl.TextPost {
	authorName := storage.UserNameFromID(post.AuthorID)
	return tmpl.TextPost{
		Location:     PostURLFromID(post.ID),
		Title:        post.Title,
		Text:         post.Body,
		Author:       authorName,
		CreationDate: post.CreatedAt.Format("2006-01-02"),
		Votes:        strconv.Itoa(post.Score),
		Comments:     []tmpl.Comment{},
	}
}
