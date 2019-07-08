package server

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	core_discovery "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
	core_xds "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/template"

	discovery_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/discovery/v1alpha1"
)

var _ = Describe("Reconcile", func() {
	Describe("reconciler", func() {

		var xdsContext core_xds.XdsContext

		BeforeEach(func() {
			xdsContext = core_xds.NewXdsContext()
		})

		It("should generate a Snaphot per Envoy Node", func() {
			// setup
			r := &reconciler{&templateSnapshotGenerator{
				ProxyTemplateResolver: &simpleProxyTemplateResolver{
					ResourceStore:        memory.NewStore(),
					DefaultProxyTemplate: template.TransparentProxyTemplate,
				},
			}, &simpleSnapshotCacher{xdsContext.Hasher(), xdsContext.Cache()}}

			// given
			info := &core_discovery.WorkloadInfo{
				Workload: &discovery_proto.Workload{
					Id: &discovery_proto.Id{
						Namespace: "example",
						Name:      "demo",
					},
					Meta: &discovery_proto.Meta{
						Labels: map[string]string{
							"app": "demo",
						},
					},
				},
				Desc: &core_discovery.WorkloadDescription{
					Version: "v1",
					Endpoints: []core_discovery.WorkloadEndpoint{
						{Address: "192.168.0.1", Port: 8080},
					},
				},
			}

			// when
			err := r.OnWorkloadUpdate(info)
			Expect(err).ToNot(HaveOccurred())

			// then
			Eventually(func() bool {
				_, err := xdsContext.Cache().GetSnapshot("demo.example")
				return err == nil
			}, "1s", "1ms").Should(BeTrue())
		})
	})
})
