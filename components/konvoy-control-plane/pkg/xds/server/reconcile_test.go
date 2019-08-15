package server

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
	test_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/resources/model"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/template"
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
			dataplane := &mesh_core.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Mesh:      "pilot",
					Namespace: "example",
					Name:      "demo",
					Version:   "v1",
				},
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Interface: "192.168.0.1:8080:8080",
							},
						},
					},
				},
			}

			// when
			err := r.OnDataplaneUpdate(dataplane)
			Expect(err).ToNot(HaveOccurred())

			// then
			Eventually(func() bool {
				_, err := xdsContext.Cache().GetSnapshot("demo.example.pilot")
				return err == nil
			}, "1s", "1ms").Should(BeTrue())
		})
	})
})
