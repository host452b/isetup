//go:build windows

package picker

import (
	"os"
	"os/signal"
)

// notifyResize is a no-op on Windows: there is no SIGWINCH equivalent.
// The picker won't auto-reflow on resize; a keypress triggers a redraw via the
// normal event loop.
func notifyResize(ch chan<- os.Signal) {}

func notifyInterrupt(ch chan<- os.Signal) {
	signal.Notify(ch, os.Interrupt)
}
