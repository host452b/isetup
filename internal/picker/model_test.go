package picker

import (
	"testing"

	"github.com/host452b/isetup/internal/config"
	"github.com/host452b/isetup/internal/detector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func linuxAptInfo() *detector.SystemInfo {
	return &detector.SystemInfo{
		OS:          "linux",
		Arch:        "amd64",
		PkgManagers: []string{"apt"},
		GPU:         detector.GPUInfo{Detected: false},
	}
}

func TestNew_DefaultCheckState(t *testing.T) {
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"00-base": {Tools: []config.Tool{
				{Name: "git", Apt: "git"},
				{Name: "only-brew", Brew: "x"},
			}},
			"07-gpu": {When: "has_gpu", Tools: []config.Tool{
				{Name: "cuda", Apt: "cuda"},
			}},
		},
	}
	m := New(cfg, linuxAptInfo())

	var base, gpu *Node
	for _, n := range m.Nodes {
		if n.Kind == KindProfile {
			switch n.Name {
			case "00-base":
				base = n
			case "07-gpu":
				gpu = n
			}
		}
	}
	require.NotNil(t, base)
	require.NotNil(t, gpu)

	assert.False(t, base.Disabled, "00-base should be enabled")
	assert.True(t, gpu.Disabled, "07-gpu should be disabled (no GPU)")
	assert.False(t, base.Expanded, "profiles start collapsed")
	assert.False(t, gpu.Expanded)

	require.Len(t, base.ChildIdxs, 2)
	git := m.Nodes[base.ChildIdxs[0]]
	onlyBrew := m.Nodes[base.ChildIdxs[1]]

	assert.Equal(t, "git", git.Name)
	assert.False(t, git.Disabled)
	assert.Equal(t, Checked, git.Check)
	assert.Equal(t, "apt", git.Method)

	assert.Equal(t, "only-brew", onlyBrew.Name)
	assert.True(t, onlyBrew.Disabled)
	assert.Equal(t, Unchecked, onlyBrew.Check)
	assert.Equal(t, "", onlyBrew.Method)

	// base aggregate: only git is selectable and checked → Checked.
	assert.Equal(t, Checked, base.Check)

	cuda := m.Nodes[gpu.ChildIdxs[0]]
	assert.True(t, cuda.Disabled, "cuda inherits profile disabled")
	assert.Equal(t, Unchecked, cuda.Check)
}

func TestNew_ProfilesAreSortedAlphabetically(t *testing.T) {
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"02-git":  {Tools: []config.Tool{{Name: "gh", Apt: "gh"}}},
			"00-base": {Tools: []config.Tool{{Name: "git", Apt: "git"}}},
			"01-lang": {Tools: []config.Tool{{Name: "go", Apt: "golang"}}},
		},
	}
	m := New(cfg, linuxAptInfo())

	var profileOrder []string
	for _, n := range m.Nodes {
		if n.Kind == KindProfile {
			profileOrder = append(profileOrder, n.Name)
		}
	}
	assert.Equal(t, []string{"00-base", "01-lang", "02-git"}, profileOrder)
}

func TestNew_CursorStartsAtFirstNode(t *testing.T) {
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"00-base": {Tools: []config.Tool{{Name: "git", Apt: "git"}}},
		},
	}
	m := New(cfg, linuxAptInfo())
	assert.Equal(t, 0, m.Cursor)
	assert.Equal(t, KindProfile, m.Nodes[m.Cursor].Kind)
}

func testModel() *Model {
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"a-profile": {Tools: []config.Tool{
				{Name: "t1", Apt: "t1"},
				{Name: "t2", Apt: "t2"},
			}},
			"b-profile": {Tools: []config.Tool{
				{Name: "t3", Apt: "t3"},
			}},
		},
	}
	return New(cfg, linuxAptInfo())
}

func TestVisibleIndices_AllCollapsed(t *testing.T) {
	m := testModel()
	vis := m.visibleIndices()
	// 2 profile rows only, children hidden.
	assert.Len(t, vis, 2)
	assert.Equal(t, KindProfile, m.Nodes[vis[0]].Kind)
	assert.Equal(t, KindProfile, m.Nodes[vis[1]].Kind)
}

func TestMoveDown_StopsAtBottom(t *testing.T) {
	m := testModel()
	m.Cursor = 0
	m.MoveDown() // to b-profile
	second := m.Cursor
	m.MoveDown() // no-op at end
	assert.Equal(t, second, m.Cursor)
}

func TestMoveUp_StopsAtTop(t *testing.T) {
	m := testModel()
	m.Cursor = 0
	m.MoveUp()
	assert.Equal(t, 0, m.Cursor)
}

func TestMoveDown_SkipsCollapsedChildren(t *testing.T) {
	m := testModel()
	m.Cursor = 0 // a-profile (collapsed)
	m.MoveDown() // should jump over t1, t2 straight to b-profile
	assert.Equal(t, KindProfile, m.Nodes[m.Cursor].Kind)
	assert.Equal(t, "b-profile", m.Nodes[m.Cursor].Name)
}

func TestMoveDown_VisitsExpandedChildren(t *testing.T) {
	m := testModel()
	m.Nodes[0].Expanded = true // expand a-profile
	m.Cursor = 0
	m.MoveDown()
	assert.Equal(t, KindTool, m.Nodes[m.Cursor].Kind)
	assert.Equal(t, "t1", m.Nodes[m.Cursor].Name)
	m.MoveDown()
	assert.Equal(t, "t2", m.Nodes[m.Cursor].Name)
	m.MoveDown()
	assert.Equal(t, KindProfile, m.Nodes[m.Cursor].Kind)
	assert.Equal(t, "b-profile", m.Nodes[m.Cursor].Name)
}
