package os

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func setFileLimit(n uint64) error {
	limit := unix.Rlimit{
		Cur: n,
		Max: n,
	}

	if err := unix.Setrlimit(unix.RLIMIT_NOFILE, &limit); err != nil {
		return fmt.Errorf("failed to set open file limit to %d: %w", limit.Cur, err)
	}

	return nil
}

// RaiseFileLimit raises the soft open file limit to match the hard limit.
func RaiseFileLimit() error {
	limit := unix.Rlimit{}
	if err := unix.Getrlimit(unix.RLIMIT_NOFILE, &limit); err != nil {
		return fmt.Errorf("failed to query open file limits: %w", err)
	}

	return setFileLimit(limit.Max)
}

// CurrentFileLimit reports the current soft open file limit.
func CurrentFileLimit() (uint64, error) {
	limit := unix.Rlimit{}
	if err := unix.Getrlimit(unix.RLIMIT_NOFILE, &limit); err != nil {
		return 0, fmt.Errorf("failed to query open file limits: %w", err)
	}

	return limit.Cur, nil
}
