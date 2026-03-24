package main

import (
	"fmt"
	"os"
	"github.com/host452b/isetup/cmd"
)

func main() {
	cmd.SetDefaultTemplate(defaultTemplate)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
