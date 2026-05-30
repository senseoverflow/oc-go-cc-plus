package doctor

import (
	"context"
	"fmt"
	"strings"

	"oc-go-cc-plus/internal/config"
	"oc-go-cc-plus/internal/daemon"
	"oc-go-cc-plus/internal/models"
	"oc-go-cc-plus/internal/validate"
)

// Report summarizes diagnostic results.
type Report struct {
	ConfigPath string
	Errors     []string
	Warnings   []string
	Info       []string
}

// Run executes config validation, optional API check, and proxy status.
func Run(ctx context.Context, cfg *config.Config, cfgPath string, checkAPI bool) Report {
	report := Report{ConfigPath: cfgPath}

	for _, issue := range validate.Config(cfg) {
		line := fmt.Sprintf("[%s] %s", issue.Field, issue.Message)
		switch issue.Severity {
		case "error":
			report.Errors = append(report.Errors, line)
		case "warn":
			report.Warnings = append(report.Warnings, line)
		default:
			report.Info = append(report.Info, line)
		}
	}

	if checkAPI && cfg.APIKey != "" && !strings.HasPrefix(cfg.APIKey, "${") {
		ids, err := models.FetchRemoteIDs(ctx, cfg.APIKey)
		if err != nil {
			report.Errors = append(report.Errors, "API: "+err.Error())
		} else {
			report.Info = append(report.Info, fmt.Sprintf("API OK — %d modelli remoti", len(ids)))
		}
	}

	report.Info = append(report.Info,
		fmt.Sprintf("OpenAI endpoint: %s", cfg.OpenCodeGo.BaseURL),
		fmt.Sprintf("Anthropic endpoint: %s", cfg.OpenCodeGo.AnthropicBaseURL),
	)

	paths, err := daemon.DefaultPaths()
	if err == nil {
		if pid, err := daemon.GetPID(paths.PIDFile); err == nil && daemon.IsProcessRunning(pid) {
			report.Info = append(report.Info, fmt.Sprintf("Proxy in esecuzione (PID %d)", pid))
		} else {
			report.Warnings = append(report.Warnings, "Proxy non in esecuzione — avvia con: oc-go-cc-plus serve")
		}
	}

	return report
}

func (r Report) Healthy() bool {
	return len(r.Errors) == 0
}
