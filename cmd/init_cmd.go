package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/spf13/cobra"
)

var (
	forceInit       bool
	defaultTemplate []byte
)

func SetDefaultTemplate(data []byte) {
	defaultTemplate = data
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate default ~/.isetup.yaml config",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home dir: %w", err)
		}
		dest := filepath.Join(home, ".isetup.yaml")

		if _, err := os.Stat(dest); err == nil && !forceInit {
			return fmt.Errorf("config already exists at %s. Use --force to overwrite", dest)
		}

		if err := os.WriteFile(dest, defaultTemplate, 0644); err != nil {
			return fmt.Errorf("write config: %w", err)
		}

		fmt.Printf("Config written to %s\n", dest)
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&forceInit, "force", false, "overwrite existing config")
	rootCmd.AddCommand(initCmd)
}
