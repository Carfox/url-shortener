package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"url-shortener/internal/models"
	"url-shortener/internal/repository"
)

type UrlHandler struct {
	Repo *repository.UrlRepository
}

func (h *UrlHandler) CreateShortUrlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Bad Request", 400)
		return
	}

	var req models.UrlRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(req.Url, "http://") && !strings.HasPrefix(req.Url, "https://") {
		req.Url = "https://" + req.Url
	}

	answer, err := h.Repo.SaveUrl(req.Url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(answer)

}

func (h *UrlHandler) RedirectHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Bad Request", 400)
		return
	}

	code := r.URL.Path[1:]

	doc, err := h.Repo.GetUrlAndIncrement(code)

	if err != nil {
		http.Error(w, "URL not found", 404)
		return
	}

	http.Redirect(w, r, doc.Url, http.StatusFound)
}

func (h *UrlHandler) ShortenRouteHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")

	// "/shorten/ - /<code>/ - /stats/"
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	code := parts[2] // /<code>/

	data, err := h.Repo.GetUrl(code)
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
		var req models.UrlRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		data, err = h.Repo.UpdateUrl(code, req.Url)

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
		answer := models.UrlResponse{
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
		h.Repo.DeleteUrl(code)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// If method doesnt exist!
	http.Error(w, "Bad Request", http.StatusBadRequest)
}
