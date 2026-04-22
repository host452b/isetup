//go:build unix

package picker

import (
	"os"
	"os/signal"
	"syscall"
)

func notifyResize(ch chan<- os.Signal) {
	signal.Notify(ch, syscall.SIGWINCH)
}

func notifyInterrupt(ch chan<- os.Signal) {
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
}
