package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/host452b/isetup/internal/detector"
)

const (
	StatusSuccess = "SUCCESS"
	StatusFailed  = "FAILED"
	StatusSkipped = "SKIPPED"
)

type ToolResult struct {
	Name       string
	Profile    string
	Method     string
	Command    string
	ExitCode   int
	Duration   time.Duration
	Stdout     string
	Stderr     string
	Status     string
	Condition  string
	SkipReason string
}

type Logger struct {
	logPath string
	envPath string
}

func New(dir string) (*Logger, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}

	ts := time.Now().Format("2006-01-02T15-04-05")
	return &Logger{
		logPath: filepath.Join(dir, fmt.Sprintf("isetup-%s.log", ts)),
		envPath: filepath.Join(dir, fmt.Sprintf("isetup-%s.env.json", ts)),
	}, nil
}

func (l *Logger) LogPath() string     { return l.logPath }
func (l *Logger) EnvJSONPath() string { return l.envPath }

func (l *Logger) WriteEnvJSON(info *detector.SystemInfo, version, configPath string, configVersion int) error {
	env := map[string]interface{}{
		"os":                 info.OS,
		"arch":               info.Arch,
		"arch_label":         info.ArchLabel,
		"distro":             info.Distro,
		"kernel":             info.Kernel,
		"wsl":                info.WSL,
		"shell":              info.Shell,
		"powershell_version": info.PowerShellVersion,
		"gpu":                info.GPU,
		"pkg_managers":       info.PkgManagers,
		"env_vars": map[string]string{
			"PATH": os.Getenv("PATH"),
			"HOME": os.Getenv("HOME"),
			"LANG": os.Getenv("LANG"),
		},
		"isetup_version": version,
		"config_path":    configPath,
		"config_version": configVersion,
		"timestamp":      time.Now().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal env json: %w", err)
	}
	return os.WriteFile(l.envPath, data, 0644)
}

func (l *Logger) WriteToolResult(r ToolResult) error {
	f, err := os.OpenFile(l.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer f.Close()

	ts := time.Now().Format(time.RFC3339)

	if r.Status == StatusSkipped {
		entry := fmt.Sprintf("=====================================\n[%s] INSTALL: %s\n  Profile: %s\n  Status: SKIPPED (%s)\n",
			ts, r.Name, r.Profile, r.SkipReason)
		_, err = f.WriteString(entry)
		return err
	}

	stderr := r.Stderr
	if stderr == "" {
		stderr = "(empty)"
	}
	stdout := r.Stdout
	if stdout == "" {
		stdout = "(empty)"
	}

	entry := fmt.Sprintf("=====================================\n[%s] INSTALL: %s\n  Profile: %s\n  Method: %s\n  Command: %s\n  Exit Code: %d\n  Duration: %s\n  STDOUT: |\n    %s\n  STDERR: |\n    %s\n  Status: %s\n",
		ts, r.Name, r.Profile, r.Method, r.Command,
		r.ExitCode, r.Duration, stdout, stderr, r.Status)

	if r.Condition != "" {
		entry = fmt.Sprintf("=====================================\n[%s] INSTALL: %s\n  Profile: %s\n  Condition: %s\n  Method: %s\n  Command: %s\n  Exit Code: %d\n  Duration: %s\n  STDOUT: |\n    %s\n  STDERR: |\n    %s\n  Status: %s\n",
			ts, r.Name, r.Profile, r.Condition, r.Method, r.Command,
			r.ExitCode, r.Duration, stdout, stderr, r.Status)
	}

	_, err = f.WriteString(entry)
	return err
}
