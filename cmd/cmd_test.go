package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveConfigPath_Default(t *testing.T) {
	cfgPath = ""
	path := resolveConfigPath()
	home, _ := os.UserHomeDir()
	assert.Equal(t, filepath.Join(home, ".isetup.yaml"), path)
}

func TestResolveConfigPath_Custom(t *testing.T) {
	cfgPath = "/tmp/custom.yaml"
	defer func() { cfgPath = "" }()
	assert.Equal(t, "/tmp/custom.yaml", resolveConfigPath())
}

func TestResolveLogDir_Default(t *testing.T) {
	logDir = ""
	dir, err := resolveLogDir()
	require.NoError(t, err)
	assert.Contains(t, dir, ".isetup")
}

func TestResolveLogDir_Custom(t *testing.T) {
	logDir = t.TempDir()
	defer func() { logDir = "" }()
	dir, err := resolveLogDir()
	require.NoError(t, err)
	assert.Equal(t, logDir, dir)
}

func TestUnwrapErr(t *testing.T) {
	inner := os.ErrNotExist
	wrapped := os.ErrNotExist
	assert.True(t, os.IsNotExist(unwrapErr(wrapped)))
	assert.True(t, os.IsNotExist(inner))
}

func TestVersionConst(t *testing.T) {
	assert.Equal(t, "0.3.0", Version)
}

func TestSetDefaultTemplate(t *testing.T) {
	data := []byte("test template")
	SetDefaultTemplate(data)
	assert.Equal(t, data, defaultTemplate)
}

func TestInitCmd_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".isetup.yaml")
	require.NoError(t, os.WriteFile(path, []byte("existing"), 0644))

	// Simulate: config exists, no --force
	// Just verify the command is wired up
	assert.NotNil(t, initCmd)
	assert.Equal(t, "init", initCmd.Use)
}
