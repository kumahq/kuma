package ingress_test

import (
	"context"
	"fmt"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	model2 "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/ingress"
)

type fakeResourceManager struct {
	manager.ResourceManager
	updCounter int
}

func (f *fakeResourceManager) Update(context.Context, model.Resource, ...store.UpdateOptionsFunc) error {
	f.updCounter++
	return nil
}

var _ = Describe("Ingress Dataplane", func() {

	type testCase struct {
		dataplanes map[string][]string
		expected   string
	}
	DescribeTable("should generate ingress based on other dataplanes",
		func(given testCase) {
			dataplanes := []*core_mesh.DataplaneResource{}

			for mesh, dps := range given.dataplanes {
				for i, dp := range dps {
					dpRes := core_mesh.NewDataplaneResource()
					err := util_proto.FromYAML([]byte(dp), dpRes.Spec)
					Expect(err).ToNot(HaveOccurred())
					dpRes.SetMeta(&model2.ResourceMeta{Name: fmt.Sprintf("dp-%d", i), Mesh: mesh})
					dataplanes = append(dataplanes, dpRes)
				}
			}

			actual := ingress.GetIngressAvailableServices(dataplanes)
			actualYAML, err := yaml.Marshal(actual)
			Expect(err).ToNot(HaveOccurred())
			Expect(actualYAML).To(MatchYAML(given.expected))
		},
		Entry("base", testCase{
			dataplanes: map[string][]string{
				"default": {
					`
                networking:
                  inbound:
                    - address: 127.0.0.1
                      port: 1010
                      servicePort: 2020
                      tags:
                        service: backend
                        version: "1"
                        region: eu
`,
					`
                networking:
                  inbound:
                    - address: 127.0.0.1
                      port: 1010
                      servicePort: 2020
                      tags:
                        service: backend
                        version: "2"
                        region: us
`,
					`
                networking:
                  inbound:
                    - address: 127.0.0.1
                      port: 1010
                      servicePort: 2020
                      tags:
                        service: backend
                        version: "2"
                        region: us
`,
				},
			},
			expected: `
            - instances: 1
              mesh: default
              tags:
                service: backend
                region: eu
                version: "1"
            - instances: 2
              mesh: default
              tags:
                service: backend
                region: us
                version: "2"
`,
		}),
		Entry("multi-mesh", testCase{
			dataplanes: map[string][]string{
				"mesh1": {
					`
            networking:
              inbound:
                - address: 127.0.0.1
                  port: 1010
                  servicePort: 2020
                  tags:
                    service: b1
`,
					`
            networking:
              inbound:
                - address: 127.0.0.1
                  port: 1010
                  servicePort: 2020
                  tags:
                    service: b2
`,
				},
				"mesh2": {
					`
            networking:
              inbound:
                - address: 127.0.0.1
                  port: 1010
                  servicePort: 2020
                  tags:
                    service: b1
`,
				},
			},
			expected: `
            - instances: 1
              mesh: mesh1
              tags:
                service: b1
            - instances: 1
              mesh: mesh1
              tags:
                service: b2
            - instances: 1
              mesh: mesh2
              tags:
                service: b1
`,
		}),
	)

	It("should not update store if ingress haven't changed", func() {
		ctx := context.Background()
		mgr := &fakeResourceManager{}

		ing := &core_mesh.ZoneIngressResource{
			Spec: &mesh_proto.ZoneIngress{
				Networking: &mesh_proto.ZoneIngress_Networking{
					Port: 10001,
				},
				AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
					{
						Instances: 1,
						Tags: map[string]string{
							"service": "backend",
							"version": "v1",
							"region":  "eu",
						},
						Mesh: "mesh1",
					},
					{
						Instances: 2,
						Tags: map[string]string{
							"service": "web",
							"version": "v2",
							"region":  "us",
						},
						Mesh: "mesh1",
					},
				},
			},
		}

		others := []*core_mesh.DataplaneResource{
			{
				Meta: &model2.ResourceMeta{Mesh: "mesh1"},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags: map[string]string{
									"service": "backend",
									"version": "v1",
									"region":  "eu",
								},
							},
						},
					},
				},
			},
			{
				Meta: &model2.ResourceMeta{Mesh: "mesh1"},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags: map[string]string{
									"service": "web",
									"version": "v2",
									"region":  "us",
								},
							},
						},
					},
				},
			},
			{
				Meta: &model2.ResourceMeta{Mesh: "mesh1"},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags: map[string]string{
									"service": "web",
									"version": "v2",
									"region":  "us",
								},
							},
						},
					},
				},
			},
		}
		err := ingress.UpdateAvailableServices(ctx, mgr, ing, others)
		Expect(err).ToNot(HaveOccurred())
		Expect(mgr.updCounter).To(Equal(0))
	})

	It("should generate available services for multiple meshes", func() {
		dataplanes := []*core_mesh.DataplaneResource{
			{
				Meta: &model2.ResourceMeta{Mesh: "mesh1"},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags: map[string]string{
									"service": "backend",
									"version": "v1",
									"region":  "eu",
								},
							},
						},
					},
				},
			},
			{
				Meta: &model2.ResourceMeta{Mesh: "mesh2"},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags: map[string]string{
									"service": "web",
									"version": "v2",
									"region":  "us",
								},
							},
						},
					},
				},
			},
			{
				Meta: &model2.ResourceMeta{Mesh: "mesh2"},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags: map[string]string{
									"service": "web",
									"version": "v1",
									"region":  "eu",
								},
							},
						},
					},
				},
			},
		}
		expectedAvailableServices := []*mesh_proto.ZoneIngress_AvailableService{
			{
				Instances: 1,
				Tags: map[string]string{
					"service": "backend",
					"version": "v1",
					"region":  "eu",
				},
				Mesh: "mesh1",
			},
			{
				Instances: 1,
				Tags: map[string]string{
					"service": "web",
					"version": "v1",
					"region":  "eu",
				},
				Mesh: "mesh2",
			},
			{
				Instances: 1,
				Tags: map[string]string{
					"service": "web",
					"version": "v2",
					"region":  "us",
				},
				Mesh: "mesh2",
			},
		}

		actual := ingress.GetIngressAvailableServices(dataplanes)
		Expect(actual).To(Equal(expectedAvailableServices))
	})

	It("should generate available services for multiple meshes with the same tags", func() {
		dataplanes := []*core_mesh.DataplaneResource{
			{
				Meta: &model2.ResourceMeta{Mesh: "mesh1"},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags: map[string]string{
									"service": "backend",
									"version": "v1",
									"region":  "eu",
								},
							},
						},
					},
				},
			},
			{
				Meta: &model2.ResourceMeta{Mesh: "mesh2"},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags: map[string]string{
									"service": "backend",
									"version": "v1",
									"region":  "eu",
								},
							},
						},
					},
				},
			},
		}
		expectedAvailableServices := []*mesh_proto.ZoneIngress_AvailableService{
			{
				Instances: 1,
				Tags: map[string]string{
					"service": "backend",
					"version": "v1",
					"region":  "eu",
				},
				Mesh: "mesh1",
			},
			{
				Instances: 1,
				Tags: map[string]string{
					"service": "backend",
					"version": "v1",
					"region":  "eu",
				},
				Mesh: "mesh2",
			},
		}

		actual := ingress.GetIngressAvailableServices(dataplanes)
		Expect(actual).To(Equal(expectedAvailableServices))
	})
})
