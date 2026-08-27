package main

import (
	"testing"

	"github.com/host452b/isetup/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultTemplateIncludesOfficialHerdrInstallers(t *testing.T) {
	cfg, err := config.LoadFromBytes(defaultTemplate)
	require.NoError(t, err)

	aiTools, ok := cfg.Profiles["04-ai-tools"]
	require.True(t, ok, "default template should contain the 04-ai-tools profile")

	var herdr *config.Tool
	for i := range aiTools.Tools {
		if aiTools.Tools[i].Name == "herdr" {
			herdr = &aiTools.Tools[i]
			break
		}
	}
	require.NotNil(t, herdr, "04-ai-tools should contain herdr")

	assert.Equal(t, "curl", herdr.DependsOn)
	assert.Equal(t, "curl -fsSL https://herdr.dev/install.sh | sh", herdr.Shell.Unix)
	assert.Equal(t, `powershell -ExecutionPolicy Bypass -c "irm https://herdr.dev/install.ps1 | iex"`, herdr.Shell.Windows)
}
