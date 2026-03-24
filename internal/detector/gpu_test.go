package detector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectGPU_DoesNotPanic(t *testing.T) {
	info := DetectGPU()
	assert.IsType(t, GPUInfo{}, info)
}

func TestParseNvidiaSMI(t *testing.T) {
	output := "NVIDIA GeForce RTX 4090, 545.23.08"
	model, driver := parseNvidiaSMI(output)
	assert.Equal(t, "NVIDIA GeForce RTX 4090", model)
	assert.Equal(t, "545.23.08", driver)
}

func TestParseNvidiaSMI_Empty(t *testing.T) {
	model, driver := parseNvidiaSMI("")
	assert.Empty(t, model)
	assert.Empty(t, driver)
}
