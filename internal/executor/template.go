package executor

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/isetup-dev/isetup/internal/detector"
)

type templateVars struct {
	Arch   string
	OS     string
	Distro string
	Home   string
}

func Interpolate(cmd string, info *detector.SystemInfo) (string, error) {
	tmpl, err := template.New("cmd").Option("missingkey=error").Parse(cmd)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	vars := templateVars{
		Arch:   info.ArchLabel,
		OS:     info.OS,
		Distro: info.Distro,
		Home:   info.Home,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}
