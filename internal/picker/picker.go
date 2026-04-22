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
