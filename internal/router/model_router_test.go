package router

import (
	"testing"

	"oc-go-cc-plus/internal/config"
)

func newTestAtomicConfig(cfg *config.Config) *config.AtomicConfig {
	return config.NewAtomicConfig(cfg, "/tmp/test-config.json")
}

func TestRoute_RespectRequestedModel_BypassesScenarioRouting(t *testing.T) {
	cfg := &config.Config{
		RespectRequestedModel: true,
		Models: map[string]config.ModelConfig{
			"default": {
				Provider: "opencode-go",
				ModelID:  "kimi-k2.6",
			},
			"kimi-k2.6": {
				Provider:         "opencode-go",
				ModelID:          "kimi-k2.6",
				Temperature:      0.7,
				MaxTokens:        4096,
				ContextThreshold: 80000,
			},
		},
		Fallbacks: map[string][]config.ModelConfig{
			"default": {
				{Provider: "opencode-go", ModelID: "qwen3.5-plus"},
			},
		},
	}

	router := NewModelRouter(newTestAtomicConfig(cfg))

	// Complex message that would normally route to GLM-5.1
	messages := []MessageContent{
		{Role: "user", Content: "Architect a new microservice"},
	}

	result, err := router.Route(messages, 100, "kimi-k2.6")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Primary.ModelID != "kimi-k2.6" {
		t.Errorf("expected model kimi-k2.6, got %s", result.Primary.ModelID)
	}
	if result.Primary.Temperature != 0.7 {
		t.Errorf("expected temperature 0.7, got %f", result.Primary.Temperature)
	}
	if result.Primary.MaxTokens != 4096 {
		t.Errorf("expected max_tokens 4096, got %d", result.Primary.MaxTokens)
	}
	if result.Scenario != ScenarioDefault {
		t.Errorf("expected ScenarioDefault, got %s", result.Scenario)
	}
}

func TestRoute_RespectRequestedModel_False_UsesScenarioRouting(t *testing.T) {
	cfg := &config.Config{
		RespectRequestedModel: false,
		Models: map[string]config.ModelConfig{
			"default": {ModelID: "kimi-k2.6"},
			"complex": {ModelID: "glm-5.1"},
		},
		Fallbacks: map[string][]config.ModelConfig{
			"default": {{ModelID: "qwen3.5-plus"}},
		},
	}

	router := NewModelRouter(newTestAtomicConfig(cfg))

	messages := []MessageContent{
		{Role: "user", Content: "Architect a new microservice"},
	}

	result, err := router.Route(messages, 100, "kimi-k2.6")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should use scenario routing, not the requested model
	if result.Primary.ModelID != "glm-5.1" {
		t.Errorf("expected scenario-routed model glm-5.1, got %s", result.Primary.ModelID)
	}
}

func TestRoute_RespectRequestedModel_EmptyModel_FallsThrough(t *testing.T) {
	cfg := &config.Config{
		RespectRequestedModel: true,
		Models: map[string]config.ModelConfig{
			"default": {ModelID: "kimi-k2.6"},
		},
		Fallbacks: map[string][]config.ModelConfig{
			"default": {{ModelID: "qwen3.5-plus"}},
		},
	}

	router := NewModelRouter(newTestAtomicConfig(cfg))

	messages := []MessageContent{
		{Role: "user", Content: "Hello"},
	}

	result, err := router.Route(messages, 100, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty model should fall through to scenario routing
	if result.Primary.ModelID != "kimi-k2.6" {
		t.Errorf("expected default model kimi-k2.6, got %s", result.Primary.ModelID)
	}
}

func TestRoute_RespectRequestedModel_UnknownModel_UsesDefaults(t *testing.T) {
	cfg := &config.Config{
		RespectRequestedModel: true,
		Models: map[string]config.ModelConfig{
			"default": {
				Provider:    "opencode-go",
				ModelID:     "kimi-k2.6",
				Temperature: 0.5,
				MaxTokens:   8192,
			},
		},
		Fallbacks: map[string][]config.ModelConfig{
			"default": {{ModelID: "qwen3.5-plus"}},
		},
	}

	router := NewModelRouter(newTestAtomicConfig(cfg))

	messages := []MessageContent{
		{Role: "user", Content: "Hello"},
	}

	result, err := router.Route(messages, 100, "some-unknown-model")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Primary.ModelID != "some-unknown-model" {
		t.Errorf("expected model some-unknown-model, got %s", result.Primary.ModelID)
	}
	// Unknown model should inherit default temperature/max_tokens
	if result.Primary.Temperature != 0.5 {
		t.Errorf("expected inherited temperature 0.5, got %f", result.Primary.Temperature)
	}
	if result.Primary.MaxTokens != 8192 {
		t.Errorf("expected inherited max_tokens 8192, got %d", result.Primary.MaxTokens)
	}
}

func TestRouteForStreaming_RespectRequestedModel_BypassesScenarioRouting(t *testing.T) {
	cfg := &config.Config{
		RespectRequestedModel: true,
		Models: map[string]config.ModelConfig{
			"default": {ModelID: "qwen3.6-plus"},
			"kimi-k2.6": {
				Provider: "opencode-go",
				ModelID:  "kimi-k2.6",
			},
		},
		Fallbacks: map[string][]config.ModelConfig{
			"default": {{ModelID: "qwen3.5-plus"}},
		},
	}

	router := NewModelRouter(newTestAtomicConfig(cfg))

	messages := []MessageContent{
		{Role: "user", Content: "Hello"},
	}

	result := router.RouteForStreaming(messages, 100, "kimi-k2.6")
	if result.Primary.ModelID != "kimi-k2.6" {
		t.Errorf("expected model kimi-k2.6, got %s", result.Primary.ModelID)
	}
	if result.Scenario != ScenarioDefault {
		t.Errorf("expected ScenarioDefault, got %s", result.Scenario)
	}
}

func TestRouteForStreaming_RespectRequestedModel_False_UsesScenarioRouting(t *testing.T) {
	cfg := &config.Config{
		RespectRequestedModel: false,
		Models: map[string]config.ModelConfig{
			"default": {ModelID: "qwen3.6-plus"},
		},
		Fallbacks: map[string][]config.ModelConfig{
			"default": {{ModelID: "qwen3.5-plus"}},
		},
	}

	router := NewModelRouter(newTestAtomicConfig(cfg))

	messages := []MessageContent{
		{Role: "user", Content: "Hello"},
	}

	result := router.RouteForStreaming(messages, 100, "kimi-k2.6")
	// Should use streaming scenario routing, not the requested model
	if result.Primary.ModelID != "qwen3.6-plus" {
		t.Errorf("expected streaming model qwen3.6-plus, got %s", result.Primary.ModelID)
	}
}

func TestResolveRequestedModel_UsesFallbacks(t *testing.T) {
	cfg := &config.Config{
		RespectRequestedModel: true,
		Models: map[string]config.ModelConfig{
			"kimi-k2.6": {ModelID: "kimi-k2.6"},
		},
		Fallbacks: map[string][]config.ModelConfig{
			"default": {
				{Provider: "opencode-go", ModelID: "qwen3.5-plus"},
				{Provider: "opencode-go", ModelID: "glm-5.1"},
			},
		},
	}

	router := NewModelRouter(newTestAtomicConfig(cfg))

	result, ok := router.resolveRequestedModel(cfg, "kimi-k2.6", 100)
	if !ok {
		t.Fatal("expected resolveRequestedModel to match")
	}
	if len(result.Fallbacks) != 2 {
		t.Errorf("expected 2 fallbacks, got %d", len(result.Fallbacks))
	}
	if result.Fallbacks[0].ModelID != "qwen3.5-plus" {
		t.Errorf("expected first fallback qwen3.5-plus, got %s", result.Fallbacks[0].ModelID)
	}
}

func TestRoute_RespectRequestedModel_LongContextOverridesRequestedModel(t *testing.T) {
	cfg := &config.Config{
		RespectRequestedModel: true,
		Models: map[string]config.ModelConfig{
			"default":      {ModelID: "deepseek-v4-pro"},
			"long_context": {ModelID: "deepseek-v4-pro", ContextThreshold: 80000},
			"deepseek-v4-pro": {
				Provider: "opencode-go",
				ModelID:  "deepseek-v4-pro",
			},
		},
		Fallbacks: map[string][]config.ModelConfig{
			"long_context": {{ModelID: "kimi-k2.6"}},
		},
	}

	router := NewModelRouter(newTestAtomicConfig(cfg))
	result, err := router.Route(nil, 85440, "deepseek-v4-pro")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Scenario != ScenarioLongContext {
		t.Fatalf("expected long_context scenario, got %s", result.Scenario)
	}
	if result.Primary.ModelID != "deepseek-v4-pro" {
		t.Fatalf("expected deepseek-v4-pro, got %s", result.Primary.ModelID)
	}
}

func TestRoute_RespectRequestedModel_GatewayPrefix(t *testing.T) {
	cfg := &config.Config{
		RespectRequestedModel: true,
		Models: map[string]config.ModelConfig{
			"deepseek-v4-pro": {
				Provider:    "opencode-go",
				ModelID:     "deepseek-v4-pro",
				Temperature: 0.1,
				MaxTokens:   8192,
			},
		},
	}

	router := NewModelRouter(newTestAtomicConfig(cfg))
	result, err := router.Route(nil, 0, "anthropic-opencode-deepseek-v4-pro")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Primary.ModelID != "deepseek-v4-pro" {
		t.Fatalf("expected deepseek-v4-pro, got %s", result.Primary.ModelID)
	}
}
