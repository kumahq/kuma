package core

import (
	"os"
	"syscall"
)

var handledSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2}

var usr2 os.Signal = syscall.SIGUSR2
