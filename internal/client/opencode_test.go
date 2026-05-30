package client

import (
	"testing"
	"time"

	"oc-go-cc-plus/internal/config"
)

func newTestAtomicConfig(cfg *config.Config) *config.AtomicConfig {
	return config.NewAtomicConfig(cfg, "/tmp/test-config.json")
}

func TestStreamingTimeout_ScalesForLargePrompts(t *testing.T) {
	cfg := &config.Config{
		OpenCodeGo: config.OpenCodeGoConfig{TimeoutMs: 300000},
		Models: map[string]config.ModelConfig{
			"long_context": {ContextThreshold: 80000},
		},
	}
	client := NewOpenCodeClient(newTestAtomicConfig(cfg))

	if got := client.StreamingTimeout(1000); got != 5*time.Minute {
		t.Fatalf("StreamingTimeout(1000) = %v, want 5m", got)
	}
	if got := client.StreamingTimeout(85440); got != 15*time.Minute {
		t.Fatalf("StreamingTimeout(85440) = %v, want 15m", got)
	}
}

func TestIsAnthropicModelOnlyRoutesNativeAnthropicModels(t *testing.T) {
	tests := []struct {
		name    string
		modelID string
		want    bool
	}{
		{
			name:    "minimax m2.5 uses anthropic endpoint",
			modelID: "minimax-m2.5",
			want:    true,
		},
		{
			name:    "minimax m2.7 uses anthropic endpoint",
			modelID: "minimax-m2.7",
			want:    true,
		},
		{
			name:    "qwen3.6 plus uses anthropic endpoint",
			modelID: "qwen3.6-plus",
			want:    true,
		},
		{
			name:    "qwen3.7 max uses anthropic endpoint",
			modelID: "qwen3.7-max",
			want:    true,
		},
		{
			name:    "deepseek pro uses openai endpoint",
			modelID: "deepseek-v4-pro",
			want:    false,
		},
		{
			name:    "deepseek flash uses openai endpoint",
			modelID: "deepseek-v4-flash",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAnthropicModel(tt.modelID); got != tt.want {
				t.Fatalf("IsAnthropicModel(%q) = %v, want %v", tt.modelID, got, tt.want)
			}
		})
	}
}
