package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"shawty/internal/service"
)

// URLHandler manages HTTP requests related to URLs.
type URLHandler struct {
	urlService service.UrlServiceInterface
}

// NewURLHandler creates a new URLHandler.
func NewURLHandler(s service.UrlServiceInterface) *URLHandler {
	return &URLHandler{urlService: s}
}

// RegisterRoutes sets up the routes for the URL handler.
func (h *URLHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.homeHandler)
	mux.HandleFunc("/shorten", h.shortenURLHandler)
	mux.HandleFunc("/r/", h.redirectURLHandler) // Using /r/ as the prefix for redirection
}

// homeHandler provides a simple welcome message.
func (h *URLHandler) homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "Hello from Shawty URL Shortener!")
}

// ShortenURLRequest defines the expected JSON body for shortening a URL.
type ShortenURLRequest struct {
	URL string `json:"url"`
}

// ShortenURLResponse defines the JSON response for a successful shortening.
type ShortenURLResponse struct {
	ShortURL     string `json:"short_url"`
	OriginalURL  string `json:"original_url"`
	CreationDate string `json:"creation_date"`
}

// shortenURLHandler handles requests to create a new short URL.
// It expects a POST request with a JSON body like: {"url": "http://example.com"}
func (h *URLHandler) shortenURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.URL == "" {
		http.Error(w, "URL field is missing or empty in request body", http.StatusBadRequest)
		return
	}

	createdURL, err := h.urlService.CreateShortURL(r.Context(), req.URL)
	if err != nil {
		log.Printf("Error creating short URL for '%s': %v", req.URL, err)

		if errors.Is(err, service.ErrHashCollision) {
			http.Error(w, "Failed to create short URL due to a hash collision. Please try again or modify the URL slightly.", http.StatusConflict)
		} else if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
			// This case should ideally be less frequent now if service.CreateShortURL handles known duplicates by returning the existing URL.
			// This might catch other unexpected duplicate errors or if ErrDuplicateShortID from store somehow propagates directly.
			http.Error(w, "This URL may have already been shortened or a conflict occurred.", http.StatusConflict)
		} else {
			http.Error(w, "Failed to create short URL", http.StatusInternalServerError)
		}
		return
	}

	// Construct the full short URL to return to the client
	// Scheme (http/https) and Host should ideally be configurable or detected
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	fullShortURL := fmt.Sprintf("%s://%s/r/%s", scheme, r.Host, createdURL.ShortUrl)

	response := ShortenURLResponse{
		ShortURL:     fullShortURL,
		OriginalURL:  createdURL.OriginalUrl,
		CreationDate: createdURL.CreationDate.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response for short URL '%s': %v", createdURL.ShortUrl, err)
		// Cannot send http.Error here as headers might have been written
	}
}

// redirectURLHandler handles requests to redirect a short URL to its original URL.
// It expects URLs in the format /r/{shortID}
func (h *URLHandler) redirectURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed for redirection", http.StatusMethodNotAllowed)
		return
	}

	shortID := strings.TrimPrefix(r.URL.Path, "/r/")
	if shortID == "" {
		http.Error(w, "Short URL ID is missing in the path", http.StatusBadRequest)
		return
	}

	originalURL, err := h.urlService.GetOriginalURL(r.Context(), shortID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, fmt.Sprintf("Short URL '%s' not found", shortID), http.StatusNotFound)
		} else {
			log.Printf("Error retrieving original URL for short ID '%s': %v", shortID, err)
			http.Error(w, "Error retrieving URL", http.StatusInternalServerError)
		}
		return
	}

	// Ensure the original URL has a scheme for proper redirection.
	// Prepend "http://" if no scheme is present.
	// A more robust solution would involve better URL validation/parsing.
	if !strings.HasPrefix(originalURL, "http://") && !strings.HasPrefix(originalURL, "https://") {
		originalURL = "http://" + originalURL
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}
