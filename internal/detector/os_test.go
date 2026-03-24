package detector

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectOS_ReturnsCurrentOS(t *testing.T) {
	info := DetectOS()
	assert.Equal(t, runtime.GOOS, info.OS)
	assert.Equal(t, runtime.GOARCH, info.Arch)
	assert.NotEmpty(t, info.OS)
}

func TestDetectOS_WSLDefaultsFalse(t *testing.T) {
	info := DetectOS()
	if runtime.GOOS != "linux" {
		assert.False(t, info.WSL)
	}
}

func TestArchLabel(t *testing.T) {
	tests := []struct {
		goarch string
		goos   string
		want   string
	}{
		{"amd64", "linux", "x86_64"},
		{"arm64", "linux", "aarch64"},
		{"arm64", "darwin", "arm64"},
		{"amd64", "darwin", "x86_64"},
		{"amd64", "windows", "x86_64"},
	}
	for _, tt := range tests {
		t.Run(tt.goarch+"_"+tt.goos, func(t *testing.T) {
			assert.Equal(t, tt.want, ArchLabel(tt.goarch, tt.goos))
		})
	}
}
