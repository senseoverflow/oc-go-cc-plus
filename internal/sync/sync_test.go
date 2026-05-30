package sync

import (
	"testing"

	"oc-go-cc-plus/internal/config"
)

func TestModelsAddsMissingEntries(t *testing.T) {
	cfg := &config.Config{
		APIKey: "test-key",
		Models: map[string]config.ModelConfig{
			"default": {Provider: "opencode-go", ModelID: "kimi-k2.6"},
		},
	}
	result := Models(cfg, []string{"deepseek-v4-pro", "qwen3.7-max", "default"})
	if len(result.Added) != 2 {
		t.Fatalf("added = %v, want 2 entries", result.Added)
	}
	if _, ok := cfg.Models["deepseek-v4-pro"]; !ok {
		t.Fatal("expected deepseek-v4-pro entry")
	}
}
