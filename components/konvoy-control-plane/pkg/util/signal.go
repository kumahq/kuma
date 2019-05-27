package util

import (
	"os"
	"os/signal"
	"syscall"
)

func WaitStopSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
