package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"oc-go-cc-plus/internal/config"
)

func TestHandleModels(t *testing.T) {
	cfg := &config.Config{
		Models: map[string]config.ModelConfig{
			"deepseek-v4-pro": {ModelID: "deepseek-v4-pro", MaxTokens: 8192},
		},
	}
	atomic := config.NewAtomicConfig(cfg, "/tmp/test-config.json")
	h := NewModelsHandler(atomic)

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	rec := httptest.NewRecorder()
	h.HandleModels(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data []struct {
			ID          string `json:"id"`
			Type        string `json:"type"`
			DisplayName string `json:"display_name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 model, got %d", len(resp.Data))
	}
	if resp.Data[0].ID != "anthropic-opencode-deepseek-v4-pro" {
		t.Fatalf("unexpected id %q", resp.Data[0].ID)
	}
	if resp.Data[0].Type != "model" {
		t.Fatalf("unexpected type %q", resp.Data[0].Type)
	}
}

func TestHandleModelsMethodNotAllowed(t *testing.T) {
	atomic := config.NewAtomicConfig(&config.Config{}, "/tmp/test-config.json")
	h := NewModelsHandler(atomic)

	req := httptest.NewRequest(http.MethodPost, "/v1/models", nil)
	rec := httptest.NewRecorder()
	h.HandleModels(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status %d", rec.Code)
	}
}
