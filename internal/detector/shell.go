package detector

import (
	"os"
	"runtime"
	"strings"
)

func DetectShell() (shell, psVersion string) {
	switch runtime.GOOS {
	case "windows":
		psVersion = detectPowerShellVersion()
		if psVersion != "" {
			shell = "powershell"
		}
	default:
		shell = os.Getenv("SHELL")
		psVersion = detectPowerShellVersion()
	}
	return shell, psVersion
}

func detectPowerShellVersion() string {
	out := runCmd("pwsh", "-NoProfile", "-Command", "$PSVersionTable.PSVersion.ToString()")
	if out != "" {
		return parsePSVersion(out)
	}
	out = runCmd("powershell", "-NoProfile", "-Command", "$PSVersionTable.PSVersion.ToString()")
	if out != "" {
		return parsePSVersion(out)
	}
	return ""
}

func parsePSVersion(raw string) string {
	return strings.TrimSpace(raw)
}
