package nsenter

import (
	"runtime"
	"sync"
)

// NsEnter executes the passed closure under the given namespace,
// restoring the original namespace afterwards.
func NsEnter(nsList []Namespace, toRun func() error) error {
	containedCall := func() error {
		return toRun()
	}

	var wg sync.WaitGroup
	wg.Add(1)

	var innerError error
	go func() {
		defer wg.Done()
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		innerError = containedCall()
	}()
	wg.Wait()

	return innerError
}
