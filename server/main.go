package main

import (
	"flag"
	"forumapp/page"
	"forumapp/session"
	"forumapp/storage"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const HELP = ` Flags:
--port, -p
        Set the application port
--help, -h
        Print this message
`

func FileServerFilter() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Disable dir listing
		if strings.HasSuffix(r.URL.Path, "/") && r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "max-age=432000")
	})
}

func ChainedHandlers(chain ...http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, v := range chain {
			v.ServeHTTP(w, r)
		}
	})
}

func setupEndpoints(mux *http.ServeMux) {
	sessions := session.NewSessions()
	strg := &storage.Storage{}

	fs := http.FileServer(http.Dir("../content/"))
	mux.Handle("GET /content/", http.StripPrefix("/content", fs))
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../content/favicon.ico")
	})

	mux.HandleFunc("GET /feed", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../storage/feed")
	})

	mux.Handle("/", page.MainPageHandler())
	mux.HandleFunc("GET /active", page.MakeActiveHandler(sessions, strg))
	mux.HandleFunc("/login", page.MakeLoginHandler(sessions))
	mux.HandleFunc("/register", page.MakeRegisterHandler(sessions, strg))
	mux.Handle("GET /u/",
		http.StripPrefix("/u",
			http.HandlerFunc(page.MakeUserContent(sessions, strg))))
	mux.HandleFunc("POST /comment", page.MakeCommentAction(sessions, strg))
	mux.HandleFunc("/addpost", page.MakeAddPostHandler(sessions, strg))
}

func main() {
	var port string

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.StringVar(&port, "port", "80", "Application Port")
	flags.StringVar(&port, "p", "80", "Application Port")
	err := flags.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	setupEndpoints(mux)

	s := &http.Server{
		Addr:           ":" + port,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Println("Listening on port: ", port)
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %s", err.Error())
	}
}
