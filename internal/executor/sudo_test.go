package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripSudo_SimpleCommand(t *testing.T) {
	assert.Equal(t, "apt-get install -y git", StripSudo("sudo apt-get install -y git"))
}

func TestStripSudo_ChainedCommands(t *testing.T) {
	input := "sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tar.gz"
	expected := "rm -rf /usr/local/go && tar -C /usr/local -xzf /tmp/go.tar.gz"
	assert.Equal(t, expected, StripSudo(input))
}

func TestStripSudo_PipedCommand(t *testing.T) {
	input := `curl -fsSL "https://example.com/script.sh" | sudo bash`
	expected := `curl -fsSL "https://example.com/script.sh" | bash`
	assert.Equal(t, expected, StripSudo(input))
}

func TestStripSudo_TeeCommand(t *testing.T) {
	input := "cat $out | sudo tee /etc/apt/keyrings/key.gpg > /dev/null"
	expected := "cat $out | tee /etc/apt/keyrings/key.gpg > /dev/null"
	assert.Equal(t, expected, StripSudo(input))
}

func TestStripSudo_NoSudo(t *testing.T) {
	input := "curl -fsSL https://example.com | bash"
	assert.Equal(t, input, StripSudo(input))
}

func TestStripSudo_MultiLine(t *testing.T) {
	input := "sudo mkdir -p /etc/apt/keyrings\nsudo chmod go+r /etc/apt/keyrings/key.gpg"
	expected := "mkdir -p /etc/apt/keyrings\nchmod go+r /etc/apt/keyrings/key.gpg"
	assert.Equal(t, expected, StripSudo(input))
}
