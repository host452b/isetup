package executor

import (
	"testing"

	"github.com/host452b/isetup/internal/config"
	"github.com/host452b/isetup/internal/detector"
	"github.com/stretchr/testify/assert"
)

func TestResolve_LinuxApt(t *testing.T) {
	tool := config.Tool{Name: "git", Apt: "git", Brew: "git"}
	info := &detector.SystemInfo{OS: "linux", PkgManagers: []string{"apt"}}
	method, cmd := Resolve(tool, info)
	assert.Equal(t, "apt", method)
	assert.Equal(t, "sudo apt-get install -y git", cmd)
}

func TestResolve_DarwinBrew(t *testing.T) {
	tool := config.Tool{Name: "git", Apt: "git", Brew: "git"}
	info := &detector.SystemInfo{OS: "darwin", PkgManagers: []string{"brew"}}
	method, cmd := Resolve(tool, info)
	assert.Equal(t, "brew", method)
	assert.Equal(t, "brew install git", cmd)
}

func TestResolve_WindowsChoco(t *testing.T) {
	tool := config.Tool{Name: "git", Choco: "git", Winget: "Git.Git"}
	info := &detector.SystemInfo{OS: "windows", PkgManagers: []string{"choco", "winget"}}
	method, cmd := Resolve(tool, info)
	assert.Equal(t, "choco", method)
	assert.Equal(t, "choco install git -y", cmd)
}

func TestResolve_WindowsWingetFallback(t *testing.T) {
	tool := config.Tool{Name: "cursor", Winget: "Anysphere.Cursor"}
	info := &detector.SystemInfo{OS: "windows", PkgManagers: []string{"winget"}}
	method, cmd := Resolve(tool, info)
	assert.Equal(t, "winget", method)
	assert.Equal(t, "winget install --id Anysphere.Cursor -e --accept-source-agreements", cmd)
}

func TestResolve_ShellExactOS(t *testing.T) {
	tool := config.Tool{
		Name: "miniconda",
		Shell: config.Shell{Linux: "linux-specific", Darwin: "darwin-specific", Unix: "unix-fallback"},
	}
	info := &detector.SystemInfo{OS: "linux"}
	method, cmd := Resolve(tool, info)
	assert.Equal(t, "shell", method)
	assert.Equal(t, "linux-specific", cmd)
}

func TestResolve_ShellUnixFallback(t *testing.T) {
	tool := config.Tool{Name: "nvm", Shell: config.Shell{Unix: "curl install.sh | bash"}}
	info := &detector.SystemInfo{OS: "darwin"}
	method, cmd := Resolve(tool, info)
	assert.Equal(t, "shell", method)
	assert.Equal(t, "curl install.sh | bash", cmd)
}

func TestResolve_Npm(t *testing.T) {
	tool := config.Tool{Name: "codex", Npm: "@openai/codex"}
	info := &detector.SystemInfo{OS: "linux", PkgManagers: []string{"npm"}}
	method, cmd := Resolve(tool, info)
	assert.Equal(t, "npm", method)
	assert.Equal(t, "npm install -g @openai/codex", cmd)
}

func TestResolve_NpmNotAvailable(t *testing.T) {
	tool := config.Tool{Name: "codex", Npm: "@openai/codex"}
	info := &detector.SystemInfo{OS: "linux", PkgManagers: []string{"apt"}}
	method, _ := Resolve(tool, info)
	assert.Equal(t, "", method)
}

func TestResolve_Pip(t *testing.T) {
	tool := config.Tool{Name: "linters", Pip: []string{"ruff", "black"}}
	info := &detector.SystemInfo{OS: "linux", PkgManagers: []string{"pip3"}}
	method, cmd := Resolve(tool, info)
	assert.Equal(t, "pip", method)
	assert.Equal(t, "pip3 install ruff black", cmd)
}

func TestResolve_PipNotAvailable(t *testing.T) {
	tool := config.Tool{Name: "linters", Pip: []string{"ruff"}}
	info := &detector.SystemInfo{OS: "linux", PkgManagers: []string{"apt"}}
	method, _ := Resolve(tool, info)
	assert.Equal(t, "", method)
}

func TestResolve_NoMatch(t *testing.T) {
	tool := config.Tool{Name: "windows-only", Choco: "something"}
	info := &detector.SystemInfo{OS: "linux", PkgManagers: []string{"apt"}}
	method, cmd := Resolve(tool, info)
	assert.Equal(t, "", method)
	assert.Equal(t, "", cmd)
}

func TestResolve_LinuxDnf(t *testing.T) {
	tool := config.Tool{Name: "git", Dnf: "git"}
	info := &detector.SystemInfo{OS: "linux", PkgManagers: []string{"dnf"}}
	method, cmd := Resolve(tool, info)
	assert.Equal(t, "dnf", method)
	assert.Equal(t, "sudo dnf install -y git", cmd)
}

func TestResolve_LinuxPacman(t *testing.T) {
	tool := config.Tool{Name: "git", Pacman: "git"}
	info := &detector.SystemInfo{OS: "linux", PkgManagers: []string{"pacman"}}
	method, cmd := Resolve(tool, info)
	assert.Equal(t, "pacman", method)
	assert.Equal(t, "sudo pacman -S --noconfirm git", cmd)
}

func TestResolve_LinuxApt_Root(t *testing.T) {
	tool := config.Tool{Name: "git", Apt: "git"}
	info := &detector.SystemInfo{OS: "linux", PkgManagers: []string{"apt"}, IsRoot: true}
	method, cmd := Resolve(tool, info)
	assert.Equal(t, "apt", method)
	assert.Equal(t, "apt-get install -y git", cmd)
}

func TestResolve_LinuxDnf_Root(t *testing.T) {
	tool := config.Tool{Name: "git", Dnf: "git"}
	info := &detector.SystemInfo{OS: "linux", PkgManagers: []string{"dnf"}, IsRoot: true}
	method, cmd := Resolve(tool, info)
	assert.Equal(t, "dnf", method)
	assert.Equal(t, "dnf install -y git", cmd)
}

func TestResolve_LinuxPacman_Root(t *testing.T) {
	tool := config.Tool{Name: "git", Pacman: "git"}
	info := &detector.SystemInfo{OS: "linux", PkgManagers: []string{"pacman"}, IsRoot: true}
	method, cmd := Resolve(tool, info)
	assert.Equal(t, "pacman", method)
	assert.Equal(t, "pacman -S --noconfirm git", cmd)
}
