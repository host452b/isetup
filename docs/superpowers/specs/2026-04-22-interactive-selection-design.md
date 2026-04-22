# Interactive Tool Selection for `isetup install`

- **Status**: Approved
- **Author**: joe0731 (brainstormed with Claude)
- **Date**: 2026-04-22
- **Target release**: unreleased / next minor
- **Related**: `cmd/install.go`, `internal/executor/executor.go`, `template/default.yaml`

## 1. Motivation

Today, `isetup install` is all-or-nothing per profile. Users who want to install only a handful of tools from the 62-tool default template must either:

- pass `-p 00-base,01-lang-runtimes` (still installs every tool in those profiles), or
- hand-edit `~/.isetup.yaml` to remove unwanted tools.

Neither matches the "one-click, press Enter and go" feel of the one-liner install. This design adds a keyboard-driven multi-select UI that lets the user pick individual tools (or whole profiles) before install, while keeping every existing non-interactive code path byte-for-byte identical.

## 2. Goals & Non-goals

### Goals
- Tool-level and profile-level selection in a single TUI.
- Zero behavioral change when `-i` is not passed and stdin is not a TTY (CI, `curl | bash`, `-p` workflows unchanged).
- One new direct dependency max (`golang.org/x/term`).
- All state machine logic in pure, independently testable functions.

### Non-goals (v1)
- Search / filter (`/`)
- Select-all / select-none shortcuts (`a` / `n`)
- Persisting selection across runs
- Mouse support
- Theme / color configuration
- Custom keybindings
- Legacy Windows `cmd.exe` (older than Windows 10 1809 / without VT). Windows Terminal and PowerShell 5.1+ are supported.

## 3. User Experience

### 3.1 Invocation rules

| Invocation | Behavior |
|------------|----------|
| `isetup install` (no args) + TTY stdin | **Enters interactive picker automatically** |
| `isetup install` (no args) + non-TTY stdin | Current behavior: install all profiles |
| `isetup install -i` / `--interactive` | Force interactive; exit 2 if stdin is not a TTY |
| `isetup install -p A,B` | Current behavior: non-interactive, install all tools in A and B |
| `isetup install -p A,B -i` | Interactive picker restricted to profiles A and B |
| `isetup install -i --dry-run` | Interactive picker → on confirm, run the current dry-run path |
| `isetup install -i -f` | Interactive picker → `--force` applies to the selected tools |
| `curl … \| bash` → `isetup install` | Non-TTY, fallback to current all-install behavior |

The auto-enter-on-TTY rule is deliberately conservative: it only triggers when **no other install flags** are passed. Any flag (`-p`, `-f`, `--dry-run`, etc.) opts out of auto-interactive. This preserves scripted usage.

### 3.2 Main picker screen

```
isetup install · interactive mode                    linux/amd64 · apt,pip,npm

[x] ▼ 00-base                                           14/14 selected
      [x] curl                  apt
      [x] wget                  apt
      [x] git                   apt
      [x] ripgrep               apt
      [x] fzf                   apt
      [x] bat                   apt
      [ ] neovim                apt
      ...
[x] ▶ 01-lang-runtimes                                   7/ 7 selected
[x] ▶ 02-git-tools                                       5/ 5 selected
[ ] ▶ 04-ai-tools                                        0/ 7 selected
[·] ✗ 07-gpu                    no GPU detected          0/ 2 (disabled)

─────────────────────────────────────────────────────────────────────────────
↑/↓ move · Space toggle · →/← expand/collapse · Enter confirm · q quit · ? help
```

**Visual elements:**
- Cursor row: ANSI reverse video.
- Checkbox: `[x]` selected, `[ ]` unselected, `[-]` partially selected (profile), `[·]` disabled (`when` unmet).
- Arrow: `▼` expanded, `▶` collapsed, `✗` disabled.
- Profile title: bold; tool method column: dim; disabled row: dim grey.
- Header shows detected OS/arch and available package managers (read-only, informational).
- Status bar is always visible.
- `?` toggles a help overlay that describes all keys; any key dismisses it.

### 3.3 Default check state (decision E)

On opening:
- Profiles whose `when` evaluates `true` (or is absent) → **selected, collapsed**.
- Profiles whose `when` evaluates `false` → **unselected, disabled, collapsed, greyed out**. Space has no effect on disabled rows; attempts to toggle a disabled row do nothing.
- Within each enabled profile, each tool that has at least one install method resolvable on the current system → **selected**.
- Tools with no resolvable install method for the current system → **unselected**, displayed with a `⚠ no method for <os>` marker on the right. The user may still toggle them (they may know better), but the confirmation page will warn and the executor will skip with the existing `no install method` reason.

The denominator shown in the profile header (`14/14 selected`) counts only tools that are *selectable* (i.e., excludes tools with no method on current system). Rationale: matches the semantics "what will actually run".

### 3.4 Keybindings

| Key | Action |
|-----|--------|
| `↑` / `k` | Cursor up |
| `↓` / `j` | Cursor down |
| `→` / `l` | On profile row: expand. On tool row: no-op. |
| `←` / `h` | On profile row: collapse. On tool row: jump to parent profile. |
| `Space` | Toggle selection. On profile row: if any unselected → select all selectable children; else unselect all. On tool row: toggle. On disabled row: no-op. |
| `Enter` | Proceed to confirmation page. If 0 tools selected, refuse and show a red status message. |
| `q` / `Esc` / `Ctrl+C` | Exit cleanly without installing (exit 0). |
| `?` | Toggle help overlay. |

All other keys are ignored (no beep, no error).

### 3.5 Confirmation page

```
Review & Install
─────────────────────────────────────────────────────────────

You selected 4 tools:
    claude-code              shell
    codex-cli                shell
    ollama                   shell
    uv                       pip

Required dependencies (auto-added): 3 tools
    curl                     apt            → already installed, will skip
    nvm                      shell
    node-lts                 shell

Total: 7 tools will be attempted
Log: ~/.isetup/logs/isetup-2026-04-22T14-23-09.log

─────────────────────────────────────────────────────────────
[Y/Enter] Install   [E] Edit selection   [N/Esc] Cancel
```

- `Required dependencies` section is omitted entirely if the selection is already closed under `depends_on`.
- Already-in-PATH dependencies (detected via existing `IsInstalled`) are annotated `→ already installed, will skip`; they stay in the list so the user sees the full dependency picture.
- `Y` / `Enter` → restore terminal, execute with the computed tool set.
- `E` → return to picker, preserving all check state, cursor position, expansion state.
- `N` / `Esc` / `Ctrl+C` → exit 0, nothing installed, log directory not created.

## 4. Architecture

### 4.1 New package: `internal/picker/`

Four files, each with a single narrow responsibility:

| File | Responsibility | I/O? |
|------|----------------|------|
| `model.go` | `Model` struct (tree nodes, cursor index, expansion set, selection set, disabled set, help-visible flag). Pure state-transition functions: `New(cfg, info)`, `MoveUp()`, `MoveDown()`, `Toggle()`, `Expand()`, `Collapse()`, `Selection()`. | None (pure) |
| `render.go` | `Render(m *Model, width, height int) string` returning ANSI-colored frame. Respects `NO_COLOR` env var. | None (pure) |
| `input.go` | `ParseKey(buf []byte) (Event, int)` — byte prefix → key event + consumed byte count. | None (pure) |
| `picker.go` | `Run(cfg *config.Config, info *detector.SystemInfo) (*Selection, error)` — wraps `x/term.MakeRaw` / `term.Restore`, SIGWINCH handling, read loop (read → parse → update model → render). Returns `nil, nil` on clean cancel, `*Selection, nil` on confirm, `nil, err` on terminal error. | stdin/stdout, signals |
| `deps.go` | `ResolveDeps(selected []string, allTools []config.Tool) (closure, added []string)` — BFS over `DependsOn` edges. | None (pure) |

### 4.2 New data type

```go
// internal/picker/selection.go (co-located with Run)
type Selection struct {
    Tools []string // tool names selected by user + resolved dependencies
}
```

### 4.3 Changes to existing code

**`cmd/install.go`:**
- Add `interactiveFlag bool` bound to `-i` / `--interactive`.
- Before loading the profile filter, decide if interactive mode applies:
  - Explicit: `interactiveFlag == true`.
  - Auto: `profilesFlag == "" && !dryRunFlag && !forceFlag && isTerminal(os.Stdin)`.
- If interactive and stdin is not a TTY → return `&ExitError{Code: 2, Message: "interactive mode requires a TTY; remove -i or run in a terminal"}`.
- Call `picker.Run(cfg, info)`; on `nil, nil` result, exit 0 cleanly (no log file).
- On `*Selection` result, set `cfg.Settings.Force = forceFlag` (unchanged) and proceed to `executor.Execute`, passing the tool filter.

**`internal/executor/executor.go`:**
- Extend `Execute` signature:
  ```go
  func Execute(ctx, cfg, info, lg, profiles []string, tools []string, onProgress) ([]logger.ToolResult, error)
  ```
  When `tools != nil`, `collectTools` keeps only entries whose `Tool.Name` is in `tools` (profile's `when` still evaluated; `SkipReason` still carried).
- All existing callers (current `install.go` path) pass `nil` for `tools`, preserving current behavior.

**`go.mod`:**
- Add `golang.org/x/term`. Transitive deps already present via cobra (`golang.org/x/sys`).

### 4.4 Data flow

```
isetup install [-i]
  ↓
cmd/install.go
  load config → detect system → decide interactive path
  ↓ interactive branch
picker.Run(cfg, info)
  build Model from profiles + system info (default state E)
  loop: read stdin → ParseKey → update Model → Render → write stdout
  on Enter: run ResolveDeps; render confirm page; read Y/E/N
  on Y: return &Selection{Tools: closure}
  on N/Esc/q anywhere: return nil
  ↓
executor.Execute(ctx, cfg, info, lg, nil, selection.Tools, onProgress)
  collectTools filters by tool name (in addition to existing profile + when filters)
  topo sort → per-tool: bootstrap, resolve method, run, fallback, log
  ↓
summary & exit (existing)
```

## 5. Dependency resolution algorithm

Pure function in `internal/picker/deps.go`:

```go
func ResolveDeps(selected []string, all []config.Tool) (closure, added []string) {
    byName := make(map[string]config.Tool, len(all))
    for _, t := range all { byName[t.Name] = t }

    inClosure := make(map[string]bool)
    queue := append([]string(nil), selected...)
    for len(queue) > 0 {
        t := queue[0]; queue = queue[1:]
        if inClosure[t] { continue }
        inClosure[t] = true
        if tool, ok := byName[t]; ok && tool.DependsOn != "" {
            queue = append(queue, tool.DependsOn)
        }
    }

    selectedSet := make(map[string]bool, len(selected))
    for _, s := range selected { selectedSet[s] = true }

    for name := range inClosure {
        closure = append(closure, name)
        if !selectedSet[name] { added = append(added, name) }
    }
    sort.Strings(closure)
    sort.Strings(added)
    return
}
```

Behavior notes:
- Dependencies pointing outside the configured tool set (`depends_on: foo` where `foo` isn't in any profile) are **not added** to closure — the existing executor's `UnresolvedDep + IsInstalled` path handles them at run time. No change needed.
- Cycles are caught by the existing `TopoSort` downstream; the resolver doesn't need to detect them.

## 6. Error handling & edge cases

| Case | Handling |
|------|----------|
| `-i` passed, stdin not TTY | `ExitError{Code: 2, Message: "interactive mode requires a TTY; remove -i or run in a terminal"}` |
| Terminal width < 50 cols | Collapse the right-hand method column; hide status bar help text, show `? for keys`. |
| Terminal width < 30 cols | Exit with `terminal too narrow (need at least 30 cols)`, exit 2. |
| Terminal height < 10 rows | Exit with `terminal too small (need at least 10 rows)`, exit 2. |
| SIGWINCH (resize) | Re-query terminal size via `x/term.GetSize`, re-render. Model untouched. |
| SIGINT / `Ctrl+C` | Same as Esc: restore terminal, exit 0. |
| Panic in raw mode | `defer term.Restore(fd, oldState)` ensures terminal is usable on crash. |
| Tool with no install method on current system | Render with `⚠ no method for <os>`; do not auto-select; user may still toggle; executor will skip at run time with existing reason. |
| User confirms with 0 tools | Refuse, return to main picker with red status message `Nothing selected — press Space to select tools`. |
| `when` condition evaluates false | Profile is disabled (see 3.3); Space has no effect; tools inside cannot be toggled individually (the rows may still be rendered on expand, but all are checkbox `[·]`). |
| `NO_COLOR` env var set | Skip ANSI color codes, keep structural characters (`[x]`, `▼`). |

## 7. Testing strategy

Follows the existing testify-based, offline-only convention.

| File | Coverage |
|------|----------|
| `internal/picker/model_test.go` | Initial state construction (all decisions of 3.3); cursor bounds; Space on profile/tool/disabled; Expand/Collapse; `Selection()` output correctness. |
| `internal/picker/render_test.go` | Snapshot tests: fixed Model → rendered string contains expected lines. Separate test with `NO_COLOR=1`. Narrow-width variant. |
| `internal/picker/input_test.go` | Byte sequence → Event lookup table: `\x1b[A→KeyUp`, `\x1b[B→KeyDown`, `\x1b[C→KeyRight`, `\x1b[D→KeyLeft`, `\x20→KeySpace`, `\x0d→KeyEnter`, `\x03→KeyCtrlC`, `\x1b→KeyEsc` (disambiguated from CSI prefix by timeout / byte-count), `k/j/h/l/q/?` → respective events. |
| `internal/picker/deps_test.go` | Linear chain; diamond; self-already-selected; unresolved dep (depends on tool not in config); empty selection; cycle would fail upstream (not tested here). |
| `cmd/cmd_test.go` (existing file, extend) | `-i` + non-TTY → exit 2; `-i + -p` combo forwards filter correctly. |
| `e2e_test.go` | **Unchanged.** Interactive path skipped — requires a PTY which the project policy avoids bringing in. |

PTY-based end-to-end tests are explicitly **out of scope**: `picker.go`'s main loop is a thin glue layer over three independently-tested pure functions, and adding a PTY mock library would violate the minimal-deps rule.

## 8. Documentation updates

- `README.md` — add "Interactive Mode" section under Quick Start with ASCII sample and the invocation rule table.
- `README_zh.md` — synchronized translation of the same section.
- `CHANGELOG.md` — `[Unreleased] → Added: interactive tool selection via isetup install -i / --interactive; auto-enters when running in a TTY with no other flags`.
- `CLAUDE.md` — add an entry to the "What NOT to Add" section clarifying the nuance: a **full-TUI replacement** is still disallowed; the interactive picker is an **opt-in mode** layered on top of the existing CLI flows.

## 9. Rollout & risk

- **Risk of breaking CI / `curl | bash`**: Mitigated by the auto-enter rule requiring a TTY *and* no flags. `curl | bash` invokes `isetup install` with stdin piped, which fails the TTY check.
- **Risk of terminal state corruption on crash**: Mitigated by `defer term.Restore`. Additionally, `picker.Run` installs a signal handler that restores terminal state on SIGTERM/SIGQUIT before re-raising.
- **Binary size**: Expected delta < 300 KB (only `x/term` is new; `x/sys` already present transitively). Will verify with `go build -ldflags "-s -w"` before merging.
- **Rollback**: The feature is additive. If `-i` misbehaves in the wild, reverting the `cmd/install.go` flag + auto-enter block restores current behavior without touching the picker package.
