package executor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_Success(t *testing.T) {
	result := Run("echo hello", "bash")
	require.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "hello")
	assert.True(t, result.Duration > 0)
}

func TestRun_Failure(t *testing.T) {
	result := Run("exit 1", "bash")
	assert.NotEqual(t, 0, result.ExitCode)
}

func TestRun_CapturesStderr(t *testing.T) {
	result := Run("echo error >&2", "bash")
	assert.Contains(t, result.Stderr, "error")
}

func TestRun_MeasuresDuration(t *testing.T) {
	result := Run("sleep 0.1", "bash")
	assert.True(t, result.Duration >= 100*time.Millisecond)
}
