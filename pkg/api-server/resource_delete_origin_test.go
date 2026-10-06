package api_server_test

import (
	"context"
	"fmt"
	"io"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
)

var _ = Describe("Resource Endpoints, origin label on delete", func() {
	type testCase struct {
		configurer   func() *testApiServerConfigurer
		storedOrigin mesh_proto.ResourceOrigin
		status       int
		reason       string
	}

	federatedZone := func() *testApiServerConfigurer { return NewTestApiServerConfigurer().WithZone("zone-1") }
	global := func() *testApiServerConfigurer { return NewTestApiServerConfigurer().WithGlobal() }

	DescribeTable("should check the origin of the stored resource",
		func(given testCase) {
			// given
			ctx := context.Background()
			store := core_store.NewPaginationStore(memory.NewStore())
			apiServer, _, stop := StartApiServer(given.configurer().WithStore(store))
			defer stop()

			Expect(store.Create(ctx, core_mesh.NewMeshResource(), core_store.CreateByKey("mesh-1", model.NoMesh))).To(Succeed())
			mtp := v1alpha1.NewMeshTrafficPermissionResource()
			mtp.Spec = builders.MeshTrafficPermission().
				WithTargetRef(builders.TargetRefMesh()).
				AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
				Build().Spec
			labels := map[string]string{}
			if given.storedOrigin != "" {
				labels[mesh_proto.ResourceOriginLabel] = string(given.storedOrigin)
			}
			Expect(store.Create(ctx, mtp, core_store.CreateByKey("mtp-1", "mesh-1"), core_store.CreateWithLabels(labels))).To(Succeed())

			// when
			req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf(
				"http://%s/meshes/mesh-1/%s/mtp-1", apiServer.Address(), v1alpha1.MeshTrafficPermissionResourceTypeDescriptor.WsPath,
			), http.NoBody)
			Expect(err).ToNot(HaveOccurred())
			resp, err := http.DefaultClient.Do(req)
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(resp.StatusCode).To(Equal(given.status), string(body))
			err = store.Get(ctx, v1alpha1.NewMeshTrafficPermissionResource(), core_store.GetByKey("mtp-1", "mesh-1"))
			if given.reason == "" {
				Expect(core_store.IsResourceNotFound(err)).To(BeTrue())
			} else {
				Expect(string(body)).To(ContainSubstring(given.reason))
				Expect(err).ToNot(HaveOccurred())
			}
		},
		Entry("federated zone rejects deleting a Global-synced copy", testCase{
			configurer:   federatedZone,
			storedOrigin: mesh_proto.GlobalResourceOrigin,
			status:       http.StatusBadRequest,
			reason:       "the origin label must be set to 'zone'",
		}),
		Entry("federated zone deletes its own resource", testCase{
			configurer:   federatedZone,
			storedOrigin: mesh_proto.ZoneResourceOrigin,
			status:       http.StatusOK,
		}),
		Entry("global rejects deleting a zone-synced copy", testCase{
			configurer:   global,
			storedOrigin: mesh_proto.ZoneResourceOrigin,
			status:       http.StatusBadRequest,
			reason:       "the origin label must be set to 'global'",
		}),
		Entry("global deletes its own resource", testCase{
			configurer:   global,
			storedOrigin: mesh_proto.GlobalResourceOrigin,
			status:       http.StatusOK,
		}),
		Entry("global deletes a resource without an origin label", testCase{
			configurer: global,
			status:     http.StatusOK,
		}),
		Entry("disabled origin label validation still allows the delete", testCase{
			configurer: func() *testApiServerConfigurer {
				return federatedZone().WithDisableOriginLabelValidation(true)
			},
			storedOrigin: mesh_proto.GlobalResourceOrigin,
			status:       http.StatusOK,
		}),
	)
})
