// Package main is the CLI entry point for oc-go-cc-plus.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"oc-go-cc-plus/internal/config"
	"oc-go-cc-plus/internal/daemon"
	"oc-go-cc-plus/internal/doctor"
	"oc-go-cc-plus/internal/models"
	"oc-go-cc-plus/internal/presets"
	"oc-go-cc-plus/internal/server"
	"oc-go-cc-plus/internal/sync"
	"oc-go-cc-plus/internal/validate"
)

const (
	appName     = "oc-go-cc-plus"
	pidFileName = "oc-go-cc-plus.pid"
)

var version = "0.2.0-dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   appName,
		Short: "Proxy Claude Code → OpenCode Go con preset, sync e routing endpoint",
		Long: `oc-go-cc-plus estende oc-go-cc con:
  • sync automatico modelli dall'API OpenCode Go
  • preset pronti (deepseek, budget, balanced, quality)
  • routing endpoint corretto (Qwen/MiniMax → Anthropic, resto → OpenAI)
  • validazione config estesa e comando doctor

Config: ~/.config/oc-go-cc-plus/config.json`,
		Version: version,
	}

	rootCmd.AddCommand(serveCmd())
	rootCmd.AddCommand(stopCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(validateCmd())
	rootCmd.AddCommand(modelsCmd())
	rootCmd.AddCommand(syncModelsCmd())
	rootCmd.AddCommand(presetCmd())
	rootCmd.AddCommand(doctorCmd())
	rootCmd.AddCommand(autostartCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func serveCmd() *cobra.Command {
	var configPath string
	var port int
	var background bool
	var daemonize bool

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Avvia il proxy",
		RunE: func(cmd *cobra.Command, args []string) error {
			if background && !daemonize {
				return daemon.ForkIntoBackground(daemon.BackgroundOpts{ConfigPath: configPath, Port: port})
			}
			if configPath != "" {
				_ = os.Setenv("OC_GO_CC_PLUS_CONFIG", configPath)
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("config: %w", err)
			}
			if port != 0 {
				cfg.Port = port
			}

			pidPath := getPIDPath()
			if !daemonize {
				if pid, err := daemon.GetPID(pidPath); err == nil && daemon.IsProcessRunning(pid) {
					return fmt.Errorf("server già in esecuzione (PID %d)", pid)
				}
				_ = os.Remove(pidPath)
			}

			if daemonize {
				paths, err := daemon.DefaultPaths()
				if err != nil {
					return err
				}
				if err := paths.EnsureConfigDir(); err != nil {
					return err
				}
				if err := daemon.DaemonizeSetup(paths); err != nil {
					return err
				}
			} else if err := daemon.WritePID(pidPath, os.Getpid()); err != nil {
				return fmt.Errorf("PID file: %w", err)
			}
			defer func() { _ = os.Remove(pidPath) }()

			atomicCfg := config.NewAtomicConfig(cfg, config.ResolveConfigPath())
			if port != 0 {
				atomicCfg.OnReload(func(newCfg *config.Config) { newCfg.Port = port })
			}

			srv, err := server.NewServer(atomicCfg)
			if err != nil {
				return err
			}

			if cfg.HotReload {
				watchCtx := context.Background()
				go func() {
					_ = config.WatchConfig(watchCtx, atomicCfg)
				}()
			}

			fmt.Printf("Avvio %s v%s\n", appName, version)
			fmt.Printf("In ascolto su %s:%d\n", cfg.Host, cfg.Port)
			fmt.Printf("OpenAI endpoint: %s\n", cfg.OpenCodeGo.BaseURL)
			fmt.Printf("Anthropic endpoint: %s\n", cfg.OpenCodeGo.AnthropicBaseURL)
			fmt.Println()
			fmt.Println("Claude Code:")
			fmt.Printf("  export ANTHROPIC_BASE_URL=http://%s:%d\n", cfg.Host, cfg.Port)
			fmt.Println("  export ANTHROPIC_AUTH_TOKEN=unused")
			return srv.Start()
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Percorso config")
	cmd.Flags().IntVarP(&port, "port", "p", 0, "Porta")
	cmd.Flags().BoolVarP(&background, "background", "b", false, "Esegui in background")
	cmd.Flags().BoolVar(&daemonize, "_daemonize", false, "")
	_ = cmd.Flags().MarkHidden("_daemonize")
	return cmd
}

func stopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Ferma il proxy",
		RunE: func(cmd *cobra.Command, args []string) error {
			pidPath := getPIDPath()
			pid, err := daemon.GetPID(pidPath)
			if err != nil {
				return fmt.Errorf("proxy non in esecuzione")
			}
			if err := daemon.StopProcess(pid); err != nil {
				return err
			}
			fmt.Printf("Stop inviato (PID %d)\n", pid)
			_ = os.Remove(pidPath)
			return nil
		},
	}
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Stato del proxy",
		RunE: func(cmd *cobra.Command, args []string) error {
			pidPath := getPIDPath()
			pid, err := daemon.GetPID(pidPath)
			if err != nil || !daemon.IsProcessRunning(pid) {
				fmt.Println("Proxy non in esecuzione")
				_ = os.Remove(pidPath)
				return nil
			}
			fmt.Printf("Proxy in esecuzione (PID %d)\n", pid)
			return nil
		},
	}
}

func initCmd() *cobra.Command {
	var presetName string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Crea config di default",
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir := getConfigDir()
			configPath := filepath.Join(configDir, "config.json")
			if _, err := os.Stat(configPath); err == nil {
				fmt.Printf("Config già presente: %s\n", configPath)
				return nil
			}
			if err := os.MkdirAll(configDir, 0o755); err != nil {
				return err
			}
			name := presetName
			if name == "" {
				name = "balanced"
			}
			if err := presets.Apply(name, configPath); err != nil {
				return err
			}
			fmt.Printf("Config creato: %s (preset: %s)\n", configPath, name)
			fmt.Println("Imposta OC_GO_CC_PLUS_API_KEY o modifica api_key nel file.")
			return nil
		},
	}
	cmd.Flags().StringVarP(&presetName, "preset", "p", "balanced", "Preset iniziale")
	return cmd
}

func validateCmd() *cobra.Command {
	var configPath string
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Valida la configurazione",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := loadConfig(configPath)
			if err != nil {
				return err
			}
			issues := validate.Config(cfg)
			printIssues(issues)
			if validate.HasErrors(issues) {
				return fmt.Errorf("config non valida: %s", path)
			}
			fmt.Printf("\nConfig valida: %s\n", path)
			return nil
		},
	}
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Percorso config")
	return cmd
}

func modelsCmd() *cobra.Command {
	var remote bool
	cmd := &cobra.Command{
		Use:   "models",
		Short: "Elenca modelli OpenCode Go",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("%-22s %-28s %s\n", "Model ID", "Endpoint", "Note")
			fmt.Println(strings.Repeat("─", 72))
			for _, id := range models.AllKnownIDs() {
				info := models.KnownModels[id]
				note := info.DisplayName
				if info.Deprecated {
					note += " (deprecato)"
				}
				fmt.Printf("%-22s %-28s %s\n", id, models.EndpointLabel(id), note)
			}
			if !remote {
				return nil
			}
			cfg, _, err := loadConfig("")
			if err != nil {
				return fmt.Errorf("per --remote serve config con api_key: %w", err)
			}
			ids, err := models.FetchRemoteIDs(cmd.Context(), cfg.APIKey)
			if err != nil {
				return err
			}
			fmt.Printf("\nModelli remoti (%d): %s\n", len(ids), strings.Join(ids, ", "))
			return nil
		},
	}
	cmd.Flags().BoolVar(&remote, "remote", false, "Interroga anche l'API OpenCode Go")
	return cmd
}

func syncModelsCmd() *cobra.Command {
	var configPath string
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "sync-models",
		Short: "Sincronizza modelli dall'API nel config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := loadConfig(configPath)
			if err != nil {
				return err
			}
			remoteIDs, err := models.FetchRemoteIDs(cmd.Context(), cfg.APIKey)
			if err != nil {
				return fmt.Errorf("sync API: %w", err)
			}
			result := sync.Models(cfg, remoteIDs)
			fmt.Printf("Modelli remoti: %d\n", len(result.RemoteIDs))
			if len(result.Added) > 0 {
				fmt.Printf("Aggiunti (%d): %s\n", len(result.Added), strings.Join(result.Added, ", "))
			} else {
				fmt.Println("Nessun nuovo modello da aggiungere.")
			}
			if len(result.Deprecated) > 0 {
				fmt.Printf("Deprecati trovati: %s\n", strings.Join(result.Deprecated, ", "))
			}
			if len(result.Unknown) > 0 {
				fmt.Printf("Non in registry: %s\n", strings.Join(result.Unknown, ", "))
			}
			if dryRun {
				fmt.Println("Dry-run: config non scritta.")
				return nil
			}
			if err := config.Save(path, cfg); err != nil {
				return err
			}
			fmt.Printf("Config aggiornata: %s\n", path)
			return nil
		},
	}
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Percorso config")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Mostra modifiche senza scrivere")
	return cmd
}

func presetCmd() *cobra.Command {
	var configPath string
	cmd := &cobra.Command{
		Use:   "preset",
		Short: "Gestisci preset di configurazione",
	}
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Elenca preset disponibili",
		RunE: func(cmd *cobra.Command, args []string) error {
			names, err := presets.Names()
			if err != nil {
				return err
			}
			fmt.Println("Preset disponibili:")
			for _, name := range names {
				desc := presetDescription(name)
				fmt.Printf("  • %-10s %s\n", name, desc)
			}
			return nil
		},
	}
	applyCmd := &cobra.Command{
		Use:   "apply [nome]",
		Short: "Applica un preset al config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := configPath
			if path == "" {
				path = config.ResolveConfigPath()
			}
			if err := presets.Apply(args[0], path); err != nil {
				return err
			}
			fmt.Printf("Preset %q applicato a %s\n", args[0], path)
			fmt.Println("La api_key esistente è stata preservata.")
			return nil
		},
	}
	applyCmd.Flags().StringVarP(&configPath, "config", "c", "", "Percorso config")
	cmd.AddCommand(listCmd, applyCmd)
	return cmd
}

func doctorCmd() *cobra.Command {
	var configPath string
	var skipAPI bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnostica config, API e proxy",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := loadConfig(configPath)
			if err != nil {
				return err
			}
			report := doctor.Run(cmd.Context(), cfg, path, !skipAPI)
			if len(report.Errors) > 0 {
				fmt.Println("Errori:")
				for _, line := range report.Errors {
					fmt.Println("  ✗", line)
				}
			}
			if len(report.Warnings) > 0 {
				fmt.Println("Avvisi:")
				for _, line := range report.Warnings {
					fmt.Println("  !", line)
				}
			}
			if len(report.Info) > 0 {
				fmt.Println("Info:")
				for _, line := range report.Info {
					fmt.Println("  ✓", line)
				}
			}
			if !report.Healthy() {
				return fmt.Errorf("doctor ha rilevato problemi")
			}
			fmt.Println("\nTutto OK.")
			return nil
		},
	}
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Percorso config")
	cmd.Flags().BoolVar(&skipAPI, "skip-api", false, "Salta test connettività API")
	return cmd
}

func autostartCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "autostart", Short: "Auto-avvio al login"}
	cmd.AddCommand(
		&cobra.Command{Use: "enable", RunE: func(cmd *cobra.Command, _ []string) error {
			var configPath string
			var port int
			if cmd.Flags().Changed("config") {
				configPath, _ = cmd.Flags().GetString("config")
			}
			if cmd.Flags().Changed("port") {
				port, _ = cmd.Flags().GetInt("port")
			}
			return daemon.EnableAutostart(configPath, port)
		}},
		&cobra.Command{Use: "disable", RunE: func(_ *cobra.Command, _ []string) error { return daemon.DisableAutostart() }},
		&cobra.Command{Use: "status", RunE: func(_ *cobra.Command, _ []string) error { return daemon.AutostartStatus() }},
	)
	cmd.PersistentFlags().StringP("config", "c", "", "Percorso config")
	cmd.PersistentFlags().IntP("port", "p", 0, "Porta")
	return cmd
}

func loadConfig(configPath string) (*config.Config, string, error) {
	path := configPath
	if path == "" {
		path = config.ResolveConfigPath()
	} else {
		_ = os.Setenv("OC_GO_CC_PLUS_CONFIG", path)
	}
	cfg, err := config.LoadFromPath(path)
	if err != nil {
		return nil, path, err
	}
	return cfg, path, nil
}

func printIssues(issues []validate.Issue) {
	for _, issue := range issues {
		prefix := "INFO"
		switch issue.Severity {
		case "error":
			prefix = "ERROR"
		case "warn":
			prefix = "WARN"
		}
		fmt.Printf("[%s] %s: %s\n", prefix, issue.Field, issue.Message)
	}
}

func presetDescription(name string) string {
	switch name {
	case "deepseek":
		return "DeepSeek V4 Pro/Flash con max thinking"
	case "budget":
		return "Massimo risparmio (MiMo, Qwen)"
	case "balanced":
		return "Bilanciato qualità/costo (Kimi, GLM)"
	case "quality":
		return "Massima qualità (GLM-5.1, Qwen3.7 Max)"
	default:
		return ""
	}
}

func getConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "oc-go-cc-plus")
}

func getPIDPath() string {
	paths, err := daemon.DefaultPaths()
	if err != nil {
		return filepath.Join(os.TempDir(), pidFileName)
	}
	return paths.PIDFile
}
