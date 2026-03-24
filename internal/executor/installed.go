package executor

import (
	"os/exec"
	"strings"

	"github.com/isetup-dev/isetup/internal/config"
)

// IsInstalled checks if a tool is already available on the system.
func IsInstalled(tool config.Tool) bool {
	name := tool.Name

	// Check common binary names
	candidates := []string{name}

	// Add package-specific binary names
	if tool.Npm != "" {
		// npm packages: try the package name (last segment)
		parts := strings.Split(tool.Npm, "/")
		candidates = append(candidates, parts[len(parts)-1])
	}

	// Special cases where binary name differs from tool name
	switch name {
	case "node-lts":
		candidates = []string{"node"}
	case "golang":
		candidates = []string{"go"}
	case "miniconda":
		candidates = []string{"conda"}
	case "rust":
		candidates = []string{"rustc", "cargo"}
	case "typescript":
		candidates = []string{"tsc"}
	case "codex-cli":
		candidates = []string{"codex"}
	case "claude-code":
		candidates = []string{"claude"}
	case "pip-tools", "pip-build-tools", "pr-analyzers":
		// pip package groups — check if the first package is importable
		if len(tool.Pip) > 0 {
			candidates = []string{tool.Pip[0]}
		}
	case "cuda-toolkit":
		candidates = []string{"nvcc"}
	case "nvidia-driver":
		candidates = []string{"nvidia-smi"}
	case "tmux-ide":
		candidates = []string{"tmux-ide"}
	}

	for _, c := range candidates {
		if _, err := exec.LookPath(c); err == nil {
			return true
		}
	}
	return false
}
