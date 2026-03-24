package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/host452b/isetup/internal/detector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger_CreatesLogDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")
	lg, err := New(dir)
	require.NoError(t, err)
	assert.DirExists(t, dir)
	assert.NotEmpty(t, lg.LogPath())
	assert.NotEmpty(t, lg.EnvJSONPath())
}

func TestWriteEnvJSON(t *testing.T) {
	dir := t.TempDir()
	lg, err := New(dir)
	require.NoError(t, err)

	info := &detector.SystemInfo{
		OS:   "linux",
		Arch: "amd64",
	}
	err = lg.WriteEnvJSON(info, "0.1.0", "/home/user/.isetup.yaml", 1)
	require.NoError(t, err)

	data, err := os.ReadFile(lg.EnvJSONPath())
	require.NoError(t, err)

	var env map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &env))
	assert.Equal(t, "linux", env["os"])
	assert.Equal(t, "amd64", env["arch"])
	assert.Equal(t, "0.1.0", env["isetup_version"])
	assert.Equal(t, float64(1), env["config_version"])
}

func TestWriteToolResult_Success(t *testing.T) {
	dir := t.TempDir()
	lg, err := New(dir)
	require.NoError(t, err)

	result := ToolResult{
		Name:     "git",
		Profile:  "base",
		Method:   "apt",
		Command:  "sudo apt-get install -y git",
		ExitCode: 0,
		Duration: 2 * time.Second,
		Stdout:   "installed",
		Stderr:   "",
		Status:   StatusSuccess,
	}
	err = lg.WriteToolResult(result)
	require.NoError(t, err)

	data, err := os.ReadFile(lg.LogPath())
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "INSTALL: git")
	assert.Contains(t, content, "sudo apt-get install -y git")
	assert.Contains(t, content, "SUCCESS")
}

func TestWriteToolResult_Failed(t *testing.T) {
	dir := t.TempDir()
	lg, err := New(dir)
	require.NoError(t, err)

	result := ToolResult{
		Name:     "cuda",
		Profile:  "gpu",
		Method:   "apt",
		Command:  "sudo apt-get install -y cuda",
		ExitCode: 100,
		Duration: 1 * time.Second,
		Stdout:   "",
		Stderr:   "E: Unable to locate package",
		Status:   StatusFailed,
	}
	err = lg.WriteToolResult(result)
	require.NoError(t, err)

	data, err := os.ReadFile(lg.LogPath())
	require.NoError(t, err)
	assert.Contains(t, string(data), "FAILED")
	assert.Contains(t, string(data), "E: Unable to locate package")
}

func TestWriteToolResult_Skipped(t *testing.T) {
	dir := t.TempDir()
	lg, err := New(dir)
	require.NoError(t, err)

	result := ToolResult{
		Name:       "cuda",
		Profile:    "gpu",
		Status:     StatusSkipped,
		SkipReason: "condition not met: has_gpu",
	}
	err = lg.WriteToolResult(result)
	require.NoError(t, err)

	data, err := os.ReadFile(lg.LogPath())
	require.NoError(t, err)
	assert.Contains(t, string(data), "SKIPPED")
	assert.Contains(t, string(data), "condition not met")
}
