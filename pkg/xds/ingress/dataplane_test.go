package ingress_test

import (
	"context"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"

	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	"github.com/Kong/kuma/pkg/xds/ingress"
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
		dataplanes []string
		expected   string
	}
	DescribeTable("should generate ingress based on other dataplanes",
		func(given testCase) {
			dataplanes := []*core_mesh.DataplaneResource{}

			for _, dp := range given.dataplanes {
				dpRes := &core_mesh.DataplaneResource{}
				err := util_proto.FromYAML([]byte(dp), &dpRes.Spec)
				Expect(err).ToNot(HaveOccurred())
				dataplanes = append(dataplanes, dpRes)
			}

			actual := ingress.GetIngressAvailableServices(dataplanes)
			actualYAML, err := yaml.Marshal(actual)
			Expect(err).ToNot(HaveOccurred())
			Expect(actualYAML).To(MatchYAML(given.expected))
		},
		Entry("base", testCase{
			dataplanes: []string{
				`
            type: Dataplane
            name: dp-1
            mesh: default
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
            type: Dataplane
            name: dp-2
            mesh: default
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
            type: Dataplane
            name: dp-3
            mesh: default
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
			expected: `
            - instances: 1
              tags:
                service: backend
                region: eu
                version: "1"
            - instances: 2
              tags:
                service: backend
                region: us
                version: "2"
`,
		}))

	It("should not update store if ingress haven't changed", func() {
		ctx := context.Background()
		mgr := &fakeResourceManager{}

		ing := &core_mesh.DataplaneResource{
			Spec: mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Ingress: &mesh_proto.Dataplane_Networking_Ingress{
						AvailableServices: []*mesh_proto.Dataplane_Networking_Ingress_AvailableService{
							{
								Instances: 1,
								Tags: map[string]string{
									"service": "backend",
									"version": "v1",
									"region":  "eu",
								},
							},
							{
								Instances: 2,
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

		others := []*core_mesh.DataplaneResource{
			{
				Spec: mesh_proto.Dataplane{
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
				Spec: mesh_proto.Dataplane{
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
				Spec: mesh_proto.Dataplane{
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
})
