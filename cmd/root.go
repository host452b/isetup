package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var (
	cfgPath     string
	logDir      string
	timeoutFlag time.Duration
)

var rootCmd = &cobra.Command{
	Use:   "isetup",
	Short: "Cross-platform dev environment setup tool (CLI-only, no GUI apps)",
	Long: `isetup - Cross-platform dev environment setup tool

Detects your OS, hardware, and architecture, then adaptively runs
the right install commands to one-click deploy your dev environment.

Designed for command-line engineers. The default template includes
only terminal-based tools — no GUI applications.

Usage:
  isetup init                      Generate default config
  isetup install                   Interactive picker in a TTY, install-all otherwise
  isetup install -i                Force the interactive picker (arrow keys, Space, Enter)
  isetup install -p base,ai-tools  Install specific profiles (no picker)
  isetup install --dry-run         Preview without executing
  isetup list                      List profiles and tools without installing
  isetup detect                    Show system info`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "config file path (default ~/.isetup.yaml)")
	rootCmd.PersistentFlags().StringVar(&logDir, "log-dir", "", "log directory (default ~/.isetup/logs/)")
	rootCmd.PersistentFlags().DurationVar(&timeoutFlag, "timeout", 10*time.Minute, "max time per tool install (e.g. 5m, 30s)")
}

func Execute() error {
	return rootCmd.Execute()
}
