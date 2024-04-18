package core

import (
	"os"
	"syscall"
)

var handledSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM, usr2}

// "syscall" doesn't define a fake USR2 for us...
type usr2Signal struct {
}

func (usr2Signal) String() string {
	return "user defined signal 2"
}

func (usr2Signal) Signal() {}

var usr2 os.Signal = usr2Signal{}
