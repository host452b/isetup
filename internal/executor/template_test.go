package executor

import (
	"testing"

	"github.com/host452b/isetup/internal/detector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInterpolate_ArchLabel(t *testing.T) {
	info := &detector.SystemInfo{ArchLabel: "x86_64", OS: "linux", Home: "/home/user"}
	result, err := Interpolate("https://example.com/{{.Arch}}.sh", info)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/x86_64.sh", result)
}

func TestInterpolate_MultipleVars(t *testing.T) {
	info := &detector.SystemInfo{ArchLabel: "arm64", OS: "darwin", Distro: "macOS 15", Home: "/Users/user"}
	result, err := Interpolate("{{.OS}}-{{.Arch}}-{{.Home}}", info)
	require.NoError(t, err)
	assert.Equal(t, "darwin-arm64-/Users/user", result)
}

func TestInterpolate_NoVars(t *testing.T) {
	info := &detector.SystemInfo{OS: "linux"}
	result, err := Interpolate("apt-get install git", info)
	require.NoError(t, err)
	assert.Equal(t, "apt-get install git", result)
}

func TestInterpolate_InvalidTemplate(t *testing.T) {
	info := &detector.SystemInfo{}
	_, err := Interpolate("{{.Invalid}}", info)
	assert.Error(t, err)
}
