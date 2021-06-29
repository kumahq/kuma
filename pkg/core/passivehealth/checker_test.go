package passivehealth_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/passivehealth"
)

type testLastSeen struct {
	lastSeen time.Time
	delta    time.Duration
}

func (ls *testLastSeen) SetLastSeen(t time.Time, delta time.Duration) {
	ls.lastSeen = t
	ls.delta = delta
}

var _ = Describe("Passive Health Checker", func() {

	setNow := func(t time.Time) {
		core.Now = func() time.Time {
			return t
		}
	}

	It("should set lastSeen", func() {
		// given
		checker := passivehealth.NewChecker(20*time.Second, 5*time.Second)
		lastSeen := &testLastSeen{}

		// when
		t1 := time.Now()
		setNow(t1)
		checker.MarkAsAlive("key1", lastSeen)

		// then
		Expect(lastSeen.lastSeen).To(Equal(t1))
		Expect(lastSeen.delta).To(Equal(25 * time.Second))
	})

	It("should not update lastSeen before 'delta'", func() {
		// given
		checker := passivehealth.NewChecker(10*time.Second, 5*time.Second)
		lastSeen := &testLastSeen{}

		// when
		t1 := time.Now()
		setNow(t1)
		checker.MarkAsAlive("key1", lastSeen)
		// then
		Expect(lastSeen.lastSeen).To(Equal(t1))

		// when
		t2 := t1.Add(2 * time.Second)
		setNow(t2)
		checker.MarkAsAlive("key1", lastSeen)
		// then
		Expect(lastSeen.lastSeen).To(Equal(t1))

		// when
		t3 := t2.Add(10 * time.Second)
		setNow(t3)
		checker.MarkAsAlive("key1", lastSeen)
		// then
		Expect(lastSeen.lastSeen).To(Equal(t3))
	})

	It("should mark different keys independently", func() {
		// given
		checker := passivehealth.NewChecker(10*time.Second, 5*time.Second)
		lastSeen1 := &testLastSeen{}
		lastSeen2 := &testLastSeen{}

		// when
		t1 := time.Now()
		setNow(t1)
		checker.MarkAsAlive("key1", lastSeen1)
		// then
		Expect(lastSeen1.lastSeen).To(Equal(t1))

		// when
		t2 := t1.Add(7 * time.Second)
		setNow(t2)
		checker.MarkAsAlive("key2", lastSeen2)
		// then
		Expect(lastSeen2.lastSeen).To(Equal(t2))

		// when
		t3 := t2.Add(5 * time.Second)
		setNow(t3)
		checker.MarkAsAlive("key1", lastSeen1)
		checker.MarkAsAlive("key2", lastSeen2)
		// then
		Expect(lastSeen1.lastSeen).To(Equal(t3))
		Expect(lastSeen2.lastSeen).To(Equal(t2))
	})
})
