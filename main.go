package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/host452b/isetup/cmd"
)

func main() {
	cmd.SetDefaultTemplate(defaultTemplate)
	if err := cmd.Execute(); err != nil {
		var exitErr *cmd.ExitError
		if errors.As(err, &exitErr) {
			fmt.Fprintln(os.Stderr, exitErr.Message)
			os.Exit(exitErr.Code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
