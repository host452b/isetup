package detector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetect_ReturnsCompleteInfo(t *testing.T) {
	info := Detect()
	assert.NotEmpty(t, info.OS)
	assert.NotEmpty(t, info.Arch)
	assert.NotEmpty(t, info.ArchLabel)
	assert.NotNil(t, info.GPU)
	assert.NotNil(t, info.PkgManagers)
}
