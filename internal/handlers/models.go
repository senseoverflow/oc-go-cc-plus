package handlers

import (
	"encoding/json"
	"net/http"

	"oc-go-cc-plus/internal/config"
	"oc-go-cc-plus/internal/models"
)

// ModelsHandler serves Claude Code gateway model discovery.
type ModelsHandler struct {
	atomic *config.AtomicConfig
}

// NewModelsHandler creates a handler for GET /v1/models.
func NewModelsHandler(atomic *config.AtomicConfig) *ModelsHandler {
	return &ModelsHandler{atomic: atomic}
}

// HandleModels handles GET /v1/models in Anthropic-native format.
func (h *ModelsHandler) HandleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := h.atomic.Get()
	resp := models.BuildGatewayModelsResponse(cfg)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
