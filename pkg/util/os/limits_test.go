package os

import (
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sys/unix"
)

var _ = Describe("File limits", func() {
	It("should query the open file limit", func() {
		Expect(CurrentFileLimit()).Should(BeNumerically(">", 0))
	})

	It("should raise the open file limit", func() {
		if runtime.GOOS == "darwin" {
			Skip("skipping on darwin because it requires priviledges")
		}
		initialLimits := unix.Rlimit{}
		Expect(unix.Getrlimit(unix.RLIMIT_NOFILE, &initialLimits)).Should(Succeed())

		Expect(CurrentFileLimit()).Should(BeNumerically("==", initialLimits.Cur))

		Expect(RaiseFileLimit()).Should(Succeed())

		Expect(CurrentFileLimit()).Should(BeNumerically("==", initialLimits.Max))

		// Restore the original limit.
		Expect(setFileLimit(initialLimits.Cur)).Should(Succeed())
		Expect(CurrentFileLimit()).Should(BeNumerically("==", initialLimits.Cur))
	})

	It("should fail to exceed the hard file limit", func() {
		Expect(setFileLimit(uint64(1) << 63)).Should(HaveOccurred())
	})
})
