package dataplane_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/managers/apis/dataplane"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
)

var _ = Describe("WorkloadLabelValidator", func() {
	type testCase struct {
		workloadLabel string
		expectError   bool
		errorContains string
	}

	DescribeTable("should validate workload label",
		func(given testCase) {
			// given
			validator := dataplane.NewWorkloadLabelValidator()
			dp := core_mesh.NewDataplaneResource()

			labels := map[string]string{}
			if given.workloadLabel != "" {
				labels[metadata.KumaWorkload] = given.workloadLabel
			}

			dp.Meta = &test_model.ResourceMeta{
				Mesh:   model.DefaultMesh,
				Name:   "test-dp",
				Labels: labels,
			}

			Expect(dp.SetSpec(&mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "192.168.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port: 8080,
							Tags: map[string]string{
								"kuma.io/service": "test-service",
							},
						},
					},
				},
			})).To(Succeed())

			mesh := core_mesh.NewMeshResource()

			// when
			err := validator.ValidateCreate(context.Background(), model.ResourceKey{Mesh: model.DefaultMesh, Name: "test-dp"}, dp, mesh)

			// then
			if given.expectError {
				Expect(err).To(HaveOccurred())
				if given.errorContains != "" {
					Expect(err.Error()).To(ContainSubstring(given.errorContains))
				}
			} else {
				Expect(err).ToNot(HaveOccurred())
			}
		},
		Entry("valid workload label", testCase{
			workloadLabel: "my-workload",
			expectError:   false,
		}),
		Entry("no workload label", testCase{
			workloadLabel: "",
			expectError:   false,
		}),
		Entry("invalid workload label with uppercase", testCase{
			workloadLabel: "MyWorkload",
			expectError:   true,
			errorContains: "must be a valid DNS-1035 label",
		}),
		Entry("invalid workload label with special characters", testCase{
			workloadLabel: "my_workload",
			expectError:   true,
			errorContains: "must be a valid DNS-1035 label",
		}),
	)
})
