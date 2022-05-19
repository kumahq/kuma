//go:build windows

package command

import (
	"context"
	"io"
	"os/exec"
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
	// todo(jakubdyszkiewicz): do not propagate SIGTERM

	return command
}
