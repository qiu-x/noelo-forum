package pages

import (
	"net/http"
	"log"
)

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNotFound)
	_, err := w.Write([]byte("404 Not found"))
	if err != nil {
		log.Println("Failed to server 404 page")
	}
}

