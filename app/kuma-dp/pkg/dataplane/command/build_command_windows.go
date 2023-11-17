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
	command := baseBuildCommand(ctx, stdout, stderr, name, args...)
	// todo(jakubdyszkiewicz): do not propagate SIGTERM

	return command
}
