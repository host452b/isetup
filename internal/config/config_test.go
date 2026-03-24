package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromBytes_MinimalConfig(t *testing.T) {
	yaml := []byte(`
version: 1
settings:
  log_level: info
profiles:
  base:
    tools:
      - name: git
        apt: git
        brew: git
`)
	cfg, err := LoadFromBytes(yaml)
	require.NoError(t, err)
	assert.Equal(t, 1, cfg.Version)
	assert.Equal(t, "info", cfg.Settings.LogLevel)
	assert.Len(t, cfg.Profiles, 1)
	assert.Len(t, cfg.Profiles["base"].Tools, 1)
	assert.Equal(t, "git", cfg.Profiles["base"].Tools[0].Name)
	assert.Equal(t, "git", cfg.Profiles["base"].Tools[0].Apt)
	assert.Equal(t, "git", cfg.Profiles["base"].Tools[0].Brew)
}

func TestLoadFromBytes_FullToolFields(t *testing.T) {
	yaml := []byte(`
version: 1
settings:
  log_level: debug
  dry_run: true
profiles:
  dev:
    tools:
      - name: nvm
        depends_on: git
        shell:
          unix: "curl -o- https://example.com/install.sh | bash"
          windows: "irm https://example.com/install.ps1 | iex"
          linux: "apt-specific command"
          darwin: "brew-specific command"
      - name: linter
        pip:
          - ruff
          - black
      - name: codex
        npm: "@openai/codex"
      - name: cursor
        choco: cursor
        winget: Anysphere.Cursor
        dnf: cursor
        pacman: cursor
`)
	cfg, err := LoadFromBytes(yaml)
	require.NoError(t, err)

	nvm := cfg.Profiles["dev"].Tools[0]
	assert.Equal(t, "nvm", nvm.Name)
	assert.Equal(t, "git", nvm.DependsOn)
	assert.Equal(t, "curl -o- https://example.com/install.sh | bash", nvm.Shell.Unix)
	assert.Equal(t, "irm https://example.com/install.ps1 | iex", nvm.Shell.Windows)
	assert.Equal(t, "apt-specific command", nvm.Shell.Linux)
	assert.Equal(t, "brew-specific command", nvm.Shell.Darwin)

	linter := cfg.Profiles["dev"].Tools[1]
	assert.Equal(t, []string{"ruff", "black"}, linter.Pip)

	codex := cfg.Profiles["dev"].Tools[2]
	assert.Equal(t, "@openai/codex", codex.Npm)

	cursor := cfg.Profiles["dev"].Tools[3]
	assert.Equal(t, "cursor", cursor.Choco)
	assert.Equal(t, "Anysphere.Cursor", cursor.Winget)
}

func TestLoadFromBytes_ProfileWithWhen(t *testing.T) {
	yaml := []byte(`
version: 1
settings:
  log_level: info
profiles:
  gpu:
    when: has_gpu
    tools:
      - name: cuda
        apt: nvidia-cuda-toolkit
`)
	cfg, err := LoadFromBytes(yaml)
	require.NoError(t, err)
	assert.Equal(t, "has_gpu", cfg.Profiles["gpu"].When)
	assert.Len(t, cfg.Profiles["gpu"].Tools, 1)
}

func TestLoadFromBytes_ShellStringShorthand(t *testing.T) {
	yaml := []byte(`
version: 1
settings:
  log_level: info
profiles:
  base:
    tools:
      - name: tmux
        shell: "curl -fsSL https://tmux.example.com/install.sh | sh"
`)
	cfg, err := LoadFromBytes(yaml)
	require.NoError(t, err)
	tmux := cfg.Profiles["base"].Tools[0]
	assert.Equal(t, "curl -fsSL https://tmux.example.com/install.sh | sh", tmux.Shell.Unix)
}

func TestLoadFromBytes_InvalidYAML(t *testing.T) {
	yaml := []byte(`not: [valid: yaml`)
	_, err := LoadFromBytes(yaml)
	assert.Error(t, err)
}

func TestLoadFromFile(t *testing.T) {
	path := t.TempDir() + "/test.yaml"
	content := []byte(`
version: 1
settings:
  log_level: info
profiles:
  base:
    tools:
      - name: git
        apt: git
`)
	require.NoError(t, os.WriteFile(path, content, 0644))
	cfg, err := LoadFromFile(path)
	require.NoError(t, err)
	assert.Equal(t, 1, cfg.Version)
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path.yaml")
	assert.Error(t, err)
}
