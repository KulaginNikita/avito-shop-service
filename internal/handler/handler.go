package handler

import (
	"encoding/json"
	"net/http"

	"avito-shop/internal/config"
	"avito-shop/internal/models"
	"avito-shop/internal/service"

	"github.com/gorilla/mux"
)

type Handler struct {
	svc service.Service
	cfg *config.Config
}

func NewHandler(svc service.Service, cfg *config.Config) *Handler {
	return &Handler{svc: svc, cfg: cfg}
}

// ------------------- /api/auth [POST] -------------------
func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	token, err := h.svc.AuthUser(req.Username, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	resp := models.AuthResponse{Token: token}
	writeJSON(w, http.StatusOK, resp)
}

// ------------------- /api/info [GET] -------------------
func (h *Handler) GetInfo(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	info, err := h.svc.GetInfo(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, info)
}

// ------------------- /api/sendCoin [POST] -------------------
func (h *Handler) SendCoin(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	var req models.SendCoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.svc.SendCoin(userID, req.ToUser, req.Amount); err != nil {
		switch err {
		case service.ErrNotEnoughCoins:
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ------------------- /api/buy/{item} [GET] -------------------
func (h *Handler) BuyItem(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	item := mux.Vars(r)["item"]
	if item == "" {
		writeError(w, http.StatusBadRequest, "Item not specified")
		return
	}

	if err := h.svc.BuyItem(userID, item); err != nil {
		switch err {
		case service.ErrNotEnoughCoins, service.ErrInvalidItem:
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}



func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(models.ErrorResponse{
		Errors: msg,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

