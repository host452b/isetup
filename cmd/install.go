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

// ANSI color codes
const (
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorDim    = "\033[2m"
	colorReset  = "\033[0m"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install tools from config",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := resolveConfigPath()
		cfg, err := config.LoadFromFile(path)
		if err != nil {
			if os.IsNotExist(unwrapErr(err)) {
				fmt.Fprintf(os.Stderr, "%sNo config found at %s, using built-in defaults%s\n", colorDim, path, colorReset)
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
			fmt.Fprintf(os.Stderr, "%sWARN: %s%s\n", colorYellow, w, colorReset)
		}
		if len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "%sERROR: %s%s\n", colorRed, e, colorReset)
			}
			return fmt.Errorf("config validation failed with %d error(s)", len(errs))
		}

		fmt.Printf("%sDetecting system...%s\n", colorDim, colorReset)
		info := detector.Detect()
		fmt.Printf("%sOS: %s | Arch: %s | Shell: %s%s\n", colorDim, info.OS, info.Arch, info.Shell, colorReset)
		if len(info.PkgManagers) > 0 {
			fmt.Printf("%sPackage managers: %s%s\n", colorDim, strings.Join(info.PkgManagers, ", "), colorReset)
		}
		if info.GPU.Detected {
			fmt.Printf("%sGPU: %s%s\n", colorDim, info.GPU.Model, colorReset)
		}
		fmt.Println()

		logPath, err := resolveLogDir()
		if err != nil {
			return err
		}
		lg, err := logger.New(logPath)
		if err != nil {
			return fmt.Errorf("setup logger: %w", err)
		}

		if err := lg.WriteEnvJSON(info, Version, path, cfg.Version); err != nil {
			fmt.Fprintf(os.Stderr, "%sWARN: failed to write env.json: %v%s\n", colorYellow, err, colorReset)
		}

		var profiles []string
		if profilesFlag != "" {
			profiles = strings.Split(profilesFlag, ",")
		}

		if cfg.Settings.DryRun {
			fmt.Printf("%sDRY RUN — commands will be printed but not executed%s\n\n", colorCyan, colorReset)
		}

		// Real-time progress callback
		onProgress := func(ev executor.ProgressEvent) {
			step := fmt.Sprintf("[%d/%d]", ev.Index+1, ev.Total)
			switch ev.Phase {
			case "start":
				fmt.Printf("%s%s Installing %s%s (%s: %s)...\n",
					colorDim, step, colorReset, ev.Name, ev.Method, truncate(ev.Command, 60))
			case "done":
				if ev.Result == nil {
					return
				}
				switch ev.Result.Status {
				case logger.StatusSuccess:
					fmt.Printf("%s%s %s%-20s PASS%s    (%-6s) %s\n",
						colorDim, step, colorGreen, ev.Name, colorReset, ev.Result.Method, ev.Result.Duration)
				case logger.StatusFailed:
					fmt.Printf("%s%s %s%-20s FAILED%s  (%-6s) %s\n",
						colorDim, step, colorRed, ev.Name, colorReset, ev.Result.Method, ev.Result.Duration)
					if ev.Result.Stderr != "" {
						// Print first line of stderr for quick debug
						lines := strings.SplitN(ev.Result.Stderr, "\n", 2)
						fmt.Printf("       %s%s%s\n", colorDim, truncate(lines[0], 80), colorReset)
					}
				case logger.StatusSkipped:
					fmt.Printf("%s%s %s%-20s SKIP%s    %s\n",
						colorDim, step, colorYellow, ev.Name, colorReset, ev.Result.SkipReason)
				}
			}
		}

		results, err := executor.Execute(cfg, info, lg, profiles, onProgress)
		if err != nil {
			return err
		}

		// Summary
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

		fmt.Println()
		fmt.Println("─────────────────────────────")
		fmt.Printf("Installed: %s%d%s | Failed: %s%d%s | Skipped: %s%d%s\n",
			colorGreen, success, colorReset, colorRed, failed, colorReset, colorYellow, skipped, colorReset)
		fmt.Printf("Log: %s\n", lg.LogPath())

		if failed > 0 {
			return fmt.Errorf("%d tool(s) failed to install", failed)
		}
		return nil
	},
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
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
