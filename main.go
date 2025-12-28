package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

// VARIABLES
var (
	urlStore   = make(map[string]UrlData)
	storeMutex sync.RWMutex
	letters    = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	currentID  = 0
)

// STRUCTS
type UrlData struct {
	ID        int       `json:"id"`
	Url       string    `json:"url"`
	ShortCode string    `json:"short_code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Clicks    int       `json:"access_count"`
}

type UrlResponse struct {
	ID        int       `json:"id"`
	Url       string    `json:"url"`
	ShortCode string    `json:"short_code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

func saveUrl(url string) (UrlData, error) {
	storeMutex.Lock()
	defer storeMutex.Unlock()

	for {

		code := shortUrlGenerator(6)

		if _, exists := urlStore[code]; exists {
			continue // colisión → reintentar
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

}

func updateUrl(code string, newUrl string) (UrlData, bool) {
	storeMutex.Lock()
	defer storeMutex.Unlock()

	data, ok := urlStore[code]
	if !ok {
		return UrlData{}, false
	}

	data.Url = newUrl
	data.UpdatedAt = time.Now()
	urlStore[code] = data

	return data, true
}

func deleteUrl(code string) {
	storeMutex.Lock()
	defer storeMutex.Unlock()

	delete(urlStore, code)

}

func getUrl(code string) (UrlData, bool) {
	storeMutex.RLock()
	defer storeMutex.RUnlock()

	data, ok := urlStore[code]
	return data, ok
}

func getAndIncrementClick(code string) (UrlData, bool) {
	storeMutex.Lock()
	defer storeMutex.Unlock()

	data, ok := urlStore[code]
	if !ok {
		return UrlData{}, false
	}

	data.Clicks++
	urlStore[code] = data

	return data, true
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

	if !strings.HasPrefix(req.Url, "http://") && !strings.HasPrefix(req.Url, "https://") {
		req.Url = "https://" + req.Url
	}

	answer, err := saveUrl(req.Url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(answer)

}

func redirectHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Bad Request", 400)
		return
	}

	code := r.URL.Path[1:]

	data, ok := getAndIncrementClick(code)

	if !ok {
		http.Error(w, "URL not found!", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, data.Url, http.StatusFound)
}

func shortenRouteHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")

	// "/shorten/ - /<code>/ - /stats/"
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	code := parts[2] // /<code>/

	data, ok := getUrl(code)
	if !ok {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	// GET shorten/<code>/stats
	if len(parts) > 3 && parts[3] == "stats" {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
		return
	}

	// PUT shorten/<code>/stats\
	if r.Method == http.MethodPut {
		var req UrlRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		data, ok = updateUrl(code, req.Url)

		if !ok {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
		return
	}

	//GET /shorten/<code>
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		answer := UrlResponse{
			ID:        data.ID,
			Url:       data.Url,
			ShortCode: data.ShortCode,
			CreatedAt: data.CreatedAt,
			UpdatedAt: data.UpdatedAt,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(answer)
		return
	}

	if r.Method == http.MethodDelete {
		deleteUrl(code)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// If method doesnt exist!
	http.Error(w, "Bad Request", http.StatusBadRequest)
}

func main() {
	http.HandleFunc("/shorten", createShortUrlHandler)
	http.HandleFunc("/shorten/", shortenRouteHandler)
	http.HandleFunc("/", redirectHandler)
	fmt.Println("Servidor corriendo en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
