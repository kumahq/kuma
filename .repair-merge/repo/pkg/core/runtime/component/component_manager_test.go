package component_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/core/runtime/component"
	leader_memory "github.com/kumahq/kuma/v3/pkg/plugins/leader/memory"
)

var _ = Describe("Component Manager", func() {
	Context("Component Manager is running", func() {
		var manager component.Manager
		var stopCh chan struct{}

		BeforeAll(func() {
			// given
			manager = component.NewManager(leader_memory.NewNeverLeaderElector())
			chComponentBeforeStart := make(chan int)
			err := manager.Add(component.ComponentFunc(func(_ <-chan struct{}) error {
				close(chComponentBeforeStart)
				return nil
			}))

			// when component manager is started
			stopCh = make(chan struct{})
			go func() {
				defer GinkgoRecover()
				Expect(manager.Start(stopCh)).To(Succeed())
			}()

			// then component added before Start() runs
			Expect(err).ToNot(HaveOccurred())
			Eventually(chComponentBeforeStart, "30s", "50ms").Should(BeClosed())
		})

		AfterAll(func() {
			close(stopCh)
		})

		It("should be able to add component in runtime", func() {
			// when component is added after Start()
			chComponentAfterStart := make(chan int)
			err := manager.Add(component.ComponentFunc(func(_ <-chan struct{}) error {
				close(chComponentAfterStart)
				return nil
			}))

			// then it runs
			Expect(err).ToNot(HaveOccurred())
			Eventually(chComponentAfterStart, "30s", "50ms").Should(BeClosed())
		})

		It("should not be able to add leader component", func() {
			// when leader component is added after Start()
			err := manager.Add(component.LeaderComponentFunc(func(_ <-chan struct{}) error {
				return nil
			}))

			// then
			Expect(err).To(Equal(component.LeaderComponentAddAfterStartErr))
		})
	})
}, Ordered)

var _ = Describe("Component supervision", func() {
	var stopCh chan struct{}

	BeforeEach(func() {
		stopCh = make(chan struct{})
	})

	AfterEach(func() {
		select {
		case <-stopCh:
		default:
			close(stopCh)
		}
	})

	It("returns error when non-leader component panics", func() {
		mgr := component.NewManager(leader_memory.NewNeverLeaderElector())
		Expect(mgr.Add(component.ComponentFunc(func(_ <-chan struct{}) error {
			panic("boom")
		}))).To(Succeed())

		err := mgr.Start(stopCh)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("panicked"))
		Expect(err.Error()).To(ContainSubstring("boom"))
	})

	It("returns nil on clean stop of non-leader component", func() {
		started := make(chan struct{})
		mgr := component.NewManager(leader_memory.NewNeverLeaderElector())
		Expect(mgr.Add(component.ComponentFunc(func(stop <-chan struct{}) error {
			close(started)
			<-stop
			return nil
		}))).To(Succeed())

		done := make(chan error, 1)
		go func() { done <- mgr.Start(stopCh) }()
		Eventually(started, "5s", "10ms").Should(BeClosed())
		close(stopCh)
		Eventually(done, "5s", "10ms").Should(Receive(BeNil()))
	})

	It("returns error when non-leader component exits with error", func() {
		mgr := component.NewManager(leader_memory.NewNeverLeaderElector())
		Expect(mgr.Add(component.ComponentFunc(func(_ <-chan struct{}) error {
			return errors.New("component failure")
		}))).To(Succeed())

		err := mgr.Start(stopCh)
		Expect(err).To(MatchError("component failure"))
	})

	It("returns error when leader component panics", func() {
		mgr := component.NewManager(leader_memory.NewAlwaysLeaderElector())
		Expect(mgr.Add(component.LeaderComponentFunc(func(_ <-chan struct{}) error {
			panic("leader boom")
		}))).To(Succeed())

		err := mgr.Start(stopCh)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("panicked"))
		Expect(err.Error()).To(ContainSubstring("leader boom"))
	})

	It("includes component name in panic error for NamedComponent", func() {
		named := &namedTestComponent{
			name: "my-named-comp",
			fn:   func(_ <-chan struct{}) error { panic("named panic") },
		}
		mgr := component.NewManager(leader_memory.NewNeverLeaderElector())
		Expect(mgr.Add(named)).To(Succeed())

		err := mgr.Start(stopCh)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("my-named-comp"))
	})

	It("does not deadlock when a second component panics after manager has returned", func() {
		// Component B blocks on <-stop (which is internalDone). internalDone closes only after
		// component A's panic is consumed by manager.Start and it returns — so B panics
		// deterministically after manager has exited, exercising the errCh default path.
		mgr := component.NewManager(leader_memory.NewNeverLeaderElector())
		Expect(mgr.Add(component.ComponentFunc(func(_ <-chan struct{}) error {
			panic("first boom")
		}))).To(Succeed())
		Expect(mgr.Add(component.ComponentFunc(func(stop <-chan struct{}) error {
			<-stop
			panic("second boom")
		}))).To(Succeed())

		err := mgr.Start(stopCh)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("first boom"))
	})

	It("returns error when runtime-added component panics", func() {
		mgr := component.NewManager(leader_memory.NewNeverLeaderElector())
		blocker := make(chan struct{})
		Expect(mgr.Add(component.ComponentFunc(func(stop <-chan struct{}) error {
			close(blocker)
			<-stop
			return nil
		}))).To(Succeed())

		done := make(chan error, 1)
		go func() { done <- mgr.Start(stopCh) }()

		// wait until manager is running before adding the panicking component at runtime
		Eventually(blocker, "5s", "10ms").Should(BeClosed())
		Expect(mgr.Add(component.ComponentFunc(func(_ <-chan struct{}) error {
			panic("runtime boom")
		}))).To(Succeed())

		Eventually(done, "5s", "10ms").Should(Receive(HaveOccurred()))
	})
})

type namedTestComponent struct {
	name string
	fn   func(<-chan struct{}) error
}

func (n *namedTestComponent) Start(stop <-chan struct{}) error { return n.fn(stop) }
func (n *namedTestComponent) NeedLeaderElection() bool         { return false }
func (n *namedTestComponent) Name() string                     { return n.name }
