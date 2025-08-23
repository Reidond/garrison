package utils

import (
	"os"
	"os/signal"
	"syscall"
)

// SetupSignalHandler wires OS signals to invoke cancelFunc.
func SetupSignalHandler(cancelFunc func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		cancelFunc()
	}()
}
