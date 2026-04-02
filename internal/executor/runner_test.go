package executor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_Success(t *testing.T) {
	result := Run(context.Background(), "echo hello", "bash")
	require.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "hello")
	assert.True(t, result.Duration > 0)
}

func TestRun_Failure(t *testing.T) {
	result := Run(context.Background(), "exit 1", "bash")
	assert.NotEqual(t, 0, result.ExitCode)
}

func TestRun_CapturesStderr(t *testing.T) {
	result := Run(context.Background(), "echo error >&2", "bash")
	assert.Contains(t, result.Stderr, "error")
}

func TestRun_MeasuresDuration(t *testing.T) {
	result := Run(context.Background(), "sleep 0.1", "bash")
	assert.True(t, result.Duration >= 100*time.Millisecond)
}

func TestRun_RespectsTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	result := Run(ctx, "sleep 10", "bash")
	assert.NotEqual(t, 0, result.ExitCode)
	assert.Less(t, result.Duration, 2*time.Second)
}

func TestRun_SuccessWithinTimeout(t *testing.T) {
	ctx := context.Background()
	result := Run(ctx, "echo hello", "bash")
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "hello", result.Stdout)
}

func TestRun_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := Run(ctx, "echo hello", "bash")
	assert.NotEqual(t, 0, result.ExitCode)
}
