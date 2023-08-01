//go:build linux

package cgroups

import (
	"path/filepath"
	"sync"

	"golang.org/x/sys/unix"
)

// TAKEN FROM https://github.com/containerd/cgroups/blob/v1.1.0/utils.go
// to get rid of dependency on containerd because of it's various CVEs

// CGMode is the cgroups mode of the host system
type CGMode int

const unifiedMountpoint = "/sys/fs/cgroup"

const (
	// Unavailable cgroup mountpoint
	Unavailable CGMode = iota
	// Legacy cgroups v1
	Legacy
	// Hybrid with cgroups v1 and v2 controllers mounted
	Hybrid
	// Unified with only cgroups v2 mounted
	Unified
)

var (
	checkMode sync.Once
	cgMode    CGMode
)

// Mode returns the cgroups mode running on the host
func Mode() CGMode {
	checkMode.Do(func() {
		var st unix.Statfs_t
		if err := unix.Statfs(unifiedMountpoint, &st); err != nil {
			cgMode = Unavailable
			return
		}
		switch st.Type {
		case unix.CGROUP2_SUPER_MAGIC:
			cgMode = Unified
		default:
			cgMode = Legacy
			if err := unix.Statfs(filepath.Join(unifiedMountpoint, "unified"), &st); err != nil {
				return
			}
			if st.Type == unix.CGROUP2_SUPER_MAGIC {
				cgMode = Hybrid
			}
		}
	})
	return cgMode
}
