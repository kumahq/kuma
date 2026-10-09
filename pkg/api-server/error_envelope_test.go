package api_server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

var _ = Describe("Error envelope", Ordered, func() {
	var apiServer *api_server.ApiServer
	stop := func() {}
	BeforeAll(func() {
		resourceStore := memory.NewStore()
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer().WithGlobal().WithStore(store.NewPaginationStore(resourceStore)))
		mesh := core_mesh.NewMeshResource()
		mesh.SetMeta(&test_model.ResourceMeta{Name: "default", Mesh: model.NoMesh})
		Expect(resourceStore.Create(context.Background(), mesh, store.CreateByKey("default", model.NoMesh))).To(Succeed())
	})
	AfterAll(func() {
		stop()
	})

	assertEnvelope := func(path string, expectedStatus int) map[string]any {
		resp, err := http.Get(fmt.Sprintf("http://%s%s", apiServer.Address(), path))
		Expect(err).ToNot(HaveOccurred())
		defer resp.Body.Close()
		Expect(resp).To(HaveHTTPStatus(expectedStatus), path)
		Expect(resp).To(HaveHTTPHeaderWithValue("Content-Type", ContainSubstring("application/json")), path)
		body, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		envelope := map[string]any{}
		Expect(json.Unmarshal(body, &envelope)).To(Succeed(), string(body))
		Expect(envelope).To(HaveKeyWithValue("status", float64(expectedStatus)), path)
		Expect(envelope).To(HaveKey("title"), path)
		Expect(envelope).To(HaveKey("type"), path)
		Expect(envelope).To(HaveKey("detail"), path)
		return envelope
	}

	It("returns the error envelope for unknown resource types", func() {
		assertEnvelope("/meshes/default/not-a-type", http.StatusNotFound)
	})

	It("returns the error envelope for unknown paths", func() {
		assertEnvelope("/no-such-endpoint", http.StatusNotFound)
	})

	It("returns 404 for a KRI with extra segments", func() {
		envelope := assertEnvelope("/_kri/kri_mal_default___ma-1_extra_more", http.StatusNotFound)
		Expect(envelope).ToNot(HaveKey("instance"))
	})

	It("returns 400 for a malformed KRI", func() {
		assertEnvelope("/_kri/garbage", http.StatusBadRequest)
	})

	It("still resolves a valid KRI", func() {
		resp, err := http.Get(fmt.Sprintf("http://%s/_kri/kri_m____default_", apiServer.Address()))
		Expect(err).ToNot(HaveOccurred())
		defer resp.Body.Close()
		Expect(resp).To(HaveHTTPStatus(http.StatusOK))
	})
})
