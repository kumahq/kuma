package clusterid_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/clusterid"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/test/runtime"
)

var _ = Describe("Cluster ID", func() {

	var stop chan struct{}

	AfterEach(func() {
		if stop != nil {
			close(stop)
		}
	})

	It("should create and set cluster ID", func() {
		// given runtime with cluster ID components
		cfg := kuma_cp.DefaultConfig()
		builder, err := runtime.BuilderFor(context.Background(), cfg)
		Expect(err).ToNot(HaveOccurred())
		runtime, err := builder.Build()
		Expect(err).ToNot(HaveOccurred())

		err = clusterid.Setup(runtime)
		Expect(err).ToNot(HaveOccurred())

		// when runtime is started
		stop = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := runtime.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()

		// then cluster ID is created and set in Runtime object
		Eventually(runtime.GetClusterId, "5s", "100ms").ShouldNot(BeEmpty())
	})
})
