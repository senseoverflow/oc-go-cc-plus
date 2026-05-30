package sync

import (
	"encoding/json"
	"sort"

	"oc-go-cc-plus/internal/config"
	"oc-go-cc-plus/internal/models"
)

// Result summarizes a sync-models operation.
type Result struct {
	RemoteIDs  []string
	Added      []string
	AlreadyHad []string
	Unknown    []string
	Deprecated []string
}

// Models adds named model entries for remote models missing from cfg.Models.
func Models(cfg *config.Config, remoteIDs []string) Result {
	result := Result{RemoteIDs: remoteIDs}
	scenarioSet := make(map[string]struct{}, len(models.ScenarioKeys))
	for _, key := range models.ScenarioKeys {
		scenarioSet[key] = struct{}{}
	}

	for _, id := range remoteIDs {
		if _, isScenario := scenarioSet[id]; isScenario {
			continue
		}
		if _, exists := cfg.Models[id]; exists {
			result.AlreadyHad = append(result.AlreadyHad, id)
			continue
		}
		if info, known := models.KnownModels[id]; known && info.Deprecated {
			result.Deprecated = append(result.Deprecated, id)
		}
		cfg.Models[id] = defaultModelEntry(id)
		result.Added = append(result.Added, id)
	}

	for id := range cfg.Models {
		if _, known := models.KnownModels[id]; !known {
			if _, isScenario := scenarioSet[id]; !isScenario {
				result.Unknown = append(result.Unknown, id)
			}
		}
	}

	sort.Strings(result.Added)
	sort.Strings(result.AlreadyHad)
	sort.Strings(result.Unknown)
	sort.Strings(result.Deprecated)
	return result
}

func defaultModelEntry(modelID string) config.ModelConfig {
	entry := config.ModelConfig{
		Provider:    "opencode-go",
		ModelID:     modelID,
		Temperature: 0.7,
		MaxTokens:   4096,
	}
	if modelID == "deepseek-v4-pro" || modelID == "deepseek-v4-flash" {
		entry.Temperature = 0.1
		entry.MaxTokens = 8192
		entry.ReasoningEffort = "max"
		entry.Thinking = json.RawMessage(`{"type":"enabled"}`)
	}
	if modelID == "minimax-m2.5" || modelID == "minimax-m2.7" {
		entry.MaxTokens = 16384
	}
	return entry
}
