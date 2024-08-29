//go:build !linux
// +build !linux

package config

import (
	"bytes"
	"context"
	"os/exec"
)

func (c InitializedExecutable) Exec(
	ctx context.Context,
	args ...string,
) (*bytes.Buffer, *bytes.Buffer, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	// #nosec G204
	cmd := exec.CommandContext(ctx, c.Path, append(c.args, args...)...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, nil, handleRunError(err, &stderr)
	}

	return &stdout, &stderr, nil
}
