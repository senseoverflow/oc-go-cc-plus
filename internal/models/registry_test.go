package models

import "testing"

func TestUsesAnthropicEndpoint(t *testing.T) {
	tests := map[string]bool{
		"minimax-m2.5":      true,
		"minimax-m2.7":      true,
		"qwen3.6-plus":      true,
		"qwen3.7-max":       true,
		"deepseek-v4-pro":   false,
		"deepseek-v4-flash": false,
		"kimi-k2.6":         false,
	}
	for id, want := range tests {
		if got := UsesAnthropicEndpoint(id); got != want {
			t.Fatalf("%s: got %v want %v", id, got, want)
		}
	}
}
