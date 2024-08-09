//go:build !linux
// +build !linux

package config

import (
	"bytes"
	"context"
	"os/exec"
)

func execCmd(
	ctx context.Context,
	_ Logger,
	_ bool,
	_ bool,
	path string,
	args ...string,
) (*bytes.Buffer, *bytes.Buffer, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	// #nosec G204
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, nil, handleRunError(err, &stderr)
	}

	return &stdout, &stderr, nil
}
