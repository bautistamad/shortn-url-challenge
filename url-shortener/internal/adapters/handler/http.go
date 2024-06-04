package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"url-shortener/internal/domain/ports"

	"github.com/gorilla/mux"
)

type HTTPHandler struct {
	router           *mux.Router
	shortenerService ports.ShorternService
}

func NewHTTPHandler(
	router *mux.Router,
	shortenerService ports.ShorternService,
) *HTTPHandler {
	return &HTTPHandler{
		router:           router,
		shortenerService: shortenerService,
	}
}

func (h *HTTPHandler) RegisterRoutes() {
	h.router.HandleFunc("/shorten", h.handleShortenURL).Methods("POST")
	h.router.HandleFunc("/{shortURL}", h.handleRedirect).Methods("GET")
	h.router.HandleFunc("/url/{shortURL}", h.handleDeleteURL).Methods("DELETE")
	h.router.HandleFunc("/url/{shortURL}/stats", h.handleGetUrlStats).Methods("GET")
}

func (h *HTTPHandler) handleShortenURL(w http.ResponseWriter, r *http.Request) {
	var request struct {
		LongURL string `json:"long_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	shortURL, err := h.shortenerService.CreateShortUrl(request.LongURL)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := struct {
		ShortURL string `json:"short_url"`
	}{ShortURL: shortURL}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *HTTPHandler) handleRedirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortURL := vars["shortURL"]

	longURL, err := h.shortenerService.GetLongUrl(shortURL)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}

func (h *HTTPHandler) handleDeleteURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortURL := vars["shortURL"]

	// Eliminar la URL
	deletedUrl, err := h.shortenerService.DeleteUrl(shortURL)
	if err != nil {
		if errors.Is(err, ports.ErrUrlNotFound) {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Message    string `json:"message"`
		DeletedURL string `json:"deleted_url,omitempty"`
	}{
		Message:    "URL deleted successfully",
		DeletedURL: deletedUrl,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *HTTPHandler) handleGetUrlStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortURL := vars["shortURL"]

	// Obtener las estadísticas de la URL
	urlStats, err := h.shortenerService.GetUrlStats(shortURL)
	if err != nil {
		if errors.Is(err, ports.ErrUrlNotFound) {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(urlStats); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
