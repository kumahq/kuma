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
	var streamIDCounter core_xds.StreamID

	BeforeEach(func() {
		memStore := memory.NewStore()
		resManager = core_manager.NewResourceManager(memStore)
		validator = NewWorkloadLabelValidator(resManager)
		streamIDCounter = 1

		err := resManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	})

	createMeshIdentity := func(name, mesh string, selector map[string]string, spiffePath *string) {
		var spiffeID *meshidentity_api.SpiffeID
		if spiffePath != nil {
			spiffeID = &meshidentity_api.SpiffeID{Path: spiffePath}
		}

		mi := &meshidentity_api.MeshIdentityResource{
			Meta: &test_model.ResourceMeta{Name: name, Mesh: mesh},
			Spec: &meshidentity_api.MeshIdentity{
				Selector: &meshidentity_api.Selector{
					Dataplane: &common_api.LabelSelector{MatchLabels: &selector},
				},
				SpiffeID: spiffeID,
			},
		}
		err := resManager.Create(context.Background(), mi, core_store.CreateByKey(name, mesh))
		Expect(err).ToNot(HaveOccurred())
	}

	createDataplane := func(name, mesh string, labels map[string]string) *core_mesh.DataplaneResource {
		return &core_mesh.DataplaneResource{
			Meta: &test_model.ResourceMeta{Name: name, Mesh: mesh, Labels: labels},
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "127.0.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:        8080,
							ServicePort: 8081,
							Tags:        map[string]string{"kuma.io/service": labels["kuma.io/service"]},
						},
					},
				},
			},
		}
	}

	createGatewayDataplane := func(name, mesh string, labels map[string]string) *core_mesh.DataplaneResource {
		return &core_mesh.DataplaneResource{
			Meta: &test_model.ResourceMeta{Name: name, Mesh: mesh, Labels: labels},
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "127.0.0.1",
					Gateway: &mesh_proto.Dataplane_Networking_Gateway{
						Type: mesh_proto.Dataplane_Networking_Gateway_BUILTIN,
						Tags: map[string]string{"kuma.io/service": labels["kuma.io/service"]},
					},
				},
			},
		}
	}

	validateConnection := func(resource core_model.Resource, proxyType mesh_proto.ProxyType, mesh, name string) error {
		md := core_xds.DataplaneMetadata{Resource: resource, ProxyType: proxyType}
		err := validator.OnProxyConnected(streamIDCounter, core_model.ResourceKey{Mesh: mesh, Name: name}, context.Background(), md)
		streamIDCounter++
		return err
	}

	type testCase struct {
		meshIdentities   func()
		dataplaneLabels  map[string]string
		dataplaneService string
		isGateway        bool
		expectError      bool
		errorSubstrings  []string
	}

	DescribeTable("dataplane validation scenarios",
		func(tc testCase) {
			if tc.meshIdentities != nil {
				tc.meshIdentities()
			}

			dpLabels := map[string]string{"kuma.io/service": tc.dataplaneService}
			for k, v := range tc.dataplaneLabels {
				dpLabels[k] = v
			}

			var dp *core_mesh.DataplaneResource
			if tc.isGateway {
				dp = createGatewayDataplane("test-dp", "default", dpLabels)
			} else {
				dp = createDataplane("test-dp", "default", dpLabels)
			}
			err := validateConnection(dp, mesh_proto.DataplaneProxyType, "default", "test-dp")

			if tc.expectError {
				Expect(err).To(HaveOccurred())
				for _, substr := range tc.errorSubstrings {
					Expect(err.Error()).To(ContainSubstring(substr))
				}
			} else {
				Expect(err).ToNot(HaveOccurred())
			}
		},
		Entry("should allow connection when dataplane has workload label", testCase{
			meshIdentities: func() {
				createMeshIdentity("mi-with-workload-label", "default",
					map[string]string{"kuma.io/service": "web"},
					pointer.To(`/workload/{{ label "kuma.io/workload" }}`))
			},
			dataplaneService: "web",
			dataplaneLabels:  map[string]string{metadata.KumaWorkload: "my-workload"},
			expectError:      false,
		}),
		Entry("should deny connection when dataplane is missing workload label", testCase{
			meshIdentities: func() {
				createMeshIdentity("mi-with-workload-label", "default",
					map[string]string{"kuma.io/service": "web"},
					pointer.To(`/workload/{{ label "kuma.io/workload" }}`))
			},
			dataplaneService: "web",
			dataplaneLabels:  map[string]string{},
			expectError:      true,
			errorSubstrings:  []string{"missing required label 'kuma.io/workload'", "mi-with-workload-label"},
		}),
		Entry("should allow connection when MeshIdentity does not use workload label", testCase{
			meshIdentities: func() {
				createMeshIdentity("mi-without-workload-label", "default",
					map[string]string{"kuma.io/service": "backend"},
					pointer.To(`/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}`))
			},
			dataplaneService: "backend",
			dataplaneLabels:  map[string]string{},
			expectError:      false,
		}),
		Entry("should allow connection when no MeshIdentity applies", testCase{
			meshIdentities:   nil,
			dataplaneService: "other",
			dataplaneLabels:  map[string]string{},
			expectError:      false,
		}),
		Entry("should allow connection when MeshIdentity has nil SpiffeID", testCase{
			meshIdentities: func() {
				createMeshIdentity("mi-nil-spiffeid", "default",
					map[string]string{"kuma.io/service": "test"}, nil)
			},
			dataplaneService: "test",
			dataplaneLabels:  map[string]string{},
			expectError:      false,
		}),
		Entry("should allow connection when MeshIdentity has empty path", testCase{
			meshIdentities: func() {
				createMeshIdentity("mi-empty-path", "default",
					map[string]string{"kuma.io/service": "empty"},
					pointer.To(""))
			},
			dataplaneService: "empty",
			dataplaneLabels:  map[string]string{},
			expectError:      false,
		}),
		Entry("should handle whitespace variations in workload label template", testCase{
			meshIdentities: func() {
				createMeshIdentity("mi-whitespace", "default",
					map[string]string{"kuma.io/service": "whitespace-test"},
					pointer.To(`/workload/{{  label  "kuma.io/workload"  }}`))
			},
			dataplaneService: "whitespace-test",
			dataplaneLabels:  map[string]string{},
			expectError:      true,
			errorSubstrings:  []string{"missing required label 'kuma.io/workload'"},
		}),
		Entry("should allow gateway dataplane even without workload label", testCase{
			meshIdentities: func() {
				createMeshIdentity("mi-gateway", "default",
					map[string]string{"kuma.io/service": "gateway"},
					pointer.To(`/workload/{{ label "kuma.io/workload" }}`))
			},
			dataplaneService: "gateway",
			dataplaneLabels:  map[string]string{},
			isGateway:        true,
			expectError:      false,
		}),
	)

	Context("with multiple MeshIdentities", func() {
		BeforeEach(func() {
			createMeshIdentity("mi-less-specific", "default",
				map[string]string{"kuma.io/service": "api"},
				pointer.To(`/service/{{ label "kuma.io/service" }}`))

			createMeshIdentity("mi-more-specific", "default",
				map[string]string{"kuma.io/service": "api", "version": "v2"},
				pointer.To(`/workload/{{ label "kuma.io/workload" }}`))
		})

		It("should use best match and require workload label", func() {
			dp := createDataplane("api-v2-01", "default", map[string]string{
				"kuma.io/service": "api",
				"version":         "v2",
			})

			err := validateConnection(dp, mesh_proto.DataplaneProxyType, "default", "api-v2-01")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("mi-more-specific"))
		})
	})

	DescribeTable("non-dataplane proxy types",
		func(proxyType mesh_proto.ProxyType, createResource func(string) core_model.Resource) {
			name := "proxy-01"
			resource := createResource(name)
			err := validateConnection(resource, proxyType, core_model.NoMesh, name)
			Expect(err).ToNot(HaveOccurred())
		},
		Entry("should allow ingress proxy", mesh_proto.IngressProxyType,
			func(name string) core_model.Resource {
				return &core_mesh.ZoneIngressResource{
					Meta: &test_model.ResourceMeta{Name: name, Mesh: core_model.NoMesh},
					Spec: &mesh_proto.ZoneIngress{
						Networking: &mesh_proto.ZoneIngress_Networking{Address: "1.1.1.1", Port: 10001},
					},
				}
			}),
		Entry("should allow egress proxy", mesh_proto.EgressProxyType,
			func(name string) core_model.Resource {
				return &core_mesh.ZoneEgressResource{
					Meta: &test_model.ResourceMeta{Name: name, Mesh: core_model.NoMesh},
					Spec: &mesh_proto.ZoneEgress{
						Networking: &mesh_proto.ZoneEgress_Networking{Address: "1.1.1.1", Port: 10002},
					},
				}
			}),
	)
})
