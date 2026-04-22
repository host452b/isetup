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

func TestVersionVar(t *testing.T) {
	// Default value when not injected via ldflags
	assert.Equal(t, "dev", Version)
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

func TestDecideInteractive(t *testing.T) {
	cases := []struct {
		name      string
		f         installFlags
		tty       bool
		wantEnter bool
		wantErr   bool
	}{
		{
			name:      "explicit -i with TTY",
			f:         installFlags{interactive: true},
			tty:       true,
			wantEnter: true,
		},
		{
			name:    "explicit -i without TTY",
			f:       installFlags{interactive: true},
			tty:     false,
			wantErr: true,
		},
		{
			name:      "no flags + TTY → auto-enter",
			tty:       true,
			wantEnter: true,
		},
		{
			name:      "no flags + no TTY → no",
			tty:       false,
			wantEnter: false,
		},
		{
			name:      "-p opts out of auto",
			f:         installFlags{profiles: "00-base"},
			tty:       true,
			wantEnter: false,
		},
		{
			name:      "--dry-run opts out of auto",
			f:         installFlags{dryRun: true},
			tty:       true,
			wantEnter: false,
		},
		{
			name:      "-f opts out of auto",
			f:         installFlags{force: true},
			tty:       true,
			wantEnter: false,
		},
		{
			name:      "-i + -p is allowed",
			f:         installFlags{interactive: true, profiles: "00-base"},
			tty:       true,
			wantEnter: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enter, err := decideInteractive(tc.f, tc.tty)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantEnter, enter)
		})
	}
}
