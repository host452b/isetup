package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/host452b/isetup/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles and tools in config",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := resolveConfigPath()
		cfg, err := config.LoadFromFile(path)
		if err != nil {
			return err
		}

		names := make([]string, 0, len(cfg.Profiles))
		for name := range cfg.Profiles {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, profName := range names {
			prof := cfg.Profiles[profName]
			header := profName
			if prof.When != "" {
				header += fmt.Sprintf(" (when: %s)", prof.When)
			}
			fmt.Printf("\n[%s]\n", header)
			for _, tool := range prof.Tools {
				dep := ""
				if tool.DependsOn != "" {
					dep = fmt.Sprintf(" → depends_on: %s", tool.DependsOn)
				}
				fmt.Printf("  - %s%s\n", tool.Name, dep)
			}
		}
		fmt.Println()
		return nil
	},
}

func resolveConfigPath() string {
	if cfgPath != "" {
		return cfgPath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".isetup.yaml"
	}
	return filepath.Join(home, ".isetup.yaml")
}

func init() {
	rootCmd.AddCommand(listCmd)
}
