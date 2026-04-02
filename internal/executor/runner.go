package executor

import (
	"bytes"
	"context"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type RunResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}

func Run(ctx context.Context, command, shell string) RunResult {
	start := time.Now()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		psExe := "powershell"
		if _, err := exec.LookPath("pwsh"); err == nil {
			psExe = "pwsh"
		}
		cmd = exec.CommandContext(ctx, psExe, "-NoProfile", "-Command", command)
	} else {
		if shell == "" {
			shell = "bash"
		}
		cmd = exec.CommandContext(ctx, shell, "-c", command)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return RunResult{
		ExitCode: exitCode,
		Stdout:   strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
		Duration: duration,
	}
}
