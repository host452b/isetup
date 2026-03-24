package executor

import (
	"fmt"
	"os"
	"strings"

	"github.com/isetup-dev/isetup/internal/config"
	"github.com/isetup-dev/isetup/internal/detector"
)

func Resolve(tool config.Tool, info *detector.SystemInfo) (string, string) {
	// Priority 1: shell with exact OS match
	if cmd := resolveShell(tool.Shell, info.OS); cmd != "" {
		return "shell", cmd
	}

	// Priority 2: system package manager
	method, cmd := resolvePkgMgr(tool, info)
	if method != "" {
		return method, cmd
	}

	// Priority 3: pip
	if len(tool.Pip) > 0 {
		pipCmd := resolvePip(tool.Pip, info)
		if pipCmd != "" {
			return "pip", pipCmd
		}
	}

	// Priority 4: npm
	if tool.Npm != "" {
		if hasPkgMgr(info.PkgManagers, "npm") {
			return "npm", fmt.Sprintf("npm install -g %s", tool.Npm)
		}
	}

	return "", ""
}

func resolveShell(shell config.Shell, goos string) string {
	switch goos {
	case "linux":
		if shell.Linux != "" {
			return shell.Linux
		}
	case "darwin":
		if shell.Darwin != "" {
			return shell.Darwin
		}
	case "windows":
		if shell.Windows != "" {
			return shell.Windows
		}
	}
	if goos != "windows" && shell.Unix != "" {
		return shell.Unix
	}
	return ""
}

func resolvePkgMgr(tool config.Tool, info *detector.SystemInfo) (string, string) {
	switch info.OS {
	case "linux":
		if tool.Apt != "" && hasPkgMgr(info.PkgManagers, "apt") {
			return "apt", fmt.Sprintf("sudo apt-get install -y %s", tool.Apt)
		}
		if tool.Dnf != "" && hasPkgMgr(info.PkgManagers, "dnf") {
			return "dnf", fmt.Sprintf("sudo dnf install -y %s", tool.Dnf)
		}
		if tool.Pacman != "" && hasPkgMgr(info.PkgManagers, "pacman") {
			return "pacman", fmt.Sprintf("sudo pacman -S --noconfirm %s", tool.Pacman)
		}
	case "darwin":
		if tool.Brew != "" && hasPkgMgr(info.PkgManagers, "brew") {
			return "brew", fmt.Sprintf("brew install %s", tool.Brew)
		}
	case "windows":
		if tool.Choco != "" && hasPkgMgr(info.PkgManagers, "choco") {
			return "choco", fmt.Sprintf("choco install %s -y", tool.Choco)
		}
		if tool.Winget != "" && hasPkgMgr(info.PkgManagers, "winget") {
			return "winget", fmt.Sprintf("winget install --id %s -e --accept-source-agreements", tool.Winget)
		}
	}
	return "", ""
}

func resolvePip(packages []string, info *detector.SystemInfo) string {
	condaPip := fmt.Sprintf("%s/miniconda3/bin/pip", info.Home)
	if _, err := os.Stat(condaPip); err == nil {
		return fmt.Sprintf("%s install %s", condaPip, strings.Join(packages, " "))
	}
	if hasPkgMgr(info.PkgManagers, "pip3") {
		return fmt.Sprintf("pip3 install %s", strings.Join(packages, " "))
	}
	if hasPkgMgr(info.PkgManagers, "pip") {
		return fmt.Sprintf("pip install %s", strings.Join(packages, " "))
	}
	return ""
}

func hasPkgMgr(managers []string, name string) bool {
	for _, m := range managers {
		if m == name {
			return true
		}
	}
	return false
}
