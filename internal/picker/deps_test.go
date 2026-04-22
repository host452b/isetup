package picker

import (
	"testing"

	"github.com/host452b/isetup/internal/config"
	"github.com/stretchr/testify/assert"
)

var testTools = []config.Tool{
	{Name: "curl"},
	{Name: "nvm", DependsOn: "curl"},
	{Name: "node-lts", DependsOn: "nvm"},
	{Name: "claude-code", DependsOn: "node-lts"},
	{Name: "codex-cli", DependsOn: "node-lts"},
	{Name: "unrelated"},
	{Name: "needs-missing", DependsOn: "not-in-config"},
}

func TestResolveDeps_Empty(t *testing.T) {
	closure, added := ResolveDeps(nil, testTools)
	assert.Empty(t, closure)
	assert.Empty(t, added)
}

func TestResolveDeps_SingleNoDep(t *testing.T) {
	closure, added := ResolveDeps([]string{"unrelated"}, testTools)
	assert.Equal(t, []string{"unrelated"}, closure)
	assert.Empty(t, added)
}

func TestResolveDeps_LinearChain(t *testing.T) {
	closure, added := ResolveDeps([]string{"claude-code"}, testTools)
	assert.Equal(t, []string{"claude-code", "curl", "node-lts", "nvm"}, closure)
	assert.Equal(t, []string{"curl", "node-lts", "nvm"}, added)
}

func TestResolveDeps_DiamondFromBelow(t *testing.T) {
	closure, added := ResolveDeps([]string{"claude-code", "codex-cli"}, testTools)
	assert.Equal(t, []string{"claude-code", "codex-cli", "curl", "node-lts", "nvm"}, closure)
	assert.Equal(t, []string{"curl", "node-lts", "nvm"}, added)
}

func TestResolveDeps_AlreadyHasDeps(t *testing.T) {
	closure, added := ResolveDeps([]string{"node-lts", "nvm", "curl"}, testTools)
	assert.Equal(t, []string{"curl", "node-lts", "nvm"}, closure)
	assert.Empty(t, added)
}

func TestResolveDeps_UnresolvedDep(t *testing.T) {
	closure, added := ResolveDeps([]string{"needs-missing"}, testTools)
	assert.Equal(t, []string{"needs-missing"}, closure)
	assert.Empty(t, added)
}

func TestResolveDeps_Cycle(t *testing.T) {
	cycleTools := []config.Tool{
		{Name: "A", DependsOn: "B"},
		{Name: "B", DependsOn: "A"},
	}
	closure, _ := ResolveDeps([]string{"A"}, cycleTools)
	assert.ElementsMatch(t, []string{"A", "B"}, closure)
}
