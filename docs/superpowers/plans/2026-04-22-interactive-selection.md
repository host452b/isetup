# Interactive Tool Selection Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship `isetup install -i / --interactive`, a keyboard-driven two-level (profile + tool) multi-select TUI that feeds the existing executor.

**Architecture:** New `internal/picker/` package split into pure-function layers (`deps.go`, `input.go`, `model.go`, `render.go`) plus a thin I/O glue (`picker.go`) built on `golang.org/x/term`. `cmd/install.go` gets a new flag and an auto-enter path when stdin is a TTY and no other install flags are present. `internal/executor.Execute` gets an additional `toolFilter []string` parameter so the picker can restrict the install to user-selected tools without mutating the config.

**Tech Stack:** Go 1.22, cobra, testify, `golang.org/x/term` (new dep; Go-team maintained).

**Spec:** `docs/superpowers/specs/2026-04-22-interactive-selection-design.md`

---

## File Structure

**New files (all in `internal/picker/`):**
| File | Responsibility |
|------|----------------|
| `deps.go` | `ResolveDeps(selected, all)` — BFS closure over `depends_on`. Pure. |
| `deps_test.go` | Tests for `ResolveDeps`. |
| `input.go` | `ParseKey([]byte) (Event, int)` — byte prefix → key event. Pure. |
| `input_test.go` | Table-driven tests for `ParseKey`. |
| `model.go` | `Model`, `Node`, `Kind`, `CheckState`, `Phase`, `Selection`; `New`, `MoveUp`, `MoveDown`, `Toggle`, `Expand`, `Collapse`, `Selection()`, helpers. Pure. |
| `model_test.go` | Tests for state machine. |
| `render.go` | `Render(m, width, height) string`, confirm page, help overlay. Pure. |
| `render_test.go` | Snapshot-style tests for render output. |
| `picker.go` | `Run(cfg, info) (*Selection, error)` — terminal raw mode, read loop, signal handling. Only I/O-bearing file. |

**Modified:**
- `cmd/install.go` — add `-i`/`--interactive` flag, auto-enter logic, call `picker.Run`, pass selection to executor.
- `cmd/cmd_test.go` — test for `decideInteractive` helper.
- `internal/executor/executor.go` — add `toolFilter []string` to `Execute` and `collectTools`.
- `internal/executor/executor_test.go` — update existing call sites, add `TestExecute_ToolFilter`.
- `go.mod`, `go.sum` — `golang.org/x/term` added.
- `README.md`, `README_zh.md`, `CHANGELOG.md`, `CLAUDE.md` — documentation.

---

## Task 1: Add `golang.org/x/term` dependency

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Add the dependency**

```bash
go get golang.org/x/term
go mod tidy
```

- [ ] **Step 2: Verify build still works**

```bash
go build ./...
```
Expected: no output, exit 0.

- [ ] **Step 3: Verify all existing tests still pass**

```bash
go test ./...
```
Expected: `ok` for every package.

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum
git commit -m "feat: add golang.org/x/term dependency for interactive picker"
```

---

## Task 2: Extend `executor.Execute` with `toolFilter`

**Files:**
- Modify: `internal/executor/executor.go` (signature of `Execute` and `collectTools`)
- Modify: `internal/executor/executor_test.go` (update call sites, add test)
- Modify: `cmd/install.go:205` (update call site)

- [ ] **Step 1: Write a failing test for tool filtering**

In `internal/executor/executor_test.go`:

1. Add `"sort"` to the existing `import ( ... )` block.
2. Append the new test function at the bottom:

```go
func TestExecute_ToolFilter(t *testing.T) {
	cfg := &config.Config{
		Version:  1,
		Settings: config.Settings{DryRun: true, Force: true},
		Profiles: map[string]config.Profile{
			"base": {Tools: []config.Tool{
				{Name: "git", Apt: "git"},
				{Name: "curl", Apt: "curl"},
				{Name: "wget", Apt: "wget"},
			}},
		},
	}
	info := testSystemInfo()
	lg, err := logger.New(t.TempDir())
	require.NoError(t, err)

	results, err := Execute(context.Background(), cfg, info, lg, nil, []string{"git", "wget"}, nil)
	require.NoError(t, err)
	require.Len(t, results, 2)
	names := []string{results[0].Name, results[1].Name}
	sort.Strings(names)
	assert.Equal(t, []string{"git", "wget"}, names)
}
```

- [ ] **Step 2: Update existing executor test call sites to match the new signature**

In `internal/executor/executor_test.go`, replace every `Execute(ctx, cfg, info, lg, <profiles>, nil)` with `Execute(ctx, cfg, info, lg, <profiles>, nil, nil)` (insert `nil` for the new 6th arg). Specifically lines 36, 57, 75, 96, 116.

- [ ] **Step 3: Run tests and confirm compile failure**

Run:
```bash
go test ./internal/executor/...
```
Expected: compile error `too many arguments in call to Execute` or similar — because `Execute` does not yet accept the new parameter.

- [ ] **Step 4: Update `Execute` and `collectTools` signatures**

In `internal/executor/executor.go`, change the signature of `Execute`:

```go
func Execute(ctx context.Context, cfg *config.Config, info *detector.SystemInfo, lg *logger.Logger, profiles []string, toolFilter []string, onProgress ProgressCallback) ([]logger.ToolResult, error) {
	if !cfg.Settings.DryRun {
		Bootstrap(ctx, info, lg)
	}

	entries := collectTools(cfg, info, profiles, toolFilter)
	// ... rest unchanged ...
```

Change the signature and body of `collectTools`:

```go
func collectTools(cfg *config.Config, info *detector.SystemInfo, profileFilter, toolFilter []string) []ToolEntry {
	selected := cfg.Profiles
	if profileFilter != nil {
		selected = make(map[string]config.Profile)
		for _, name := range profileFilter {
			if p, ok := cfg.Profiles[name]; ok {
				selected[name] = p
			}
		}
	}

	toolSet := map[string]bool(nil)
	if toolFilter != nil {
		toolSet = make(map[string]bool, len(toolFilter))
		for _, n := range toolFilter {
			toolSet[n] = true
		}
	}

	names := make([]string, 0, len(selected))
	for name := range selected {
		names = append(names, name)
	}
	sort.Strings(names)

	var entries []ToolEntry
	for _, profName := range names {
		prof := selected[profName]
		skipReason := ""
		if prof.When != "" && !evaluateCondition(prof.When, info) {
			skipReason = fmt.Sprintf("condition not met: %s", prof.When)
		}

		for _, tool := range prof.Tools {
			if toolSet != nil && !toolSet[tool.Name] {
				continue
			}
			entries = append(entries, ToolEntry{
				Tool:       tool,
				Profile:    profName,
				SkipReason: skipReason,
			})
		}
	}
	return entries
}
```

- [ ] **Step 5: Update the existing call site in `cmd/install.go:205`**

Change:
```go
results, err := executor.Execute(ctx, cfg, info, lg, profiles, onProgress)
```
To:
```go
results, err := executor.Execute(ctx, cfg, info, lg, profiles, nil, onProgress)
```

- [ ] **Step 6: Run all tests**

```bash
go test ./...
```
Expected: all pass including `TestExecute_ToolFilter`.

- [ ] **Step 7: Commit**

```bash
git add internal/executor/executor.go internal/executor/executor_test.go cmd/install.go
git commit -m "feat(executor): add toolFilter parameter to Execute for picker integration"
```

---

## Task 3: Dependency resolver (`picker/deps.go`)

**Files:**
- Create: `internal/picker/deps.go`
- Create: `internal/picker/deps_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/picker/deps_test.go`:

```go
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
	// Two tools sharing a common ancestor.
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
	// Unknown dep is NOT pulled into the closure; executor handles it at runtime.
	assert.Equal(t, []string{"needs-missing"}, closure)
	assert.Empty(t, added)
}

func TestResolveDeps_Cycle(t *testing.T) {
	// A→B, B→A. ResolveDeps must terminate.
	cycleTools := []config.Tool{
		{Name: "A", DependsOn: "B"},
		{Name: "B", DependsOn: "A"},
	}
	closure, _ := ResolveDeps([]string{"A"}, cycleTools)
	assert.ElementsMatch(t, []string{"A", "B"}, closure)
}
```

- [ ] **Step 2: Run the tests to confirm they fail**

```bash
go test ./internal/picker/... -run TestResolveDeps -v
```
Expected: `package picker is not in std` or `undefined: ResolveDeps` (depending on whether the package compiles).

- [ ] **Step 3: Write the implementation**

Create `internal/picker/deps.go`:

```go
package picker

import (
	"sort"

	"github.com/host452b/isetup/internal/config"
)

// ResolveDeps computes the transitive closure of `selected` under the
// `DependsOn` relation defined by `all`. Dependencies that name tools not
// present in `all` are dropped from the closure (the executor's runtime path
// handles "already on system" detection for those).
//
// Returns:
//   closure — all tools that must be considered, sorted by name.
//   added   — tools added by dependency resolution (closure − selected), sorted.
func ResolveDeps(selected []string, all []config.Tool) (closure, added []string) {
	byName := make(map[string]config.Tool, len(all))
	for _, t := range all {
		byName[t.Name] = t
	}

	inClosure := make(map[string]bool)
	queue := append([]string(nil), selected...)
	for len(queue) > 0 {
		t := queue[0]
		queue = queue[1:]
		if inClosure[t] {
			continue
		}
		inClosure[t] = true
		if tool, ok := byName[t]; ok && tool.DependsOn != "" {
			queue = append(queue, tool.DependsOn)
		}
	}

	selectedSet := make(map[string]bool, len(selected))
	for _, s := range selected {
		selectedSet[s] = true
	}

	for name := range inClosure {
		closure = append(closure, name)
		if !selectedSet[name] {
			added = append(added, name)
		}
	}
	sort.Strings(closure)
	sort.Strings(added)
	return
}
```

- [ ] **Step 4: Run tests and verify all pass**

```bash
go test ./internal/picker/... -run TestResolveDeps -v
```
Expected: all seven tests `PASS`.

- [ ] **Step 5: Commit**

```bash
git add internal/picker/deps.go internal/picker/deps_test.go
git commit -m "feat(picker): ResolveDeps for dependency closure computation"
```

---

## Task 4: Key parsing (`picker/input.go`)

**Files:**
- Create: `internal/picker/input.go`
- Create: `internal/picker/input_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/picker/input_test.go`:

```go
package picker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseKey_Empty(t *testing.T) {
	ev, n := ParseKey([]byte{})
	assert.Equal(t, EventNone, ev)
	assert.Equal(t, 0, n)
}

func TestParseKey_Arrows(t *testing.T) {
	cases := []struct {
		name  string
		input []byte
		want  Event
	}{
		{"up", []byte{0x1b, '[', 'A'}, EventUp},
		{"down", []byte{0x1b, '[', 'B'}, EventDown},
		{"right", []byte{0x1b, '[', 'C'}, EventRight},
		{"left", []byte{0x1b, '[', 'D'}, EventLeft},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ev, n := ParseKey(tc.input)
			assert.Equal(t, tc.want, ev)
			assert.Equal(t, 3, n)
		})
	}
}

func TestParseKey_VimKeys(t *testing.T) {
	for b, want := range map[byte]Event{
		'k': EventUp, 'j': EventDown, 'h': EventLeft, 'l': EventRight,
	} {
		ev, n := ParseKey([]byte{b})
		assert.Equal(t, want, ev, "byte %q", b)
		assert.Equal(t, 1, n)
	}
}

func TestParseKey_ControlKeys(t *testing.T) {
	for b, want := range map[byte]Event{
		0x0d: EventEnter, // CR
		0x0a: EventEnter, // LF
		0x20: EventSpace,
		0x03: EventCtrlC,
		'q':  EventQ,
		'y':  EventY,
		'Y':  EventY,
		'n':  EventN,
		'N':  EventN,
		'e':  EventE,
		'E':  EventE,
		'?':  EventQuestion,
	} {
		ev, n := ParseKey([]byte{b})
		assert.Equal(t, want, ev, "byte 0x%02x", b)
		assert.Equal(t, 1, n)
	}
}

func TestParseKey_IncompleteEscape(t *testing.T) {
	// Lone ESC: caller may need to read more bytes to disambiguate.
	ev, n := ParseKey([]byte{0x1b})
	assert.Equal(t, EventIncomplete, ev)
	assert.Equal(t, 0, n)
}

func TestParseKey_IncompleteCSI(t *testing.T) {
	// ESC + '[' but nothing after: incomplete.
	ev, n := ParseKey([]byte{0x1b, '['})
	assert.Equal(t, EventIncomplete, ev)
	assert.Equal(t, 0, n)
}

func TestParseKey_BareEsc(t *testing.T) {
	// ESC followed by a non-CSI byte is a bare Esc, consume 1 byte.
	ev, n := ParseKey([]byte{0x1b, 'x'})
	assert.Equal(t, EventEsc, ev)
	assert.Equal(t, 1, n)
}

func TestParseKey_UnknownCSI(t *testing.T) {
	ev, n := ParseKey([]byte{0x1b, '[', 'Z'})
	assert.Equal(t, EventNone, ev)
	assert.Equal(t, 3, n)
}

func TestParseKey_UnknownByte(t *testing.T) {
	ev, n := ParseKey([]byte{'x'})
	assert.Equal(t, EventNone, ev)
	assert.Equal(t, 1, n)
}
```

- [ ] **Step 2: Run to confirm fail**

```bash
go test ./internal/picker/... -run TestParseKey -v
```
Expected: `undefined: ParseKey`, `EventUp`, etc.

- [ ] **Step 3: Write implementation**

Create `internal/picker/input.go`:

```go
package picker

// Event represents a keyboard event parsed from raw stdin bytes.
type Event int

const (
	EventNone       Event = iota // Unrecognized or empty buffer; consume consumed bytes and keep reading.
	EventIncomplete              // Buffer is a prefix of an escape sequence; caller should read more bytes.
	EventUp
	EventDown
	EventLeft
	EventRight
	EventSpace
	EventEnter
	EventEsc
	EventCtrlC
	EventQ
	EventY
	EventN
	EventE
	EventQuestion
)

// ParseKey inspects the prefix of buf and returns the event it represents
// together with the number of bytes consumed. When the prefix is an unfinished
// escape sequence, it returns (EventIncomplete, 0) and the caller must read
// more bytes before calling again.
func ParseKey(buf []byte) (Event, int) {
	if len(buf) == 0 {
		return EventNone, 0
	}
	switch buf[0] {
	case 0x1b: // ESC
		if len(buf) == 1 {
			return EventIncomplete, 0
		}
		if buf[1] == '[' {
			if len(buf) < 3 {
				return EventIncomplete, 0
			}
			switch buf[2] {
			case 'A':
				return EventUp, 3
			case 'B':
				return EventDown, 3
			case 'C':
				return EventRight, 3
			case 'D':
				return EventLeft, 3
			}
			return EventNone, 3
		}
		return EventEsc, 1
	case 0x03:
		return EventCtrlC, 1
	case 0x0d, 0x0a:
		return EventEnter, 1
	case 0x20:
		return EventSpace, 1
	case 'k':
		return EventUp, 1
	case 'j':
		return EventDown, 1
	case 'h':
		return EventLeft, 1
	case 'l':
		return EventRight, 1
	case 'q':
		return EventQ, 1
	case 'y', 'Y':
		return EventY, 1
	case 'n', 'N':
		return EventN, 1
	case 'e', 'E':
		return EventE, 1
	case '?':
		return EventQuestion, 1
	}
	return EventNone, 1
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/picker/... -run TestParseKey -v
```
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add internal/picker/input.go internal/picker/input_test.go
git commit -m "feat(picker): ParseKey for raw stdin event recognition"
```

---

## Task 5: Model types and `New()` constructor

**Files:**
- Create: `internal/picker/model.go`
- Create: `internal/picker/model_test.go`

- [ ] **Step 1: Write failing test for initial state (Option E)**

Create `internal/picker/model_test.go`:

```go
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
```

- [ ] **Step 2: Confirm test fails**

```bash
go test ./internal/picker/... -run TestNew -v
```
Expected: undefined identifiers.

- [ ] **Step 3: Write model types and `New`**

Create `internal/picker/model.go`:

```go
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
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/picker/... -run TestNew -v
```
Expected: all three pass.

- [ ] **Step 5: Commit**

```bash
git add internal/picker/model.go internal/picker/model_test.go
git commit -m "feat(picker): Model types and New constructor with default check state"
```

---

## Task 6: Cursor navigation

**Files:**
- Modify: `internal/picker/model.go` (add methods)
- Modify: `internal/picker/model_test.go` (add tests)

- [ ] **Step 1: Write failing tests**

Append to `internal/picker/model_test.go`:

```go
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
```

- [ ] **Step 2: Confirm fail**

```bash
go test ./internal/picker/... -run "TestVisibleIndices|TestMove" -v
```

Expected: undefined methods.

- [ ] **Step 3: Add navigation methods to `model.go`**

Append to `internal/picker/model.go`:

```go
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
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/picker/... -run "TestVisibleIndices|TestMove" -v
```
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add internal/picker/model.go internal/picker/model_test.go
git commit -m "feat(picker): cursor navigation with collapsed/expanded visibility"
```

---

## Task 7: Toggle logic

**Files:**
- Modify: `internal/picker/model.go`
- Modify: `internal/picker/model_test.go`

- [ ] **Step 1: Write failing tests**

Append to `internal/picker/model_test.go`:

```go
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
```

- [ ] **Step 2: Confirm fail**

```bash
go test ./internal/picker/... -run TestToggle -v
```

- [ ] **Step 3: Implement `Toggle`**

Append to `internal/picker/model.go`:

```go
// Toggle flips the check state at the cursor. Tool rows flip between Checked
// and Unchecked. Profile rows apply the "if any unselected → select all, else
// unselect all" rule, ignoring disabled children. Disabled rows are no-ops.
func (m *Model) Toggle() {
	m.StatusMsg = ""
	n := m.Nodes[m.Cursor]
	if n.Disabled {
		return
	}
	if n.Kind == KindTool {
		if n.Check == Checked {
			n.Check = Unchecked
		} else {
			n.Check = Checked
		}
		parent := m.Nodes[n.ParentIdx]
		parent.Check = profileAggregate(m, parent)
		return
	}
	// Profile: if any selectable child is unchecked → check all; else uncheck all.
	target := Unchecked
	for _, ci := range n.ChildIdxs {
		c := m.Nodes[ci]
		if c.Disabled {
			continue
		}
		if c.Check != Checked {
			target = Checked
			break
		}
	}
	for _, ci := range n.ChildIdxs {
		c := m.Nodes[ci]
		if c.Disabled {
			continue
		}
		c.Check = target
	}
	n.Check = profileAggregate(m, n)
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/picker/... -run TestToggle -v
```
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add internal/picker/model.go internal/picker/model_test.go
git commit -m "feat(picker): Toggle with profile aggregate and disabled-safe behavior"
```

---

## Task 8: Expand / Collapse

**Files:**
- Modify: `internal/picker/model.go`
- Modify: `internal/picker/model_test.go`

- [ ] **Step 1: Write failing tests**

Append to `internal/picker/model_test.go`:

```go
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
```

- [ ] **Step 2: Confirm fail**

```bash
go test ./internal/picker/... -run "TestExpand|TestCollapse" -v
```

- [ ] **Step 3: Implement Expand/Collapse**

Append to `internal/picker/model.go`:

```go
// Expand opens the profile at the cursor. No effect on tool rows.
func (m *Model) Expand() {
	m.StatusMsg = ""
	n := m.Nodes[m.Cursor]
	if n.Kind == KindProfile {
		n.Expanded = true
	}
}

// Collapse folds the profile at the cursor. If the cursor is on a tool row,
// it jumps to the parent profile without collapsing (the user may press
// collapse again to fold).
func (m *Model) Collapse() {
	m.StatusMsg = ""
	n := m.Nodes[m.Cursor]
	if n.Kind == KindProfile {
		n.Expanded = false
		return
	}
	m.Cursor = n.ParentIdx
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/picker/... -run "TestExpand|TestCollapse" -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/picker/model.go internal/picker/model_test.go
git commit -m "feat(picker): Expand/Collapse with tool-row parent jump"
```

---

## Task 9: Selection output helpers

**Files:**
- Modify: `internal/picker/model.go`
- Modify: `internal/picker/model_test.go`

- [ ] **Step 1: Write failing tests**

Append to `internal/picker/model_test.go`:

```go
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
```

- [ ] **Step 2: Confirm fail**

```bash
go test ./internal/picker/... -run "TestSelection|TestAllTool|TestHasSelection" -v
```

- [ ] **Step 3: Implement helpers and Selection type**

Append to `internal/picker/model.go`:

```go
// Selection returns the names of currently-checked, enabled tool rows, in
// display order (grouped by profile, profile-sorted alphabetically).
func (m *Model) Selection() []string {
	var out []string
	for _, n := range m.Nodes {
		if n.Kind == KindTool && n.Check == Checked && !n.Disabled {
			out = append(out, n.Name)
		}
	}
	return out
}

// AllToolConfigs returns every tool config in the model in a flat slice, for
// passing to ResolveDeps.
func (m *Model) AllToolConfigs() []config.Tool {
	var out []config.Tool
	for _, n := range m.Nodes {
		if n.Kind == KindTool && n.Tool != nil {
			out = append(out, *n.Tool)
		}
	}
	return out
}

// HasSelection reports whether at least one selectable tool is checked.
func (m *Model) HasSelection() bool {
	for _, n := range m.Nodes {
		if n.Kind == KindTool && n.Check == Checked && !n.Disabled {
			return true
		}
	}
	return false
}

// SelectionResult is returned by Run when the user confirms.
type SelectionResult struct {
	// Tools is the full dependency closure of the user's selection, sorted.
	Tools []string
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/picker/... -v
```
Expected: all model tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/picker/model.go internal/picker/model_test.go
git commit -m "feat(picker): Selection/AllToolConfigs/HasSelection helpers + SelectionResult"
```

---

## Task 10: Main picker rendering

**Files:**
- Create: `internal/picker/render.go`
- Create: `internal/picker/render_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/picker/render_test.go`:

```go
package picker

import (
	"os"
	"strings"
	"testing"

	"github.com/host452b/isetup/internal/config"
	"github.com/stretchr/testify/assert"
)

func renderableTestModel() *Model {
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
	return New(cfg, linuxAptInfo())
}

func TestRender_ContainsProfileNames(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	out := Render(m, 80, 24)
	assert.Contains(t, out, "00-base")
	assert.Contains(t, out, "07-gpu")
}

func TestRender_ExpandedShowsTools(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	// Expand 00-base.
	for _, n := range m.Nodes {
		if n.Kind == KindProfile && n.Name == "00-base" {
			n.Expanded = true
		}
	}
	out := Render(m, 80, 24)
	assert.Contains(t, out, "git")
	assert.Contains(t, out, "only-brew")
}

func TestRender_CollapsedHidesTools(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	out := Render(m, 80, 24)
	assert.NotContains(t, out, "  git") // tool indent wouldn't appear when collapsed
}

func TestRender_CheckboxesForStates(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	out := Render(m, 80, 24)
	assert.Contains(t, out, "[x]", "checked profile")
	assert.Contains(t, out, "[·]", "disabled profile")
}

func TestRender_DisabledProfileShowsMarker(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	out := Render(m, 80, 24)
	assert.Contains(t, out, "no GPU detected", "disabled profile displays the reason")
}

func TestRender_NoColorByDefault(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	out := Render(m, 80, 24)
	assert.NotContains(t, out, "\x1b[", "NO_COLOR=1 strips ANSI")
}

func TestRender_ColorWhenEnabled(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	m := renderableTestModel()
	out := Render(m, 80, 24)
	assert.Contains(t, out, "\x1b[", "ANSI codes present when color is enabled")
}

func TestRender_StatusBarIncludesKeys(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	out := Render(m, 80, 24)
	assert.Contains(t, out, "Space")
	assert.Contains(t, out, "Enter")
	assert.Contains(t, out, "q")
}

func TestRender_NarrowWidthDropsMethodColumn(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	for _, n := range m.Nodes {
		if n.Kind == KindProfile && n.Name == "00-base" {
			n.Expanded = true
		}
	}
	wide := Render(m, 80, 24)
	narrow := Render(m, 40, 24)
	// Wide shows "apt" next to git; narrow drops it.
	assert.Contains(t, wide, "apt")
	assert.NotContains(t, narrow, "apt")
	// Both still show the tool name.
	assert.Contains(t, narrow, "git")
}

func TestRender_StatusMessageAppears(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	m.StatusMsg = "Nothing selected — press Space to select tools"
	out := Render(m, 80, 24)
	assert.Contains(t, out, "Nothing selected")
}

func TestRender_HelpOverlay(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := renderableTestModel()
	m.HelpOpen = true
	out := Render(m, 80, 24)
	// Help overlay should include more detailed key documentation.
	assert.Contains(t, strings.ToLower(out), "help")
	assert.Contains(t, out, "↑/↓")
	assert.Contains(t, out, "Space")
}
```

- [ ] **Step 2: Confirm fail**

```bash
go test ./internal/picker/... -run TestRender -v
```

- [ ] **Step 3: Implement `Render` for picker phase**

Create `internal/picker/render.go`:

```go
package picker

import (
	"fmt"
	"os"
	"strings"
)

const (
	ansiReset   = "\x1b[0m"
	ansiReverse = "\x1b[7m"
	ansiBold    = "\x1b[1m"
	ansiDim     = "\x1b[2m"
)

func useColor() bool {
	return os.Getenv("NO_COLOR") == ""
}

func col(code, s string) string {
	if !useColor() {
		return s
	}
	return code + s + ansiReset
}

// Render produces the full screen text for the model. Caller is responsible
// for clearing the terminal before writing.
func Render(m *Model, width, height int) string {
	if m.Phase == PhaseConfirm {
		return renderConfirm(m, width, height)
	}
	return renderPicker(m, width, height)
}

func renderPicker(m *Model, width, height int) string {
	var b strings.Builder
	b.WriteString(renderHeader(m, width))
	b.WriteString("\n\n")

	vis := m.visibleIndices()
	cursorPos := indexOf(vis, m.Cursor)
	narrow := width < 50
	for i, idx := range vis {
		b.WriteString(renderRow(m, idx, i == cursorPos, narrow, width))
		b.WriteString("\n")
	}

	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n")
	if m.StatusMsg != "" {
		b.WriteString(col("\x1b[31m", m.StatusMsg))
		b.WriteString("\n")
	}
	b.WriteString(renderStatusBar(width))
	if m.HelpOpen {
		b.WriteString("\n\n")
		b.WriteString(renderHelpOverlay(width))
	}
	return b.String()
}

func renderHeader(m *Model, width int) string {
	left := col(ansiBold, "isetup install · interactive mode")
	right := ""
	if m.Info != nil {
		right = col(ansiDim, fmt.Sprintf("%s/%s · %s", m.Info.OS, m.Info.Arch, strings.Join(m.Info.PkgManagers, ",")))
	}
	pad := width - visualLen(left) - visualLen(right)
	if pad < 1 {
		pad = 1
	}
	return left + strings.Repeat(" ", pad) + right
}

func renderRow(m *Model, idx int, cursor bool, narrow bool, width int) string {
	n := m.Nodes[idx]
	var line string
	if n.Kind == KindProfile {
		line = renderProfileRow(m, n, narrow)
	} else {
		line = renderToolRow(n, narrow)
	}
	if cursor {
		return col(ansiReverse, line)
	}
	return line
}

func renderProfileRow(m *Model, n *Node, narrow bool) string {
	box := checkbox(n)
	arrow := "▶"
	if n.Expanded {
		arrow = "▼"
	}
	if n.Disabled {
		arrow = "✗"
	}

	name := col(ansiBold, n.Name)
	if n.Disabled {
		name = col(ansiDim, n.Name)
	}
	prefix := fmt.Sprintf("%s %s %s", box, arrow, name)

	suffix := ""
	if n.Disabled {
		suffix = col(ansiDim, "  no GPU detected") // only condition supported today
	} else {
		selected, total := profileCounts(m, n)
		suffix = col(ansiDim, fmt.Sprintf("  %d/%d selected", selected, total))
	}
	return prefix + suffix
}

func renderToolRow(n *Node, narrow bool) string {
	box := checkbox(n)
	name := n.Name
	if n.Disabled {
		name = col(ansiDim, name)
	}
	row := fmt.Sprintf("      %s %s", box, name)
	if !narrow && n.Method != "" {
		row += col(ansiDim, "  "+n.Method)
	}
	if n.Disabled && n.Method == "" {
		row += col(ansiDim, "  ⚠ no method")
	}
	return row
}

func checkbox(n *Node) string {
	if n.Disabled {
		return "[·]"
	}
	switch n.Check {
	case Checked:
		return "[x]"
	case Partial:
		return "[-]"
	default:
		return "[ ]"
	}
}

func profileCounts(m *Model, p *Node) (selected, total int) {
	for _, ci := range p.ChildIdxs {
		c := m.Nodes[ci]
		if c.Disabled {
			continue
		}
		total++
		if c.Check == Checked {
			selected++
		}
	}
	return
}

func renderStatusBar(width int) string {
	bar := "↑/↓ move · Space toggle · →/← expand/collapse · Enter confirm · q quit · ? help"
	return col(ansiDim, bar)
}

func renderHelpOverlay(width int) string {
	lines := []string{
		"Help",
		"  ↑/↓ or k/j        Move cursor",
		"  →/l               Expand profile",
		"  ←/h               Collapse profile (from tool: jump to parent)",
		"  Space             Toggle selection (profile: all children)",
		"  Enter             Confirm selection and proceed",
		"  q / Esc / Ctrl+C  Quit without installing",
		"  ?                 Toggle this help",
	}
	return col(ansiDim, strings.Join(lines, "\n"))
}

// visualLen ignores ANSI escape sequences when measuring width.
func visualLen(s string) int {
	n := 0
	inEscape := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inEscape {
			if c == 'm' {
				inEscape = false
			}
			continue
		}
		if c == 0x1b {
			inEscape = true
			continue
		}
		n++
	}
	return n
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/picker/... -run TestRender -v
```
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add internal/picker/render.go internal/picker/render_test.go
git commit -m "feat(picker): main picker rendering with checkbox, expand, narrow adaptation"
```

---

## Task 11: Confirm page rendering

**Files:**
- Modify: `internal/picker/render.go`
- Modify: `internal/picker/render_test.go`

- [ ] **Step 1: Write failing tests**

Append to `internal/picker/render_test.go`:

```go
func TestRenderConfirm_ListsSelectedAndDeps(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"00-base": {Tools: []config.Tool{
				{Name: "curl", Apt: "curl"},
			}},
			"04-ai": {Tools: []config.Tool{
				{Name: "claude-code", DependsOn: "curl", Apt: "claude-code"},
			}},
		},
	}
	m := New(cfg, linuxAptInfo())
	// Uncheck curl so we can verify it comes back via dep resolution.
	for _, n := range m.Nodes {
		if n.Name == "curl" {
			n.Check = Unchecked
		}
	}
	for _, n := range m.Nodes {
		if n.Kind == KindProfile {
			n.Check = profileAggregate(m, n)
		}
	}
	m.Phase = PhaseConfirm
	out := Render(m, 80, 24)

	assert.Contains(t, out, "claude-code")
	assert.Contains(t, out, "curl")
	assert.Contains(t, out, "Required dependencies")
	assert.Contains(t, out, "[Y/Enter]")
	assert.Contains(t, out, "[E]")
	assert.Contains(t, out, "[N/Esc]")
}

func TestRenderConfirm_NoDepsSection(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"00-base": {Tools: []config.Tool{
				{Name: "git", Apt: "git"},
			}},
		},
	}
	m := New(cfg, linuxAptInfo())
	m.Phase = PhaseConfirm
	out := Render(m, 80, 24)

	assert.Contains(t, out, "git")
	assert.NotContains(t, out, "Required dependencies", "no deps to add → omit section")
}
```

- [ ] **Step 2: Confirm fail**

```bash
go test ./internal/picker/... -run TestRenderConfirm -v
```

- [ ] **Step 3: Implement `renderConfirm`**

Append to `internal/picker/render.go`:

```go
// renderConfirm produces the confirmation page. It calls ResolveDeps to
// compute the closure and renders the delta between user selection and
// auto-added dependencies.
func renderConfirm(m *Model, width, height int) string {
	selected := m.Selection()
	all := m.AllToolConfigs()
	closure, added := ResolveDeps(selected, all)

	// Build method lookup for enriching each line.
	methodByName := make(map[string]string, len(m.Nodes))
	for _, n := range m.Nodes {
		if n.Kind == KindTool {
			methodByName[n.Name] = n.Method
		}
	}

	var b strings.Builder
	b.WriteString(col(ansiBold, "Review & Install"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n\n")

	addedSet := make(map[string]bool, len(added))
	for _, a := range added {
		addedSet[a] = true
	}

	fmt.Fprintf(&b, "You selected %d tool(s):\n", len(selected))
	for _, name := range selected {
		fmt.Fprintf(&b, "    %-24s %s\n", name, col(ansiDim, methodByName[name]))
	}

	if len(added) > 0 {
		fmt.Fprintf(&b, "\nRequired dependencies (auto-added): %d tool(s)\n", len(added))
		for _, name := range added {
			fmt.Fprintf(&b, "    %-24s %s\n", name, col(ansiDim, methodByName[name]))
		}
	}

	fmt.Fprintf(&b, "\nTotal: %d tool(s) will be attempted\n", len(closure))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n")
	b.WriteString(col(ansiBold, "[Y/Enter] Install"))
	b.WriteString("   ")
	b.WriteString("[E] Edit selection")
	b.WriteString("   ")
	b.WriteString("[N/Esc] Cancel")
	return b.String()
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/picker/... -run TestRenderConfirm -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/picker/render.go internal/picker/render_test.go
git commit -m "feat(picker): confirm page rendering with dependency delta"
```

---

## Task 12: `picker.Run` — terminal glue

**Files:**
- Create: `internal/picker/picker.go`

This is the only file with real I/O. Logic below the Run-level is already tested; this file just wires signals, raw mode, and the read loop.

- [ ] **Step 1: Write `handleEvent` tests (pure function, testable without a TTY)**

Append to `internal/picker/model_test.go`:

```go
func TestHandleEvent_EnterWithSelectionMovesToConfirm(t *testing.T) {
	m := testModel()
	done, sel := handleEvent(m, EventEnter)
	assert.False(t, done)
	assert.Nil(t, sel)
	assert.Equal(t, PhaseConfirm, m.Phase)
}

func TestHandleEvent_EnterWithoutSelectionSetsStatus(t *testing.T) {
	m := testModel()
	for _, n := range m.Nodes {
		if n.Kind == KindTool {
			n.Check = Unchecked
		}
	}
	done, _ := handleEvent(m, EventEnter)
	assert.False(t, done)
	assert.Equal(t, PhasePick, m.Phase)
	assert.Contains(t, m.StatusMsg, "Nothing selected")
}

func TestHandleEvent_ConfirmYReturnsSelection(t *testing.T) {
	m := testModel()
	m.Phase = PhaseConfirm
	done, sel := handleEvent(m, EventY)
	assert.True(t, done)
	require.NotNil(t, sel)
	assert.ElementsMatch(t, []string{"t1", "t2", "t3"}, sel.Tools)
}

func TestHandleEvent_ConfirmEditGoesBack(t *testing.T) {
	m := testModel()
	m.Phase = PhaseConfirm
	done, sel := handleEvent(m, EventE)
	assert.False(t, done)
	assert.Nil(t, sel)
	assert.Equal(t, PhasePick, m.Phase)
}

func TestHandleEvent_ConfirmCancelExits(t *testing.T) {
	m := testModel()
	m.Phase = PhaseConfirm
	done, sel := handleEvent(m, EventN)
	assert.True(t, done)
	assert.Nil(t, sel)
}

func TestHandleEvent_EscInPickExits(t *testing.T) {
	m := testModel()
	done, sel := handleEvent(m, EventEsc)
	assert.True(t, done)
	assert.Nil(t, sel)
}

func TestHandleEvent_QuestionTogglesHelp(t *testing.T) {
	m := testModel()
	assert.False(t, m.HelpOpen)
	handleEvent(m, EventQuestion)
	assert.True(t, m.HelpOpen)
	handleEvent(m, EventQuestion)
	assert.False(t, m.HelpOpen)
}
```

(Uses `require` — add its import to the test file if not already present; the existing test file already imports both `assert` and `require`.)

- [ ] **Step 2: Confirm fail**

```bash
go test ./internal/picker/... -run TestHandleEvent -v
```

- [ ] **Step 3: Implement `handleEvent` and `Run`**

Create `internal/picker/picker.go`:

```go
package picker

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/host452b/isetup/internal/config"
	"github.com/host452b/isetup/internal/detector"
	"golang.org/x/term"
)

// Run presents the interactive picker on the current terminal and returns the
// user's final selection. Returns (nil, nil) if the user cancelled (Esc/q/N).
// Returns (nil, error) on terminal setup failures.
func Run(cfg *config.Config, info *detector.SystemInfo) (*SelectionResult, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return nil, fmt.Errorf("stdin is not a terminal")
	}
	width, height, err := term.GetSize(fd)
	if err != nil {
		return nil, fmt.Errorf("get terminal size: %w", err)
	}
	if width < 30 {
		return nil, fmt.Errorf("terminal too narrow (need at least 30 cols, have %d)", width)
	}
	if height < 10 {
		return nil, fmt.Errorf("terminal too small (need at least 10 rows, have %d)", height)
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, fmt.Errorf("enter raw mode: %w", err)
	}
	defer term.Restore(fd, oldState)

	resizeCh := make(chan os.Signal, 1)
	signal.Notify(resizeCh, syscall.SIGWINCH)
	defer signal.Stop(resizeCh)

	intCh := make(chan os.Signal, 1)
	signal.Notify(intCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(intCh)

	bytesCh := make(chan []byte, 8)
	errCh := make(chan error, 1)
	stopReader := make(chan struct{})
	defer close(stopReader)

	go func() {
		buf := make([]byte, 16)
		for {
			select {
			case <-stopReader:
				return
			default:
			}
			n, err := os.Stdin.Read(buf)
			if err != nil {
				errCh <- err
				return
			}
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			select {
			case bytesCh <- chunk:
			case <-stopReader:
				return
			}
		}
	}()

	m := New(cfg, info)
	pending := make([]byte, 0, 16)
	drawScreen(m, width, height)

	for {
		select {
		case <-intCh:
			return nil, nil
		case <-resizeCh:
			if w, h, err := term.GetSize(fd); err == nil {
				width, height = w, h
			}
			drawScreen(m, width, height)
		case err := <-errCh:
			return nil, fmt.Errorf("read stdin: %w", err)
		case chunk := <-bytesCh:
			pending = append(pending, chunk...)
			for {
				ev, consumed := ParseKey(pending)
				if ev == EventIncomplete {
					break
				}
				if consumed == 0 {
					pending = pending[:0]
					break
				}
				pending = pending[consumed:]
				if ev == EventNone {
					continue
				}
				done, sel := handleEvent(m, ev)
				if done {
					return sel, nil
				}
				drawScreen(m, width, height)
			}
		}
	}
}

// handleEvent applies the event to the model and reports whether the run
// should terminate and, if so, with what selection.
func handleEvent(m *Model, ev Event) (bool, *SelectionResult) {
	if m.Phase == PhaseConfirm {
		switch ev {
		case EventY, EventEnter:
			selected := m.Selection()
			closure, _ := ResolveDeps(selected, m.AllToolConfigs())
			return true, &SelectionResult{Tools: closure}
		case EventE:
			m.Phase = PhasePick
		case EventN, EventEsc, EventCtrlC, EventQ:
			return true, nil
		}
		return false, nil
	}
	switch ev {
	case EventUp:
		m.MoveUp()
	case EventDown:
		m.MoveDown()
	case EventLeft:
		m.Collapse()
	case EventRight:
		m.Expand()
	case EventSpace:
		m.Toggle()
	case EventQuestion:
		m.HelpOpen = !m.HelpOpen
	case EventEnter:
		if m.HasSelection() {
			m.Phase = PhaseConfirm
			m.StatusMsg = ""
		} else {
			m.StatusMsg = "Nothing selected — press Space to select tools"
		}
	case EventEsc, EventCtrlC, EventQ:
		return true, nil
	}
	return false, nil
}

// drawScreen clears the terminal and writes the current render.
func drawScreen(m *Model, width, height int) {
	// Clear screen and move cursor to top-left.
	os.Stdout.WriteString("\x1b[2J\x1b[H")
	os.Stdout.WriteString(Render(m, width, height))
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/picker/... -v
```
Expected: all pass. (`Run` itself is exercised only via manual smoke test; its pure helpers and state machine are fully covered.)

- [ ] **Step 5: Manual smoke test**

```bash
go build -o /tmp/isetup-picker-smoke .
/tmp/isetup-picker-smoke install --help  # sanity check: the binary still launches
```

(Full interactive test comes in Task 13 once `-i` is wired.)

- [ ] **Step 6: Commit**

```bash
git add internal/picker/picker.go internal/picker/model_test.go
git commit -m "feat(picker): Run with raw-mode event loop, handleEvent state dispatcher"
```

---

## Task 13: Wire `-i` / `--interactive` into `cmd/install.go`

**Files:**
- Modify: `cmd/install.go`
- Modify: `cmd/cmd_test.go`

- [ ] **Step 1: Write failing test for the decision helper**

`cmd/cmd_test.go` already exists in package `cmd` and imports `testify/assert` + `testify/require`. Append this test at the bottom:

```go
func TestDecideInteractive(t *testing.T) {
	cases := []struct {
		name      string
		f         installFlags
		tty       bool
		wantEnter bool
		wantErr   bool
	}{
		{
			name:      "explicit -i with TTY",
			f:         installFlags{interactive: true},
			tty:       true,
			wantEnter: true,
		},
		{
			name:    "explicit -i without TTY",
			f:       installFlags{interactive: true},
			tty:     false,
			wantErr: true,
		},
		{
			name:      "no flags + TTY → auto-enter",
			tty:       true,
			wantEnter: true,
		},
		{
			name:      "no flags + no TTY → no",
			tty:       false,
			wantEnter: false,
		},
		{
			name:      "-p opts out of auto",
			f:         installFlags{profiles: "00-base"},
			tty:       true,
			wantEnter: false,
		},
		{
			name:      "--dry-run opts out of auto",
			f:         installFlags{dryRun: true},
			tty:       true,
			wantEnter: false,
		},
		{
			name:      "-f opts out of auto",
			f:         installFlags{force: true},
			tty:       true,
			wantEnter: false,
		},
		{
			name:      "-i + -p is allowed",
			f:         installFlags{interactive: true, profiles: "00-base"},
			tty:       true,
			wantEnter: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enter, err := decideInteractive(tc.f, tc.tty)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantEnter, enter)
		})
	}
}
```

- [ ] **Step 2: Confirm fail**

```bash
go test ./cmd/... -run TestDecideInteractive -v
```
Expected: undefined `decideInteractive`, `installFlags`.

- [ ] **Step 3: Implement the decision helper and wire the flag**

Edit `cmd/install.go`:

- Add `interactiveFlag` to the flag vars block near the top:

```go
var (
	profilesFlag    string
	dryRunFlag      bool
	forceFlag       bool
	interactiveFlag bool
)
```

- Add the helper and the `installFlags` struct below `func truncate(...)`:

```go
type installFlags struct {
	interactive bool
	profiles    string
	dryRun      bool
	force       bool
}

// decideInteractive returns whether the install command should enter the
// interactive picker and a non-nil error if the user passed -i without a TTY.
// Auto-enter applies only when no other install flag is set; any of -p,
// --dry-run, or -f opts out of auto mode.
func decideInteractive(f installFlags, stdinIsTTY bool) (bool, error) {
	if f.interactive {
		if !stdinIsTTY {
			return false, &ExitError{Code: ExitConfigError, Message: "interactive mode requires a TTY; remove -i or run in a terminal"}
		}
		return true, nil
	}
	auto := f.profiles == "" && !f.dryRun && !f.force && stdinIsTTY
	return auto, nil
}

func isTerminalStdin() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
```

- Add the `golang.org/x/term` and `picker` imports to `cmd/install.go`:

```go
import (
	// ...
	"github.com/host452b/isetup/internal/picker"
	"golang.org/x/term"
)
```

- Register the flag in `init()`:

```go
installCmd.Flags().BoolVarP(&interactiveFlag, "interactive", "i", false, "pick tools interactively with arrow keys")
```

- In the `RunE` body, after `cfg` is loaded and after `info` is detected (i.e., after line ~102 where the detection prints finish), insert the picker path:

```go
var toolFilter []string

flagState := installFlags{
	interactive: interactiveFlag,
	profiles:    profilesFlag,
	dryRun:      dryRunFlag,
	force:       forceFlag,
}
enterInteractive, err := decideInteractive(flagState, isTerminalStdin())
if err != nil {
	return err
}

if enterInteractive {
	cfgForPicker := cfg
	if len(profiles) > 0 {
		narrowed := &config.Config{
			Version:  cfg.Version,
			Settings: cfg.Settings,
			Profiles: make(map[string]config.Profile, len(profiles)),
		}
		for _, p := range profiles {
			narrowed.Profiles[p] = cfg.Profiles[p]
		}
		cfgForPicker = narrowed
	}
	sel, err := picker.Run(cfgForPicker, info)
	if err != nil {
		return fmt.Errorf("picker: %w", err)
	}
	if sel == nil {
		return nil
	}
	toolFilter = sel.Tools
}
```

- Replace the executor call:

```go
results, err := executor.Execute(ctx, cfg, info, lg, profiles, toolFilter, onProgress)
```

- [ ] **Step 4: Run tests**

```bash
go test ./...
```
Expected: all pass.

- [ ] **Step 5: Manual smoke test (interactive)**

```bash
go build -o /tmp/isetup-i .
/tmp/isetup-i install -i --dry-run
# Expected: picker opens, arrow keys move cursor, space toggles, Enter goes to
# confirm, Y proceeds to dry-run output, q/Esc exits cleanly.
```

Also verify the non-TTY error path:
```bash
echo | /tmp/isetup-i install -i
# Expected: exit code 2, message "interactive mode requires a TTY".
```

And the CI path:
```bash
echo | /tmp/isetup-i install --dry-run -p 00-base
# Expected: behaves exactly like before the change (no picker).
```

- [ ] **Step 6: Commit**

```bash
git add cmd/install.go cmd/cmd_test.go
git commit -m "feat(install): -i/--interactive flag and auto-enter on TTY"
```

---

## Task 14: Documentation updates

**Files:**
- Modify: `README.md`
- Modify: `README_zh.md`
- Modify: `CHANGELOG.md`
- Modify: `CLAUDE.md`

- [ ] **Step 1: Add Interactive Mode section to README.md**

In `README.md`, after the "Quick Start" section and before "Default Tools", insert:

```markdown
## Interactive Mode

Run `isetup install -i` (or just `isetup install` in a TTY with no other flags)
to pick tools with arrow keys.

```
isetup install · interactive mode                    linux/amd64 · apt,pip,npm

[x] ▼ 00-base                                           14/14 selected
      [x] curl                  apt
      [x] git                   apt
      [ ] neovim                apt
[x] ▶ 01-lang-runtimes                                   7/ 7 selected
[ ] ▶ 04-ai-tools                                        0/ 7 selected
[·] ✗ 07-gpu                    no GPU detected          0/ 2 (disabled)

─────────────────────────────────────────────────────────────────────────────
↑/↓ move · Space toggle · →/← expand/collapse · Enter confirm · q quit · ? help
```

Keys:

| Key | Action |
|-----|--------|
| `↑` / `↓` / `k` / `j` | Move cursor |
| `→` / `l` | Expand profile |
| `←` / `h` | Collapse profile; from a tool row, jump to parent profile |
| `Space` | Toggle selection (on a profile: all selectable children at once) |
| `Enter` | Go to confirmation page |
| `Y` / `Enter` on confirm | Start install |
| `E` on confirm | Return to picker, preserving selection |
| `q` / `Esc` / `Ctrl+C` | Exit without installing |
| `?` | Toggle help overlay |

Defaults on open:
- Profiles whose `when:` condition is satisfied on the current system → selected, collapsed.
- Profiles whose `when:` is unmet (e.g., `has_gpu` on a laptop) → disabled (cannot be toggled).
- Tools with no install method on the current system → unchecked, shown with `⚠ no method`.

Dependencies are resolved at confirm time: selecting `claude-code` automatically pulls in `node-lts`, `nvm`, and `curl` — they appear on the confirm page under "Required dependencies".

The interactive flow opts **out** when stdin isn't a TTY (CI, `curl | bash`) or when any install flag is passed (`-p`, `-f`, `--dry-run`). To force interactive inside a TTY even with those flags, pass `-i` explicitly.
```

- [ ] **Step 2: Mirror the section in README_zh.md**

Add an equivalent `## 交互模式` section at the same location in `README_zh.md`, with the same mockup and key table translated into Chinese.

- [ ] **Step 3: Add CHANGELOG entry**

In `CHANGELOG.md`, under the `[Unreleased]` → `### Added` section, add:

```markdown
- **Interactive tool selection**: `isetup install -i` (or `isetup install` in a TTY with no other flags) opens a keyboard-driven picker with profile + tool two-level tree. Arrow keys navigate, Space toggles, Enter confirms. Dependencies are auto-added at the confirm step. CI and `curl | bash` flows are unchanged (non-TTY → current behavior).
```

- [ ] **Step 4: Clarify the CLAUDE.md rule**

In `CLAUDE.md`, modify the "What NOT to Add" section bullet to clarify nuance:

```markdown
- GUI / TUI installer as a **replacement** for the CLI — isetup stays CLI-first. An opt-in interactive picker layered on top of `isetup install` is allowed (see `isetup install -i`); the non-interactive flows (`-p`, `--dry-run`, `curl | bash`) must remain byte-for-byte identical.
```

- [ ] **Step 5: Verify the whole tree builds and tests pass**

```bash
go build ./...
go test ./...
```

- [ ] **Step 6: Commit**

```bash
git add README.md README_zh.md CHANGELOG.md CLAUDE.md
git commit -m "docs: interactive mode section, CHANGELOG, CLAUDE.md rule clarification"
```

---

## Self-Review Summary (filled out by plan author)

**Spec coverage check:**
- §3.1 Invocation rules → Task 13 (`decideInteractive` + wiring).
- §3.2 Main picker screen → Tasks 5–10.
- §3.3 Default check state (Option E) → Task 5 (`New`), Task 10 (render).
- §3.4 Keybindings → Task 4 (input), Task 12 (`handleEvent`).
- §3.5 Confirmation page → Task 11 (render), Task 12 (handleEvent + ResolveDeps).
- §4 Architecture (packages/files) → Tasks 3–12.
- §4.2 `SelectionResult` type → Task 9.
- §4.3 `cmd/install.go` + `executor.Execute` signature → Tasks 2 and 13.
- §4.4 Data flow → Tasks 12–13 (end-to-end wiring).
- §5 Dependency resolution → Task 3.
- §6 Error handling → Task 12 (small-terminal checks, SIGWINCH, Ctrl+C), Task 13 (non-TTY).
- §7 Testing strategy → Tasks 3–13 all use TDD with concrete test code.
- §8 Docs → Task 14.
- §9 Rollout risk → mitigated: non-TTY path short-circuits in `decideInteractive` tests.

No spec section is unmapped.

**Placeholder scan:** No TBD/TODO; every code-bearing step contains complete code; every test step contains concrete test bodies and expected outcomes.

**Type consistency:** `SelectionResult{Tools []string}` used consistently in Tasks 9, 12, 13. `installFlags` defined in Task 13 and used in its own test table. `Model`/`Node`/`Kind`/`CheckState`/`Phase`/`Event` names are stable across all tasks.
