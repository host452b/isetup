package detector

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectShell_ReturnsString(t *testing.T) {
	shell, _ := DetectShell()
	if runtime.GOOS != "windows" {
		assert.NotEmpty(t, shell)
	}
}

func TestParsePowerShellVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"7.4.1", "7.4.1"},
		{"5.1.22621.4111", "5.1.22621.4111"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, parsePSVersion(tt.input))
		})
	}
}
