package validate

import (
	"fmt"
	"strings"

	"oc-go-cc-plus/internal/config"
	"oc-go-cc-plus/internal/models"
)

// Issue describes a config problem with severity.
type Issue struct {
	Severity string // error, warn, info
	Field    string
	Message  string
}

// Config runs extended validation and returns all issues found.
func Config(cfg *config.Config) []Issue {
	var issues []Issue

	if cfg.APIKey == "" {
		issues = append(issues, Issue{
			Severity: "error",
			Field:    "api_key",
			Message:  "api_key is required (OC_GO_CC_PLUS_API_KEY / OC_GO_CC_API_KEY)",
		})
	} else if strings.HasPrefix(cfg.APIKey, "${") {
		issues = append(issues, Issue{
			Severity: "error",
			Field:    "api_key",
			Message:  "api_key env var is not set",
		})
	}

	if cfg.Port < 1 || cfg.Port > 65535 {
		issues = append(issues, Issue{
			Severity: "error",
			Field:    "port",
			Message:  fmt.Sprintf("port must be 1-65535, got %d", cfg.Port),
		})
	}

	if cfg.OpenCodeGo.BaseURL == "" {
		issues = append(issues, Issue{Severity: "error", Field: "opencode_go.base_url", Message: "base_url is required"})
	}
	if cfg.OpenCodeGo.AnthropicBaseURL == "" {
		issues = append(issues, Issue{Severity: "error", Field: "opencode_go.anthropic_base_url", Message: "anthropic_base_url is required"})
	}

	scenarioSet := make(map[string]struct{}, len(models.ScenarioKeys))
	for _, key := range models.ScenarioKeys {
		scenarioSet[key] = struct{}{}
	}

	for _, key := range models.ScenarioKeys {
		modelCfg, ok := cfg.Models[key]
		if !ok {
			issues = append(issues, Issue{
				Severity: "warn",
				Field:    "models." + key,
				Message:  "scenario not configured",
			})
			continue
		}
		issues = append(issues, validateModelRef(key, modelCfg)...)
	}

	for name, modelCfg := range cfg.Models {
		if _, isScenario := scenarioSet[name]; isScenario {
			continue
		}
		if modelCfg.ModelID == "" {
			issues = append(issues, Issue{Severity: "error", Field: "models." + name, Message: "model_id is empty"})
			continue
		}
		issues = append(issues, validateModelRef(name, modelCfg)...)
	}

	for scenario, chain := range cfg.Fallbacks {
		for i, modelCfg := range chain {
			field := fmt.Sprintf("fallbacks.%s[%d]", scenario, i)
			if modelCfg.ModelID == "" {
				issues = append(issues, Issue{Severity: "error", Field: field, Message: "model_id is empty"})
				continue
			}
			for _, issue := range validateModelRef(field, modelCfg) {
				issue.Field = field
				issues = append(issues, issue)
			}
		}
	}

	return issues
}

func validateModelRef(field string, modelCfg config.ModelConfig) []Issue {
	var issues []Issue
	id := modelCfg.ModelID
	if id == "" {
		return issues
	}

	info, known := models.KnownModels[id]
	if !known {
		issues = append(issues, Issue{
			Severity: "warn",
			Field:    field,
			Message:  fmt.Sprintf("model %q not in registry (may still work)", id),
		})
		return issues
	}

	if info.Deprecated {
		msg := fmt.Sprintf("model %q is deprecated", id)
		if info.Replacement != "" {
			msg += fmt.Sprintf("; consider %q", info.Replacement)
		}
		issues = append(issues, Issue{Severity: "warn", Field: field, Message: msg})
	}

	issues = append(issues, Issue{
		Severity: "info",
		Field:    field,
		Message:  fmt.Sprintf("%s → %s", id, models.EndpointLabel(id)),
	})
	return issues
}

// HasErrors returns true if any issue is an error.
func HasErrors(issues []Issue) bool {
	for _, issue := range issues {
		if issue.Severity == "error" {
			return true
		}
	}
	return false
}
