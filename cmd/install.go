package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/isetup-dev/isetup/internal/config"
	"github.com/isetup-dev/isetup/internal/detector"
	"github.com/isetup-dev/isetup/internal/executor"
	"github.com/isetup-dev/isetup/internal/logger"
	"github.com/spf13/cobra"
)

var (
	profilesFlag string
	dryRunFlag   bool
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install tools from config",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := resolveConfigPath()
		cfg, err := config.LoadFromFile(path)
		if err != nil {
			// No config file — fall back to embedded default template
			if os.IsNotExist(unwrapErr(err)) {
				fmt.Fprintf(os.Stderr, "No config found at %s, using built-in defaults\n", path)
				cfg, err = config.LoadFromBytes(defaultTemplate)
				if err != nil {
					return fmt.Errorf("load default template: %w", err)
				}
			} else {
				return fmt.Errorf("load config: %w", err)
			}
		}

		if dryRunFlag {
			cfg.Settings.DryRun = true
		}

		errs, warns := config.Validate(cfg)
		for _, w := range warns {
			fmt.Fprintf(os.Stderr, "WARN: %s\n", w)
		}
		if len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "ERROR: %s\n", e)
			}
			return fmt.Errorf("config validation failed with %d error(s)", len(errs))
		}

		info := detector.Detect()

		logPath, err := resolveLogDir()
		if err != nil {
			return err
		}
		lg, err := logger.New(logPath)
		if err != nil {
			return fmt.Errorf("setup logger: %w", err)
		}

		if err := lg.WriteEnvJSON(info, Version, path, cfg.Version); err != nil {
			fmt.Fprintf(os.Stderr, "WARN: failed to write env.json: %v\n", err)
		}

		var profiles []string
		if profilesFlag != "" {
			profiles = strings.Split(profilesFlag, ",")
		}

		if cfg.Settings.DryRun {
			fmt.Println("DRY RUN — commands will be printed but not executed")
			fmt.Println()
		}

		results, err := executor.Execute(cfg, info, lg, profiles)
		if err != nil {
			return err
		}

		// ANSI color codes
		green := "\033[32m"
		red := "\033[31m"
		yellow := "\033[33m"
		reset := "\033[0m"

		fmt.Println()
		success, failed, skipped := 0, 0, 0
		for _, r := range results {
			switch r.Status {
			case logger.StatusSuccess:
				success++
			case logger.StatusFailed:
				failed++
			case logger.StatusSkipped:
				skipped++
			}
		}

		// Print summary: all tools with colored status
		for _, r := range results {
			switch r.Status {
			case logger.StatusSuccess:
				fmt.Printf("%s%-20s PASS%s    (%-6s) %s\n", green, r.Name, reset, r.Method, r.Duration)
			case logger.StatusFailed:
				fmt.Printf("%s%-20s FAILED%s  (%-6s) %s  → see log\n", red, r.Name, reset, r.Method, r.Duration)
			case logger.StatusSkipped:
				fmt.Printf("%s%-20s SKIP%s    %s\n", yellow, r.Name, reset, r.SkipReason)
			}
		}

		fmt.Println("─────────────────────────────")
		fmt.Printf("Installed: %s%d%s | Failed: %s%d%s | Skipped: %s%d%s\n",
			green, success, reset, red, failed, reset, yellow, skipped, reset)
		fmt.Printf("Log: %s\n", lg.LogPath())

		if failed > 0 {
			return fmt.Errorf("%d tool(s) failed to install", failed)
		}
		return nil
	},
}

func unwrapErr(err error) error {
	for {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}

func resolveLogDir() (string, error) {
	if logDir != "" {
		return logDir, nil
	}
	home, err := os.UserHomeDir()
	if err == nil {
		primary := filepath.Join(home, ".isetup", "logs")
		if err := os.MkdirAll(primary, 0755); err == nil {
			return primary, nil
		}
	}
	fallback := "./isetup-logs"
	if err := os.MkdirAll(fallback, 0755); err != nil {
		return "", fmt.Errorf("cannot create log directory: tried ~/.isetup/logs/ and ./isetup-logs/")
	}
	return fallback, nil
}

func init() {
	installCmd.Flags().StringVarP(&profilesFlag, "profiles", "p", "", "comma-separated list of profiles to install")
	installCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "print commands without executing")
	rootCmd.AddCommand(installCmd)
}
