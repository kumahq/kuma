package netns

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// This article is very helpful to understand the logic of why we have to use
// runtime.LockOSThread() in places where we are working with the Linux network
// namespaces:
// https://www.ardanlabs.com/blog/2018/08/scheduling-in-go-part2.html

// CLONE_NEWNET requires Linux Kernel 3.0+

type NetNS struct {
	name              string
	ns                netns.NsHandle
	originalNS        netns.NsHandle
	veth              *Veth
	beforeExecFuncs   []func() error
	sharedLinkAddress *netlink.Addr
}

func (ns *NetNS) Name() string {
	return ns.name
}

func (ns *NetNS) Veth() *Veth {
	return ns.veth
}

func (ns *NetNS) Set() error {
	if err := netns.Set(ns.ns); err != nil {
		return fmt.Errorf("cannot switch to the network namespace %q: %s", ns.ns.String(), err)
	}

	return nil
}

func (ns *NetNS) Unset() error {
	if err := netns.Set(ns.originalNS); err != nil {
		return fmt.Errorf(
			"cannot switch to the original network namespace %q: %s",
			ns.originalNS.String(),
			err,
		)
	}

	return nil
}

// UnsafeExec will execute provided callback function in the created network namespace
// from the *NetNS. It was named UnsafeExec instead of Exec as you have to be very
// cautious and remember to not spawn new goroutines inside provided callback (more
// info in warning below)
//
// WARNING!:
//
//	Don't spawn new goroutines inside callback functions as the one inside UnsafeExec
//	function have exclusive access to the current network namespace, and you should
//	assume, that any new goroutine will be placed in the different namespace
func (ns *NetNS) UnsafeExec(callback func(), beforeCallbackFuncs ...func() error) <-chan error {
	return ns.UnsafeExecInLoop(1, 0, callback, beforeCallbackFuncs...)
}

// UnsafeExecInLoop will execute provided callback function inside the created
// network namespace in a loop. It was named UnsafeExecInLoop instead of ExecInLoop
// as you have to be very cautious and remember to not spawn new goroutines
// inside provided callback (more info in warning below)
//
// WARNING!:
//
//	Don't spawn new goroutines inside callback functions as the one inside UnsafeExecInLoop
//	function have exclusive access to the current network namespace, and you should
//	assume, that any new goroutine will be placed in the different namespace
func (ns *NetNS) UnsafeExecInLoop(
	numOfIterations uint,
	delay time.Duration,
	callback func(),
	beforeCallbackFuncs ...func() error,
) <-chan error {
	done := make(chan error)

	go func() {
		defer ginkgo.GinkgoRecover()
		defer close(done)

		runtime.LockOSThread()

		if err := ns.Set(); err != nil {
			done <- fmt.Errorf("cannot set the namespace %q: %s", ns.name, err)
		}
		defer ns.Unset() //nolint:errcheck

		for _, fn := range append(ns.beforeExecFuncs, beforeCallbackFuncs...) {
			if err := fn(); err != nil {
				done <- err
			}
		}

		for i := 0; i < int(numOfIterations); i++ {
			callback()

			time.Sleep(delay)
		}
	}()

	return done
}

func (ns *NetNS) Cleanup() error {
	if ns == nil {
		return nil
	}

	done := make(chan error)

	// It's necessary to run the code in separate goroutine to lock the os thread
	// to pin the network namespaces for our purposes
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		var errs []string

		if ns.originalNS.IsOpen() {
			if err := ns.originalNS.Close(); err != nil {
				errs = append(errs, fmt.Sprintf("cannot close the original namespace fd: %s", err))
			}
		}

		if ns.ns.IsOpen() {
			if err := ns.ns.Close(); err != nil {
				errs = append(errs, fmt.Sprintf("cannot close the network namespace fd: %s", err))
			}
		}

		if err := netNsDeleteNamed(ns.Name()); err != nil {
			errs = append(errs, fmt.Sprintf("cannot delete network namespace: %s", err))
		}

		veth := ns.Veth().Veth()
		if err := netlink.LinkDel(veth); err != nil {
			errs = append(errs, fmt.Sprintf("cannot delete veth interface %q: %s", veth.Name, err))
		}

		if len(errs) > 0 {
			done <- fmt.Errorf("cleanup failed:\n  - %s", strings.Join(errs, "\n  - "))
		}

		close(done)
	}()

	return <-done
}

func (ns *NetNS) SharedLinkAddress() *netlink.Addr {
	return ns.sharedLinkAddress
}
