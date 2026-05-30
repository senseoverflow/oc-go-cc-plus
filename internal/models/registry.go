// Package models provides OpenCode Go model metadata, endpoint routing, and sync.
package models

// EndpointType identifies which OpenCode Go API surface a model uses.
type EndpointType string

const (
	EndpointOpenAI          EndpointType = "openai"
	EndpointAnthropicNative EndpointType = "anthropic"
)

const (
	OpenAIBaseURL      = "https://opencode.ai/zen/go/v1/chat/completions"
	AnthropicBaseURL   = "https://opencode.ai/zen/go/v1/messages"
	ModelsListURL      = "https://opencode.ai/zen/go/v1/models"
)

// Info describes a known OpenCode Go model.
type Info struct {
	ID           string
	DisplayName  string
	Endpoint     EndpointType
	Deprecated   bool
	Replacement  string
}

// KnownModels is the curated registry aligned with OpenCode Go documentation.
// Sync-models merges live API results with this list.
var KnownModels = map[string]Info{
	"deepseek-v4-pro":   {ID: "deepseek-v4-pro", DisplayName: "DeepSeek V4 Pro", Endpoint: EndpointOpenAI},
	"deepseek-v4-flash": {ID: "deepseek-v4-flash", DisplayName: "DeepSeek V4 Flash", Endpoint: EndpointOpenAI},
	"glm-5":             {ID: "glm-5", DisplayName: "GLM-5", Endpoint: EndpointOpenAI},
	"glm-5.1":           {ID: "glm-5.1", DisplayName: "GLM-5.1", Endpoint: EndpointOpenAI},
	"kimi-k2.5":         {ID: "kimi-k2.5", DisplayName: "Kimi K2.5", Endpoint: EndpointOpenAI},
	"kimi-k2.6":         {ID: "kimi-k2.6", DisplayName: "Kimi K2.6", Endpoint: EndpointOpenAI},
	"mimo-v2.5":         {ID: "mimo-v2.5", DisplayName: "MiMo V2.5", Endpoint: EndpointOpenAI},
	"mimo-v2.5-pro":     {ID: "mimo-v2.5-pro", DisplayName: "MiMo V2.5 Pro", Endpoint: EndpointOpenAI},
	"minimax-m2.5":      {ID: "minimax-m2.5", DisplayName: "MiniMax M2.5", Endpoint: EndpointAnthropicNative},
	"minimax-m2.7":      {ID: "minimax-m2.7", DisplayName: "MiniMax M2.7", Endpoint: EndpointAnthropicNative},
	"qwen3.6-plus":      {ID: "qwen3.6-plus", DisplayName: "Qwen3.6 Plus", Endpoint: EndpointAnthropicNative},
	"qwen3.7-max":       {ID: "qwen3.7-max", DisplayName: "Qwen3.7 Max", Endpoint: EndpointAnthropicNative},
	"qwen3.5-plus":      {ID: "qwen3.5-plus", DisplayName: "Qwen3.5 Plus", Endpoint: EndpointOpenAI, Deprecated: true, Replacement: "mimo-v2.5"},
	"mimo-v2-pro":       {ID: "mimo-v2-pro", DisplayName: "MiMo V2 Pro", Endpoint: EndpointOpenAI, Deprecated: true, Replacement: "mimo-v2.5-pro"},
	"mimo-v2-omni":      {ID: "mimo-v2-omni", DisplayName: "MiMo V2 Omni", Endpoint: EndpointOpenAI, Deprecated: true, Replacement: "mimo-v2.5"},
}

// ScenarioKeys are the routing scenarios oc-go-cc-plus understands.
var ScenarioKeys = []string{"background", "default", "long_context", "think", "complex", "fast"}

// UsesAnthropicEndpoint reports whether requests for modelID should use /v1/messages.
func UsesAnthropicEndpoint(modelID string) bool {
	if info, ok := KnownModels[modelID]; ok {
		return info.Endpoint == EndpointAnthropicNative
	}
	return false
}

// EndpointLabel returns a human-readable endpoint description.
func EndpointLabel(modelID string) string {
	if UsesAnthropicEndpoint(modelID) {
		return "Anthropic (/v1/messages)"
	}
	return "OpenAI (/v1/chat/completions)"
}

// AllKnownIDs returns sorted known model IDs (non-deprecated first).
func AllKnownIDs() []string {
	ids := make([]string, 0, len(KnownModels))
	for id, info := range KnownModels {
		if !info.Deprecated {
			ids = append(ids, id)
		}
	}
	for id, info := range KnownModels {
		if info.Deprecated {
			ids = append(ids, id)
		}
	}
	return ids
}
