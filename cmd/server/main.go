package main

import (
	"fmt"
	"log"
	"net/http"
	"url-shortener/internal/config"
	"url-shortener/internal/handlers"
	"url-shortener/internal/repository"
)

func main() {
	dbCollect := config.InitDB()
	repo := &repository.UrlRepository{
		Collection: dbCollect,
	}
	customHandlers := &handlers.UrlHandler{
		Repo: repo,
	}
	http.HandleFunc("/shorten", customHandlers.CreateShortUrlHandler)
	http.HandleFunc("/shorten/", customHandlers.ShortenRouteHandler)
	http.HandleFunc("/", customHandlers.RedirectHandler)
	fmt.Println("Servidor corriendo en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
