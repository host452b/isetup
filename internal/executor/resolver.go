package executor

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/host452b/isetup/internal/config"
	"github.com/host452b/isetup/internal/detector"
)

// safePkgRe allows: alphanumeric, hyphens, dots, underscores, slashes (npm scopes),
// at-signs, plus, equals, colons. Rejects shell metacharacters.
var safePkgRe = regexp.MustCompile(`^[a-zA-Z0-9@_.+:/-]+([=<>!]=?[a-zA-Z0-9._+-]+)?$`)

func isSafePkgName(name string) bool {
	return safePkgRe.MatchString(name)
}

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
		if hasPkgMgr(info.PkgManagers, "npm") && isSafePkgName(tool.Npm) {
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

func sudoPrefix(info *detector.SystemInfo) string {
	if info.IsRoot {
		return ""
	}
	return "sudo "
}

func resolvePkgMgr(tool config.Tool, info *detector.SystemInfo) (string, string) {
	sudo := sudoPrefix(info)
	switch info.OS {
	case "linux":
		if tool.Apt != "" && isSafePkgName(tool.Apt) {
			if hasPkgMgr(info.PkgManagers, "apt") {
				return "apt", fmt.Sprintf("%sapt install -y %s", sudo, tool.Apt)
			}
			if hasPkgMgr(info.PkgManagers, "apt-get") {
				return "apt", fmt.Sprintf("%sapt-get install -y %s", sudo, tool.Apt)
			}
		}
		if tool.Dnf != "" && hasPkgMgr(info.PkgManagers, "dnf") && isSafePkgName(tool.Dnf) {
			return "dnf", fmt.Sprintf("%sdnf install -y %s", sudo, tool.Dnf)
		}
		if tool.Pacman != "" && hasPkgMgr(info.PkgManagers, "pacman") && isSafePkgName(tool.Pacman) {
			return "pacman", fmt.Sprintf("%spacman -S --noconfirm %s", sudo, tool.Pacman)
		}
	case "darwin":
		if tool.Brew != "" && hasPkgMgr(info.PkgManagers, "brew") && isSafePkgName(tool.Brew) {
			return "brew", fmt.Sprintf("brew install %s", tool.Brew)
		}
	case "windows":
		if tool.Choco != "" && hasPkgMgr(info.PkgManagers, "choco") && isSafePkgName(tool.Choco) {
			return "choco", fmt.Sprintf("choco install %s -y", tool.Choco)
		}
		if tool.Winget != "" && hasPkgMgr(info.PkgManagers, "winget") && isSafePkgName(tool.Winget) {
			return "winget", fmt.Sprintf("winget install --id %s -e --accept-source-agreements", tool.Winget)
		}
	}
	return "", ""
}

func resolvePip(packages []string, info *detector.SystemInfo) string {
	for _, pkg := range packages {
		if !isSafePkgName(pkg) {
			return ""
		}
	}
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
