package command

import (
	"context"
	"io"
	"os/exec"
	"syscall"
	"time"
)

func baseBuildCommand(
	ctx context.Context,
	stdout io.Writer,
	stderr io.Writer,
	name string,
	args ...string,
) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	command.Stdout = stdout
	command.Stderr = stderr
	command.Cancel = func() error {
		return command.Process.Signal(syscall.SIGTERM)
	}
	command.WaitDelay = time.Second * 5

	return command
}
