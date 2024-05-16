package cni

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

// PidOf Find process(es) with a specified name (string match)
// copied from https://github.com/kubernetes/kubernetes/blob/v1.24.3/pkg/util/procfs/procfs_linux.go#L99
// with small modifications
// and return their pid(s)
func pidOf(name string) ([]int, error) {
	if len(name) == 0 {
		return []int{}, errors.New("name should not be empty")
	}
	return getPids(name), nil
}

// we don't need regex so this is changed to "string"
func getPids(name string) []int {
	pids := []int{}

	dirFD, err := os.Open("/proc")
	if err != nil {
		return nil
	}
	defer dirFD.Close()

	for {
		// Read a small number at a time in case there are many entries, we don't want to
		// allocate a lot here.
		ls, err := dirFD.Readdir(10)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil
		}

		for _, entry := range ls {
			if !entry.IsDir() {
				continue
			}

			// If the directory is not a number (i.e. not a PID), skip it
			pid, err := strconv.Atoi(entry.Name())
			if err != nil {
				continue
			}

			cmdline, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "cmdline"))
			if err != nil {
				continue
			}

			// The bytes we read have '\0' as a separator for the command line
			parts := bytes.SplitN(cmdline, []byte{0}, 2)
			if len(parts) == 0 {
				continue
			}
			// Split the command line itself we are interested in just the first part
			exe := strings.FieldsFunc(string(parts[0]), func(c rune) bool {
				return unicode.IsSpace(c) || c == ':'
			})
			if len(exe) == 0 {
				continue
			}
			// Check if the name of the executable is what we are looking for
			if name == exe[0] {
				// Grab the PID from the directory path
				pids = append(pids, pid)
			}
		}
	}

	return pids
}
