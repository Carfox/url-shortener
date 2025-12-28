package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// VARIABLES
var urlStore = make(map[string]UrlData)
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
var currentID = 0

// STRUCTS
type UrlData struct {
	ID        int
	Url       string
	ShortCode string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Page struct {
	Title string
	Body  []byte
}

type UrlRequest struct {
	Url string `json:"url"`
}

//  FUNCTIONS

func shortUrlGenerator(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func saveUrl(url string, code string) (UrlData, error) {
	if _, ok := urlStore[code]; ok {
		return UrlData{}, errors.New("Internal Error, Try again!")
	}
	currentID++

	newRegister := UrlData{
		ID:        currentID,
		Url:       url,
		ShortCode: code,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	urlStore[code] = newRegister

	fmt.Println("ID saved:", currentID, "->", code)
	return newRegister, nil

}

// HANDLERS

func createShortUrlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Bad Request", 400)
		return
	}

	var req UrlRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	answer, err := saveUrl(req.Url, shortUrlGenerator(6))
	if err != nil {
		http.Error(w, err.Error(), 409)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(answer)

}

func LoadUrls() {
	fmt.Println(urlStore)
}

func main() {
	http.HandleFunc("/shorten", createShortUrlHandler)
	fmt.Println("Servidor corriendo en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
