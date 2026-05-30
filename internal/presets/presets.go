package presets

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"

	"oc-go-cc-plus/internal/config"
)

//go:embed *.json
var files embed.FS

// Names returns available preset identifiers sorted alphabetically.
func Names() ([]string, error) {
	entries, err := files.ReadDir(".")
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) > 5 && name[len(name)-5:] == ".json" {
			names = append(names, name[:len(name)-5])
		}
	}
	sort.Strings(names)
	return names, nil
}

// Load returns the preset configuration by name.
func Load(name string) (*config.Config, error) {
	data, err := files.ReadFile(name + ".json")
	if err != nil {
		return nil, fmt.Errorf("preset %q not found: %w", name, err)
	}
	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing preset %q: %w", name, err)
	}
	return &cfg, nil
}

// Apply loads a preset and writes it to destPath, preserving api_key from existing config when present.
func Apply(name, destPath string) error {
	preset, err := Load(name)
	if err != nil {
		return err
	}

	if existing, err := config.LoadFromPath(destPath); err == nil && existing.APIKey != "" {
		preset.APIKey = existing.APIKey
	}

	return config.Save(destPath, preset)
}
