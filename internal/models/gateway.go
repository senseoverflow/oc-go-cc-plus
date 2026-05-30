package models

import (
	"sort"
	"strings"
	"time"

	"oc-go-cc-plus/internal/config"
)

const (
	// GatewayModelPrefix makes OpenCode Go models discoverable by Claude Code,
	// which only lists gateway models whose ID starts with "claude" or "anthropic".
	GatewayModelPrefix = "anthropic-opencode-"
)

// GatewayModel describes a model entry returned by GET /v1/models.
type GatewayModel struct {
	ID          string
	DisplayName string
	MaxTokens   int
	Thinking    bool
}

// ToGatewayModelID maps an OpenCode Go model ID to a Claude Code discoverable ID.
func ToGatewayModelID(modelID string) string {
	if modelID == "" {
		return ""
	}
	if strings.HasPrefix(modelID, GatewayModelPrefix) {
		return modelID
	}
	return GatewayModelPrefix + modelID
}

// NormalizeRequestedModel converts a Claude Code model name to an OpenCode Go model ID.
func NormalizeRequestedModel(requested string) string {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		return ""
	}
	if strings.HasPrefix(requested, GatewayModelPrefix) {
		return strings.TrimPrefix(requested, GatewayModelPrefix)
	}
	return requested
}

// ListGatewayModels returns configured OpenCode Go models for Claude Code discovery.
func ListGatewayModels(cfg *config.Config) []GatewayModel {
	scenarioSet := make(map[string]struct{}, len(ScenarioKeys))
	for _, key := range ScenarioKeys {
		scenarioSet[key] = struct{}{}
	}

	seen := make(map[string]struct{})
	out := make([]GatewayModel, 0, len(cfg.Models))

	for name, entry := range cfg.Models {
		if _, isScenario := scenarioSet[name]; isScenario {
			continue
		}
		modelID := entry.ModelID
		if modelID == "" {
			modelID = name
		}
		if _, ok := seen[modelID]; ok {
			continue
		}
		seen[modelID] = struct{}{}

		if info, ok := KnownModels[modelID]; ok && info.Deprecated {
			continue
		}

		displayName := modelID
		thinking := false
		if info, ok := KnownModels[modelID]; ok {
			displayName = info.DisplayName
		}
		displayName = displayName + " (OpenCode Go)"

		if len(entry.Thinking) > 0 && string(entry.Thinking) != "null" {
			thinking = true
		}
		maxTokens := entry.MaxTokens
		if maxTokens <= 0 {
			maxTokens = 8192
		}

		out = append(out, GatewayModel{
			ID:          ToGatewayModelID(modelID),
			DisplayName: displayName,
			MaxTokens:   maxTokens,
			Thinking:    thinking,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

// GatewayModelsResponse is the Anthropic-native /v1/models payload.
type GatewayModelsResponse struct {
	Data    []GatewayModelInfo `json:"data"`
	FirstID string             `json:"first_id,omitempty"`
	LastID  string             `json:"last_id,omitempty"`
	HasMore bool               `json:"has_more"`
}

// GatewayModelInfo matches the Anthropic Models API shape used by Claude Code discovery.
type GatewayModelInfo struct {
	ID             string             `json:"id"`
	Type           string             `json:"type"`
	DisplayName    string             `json:"display_name"`
	CreatedAt      string             `json:"created_at"`
	MaxInputTokens int                `json:"max_input_tokens"`
	MaxTokens      int                `json:"max_tokens"`
	Capabilities   GatewayCapabilities `json:"capabilities"`
}

type GatewayCapabilities struct {
	Thinking GatewayThinkingCapability `json:"thinking"`
	Effort   GatewayEffortCapability   `json:"effort"`
}

type GatewayThinkingCapability struct {
	Supported bool `json:"supported"`
}

type GatewayEffortCapability struct {
	Supported bool `json:"supported"`
}

// BuildGatewayModelsResponse builds the Anthropic /v1/models response from config.
func BuildGatewayModelsResponse(cfg *config.Config) GatewayModelsResponse {
	models := ListGatewayModels(cfg)
	data := make([]GatewayModelInfo, 0, len(models))
	createdAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	for _, model := range models {
		thinking := GatewayThinkingCapability{Supported: model.Thinking}
		effort := GatewayEffortCapability{Supported: model.Thinking}
		data = append(data, GatewayModelInfo{
			ID:             model.ID,
			Type:           "model",
			DisplayName:    model.DisplayName,
			CreatedAt:      createdAt,
			MaxInputTokens: 200000,
			MaxTokens:      model.MaxTokens,
			Capabilities: GatewayCapabilities{
				Thinking: thinking,
				Effort:   effort,
			},
		})
	}

	resp := GatewayModelsResponse{
		Data:    data,
		HasMore: false,
	}
	if len(data) > 0 {
		resp.FirstID = data[0].ID
		resp.LastID = data[len(data)-1].ID
	}
	return resp
}
