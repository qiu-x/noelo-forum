package main

import (
	"flag"
	"forumapp/pages"
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

func main() {
	var port string

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.StringVar(&port, "port", "80", "Application Port")
	flags.StringVar(&port, "p", "80", "Application Port")
	err := flags.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	// limiter := limit.NewIPRateLimiter(0.075, 7) // For forum posts, etc...

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("../content/"))
	mux.Handle("GET /content/", http.StripPrefix("/content", fs))
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../content/favicon.ico")
	})

	mux.Handle("/", pages.MainPageHandler())
	mux.HandleFunc("GET /active", pages.ActiveSection)

	mux.HandleFunc("/login", pages.LoginHandler)
	mux.HandleFunc("/register", pages.RegisterHandler)

	mux.Handle("GET /u/", http.StripPrefix("/u", http.HandlerFunc(pages.UserContent)))

	mux.HandleFunc("POST /comment", pages.CommentAction)

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
