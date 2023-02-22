package syscall

import (
	"fmt"
	"syscall"
)

// SetUID - This is very hacky way of making one process to operate under multiple UIDs,
// which breaks POSIX semantics (ref. https://man7.org/linux/man-pages/man7/nptl.7.html)
// The go's native syscall.Setuid is designed in a way, that all threads
// will switch to provided UID, so we have to use the Linux's setuid() syscall
// directly (it doesn't honor POSIX semantics).
//
// This logic exists to potentially run tests with DNS UDP conntrack zone splitting
// enabled.
//
// ref. https://stackoverflow.com/a/66523695
func SetUID(uid uintptr) func() error {
	return func() error {
		if _, _, e := syscall.RawSyscall(syscall.SYS_SETUID, uid, 0, 0); e != 0 {
			return fmt.Errorf("cannot exec syscall.SYS_SETUID (error number: %d)", e)
		}

		return nil
	}
}
