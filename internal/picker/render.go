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

	vis := m.visibleIndices()
	cursorPos := indexOf(vis, m.Cursor)
	narrow := width < 50
	b.WriteString(renderHeader(m, width, narrow))
	b.WriteString("\n\n")
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

func renderHeader(m *Model, width int, narrow bool) string {
	left := col(ansiBold, "isetup install · interactive mode")
	right := ""
	if m.Info != nil && !narrow {
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

// renderConfirm is a stub; implemented in Task 11.
// TODO: implemented in Task 11
func renderConfirm(m *Model, width, height int) string {
	return ""
}
