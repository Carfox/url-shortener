package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// VARIABLES

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

// STRUCTS
type UrlData struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	Url       string    `json:"url" bson:"url"`
	ShortCode string    `json:"short_code" bson:"short_code"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
	Clicks    int       `json:"access_count" bson:"access_count"`
}

type UrlResponse struct {
	ID        string    `json:"id"`
	Url       string    `json:"url"`
	ShortCode string    `json:"short_code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UrlRequest struct {
	Url string `json:"url"`
}

//  FUNCTIONS

func codeGenerator(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func saveUrl(url string) (UrlData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for {
		code := codeGenerator(6)
		filter := bson.M{"_id": code}
		err := collection.FindOne(ctx, filter).Err()

		if err == nil {
			fmt.Println("Already exists a code saved", code)
			continue
		}

		if err != mongo.ErrNoDocuments {
			return UrlData{}, err
		}

		newURLRegister := UrlData{
			ID:        code,
			Url:       url,
			ShortCode: code,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, errInsert := collection.InsertOne(ctx, newURLRegister)
		if errInsert != nil {
			return UrlData{}, errInsert
		}

		fmt.Println("URL Saved:", url, "->", code)
		return newURLRegister, nil
	}

}

func updateUrl(code string, newUrl string) (UrlData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"_id": code}

	update := bson.M{
		"$set": bson.M{
			"url":        newUrl,
			"updated_at": time.Now(),
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedDoc UrlData

	err := collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedDoc)

	return updatedDoc, err
}

func deleteUrl(code string) {
	filter := bson.M{"_id": code}
	collection.FindOneAndDelete(context.TODO(), filter)
}

func getUrl(code string) (UrlData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var doc UrlData

	filter := bson.M{"_id": code}

	err := collection.FindOne(ctx, filter).Decode(&doc)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No documents found")
		} else {
			panic(err)
		}
	}

	return doc, err
}

func getUrlAndIncrement(code string) (UrlData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": code}
	update := bson.M{"$inc": bson.M{"access_count": 1}}

	var result UrlData
	err := collection.FindOneAndUpdate(ctx, filter, update).Decode(&result)

	return result, err
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

	doc, err := getUrlAndIncrement(code)

	if err != nil {
		http.Error(w, "URL not found", 404)
		return
	}

	http.Redirect(w, r, doc.Url, http.StatusFound)
}

func shortenRouteHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")

	// "/shorten/ - /<code>/ - /stats/"
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	code := parts[2] // /<code>/

	data, err := getUrl(code)
	if err != nil {
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

		data, err = updateUrl(code, req.Url)

		if err != nil {
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
			ID:        data.ShortCode,
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
	initDB()
	http.HandleFunc("/shorten", createShortUrlHandler)
	http.HandleFunc("/shorten/", shortenRouteHandler)
	http.HandleFunc("/", redirectHandler)
	fmt.Println("Servidor corriendo en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
