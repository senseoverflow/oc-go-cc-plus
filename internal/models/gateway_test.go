package models

import (
	"testing"

	"oc-go-cc-plus/internal/config"
)

func TestToGatewayModelID(t *testing.T) {
	if got := ToGatewayModelID("deepseek-v4-pro"); got != "anthropic-opencode-deepseek-v4-pro" {
		t.Fatalf("got %q", got)
	}
	if got := ToGatewayModelID("anthropic-opencode-deepseek-v4-pro"); got != "anthropic-opencode-deepseek-v4-pro" {
		t.Fatalf("idempotent: got %q", got)
	}
}

func TestNormalizeRequestedModel(t *testing.T) {
	tests := map[string]string{
		"anthropic-opencode-deepseek-v4-pro": "deepseek-v4-pro",
		"deepseek-v4-pro":                    "deepseek-v4-pro",
		"  kimi-k2.6  ":                      "kimi-k2.6",
	}
	for in, want := range tests {
		if got := NormalizeRequestedModel(in); got != want {
			t.Fatalf("%q: got %q want %q", in, got, want)
		}
	}
}

func TestListGatewayModelsSkipsScenarios(t *testing.T) {
	cfg := &config.Config{
		Models: map[string]config.ModelConfig{
			"default":         {ModelID: "deepseek-v4-pro"},
			"deepseek-v4-pro": {ModelID: "deepseek-v4-pro", MaxTokens: 8192},
			"kimi-k2.6":       {ModelID: "kimi-k2.6"},
		},
	}
	models := ListGatewayModels(cfg)
	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}
	if models[0].ID != "anthropic-opencode-deepseek-v4-pro" {
		t.Fatalf("unexpected first id: %s", models[0].ID)
	}
}

func TestBuildGatewayModelsResponse(t *testing.T) {
	cfg := &config.Config{
		Models: map[string]config.ModelConfig{
			"deepseek-v4-pro": {ModelID: "deepseek-v4-pro", MaxTokens: 8192},
		},
	}
	resp := BuildGatewayModelsResponse(cfg)
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(resp.Data))
	}
	if resp.Data[0].Type != "model" {
		t.Fatalf("expected type model, got %q", resp.Data[0].Type)
	}
	if resp.FirstID != resp.Data[0].ID || resp.LastID != resp.Data[0].ID {
		t.Fatalf("pagination ids mismatch")
	}
}
