package callbacks_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	meshidentity_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	core_manager "github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/v2/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	. "github.com/kumahq/kuma/v2/pkg/xds/server/callbacks"
)

var _ = Describe("Workload Label Validator", func() {
	var resManager core_manager.ResourceManager
	var validator *WorkloadLabelValidator

	BeforeEach(func() {
		memStore := memory.NewStore()
		resManager = core_manager.NewResourceManager(memStore)
		validator = NewWorkloadLabelValidator(resManager)

		// Create default mesh
		err := resManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	})

	Context("when MeshIdentity uses workload label in path template", func() {
		BeforeEach(func() {
			// Create MeshIdentity with workload label in path template
			mi := &meshidentity_api.MeshIdentityResource{
				Meta: &test_model.ResourceMeta{
					Name: "mi-with-workload-label",
					Mesh: "default",
				},
				Spec: &meshidentity_api.MeshIdentity{
					Selector: &meshidentity_api.Selector{
						Dataplane: &common_api.LabelSelector{
							MatchLabels: &map[string]string{
								"kuma.io/service": "web",
							},
						},
					},
					SpiffeID: &meshidentity_api.SpiffeID{
						TrustDomain: pointer.To("{{ label \"kuma.io/mesh\" }}.mesh.local"),
						Path:        pointer.To("/workload/{{ label \"kuma.io/workload\" }}"),
					},
				},
			}
			err := resManager.Create(context.Background(), mi, core_store.CreateByKey("mi-with-workload-label", "default"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should allow connection when dataplane has workload label", func() {
			// given
			dpRes := &core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "web-01",
					Mesh: "default",
					Labels: map[string]string{
						"kuma.io/service":     "web",
						metadata.KumaWorkload: "my-workload",
					},
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "127.0.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port:        8080,
								ServicePort: 8081,
								Tags: map[string]string{
									"kuma.io/service": "web",
								},
							},
						},
					},
				},
			}

			md := core_xds.DataplaneMetadata{
				Resource:  dpRes,
				ProxyType: mesh_proto.DataplaneProxyType,
			}

			// when
			err := validator.OnProxyConnected(
				core_xds.StreamID(1),
				core_model.ResourceKey{Mesh: "default", Name: "web-01"},
				context.Background(),
				md,
			)

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should deny connection when dataplane is missing workload label", func() {
			// given
			dpRes := &core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "web-02",
					Mesh: "default",
					Labels: map[string]string{
						"kuma.io/service": "web",
						// Missing kuma.io/workload label
					},
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "127.0.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port:        8080,
								ServicePort: 8081,
								Tags: map[string]string{
									"kuma.io/service": "web",
								},
							},
						},
					},
				},
			}

			md := core_xds.DataplaneMetadata{
				Resource:  dpRes,
				ProxyType: mesh_proto.DataplaneProxyType,
			}

			// when
			err := validator.OnProxyConnected(
				core_xds.StreamID(2),
				core_model.ResourceKey{Mesh: "default", Name: "web-02"},
				context.Background(),
				md,
			)

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("missing required label 'kuma.io/workload'"))
			Expect(err.Error()).To(ContainSubstring("mi-with-workload-label"))
			Expect(err.Error()).To(ContainSubstring("/workload/{{ label \"kuma.io/workload\" }}"))
		})
	})

	Context("when MeshIdentity does not use workload label", func() {
		BeforeEach(func() {
			// Create MeshIdentity without workload label in path template
			mi := &meshidentity_api.MeshIdentityResource{
				Meta: &test_model.ResourceMeta{
					Name: "mi-without-workload-label",
					Mesh: "default",
				},
				Spec: &meshidentity_api.MeshIdentity{
					Selector: &meshidentity_api.Selector{
						Dataplane: &common_api.LabelSelector{
							MatchLabels: &map[string]string{
								"kuma.io/service": "backend",
							},
						},
					},
					SpiffeID: &meshidentity_api.SpiffeID{
						Path: pointer.To("/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}"),
					},
				},
			}
			err := resManager.Create(context.Background(), mi, core_store.CreateByKey("mi-without-workload-label", "default"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should allow connection even without workload label", func() {
			// given
			dpRes := &core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "backend-01",
					Mesh: "default",
					Labels: map[string]string{
						"kuma.io/service": "backend",
						// No workload label
					},
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "127.0.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port:        8080,
								ServicePort: 8081,
								Tags: map[string]string{
									"kuma.io/service": "backend",
								},
							},
						},
					},
				},
			}

			md := core_xds.DataplaneMetadata{
				Resource:  dpRes,
				ProxyType: mesh_proto.DataplaneProxyType,
			}

			// when
			err := validator.OnProxyConnected(
				core_xds.StreamID(3),
				core_model.ResourceKey{Mesh: "default", Name: "backend-01"},
				context.Background(),
				md,
			)

			// then
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("when no MeshIdentity applies to dataplane", func() {
		It("should allow connection", func() {
			// given - no MeshIdentity created
			dpRes := &core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "other-01",
					Mesh: "default",
					Labels: map[string]string{
						"kuma.io/service": "other",
					},
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "127.0.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port:        8080,
								ServicePort: 8081,
								Tags: map[string]string{
									"kuma.io/service": "other",
								},
							},
						},
					},
				},
			}

			md := core_xds.DataplaneMetadata{
				Resource:  dpRes,
				ProxyType: mesh_proto.DataplaneProxyType,
			}

			// when
			err := validator.OnProxyConnected(
				core_xds.StreamID(4),
				core_model.ResourceKey{Mesh: "default", Name: "other-01"},
				context.Background(),
				md,
			)

			// then
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("when proxy is not a dataplane", func() {
		It("should allow ingress proxy", func() {
			// given
			ingressRes := &core_mesh.ZoneIngressResource{
				Meta: &test_model.ResourceMeta{
					Name: "ingress-01",
					Mesh: core_model.NoMesh,
				},
				Spec: &mesh_proto.ZoneIngress{
					Networking: &mesh_proto.ZoneIngress_Networking{
						Address: "1.1.1.1",
						Port:    10001,
					},
				},
			}

			md := core_xds.DataplaneMetadata{
				Resource:  ingressRes,
				ProxyType: mesh_proto.IngressProxyType,
			}

			// when
			err := validator.OnProxyConnected(
				core_xds.StreamID(5),
				core_model.ResourceKey{Mesh: core_model.NoMesh, Name: "ingress-01"},
				context.Background(),
				md,
			)

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should allow egress proxy", func() {
			// given
			egressRes := &core_mesh.ZoneEgressResource{
				Meta: &test_model.ResourceMeta{
					Name: "egress-01",
					Mesh: core_model.NoMesh,
				},
				Spec: &mesh_proto.ZoneEgress{
					Networking: &mesh_proto.ZoneEgress_Networking{
						Address: "1.1.1.1",
						Port:    10002,
					},
				},
			}

			md := core_xds.DataplaneMetadata{
				Resource:  egressRes,
				ProxyType: mesh_proto.EgressProxyType,
			}

			// when
			err := validator.OnProxyConnected(
				core_xds.StreamID(6),
				core_model.ResourceKey{Mesh: core_model.NoMesh, Name: "egress-01"},
				context.Background(),
				md,
			)

			// then
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("with multiple MeshIdentities", func() {
		BeforeEach(func() {
			// Create less specific MeshIdentity
			mi1 := &meshidentity_api.MeshIdentityResource{
				Meta: &test_model.ResourceMeta{
					Name: "mi-less-specific",
					Mesh: "default",
				},
				Spec: &meshidentity_api.MeshIdentity{
					Selector: &meshidentity_api.Selector{
						Dataplane: &common_api.LabelSelector{
							MatchLabels: &map[string]string{
								"kuma.io/service": "api",
							},
						},
					},
					SpiffeID: &meshidentity_api.SpiffeID{
						Path: pointer.To("/service/{{ label \"kuma.io/service\" }}"),
					},
				},
			}
			err := resManager.Create(context.Background(), mi1, core_store.CreateByKey("mi-less-specific", "default"))
			Expect(err).ToNot(HaveOccurred())

			// Create more specific MeshIdentity with workload label
			mi2 := &meshidentity_api.MeshIdentityResource{
				Meta: &test_model.ResourceMeta{
					Name: "mi-more-specific",
					Mesh: "default",
				},
				Spec: &meshidentity_api.MeshIdentity{
					Selector: &meshidentity_api.Selector{
						Dataplane: &common_api.LabelSelector{
							MatchLabels: &map[string]string{
								"kuma.io/service": "api",
								"version":         "v2",
							},
						},
					},
					SpiffeID: &meshidentity_api.SpiffeID{
						Path: pointer.To("/workload/{{ label \"kuma.io/workload\" }}"),
					},
				},
			}
			err = resManager.Create(context.Background(), mi2, core_store.CreateByKey("mi-more-specific", "default"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should use best match and require workload label", func() {
			// given - dataplane matching more specific MeshIdentity
			dpRes := &core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "api-v2-01",
					Mesh: "default",
					Labels: map[string]string{
						"kuma.io/service": "api",
						"version":         "v2",
						// Missing workload label
					},
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "127.0.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port:        8080,
								ServicePort: 8081,
								Tags: map[string]string{
									"kuma.io/service": "api",
								},
							},
						},
					},
				},
			}

			md := core_xds.DataplaneMetadata{
				Resource:  dpRes,
				ProxyType: mesh_proto.DataplaneProxyType,
			}

			// when
			err := validator.OnProxyConnected(
				core_xds.StreamID(7),
				core_model.ResourceKey{Mesh: "default", Name: "api-v2-01"},
				context.Background(),
				md,
			)

			// then - should fail because best match requires workload label
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("mi-more-specific"))
		})
	})

	Context("edge cases", func() {
		It("should allow connection when MeshIdentity has nil SpiffeID", func() {
			// given
			mi := &meshidentity_api.MeshIdentityResource{
				Meta: &test_model.ResourceMeta{
					Name: "mi-nil-spiffeid",
					Mesh: "default",
				},
				Spec: &meshidentity_api.MeshIdentity{
					Selector: &meshidentity_api.Selector{
						Dataplane: &common_api.LabelSelector{
							MatchLabels: &map[string]string{
								"kuma.io/service": "test",
							},
						},
					},
					SpiffeID: nil,
				},
			}
			err := resManager.Create(context.Background(), mi, core_store.CreateByKey("mi-nil-spiffeid", "default"))
			Expect(err).ToNot(HaveOccurred())

			dpRes := &core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "test-01",
					Mesh: "default",
					Labels: map[string]string{
						"kuma.io/service": "test",
					},
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "127.0.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port:        8080,
								ServicePort: 8081,
								Tags: map[string]string{
									"kuma.io/service": "test",
								},
							},
						},
					},
				},
			}

			md := core_xds.DataplaneMetadata{
				Resource:  dpRes,
				ProxyType: mesh_proto.DataplaneProxyType,
			}

			// when
			err = validator.OnProxyConnected(
				core_xds.StreamID(8),
				core_model.ResourceKey{Mesh: "default", Name: "test-01"},
				context.Background(),
				md,
			)

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should allow connection when MeshIdentity has empty path", func() {
			// given
			mi := &meshidentity_api.MeshIdentityResource{
				Meta: &test_model.ResourceMeta{
					Name: "mi-empty-path",
					Mesh: "default",
				},
				Spec: &meshidentity_api.MeshIdentity{
					Selector: &meshidentity_api.Selector{
						Dataplane: &common_api.LabelSelector{
							MatchLabels: &map[string]string{
								"kuma.io/service": "empty",
							},
						},
					},
					SpiffeID: &meshidentity_api.SpiffeID{
						Path: pointer.To(""),
					},
				},
			}
			err := resManager.Create(context.Background(), mi, core_store.CreateByKey("mi-empty-path", "default"))
			Expect(err).ToNot(HaveOccurred())

			dpRes := &core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "empty-01",
					Mesh: "default",
					Labels: map[string]string{
						"kuma.io/service": "empty",
					},
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "127.0.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port:        8080,
								ServicePort: 8081,
								Tags: map[string]string{
									"kuma.io/service": "empty",
								},
							},
						},
					},
				},
			}

			md := core_xds.DataplaneMetadata{
				Resource:  dpRes,
				ProxyType: mesh_proto.DataplaneProxyType,
			}

			// when
			err = validator.OnProxyConnected(
				core_xds.StreamID(9),
				core_model.ResourceKey{Mesh: "default", Name: "empty-01"},
				context.Background(),
				md,
			)

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle whitespace variations in workload label template", func() {
			// Test with extra whitespace
			mi := &meshidentity_api.MeshIdentityResource{
				Meta: &test_model.ResourceMeta{
					Name: "mi-whitespace",
					Mesh: "default",
				},
				Spec: &meshidentity_api.MeshIdentity{
					Selector: &meshidentity_api.Selector{
						Dataplane: &common_api.LabelSelector{
							MatchLabels: &map[string]string{
								"kuma.io/service": "whitespace-test",
							},
						},
					},
					SpiffeID: &meshidentity_api.SpiffeID{
						Path: pointer.To("/workload/{{  label  \"kuma.io/workload\"  }}"),
					},
				},
			}
			err := resManager.Create(context.Background(), mi, core_store.CreateByKey("mi-whitespace", "default"))
			Expect(err).ToNot(HaveOccurred())

			dpRes := &core_mesh.DataplaneResource{
				Meta: &test_model.ResourceMeta{
					Name: "whitespace-01",
					Mesh: "default",
					Labels: map[string]string{
						"kuma.io/service": "whitespace-test",
						// Missing workload label
					},
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "127.0.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port:        8080,
								ServicePort: 8081,
								Tags: map[string]string{
									"kuma.io/service": "whitespace-test",
								},
							},
						},
					},
				},
			}

			md := core_xds.DataplaneMetadata{
				Resource:  dpRes,
				ProxyType: mesh_proto.DataplaneProxyType,
			}

			// when - should fail with whitespace variations
			err = validator.OnProxyConnected(
				core_xds.StreamID(10),
				core_model.ResourceKey{Mesh: "default", Name: "whitespace-01"},
				context.Background(),
				md,
			)

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("missing required label 'kuma.io/workload'"))
		})
	})
})
