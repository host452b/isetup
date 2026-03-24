package detector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectPkgManagers_ReturnsSlice(t *testing.T) {
	managers := DetectPkgManagers()
	assert.IsType(t, []string{}, managers)
}

func TestIsInPath(t *testing.T) {
	assert.True(t, isInPath("go"))
	assert.False(t, isInPath("definitely_not_a_real_command_xyz"))
}
