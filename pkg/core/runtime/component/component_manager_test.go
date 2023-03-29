package component_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/runtime/component"
	leader_memory "github.com/kumahq/kuma/pkg/plugins/leader/memory"
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
