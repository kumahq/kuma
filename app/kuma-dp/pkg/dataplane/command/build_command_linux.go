//go:build linux

package command

import (
	"context"
	"io"
	"os/exec"
	"syscall"
)

func BuildCommand(
	ctx context.Context,
	stdout io.Writer,
	stderr io.Writer,
	name string,
	args ...string,
) *exec.Cmd {
	command := baseBuildCommand(ctx, stdout, stderr, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGKILL,
		// Set those attributes so the new process won't receive the signals from a parent automatically.
		Setpgid: true,
		Pgid:    0,
	}
	return command
}
