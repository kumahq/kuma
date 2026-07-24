package component

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("panicLogLimiter", func() {
	It("allows first log per component name with zero suppressed", func() {
		l := newPanicLogLimiter(10 * time.Second)
		ok, suppressed := l.shouldLog("comp-a")
		Expect(ok).To(BeTrue())
		Expect(suppressed).To(Equal(0))
	})

	It("suppresses second log within the window and counts it", func() {
		l := newPanicLogLimiter(100 * time.Millisecond)
		ok, _ := l.shouldLog("comp-a")
		Expect(ok).To(BeTrue())
		ok, _ = l.shouldLog("comp-a")
		Expect(ok).To(BeFalse())
	})

	It("reports suppressed count on next allowed log", func() {
		l := newPanicLogLimiter(10 * time.Millisecond)
		ok, _ := l.shouldLog("comp-a")
		Expect(ok).To(BeTrue())
		// two suppressed calls within the window
		ok, _ = l.shouldLog("comp-a")
		Expect(ok).To(BeFalse())
		ok, _ = l.shouldLog("comp-a")
		Expect(ok).To(BeFalse())
		time.Sleep(20 * time.Millisecond)
		ok, suppressed := l.shouldLog("comp-a")
		Expect(ok).To(BeTrue())
		Expect(suppressed).To(Equal(2))
	})

	It("allows log again after the window expires with reset count", func() {
		l := newPanicLogLimiter(10 * time.Millisecond)
		ok, _ := l.shouldLog("comp-a")
		Expect(ok).To(BeTrue())
		time.Sleep(20 * time.Millisecond)
		ok, suppressed := l.shouldLog("comp-a")
		Expect(ok).To(BeTrue())
		Expect(suppressed).To(Equal(0))
	})

	It("tracks window independently per component name", func() {
		l := newPanicLogLimiter(100 * time.Millisecond)
		ok, _ := l.shouldLog("comp-a")
		Expect(ok).To(BeTrue())
		ok, _ = l.shouldLog("comp-a")
		Expect(ok).To(BeFalse())
		ok, _ = l.shouldLog("comp-b")
		Expect(ok).To(BeTrue())
		ok, _ = l.shouldLog("comp-b")
		Expect(ok).To(BeFalse())
	})
})
