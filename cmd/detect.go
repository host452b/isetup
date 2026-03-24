package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/host452b/isetup/internal/detector"
	"github.com/spf13/cobra"
)

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Print detected system information",
	RunE: func(cmd *cobra.Command, args []string) error {
		info := detector.Detect()
		data, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(detectCmd)
}
