package picker

import (
	"sort"

	"github.com/host452b/isetup/internal/config"
	"github.com/host452b/isetup/internal/detector"
	"github.com/host452b/isetup/internal/executor"
)

// Kind distinguishes the two row types in the picker.
type Kind int

const (
	KindProfile Kind = iota
	KindTool
)

// CheckState is the tri-state check value.
type CheckState int

const (
	Unchecked CheckState = iota
	Checked
	Partial // only valid for profile rows
)

// Phase is the current top-level UI mode of the picker.
type Phase int

const (
	PhasePick Phase = iota
	PhaseConfirm
)

// Node is one row in the tree: a profile header or a tool under a profile.
type Node struct {
	Kind      Kind
	Name      string       // profile name or tool name
	Method    string       // tool rows: the resolved install method, or "" if none on current system
	Disabled  bool         // profile with when=false, or tool with no install method
	ChildIdxs []int        // profile rows: indices into Model.Nodes
	ParentIdx int          // tool rows: index of parent profile; profiles: -1
	Expanded  bool         // profile rows only
	Check     CheckState
	Tool      *config.Tool // backref for downstream dep computation
}

// Model is the full state of the picker UI.
type Model struct {
	Nodes     []*Node
	Cursor    int   // absolute index into Nodes; must always be a visible row
	HelpOpen  bool
	Phase     Phase
	StatusMsg string // transient one-line message shown in status area
	Cfg       *config.Config
	Info      *detector.SystemInfo
}

// New builds the initial Model with default check state per design decision E:
// enabled profiles are checked and collapsed; disabled profiles ("when"
// unmet) are unchecked and disabled; inside each profile, tools with an
// install method on the current system are checked, tools without are
// disabled and unchecked.
func New(cfg *config.Config, info *detector.SystemInfo) *Model {
	m := &Model{Cfg: cfg, Info: info}

	profileNames := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		profileNames = append(profileNames, name)
	}
	sort.Strings(profileNames)

	for _, pname := range profileNames {
		prof := cfg.Profiles[pname]
		profDisabled := prof.When != "" && !evaluateWhen(prof.When, info)
		profIdx := len(m.Nodes)
		pnode := &Node{
			Kind:      KindProfile,
			Name:      pname,
			ParentIdx: -1,
			Disabled:  profDisabled,
		}
		m.Nodes = append(m.Nodes, pnode)

		for i := range prof.Tools {
			tool := &prof.Tools[i]
			method, _ := executor.Resolve(*tool, info)
			disabled := profDisabled || method == ""
			check := Unchecked
			if !disabled {
				check = Checked
			}
			tnode := &Node{
				Kind:      KindTool,
				Name:      tool.Name,
				Method:    method,
				ParentIdx: profIdx,
				Disabled:  disabled,
				Check:     check,
				Tool:      tool,
			}
			pnode.ChildIdxs = append(pnode.ChildIdxs, len(m.Nodes))
			m.Nodes = append(m.Nodes, tnode)
		}

		pnode.Check = profileAggregate(m, pnode)
	}
	return m
}

// evaluateWhen mirrors executor.evaluateCondition (kept inline to avoid
// cross-package coupling for a 5-line switch). Only "has_gpu" is supported.
func evaluateWhen(when string, info *detector.SystemInfo) bool {
	switch when {
	case "has_gpu":
		return info.GPU.Detected
	default:
		return false
	}
}

// profileAggregate computes the CheckState of a profile from its selectable
// children. Disabled children are ignored.
func profileAggregate(m *Model, p *Node) CheckState {
	checked, unchecked := 0, 0
	for _, ci := range p.ChildIdxs {
		c := m.Nodes[ci]
		if c.Disabled {
			continue
		}
		if c.Check == Checked {
			checked++
		} else {
			unchecked++
		}
	}
	switch {
	case checked > 0 && unchecked == 0:
		return Checked
	case checked > 0 && unchecked > 0:
		return Partial
	default:
		return Unchecked
	}
}

// visibleIndices returns the list of node indices currently visible in the
// picker (profile rows plus tool rows inside expanded profiles), in display
// order.
func (m *Model) visibleIndices() []int {
	var out []int
	for i, n := range m.Nodes {
		if n.Kind == KindProfile {
			out = append(out, i)
			continue
		}
		if m.Nodes[n.ParentIdx].Expanded {
			out = append(out, i)
		}
	}
	return out
}

// MoveDown advances the cursor to the next visible row, or stays if already
// at the bottom. Has no effect when there are no rows.
func (m *Model) MoveDown() {
	m.StatusMsg = ""
	vis := m.visibleIndices()
	pos := indexOf(vis, m.Cursor)
	if pos >= 0 && pos < len(vis)-1 {
		m.Cursor = vis[pos+1]
	}
}

// MoveUp retreats the cursor to the previous visible row, or stays at top.
func (m *Model) MoveUp() {
	m.StatusMsg = ""
	vis := m.visibleIndices()
	pos := indexOf(vis, m.Cursor)
	if pos > 0 {
		m.Cursor = vis[pos-1]
	}
}

func indexOf(s []int, v int) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}
	return -1
}
