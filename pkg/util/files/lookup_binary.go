package files

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

type LookupPathFn = func() (string, error)

// LookupNextToCurrentExecutable looks for the binary next to the current binary
// Example: if this function is executed by /usr/bin/kuma-dp, this function will lookup for binary 'x' in /usr/bin/x
func LookupNextToCurrentExecutable(binary string) LookupPathFn {
	return func() (string, error) {
		ex, err := os.Executable()
		if err != nil {
			return "", err
		}
		return filepath.Dir(ex) + "/" + binary, nil
	}
}

// LookupInCurrentDirectory looks for the binary in the current directory
// Example: if this function is executed by /usr/bin/kuma-dp that was run in /home/kuma-dp, this function will lookup for binary 'x' in /home/kuma-dp/x
func LookupInCurrentDirectory(binary string) LookupPathFn {
	return func() (string, error) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return cwd + "/" + binary, nil
	}
}

func LookupInPath(path string) LookupPathFn {
	return func() (string, error) {
		return path, nil
	}
}

// LookupBinaryPath looks for a binary in order of passed lookup functions.
// It fails only if all lookup function does not contain a binary.
func LookupBinaryPath(pathFns ...LookupPathFn) (string, error) {
	var candidatePaths []string
	for _, candidatePathFn := range pathFns {
		candidatePath, err := candidatePathFn()
		if err != nil {
			continue
		}
		candidatePaths = append(candidatePaths, candidatePath)
		path, err := exec.LookPath(candidatePath)
		if err == nil {
			return path, nil
		}
	}

	return "", errors.Errorf("could not find binary in any of the following paths: %v", candidatePaths)
}
