package executor

import (
	"testing"

	"github.com/host452b/isetup/internal/config"
	"github.com/host452b/isetup/internal/detector"
	"github.com/host452b/isetup/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testSystemInfo() *detector.SystemInfo {
	return &detector.SystemInfo{
		OS:          "linux",
		Arch:        "amd64",
		ArchLabel:   "x86_64",
		PkgManagers: []string{"apt"},
		GPU:         detector.GPUInfo{Detected: false},
	}
}

func TestExecute_DryRun(t *testing.T) {
	cfg := &config.Config{
		Version:  1,
		Settings: config.Settings{DryRun: true, Force: true},
		Profiles: map[string]config.Profile{
			"base": {Tools: []config.Tool{{Name: "git", Apt: "git"}}},
		},
	}
	info := testSystemInfo()
	lg, err := logger.New(t.TempDir())
	require.NoError(t, err)

	results, err := Execute(cfg, info, lg, nil, nil)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "git", results[0].Name)
	assert.Equal(t, logger.StatusSuccess, results[0].Status)
	assert.Equal(t, "sudo apt-get install -y git", results[0].Command)
}

func TestExecute_ProfileFilter(t *testing.T) {
	cfg := &config.Config{
		Version:  1,
		Settings: config.Settings{DryRun: true, Force: true},
		Profiles: map[string]config.Profile{
			"base": {Tools: []config.Tool{{Name: "git", Apt: "git"}}},
			"dev":  {Tools: []config.Tool{{Name: "neovim", Apt: "neovim"}}},
		},
	}
	info := testSystemInfo()
	lg, err := logger.New(t.TempDir())
	require.NoError(t, err)

	results, err := Execute(cfg, info, lg, []string{"base"}, nil)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "git", results[0].Name)
}

func TestExecute_SkipsWhenConditionNotMet(t *testing.T) {
	cfg := &config.Config{
		Version:  1,
		Settings: config.Settings{DryRun: true, Force: true},
		Profiles: map[string]config.Profile{
			"gpu": {When: "has_gpu", Tools: []config.Tool{{Name: "cuda", Apt: "cuda"}}},
		},
	}
	info := testSystemInfo() // GPU not detected
	lg, err := logger.New(t.TempDir())
	require.NoError(t, err)

	results, err := Execute(cfg, info, lg, nil, nil)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, logger.StatusSkipped, results[0].Status)
}

func TestExecute_SkipsDependencyFailed(t *testing.T) {
	cfg := &config.Config{
		Version:  1,
		Settings: config.Settings{DryRun: false, Force: true},
		Profiles: map[string]config.Profile{
			"dev": {Tools: []config.Tool{
				{Name: "nvm"},             // no install method → skip
				{Name: "node", DependsOn: "nvm", Apt: "node"},
			}},
		},
	}
	info := testSystemInfo()
	lg, err := logger.New(t.TempDir())
	require.NoError(t, err)

	results, err := Execute(cfg, info, lg, nil, nil)
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, logger.StatusSkipped, results[0].Status)
	assert.Equal(t, logger.StatusSkipped, results[1].Status)
	assert.Contains(t, results[1].SkipReason, "dependency")
}

func TestExecute_NoMatchMethod(t *testing.T) {
	cfg := &config.Config{
		Version:  1,
		Settings: config.Settings{DryRun: true, Force: true},
		Profiles: map[string]config.Profile{
			"base": {Tools: []config.Tool{{Name: "win-only", Choco: "something"}}},
		},
	}
	info := testSystemInfo() // linux, no choco
	lg, err := logger.New(t.TempDir())
	require.NoError(t, err)

	results, err := Execute(cfg, info, lg, nil, nil)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, logger.StatusSkipped, results[0].Status)
}
