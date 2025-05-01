package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/singl3focus/load_balancer-go/internal/adapters/secondary/storage"
	"github.com/singl3focus/load_balancer-go/internal/domain"
	"github.com/singl3focus/load_balancer-go/pkg/logger"
)

type ClientHandler struct {
	logger  logger.Logger
	service domain.Storage
}

func NewClientHandler(l logger.Logger, s domain.Storage) *ClientHandler {
	return &ClientHandler{
		logger:  l,
		service: s,
	}
}

func (h *ClientHandler) AddConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key    string              `json:"key"`
		Config domain.BucketConfig `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request format", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.StoreConfig(req.Key, req.Config); err != nil {
		h.logger.Error("Failed to store config", "key", req.Key, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}

func (h *ClientHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	config, err := h.service.GetConfig(key)
	if err != nil {
		if errors.Is(err, storage.ErrDataNotFound) {
			http.Error(w, "Config not found", http.StatusNotFound)
			return
		}
		h.logger.Error("Failed to get config", "key", key, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(config)
}

func (h *ClientHandler) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteConfig(key); err != nil {
		h.logger.Error("Failed to delete config", "key", key, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (h *ClientHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key    string              `json:"key"`
		Config domain.BucketConfig `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request format", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateConfig(req.Key, req.Config); err != nil {
		h.logger.Error("Failed to update config", "key", req.Key, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}
