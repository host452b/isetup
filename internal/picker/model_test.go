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

func TestToggle_ToolFlips(t *testing.T) {
	m := testModel()
	m.Nodes[0].Expanded = true
	m.Cursor = 1 // t1
	assert.Equal(t, Checked, m.Nodes[1].Check)
	m.Toggle()
	assert.Equal(t, Unchecked, m.Nodes[1].Check)
	m.Toggle()
	assert.Equal(t, Checked, m.Nodes[1].Check)
}

func TestToggle_ToolUpdatesParentAggregate(t *testing.T) {
	m := testModel()
	m.Nodes[0].Expanded = true
	m.Cursor = 1 // t1
	m.Toggle()                            // t1 unchecked; t2 still checked
	assert.Equal(t, Partial, m.Nodes[0].Check)
	m.Cursor = 2 // t2
	m.Toggle()                            // t2 unchecked; both unchecked
	assert.Equal(t, Unchecked, m.Nodes[0].Check)
}

func TestToggle_ProfileFullChecksAll(t *testing.T) {
	m := testModel()
	// Start: both children checked. Toggle profile → all unchecked.
	m.Cursor = 0
	m.Toggle()
	assert.Equal(t, Unchecked, m.Nodes[1].Check)
	assert.Equal(t, Unchecked, m.Nodes[2].Check)
	assert.Equal(t, Unchecked, m.Nodes[0].Check)
}

func TestToggle_ProfileEmptyChecksAll(t *testing.T) {
	m := testModel()
	// First, uncheck everything manually.
	m.Nodes[1].Check = Unchecked
	m.Nodes[2].Check = Unchecked
	m.Nodes[0].Check = profileAggregate(m, m.Nodes[0])
	m.Cursor = 0
	m.Toggle()
	assert.Equal(t, Checked, m.Nodes[1].Check)
	assert.Equal(t, Checked, m.Nodes[2].Check)
	assert.Equal(t, Checked, m.Nodes[0].Check)
}

func TestToggle_ProfilePartialBecomesAllChecked(t *testing.T) {
	m := testModel()
	m.Nodes[1].Check = Unchecked
	m.Nodes[0].Check = profileAggregate(m, m.Nodes[0]) // Partial
	m.Cursor = 0
	m.Toggle()
	assert.Equal(t, Checked, m.Nodes[1].Check)
	assert.Equal(t, Checked, m.Nodes[2].Check)
	assert.Equal(t, Checked, m.Nodes[0].Check)
}

func TestToggle_DisabledRowNoop(t *testing.T) {
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"07-gpu": {When: "has_gpu", Tools: []config.Tool{
				{Name: "cuda", Apt: "cuda"},
			}},
		},
	}
	m := New(cfg, linuxAptInfo()) // no GPU → profile disabled
	m.Cursor = 0
	before := *m.Nodes[0]
	m.Toggle()
	assert.Equal(t, before, *m.Nodes[0], "toggle on disabled profile should be a no-op")
}

func TestToggle_ProfileIgnoresDisabledChildren(t *testing.T) {
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"a": {Tools: []config.Tool{
				{Name: "ok", Apt: "ok"},
				{Name: "only-brew", Brew: "x"}, // disabled on linux-apt
			}},
		},
	}
	m := New(cfg, linuxAptInfo())
	m.Cursor = 0
	m.Toggle() // currently: ok=Checked → Unchecked
	assert.Equal(t, Unchecked, m.Nodes[1].Check) // ok
	assert.Equal(t, Unchecked, m.Nodes[2].Check) // only-brew stays unchecked and disabled
	assert.True(t, m.Nodes[2].Disabled)
	assert.Equal(t, Unchecked, m.Nodes[0].Check)
}

func TestExpand_OnProfile(t *testing.T) {
	m := testModel()
	m.Cursor = 0
	m.Expand()
	assert.True(t, m.Nodes[0].Expanded)
}

func TestCollapse_OnProfile(t *testing.T) {
	m := testModel()
	m.Nodes[0].Expanded = true
	m.Cursor = 0
	m.Collapse()
	assert.False(t, m.Nodes[0].Expanded)
}

func TestExpand_OnTool_NoOp(t *testing.T) {
	m := testModel()
	m.Nodes[0].Expanded = true
	m.Cursor = 1 // t1
	m.Expand()
	assert.True(t, m.Nodes[0].Expanded) // profile unchanged
	assert.Equal(t, 1, m.Cursor)
}

func TestCollapse_OnTool_JumpsToParent(t *testing.T) {
	m := testModel()
	m.Nodes[0].Expanded = true
	m.Cursor = 2 // t2
	m.Collapse()
	assert.Equal(t, 0, m.Cursor)               // cursor jumped to a-profile
	assert.True(t, m.Nodes[0].Expanded)        // profile still expanded
}

func TestSelection_ReturnsOnlyCheckedEnabledTools(t *testing.T) {
	m := testModel()
	// Default: t1 and t3 checked (different profiles). Uncheck t1 via toggle.
	m.Nodes[1].Check = Unchecked
	m.Nodes[0].Check = profileAggregate(m, m.Nodes[0])

	sel := m.Selection()
	assert.Equal(t, []string{"t2", "t3"}, sel)
}

func TestSelection_SkipsDisabled(t *testing.T) {
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"a": {Tools: []config.Tool{
				{Name: "ok", Apt: "ok"},
				{Name: "only-brew", Brew: "x"},
			}},
		},
	}
	m := New(cfg, linuxAptInfo())
	sel := m.Selection()
	assert.Equal(t, []string{"ok"}, sel)
}

func TestAllToolConfigs_IncludesEveryTool(t *testing.T) {
	m := testModel()
	all := m.AllToolConfigs()
	names := make([]string, len(all))
	for i, t := range all {
		names[i] = t.Name
	}
	assert.ElementsMatch(t, []string{"t1", "t2", "t3"}, names)
}

func TestHasSelection_TrueWhenAnyChecked(t *testing.T) {
	m := testModel()
	assert.True(t, m.HasSelection())
	// Uncheck all.
	for _, n := range m.Nodes {
		if n.Kind == KindTool {
			n.Check = Unchecked
		}
	}
	assert.False(t, m.HasSelection())
}
