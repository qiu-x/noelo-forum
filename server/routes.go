package main

import (
	"forumapp/page"
	"forumapp/session"
	"forumapp/storage"
	"net/http"
	"strings"
)

func FileServerFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Disable dir listing
		if strings.HasSuffix(r.URL.Path, "/") && r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ServeFile(name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, name)
	})
}

func addRoutes(mux *http.ServeMux, sessions *session.Sessions, strg *storage.Storage) {
	// Static content
	fs := http.FileServer(http.Dir("../content/"))
	mux.Handle("GET /content/", http.StripPrefix("/content", FileServerFilter(fs)))
	mux.Handle("GET /favicon.ico", ServeFile("../content/favicon.ico"))
	mux.Handle("GET /feed", ServeFile("../storage/feed"))

	// Dynamic content
	mux.Handle("/", page.MainPageHandler())
	mux.Handle("GET /active", page.ActiveHandler(sessions, strg))
	mux.Handle("/u/", http.StripPrefix("/u", page.UserContentHandler(sessions, strg)))
	mux.Handle("GET /logout", page.LogoutHandler(sessions))
	//	mux.Handle("POST /comment", page.CommentAction(sessions, strg))
	mux.Handle("POST /reply", page.ReplyAction(sessions, strg))
	mux.Handle("/login", page.LoginHandler(sessions))
	mux.Handle("/register", page.RegisterHandler(sessions, strg))
	mux.Handle("/addpost", page.AddPostHandler(sessions, strg))
}
