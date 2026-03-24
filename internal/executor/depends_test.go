package executor

import (
	"testing"

	"github.com/host452b/isetup/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTopoSort_NoDeps(t *testing.T) {
	tools := []ToolEntry{
		{Tool: config.Tool{Name: "git"}, Profile: "base"},
		{Tool: config.Tool{Name: "neovim"}, Profile: "base"},
	}
	sorted, err := TopoSort(tools)
	require.NoError(t, err)
	assert.Len(t, sorted, 2)
}

func TestTopoSort_SimpleDep(t *testing.T) {
	tools := []ToolEntry{
		{Tool: config.Tool{Name: "node-lts", DependsOn: "nvm"}, Profile: "node-dev"},
		{Tool: config.Tool{Name: "nvm"}, Profile: "node-dev"},
	}
	sorted, err := TopoSort(tools)
	require.NoError(t, err)
	require.Len(t, sorted, 2)
	assert.Equal(t, "nvm", sorted[0].Tool.Name)
	assert.Equal(t, "node-lts", sorted[1].Tool.Name)
}

func TestTopoSort_CrossProfile(t *testing.T) {
	tools := []ToolEntry{
		{Tool: config.Tool{Name: "pip-tools", DependsOn: "miniconda"}, Profile: "python-dev"},
		{Tool: config.Tool{Name: "git"}, Profile: "base"},
		{Tool: config.Tool{Name: "miniconda"}, Profile: "python-dev"},
	}
	sorted, err := TopoSort(tools)
	require.NoError(t, err)
	idxMini := indexOf(sorted, "miniconda")
	idxPip := indexOf(sorted, "pip-tools")
	assert.Less(t, idxMini, idxPip)
}

func TestTopoSort_Circular(t *testing.T) {
	tools := []ToolEntry{
		{Tool: config.Tool{Name: "a", DependsOn: "b"}, Profile: "base"},
		{Tool: config.Tool{Name: "b", DependsOn: "a"}, Profile: "base"},
	}
	_, err := TopoSort(tools)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular")
}

func TestTopoSort_CircularChainOfThree(t *testing.T) {
	tools := []ToolEntry{
		{Tool: config.Tool{Name: "a", DependsOn: "b"}, Profile: "base"},
		{Tool: config.Tool{Name: "b", DependsOn: "c"}, Profile: "base"},
		{Tool: config.Tool{Name: "c", DependsOn: "a"}, Profile: "base"},
	}
	_, err := TopoSort(tools)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular")
}

func TestTopoSort_UnresolvedDep(t *testing.T) {
	tools := []ToolEntry{
		{Tool: config.Tool{Name: "node-lts", DependsOn: "nvm"}, Profile: "node-dev"},
	}
	sorted, err := TopoSort(tools)
	require.NoError(t, err)
	assert.Len(t, sorted, 1)
	assert.True(t, sorted[0].UnresolvedDep)
}

func indexOf(entries []ToolEntry, name string) int {
	for i, e := range entries {
		if e.Tool.Name == name {
			return i
		}
	}
	return -1
}
