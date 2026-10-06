package api_server_test

import (
	"context"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/v3/pkg/api-server"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
)

var _ = Describe("Cross-mesh list endpoints", Ordered, func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	stop := func() {}
	BeforeAll(func() {
		resourceStore = memory.NewStore()
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer().WithStore(store.NewPaginationStore(resourceStore)))
	})
	AfterAll(func() {
		stop()
	})

	It("returns 404 for cross-mesh lists of mesh-scoped resources", func() {
		mesh := core_mesh.NewMeshResource()
		mesh.SetMeta(&test_model.ResourceMeta{Name: "default", Mesh: model.NoMesh})
		Expect(resourceStore.Create(context.Background(), mesh, store.CreateByKey("default", model.NoMesh))).To(Succeed())

		cases := []struct {
			path   string
			status int
		}{
			{path: "/meshaccesslogs", status: http.StatusNotFound},
			{path: "/dataplanes", status: http.StatusNotFound},
			{path: "/meshservices", status: http.StatusNotFound},
			{path: "/secrets", status: http.StatusNotFound},
			{path: "/dataplane-insights", status: http.StatusNotFound},
			{path: "/dataplane-insights/_overview", status: http.StatusNotFound},
			{path: "/meshes/default/meshaccesslogs", status: http.StatusOK},
			{path: "/meshes/default/dataplanes", status: http.StatusOK},
			{path: "/meshes/default/secrets", status: http.StatusOK},
		}
		for _, tc := range cases {
			resp, err := http.Get(fmt.Sprintf("http://%s%s", apiServer.Address(), tc.path))
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(HaveHTTPStatus(tc.status), tc.path)
		}
	})
})
