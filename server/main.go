package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	// "forumapp/limit"
	"forumapp/pages"
)

const HELP = ` Flags:
--port, -p
        Set the application port
--help, -h
        Print this message
`

var (
	PORT    string
	API_KEY string
)

func init() {
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.StringVar(&PORT, "port", "80", "Application Port")
	flags.StringVar(&PORT, "p", "80", "Application Port")
	err := flags.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
}

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

func CheckMethod(t string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if t != r.Method {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
				http.StatusMethodNotAllowed)
			return
		}
	})
}

func ChainedHandlers(chain ...http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, v := range chain {
			v.ServeHTTP(w, r)
		}
	})
}

func MainPageHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	    if r.URL.Path != "/" {
	        NotFoundHandler(w, r, http.StatusNotFound)
	        return
	    }
		http.Redirect(w, r, "/active", http.StatusSeeOther)
	})
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request, i int) {
	w.Write([]byte("404 Not found"))
}

func main() {
	// limiter := limit.NewIPRateLimiter(0.075, 7) // For forum posts, etc...

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("../content/"))
	mux.Handle("/content/", http.StripPrefix("/content", fs))
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../content/favicon.ico")
	})

	mux.Handle("/", ChainedHandlers(CheckMethod("GET"), MainPageHandler()))
	mux.Handle("/active", ChainedHandlers(CheckMethod("GET"), &pages.MainPage{}))

	s := &http.Server{
		Addr:           ":" + PORT,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Println("Listening on port: ", PORT)
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %s", err.Error())
	}
}
