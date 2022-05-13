//go:build !linux && !windows

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
	command := exec.CommandContext(ctx, name, args...)
	command.Stdout = stdout
	command.Stderr = stderr
	command.SysProcAttr = &syscall.SysProcAttr{
		// Set those attributes so the new process won't receive the signals from a parent automatically.
		Setpgid: true,
		Pgid:    0,
	}

	return command
}
