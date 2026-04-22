package main

import (
	"context"
	"testing"

	"github.com/host452b/isetup/internal/config"
	"github.com/host452b/isetup/internal/detector"
	"github.com/host452b/isetup/internal/executor"
	"github.com/host452b/isetup/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_DryRunFullPipeline(t *testing.T) {
	yamlData := []byte(`
version: 1
settings:
  log_level: debug
  dry_run: true
profiles:
  base:
    tools:
      - name: git
        apt: git
        brew: git
  gpu:
    when: has_gpu
    tools:
      - name: cuda
        apt: nvidia-cuda-toolkit
`)
	cfg, err := config.LoadFromBytes(yamlData)
	require.NoError(t, err)
	cfg.Settings.Force = true // skip "already installed" checks in tests

	errs, _ := config.Validate(cfg)
	require.Empty(t, errs)

	info := detector.Detect()
	lg, err := logger.New(t.TempDir())
	require.NoError(t, err)

	results, err := executor.Execute(context.Background(), cfg, info, lg, nil, nil, nil)
	require.NoError(t, err)
	require.NotEmpty(t, results)

	gitResult := findResult(results, "git")
	require.NotNil(t, gitResult)
	assert.Equal(t, logger.StatusSuccess, gitResult.Status)
	assert.NotEmpty(t, gitResult.Command)

	cudaResult := findResult(results, "cuda")
	require.NotNil(t, cudaResult)
	assert.Contains(t, []string{logger.StatusSuccess, logger.StatusSkipped}, cudaResult.Status)
}

func findResult(results []logger.ToolResult, name string) *logger.ToolResult {
	for _, r := range results {
		if r.Name == name {
			return &r
		}
	}
	return nil
}
