package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/bajdzun/go-url-shortener/internal/domain"
	"github.com/bajdzun/go-url-shortener/internal/service"
	"go.uber.org/zap"
)

type URLHandler struct {
	service *service.URLService
	logger  *zap.Logger
}

func NewURLHandler(service *service.URLService, logger *zap.Logger) *URLHandler {
	return &URLHandler{
		service: service,
		logger:  logger,
	}
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func (h *URLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var req service.CreateURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	resp, err := h.service.CreateShortURL(r.Context(), &req)
	if err != nil {
		switch err {
		case domain.ErrInvalidURL:
			h.respondError(w, http.StatusBadRequest, "invalid URL", err.Error())
		case domain.ErrShortCodeExists:
			h.respondError(w, http.StatusConflict, "short code already exists", err.Error())
		default:
			h.logger.Error("failed to create short URL", zap.Error(err))
			h.respondError(w, http.StatusInternalServerError, "internal server error", "")
		}

		return
	}

	h.respondJSON(w, http.StatusCreated, resp)
}

func (h *URLHandler) RedirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")
	if shortCode == "" {
		h.respondError(w, http.StatusBadRequest, "short code is required", "")

		return
	}

	analytics := &domain.Analytics{
		IPAddress: h.getClientIP(r),
		UserAgent: r.UserAgent(),
		Referer:   r.Referer(),
	}

	originalURL, err := h.service.GetOriginalURL(r.Context(), shortCode, analytics)
	if err != nil {
		switch err {
		case domain.ErrURLNotFound:
			h.respondError(w, http.StatusNotFound, "URL not found", err.Error())
		case domain.ErrExpiredURL:
			h.respondError(w, http.StatusGone, "URL has expired", err.Error())
		default:
			h.logger.Error("failed to get original URL", zap.Error(err))
			h.respondError(w, http.StatusInternalServerError, "internal server error", "")
		}

		return
	}

	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}

func (h *URLHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")
	if shortCode == "" {
		h.respondError(w, http.StatusBadRequest, "short code is required", "")

		return
	}

	stats, err := h.service.GetStats(r.Context(), shortCode)
	if err != nil {
		switch err {
		case domain.ErrURLNotFound:
			h.respondError(w, http.StatusNotFound, "URL not found", err.Error())
		default:
			h.logger.Error("failed to get stats", zap.Error(err))
			h.respondError(w, http.StatusInternalServerError, "internal server error", "")
		}

		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}

func (h *URLHandler) DeleteURL(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")
	if shortCode == "" {
		h.respondError(w, http.StatusBadRequest, "short code is required", "")

		return
	}

	if err := h.service.DeleteURL(r.Context(), shortCode); err != nil {
		switch err {
		case domain.ErrURLNotFound:
			h.respondError(w, http.StatusNotFound, "URL not found", err.Error())
		default:
			h.logger.Error("failed to delete URL", zap.Error(err))
			h.respondError(w, http.StatusInternalServerError, "internal server error", "")
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *URLHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *URLHandler) respondError(w http.ResponseWriter, status int, error string, message string) {
	h.respondJSON(w, status, ErrorResponse{
		Error:   error,
		Message: message,
	})
}

func (h *URLHandler) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}

	return ip
}
