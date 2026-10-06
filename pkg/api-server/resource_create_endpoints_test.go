package api_server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config "github.com/kumahq/kuma/v3/pkg/config/api-server"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
)

var _ = Describe("Create-only resource endpoints", func() {
	request := func(address, method, path, body string) (int, http.Header, []byte) {
		GinkgoHelper()
		req, err := http.NewRequestWithContext(context.Background(), method, "http://"+address+path, bytes.NewBufferString(body))
		Expect(err).NotTo(HaveOccurred())
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())
		return resp.StatusCode, resp.Header, data
	}

	DescribeTable("creates once and preserves the original on conflict", func(collection, body, replacement string) {
		resourceStore := memory.NewStore()
		Expect(resourceStore.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("default", model.NoMesh))).To(Succeed())
		api, _, stop := StartApiServer(NewTestApiServerConfigurer().WithStore(resourceStore).WithConfigMutator(func(c *config.ApiServerConfig) { c.BasePath = "/api" }))
		defer stop()
		collection = "/api" + collection
		status, headers, data := request(api.Address(), http.MethodPost, collection, body)
		Expect(status).To(Equal(http.StatusCreated), string(data))
		Expect(headers.Get("Location")).To(Equal(collection + "/test"))
		var warnings map[string]any
		Expect(json.Unmarshal(data, &warnings)).To(Succeed())
		status, _, before := request(api.Address(), http.MethodGet, collection+"/test", "")
		Expect(status).To(Equal(http.StatusOK))
		status, headers, data = request(api.Address(), http.MethodPost, collection, replacement)
		Expect(status).To(Equal(http.StatusConflict), string(data))
		Expect(headers.Get("Location")).To(BeEmpty())
		var conflict map[string]any
		Expect(json.Unmarshal(data, &conflict)).To(Succeed())
		Expect(conflict["status"]).To(BeNumerically("==", 409))
		Expect(conflict["detail"]).To(ContainSubstring("test"))
		status, _, after := request(api.Address(), http.MethodGet, collection+"/test", "")
		Expect(status).To(Equal(http.StatusOK))
		Expect(after).To(MatchJSON(before))
		status, _, data = request(api.Address(), http.MethodPut, collection+"/test", replacement)
		Expect(status).To(Equal(http.StatusOK), string(data))
		status, _, data = request(api.Address(), http.MethodDelete, collection+"/test", "")
		Expect(status).To(Equal(http.StatusOK), string(data))
		status, _, data = request(api.Address(), http.MethodPut, collection+"/test", body)
		Expect(status).To(Equal(http.StatusCreated), string(data))
	},
		Entry("global Mesh", "/meshes", `{"type":"Mesh","name":"test","labels":{"test":"original"}}`, `{"type":"Mesh","name":"test","labels":{"test":"replacement"}}`),
		Entry("mesh-scoped secret", "/meshes/default/secrets", `{"type":"Secret","mesh":"default","name":"test","data":"b3JpZ2luYWw="}`, `{"type":"Secret","mesh":"default","name":"test","data":"cmVwbGFjZW1lbnQ="}`),
		Entry("global secret alias", "/global-secrets", `{"type":"GlobalSecret","name":"test","data":"b3JpZ2luYWw="}`, `{"type":"GlobalSecret","name":"test","data":"cmVwbGFjZW1lbnQ="}`),
	)

	DescribeTable("rejects invalid creation", func(collection, body string, expected int) {
		resourceStore := memory.NewStore()
		Expect(resourceStore.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("default", model.NoMesh))).To(Succeed())
		api, _, stop := StartApiServer(NewTestApiServerConfigurer().WithStore(resourceStore))
		defer stop()
		status, headers, data := request(api.Address(), http.MethodPost, collection, body)
		Expect(status).To(Equal(expected), string(data))
		Expect(headers.Get("Location")).To(BeEmpty())
	},
		Entry("malformed JSON", "/meshes", `{`, 400),
		Entry("missing name", "/meshes", `{"type":"Mesh"}`, 400),
		Entry("wrong type", "/meshes", `{"type":"Zone","name":"test"}`, 400),
		Entry("wrong mesh", "/meshes/default/secrets", `{"type":"Secret","mesh":"other","name":"test","data":"dGVzdA=="}`, 400),
		Entry("missing parent mesh", "/meshes/missing/secrets", `{"type":"Secret","mesh":"missing","name":"test","data":"dGVzdA=="}`, 404),
	)

	It("returns deprecation warnings when creating a resource", func() {
		resourceStore := memory.NewStore()
		Expect(resourceStore.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("default", model.NoMesh))).To(Succeed())
		api, _, stop := StartApiServer(NewTestApiServerConfigurer().WithStore(resourceStore))
		defer stop()
		status, _, data := request(api.Address(), http.MethodPost, "/meshes/default/meshratelimits",
			`{"type":"MeshRateLimit","name":"warning","mesh":"default","spec":{"targetRef":{"kind":"Dataplane"},"rules":[{"default":{"local":{"http":{"onRateLimit":{"status":123}}}}}]}}`)
		Expect(status).To(Equal(http.StatusCreated), string(data))
		var result struct {
			Warnings []string `json:"warnings"`
		}
		Expect(json.Unmarshal(data, &result)).To(Succeed())
		Expect(result.Warnings).To(ContainElement(ContainSubstring("must be 400 or higher")))
	})

	It("enforces ownership labels on creation", func() {
		resourceStore := memory.NewStore()
		Expect(resourceStore.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("default", model.NoMesh))).To(Succeed())
		api, _, stop := StartApiServer(NewTestApiServerConfigurer().WithStore(resourceStore).WithZone("zone-1"))
		defer stop()
		status, _, data := request(api.Address(), http.MethodPost, "/meshes/default/meshtrafficpermissions",
			`{"type":"MeshTrafficPermission","name":"test","mesh":"default","labels":{"kuma.io/origin":"global"},"spec":{"targetRef":{"kind":"Mesh"},"rules":[{"default":{"allow":[{"spiffeID":{"type":"Prefix","value":"spiffe://example"}}]}}]}}`)
		Expect(status).To(Equal(http.StatusBadRequest), string(data))
		Expect(string(data)).To(ContainSubstring("origin"))
	})

	It("allows only one concurrent create and preserves the winner", func() {
		api, _, stop := StartApiServer(NewTestApiServerConfigurer())
		defer stop()
		const count = 8
		statuses := make([]int, count)
		bodies := make([][]byte, count)
		start := make(chan struct{})
		var wg sync.WaitGroup
		for i := range count {
			wg.Go(func() {
				defer GinkgoRecover()
				<-start
				statuses[i], _, bodies[i] = request(api.Address(), http.MethodPost, "/meshes", fmt.Sprintf(`{"type":"Mesh","name":"race","labels":{"writer":"%d"}}`, i))
			})
		}
		close(start)
		wg.Wait()
		winner := -1
		for i, status := range statuses {
			if status == http.StatusCreated {
				Expect(winner).To(Equal(-1))
				winner = i
			} else {
				Expect(status).To(Equal(http.StatusConflict), string(bodies[i]))
			}
		}
		Expect(winner).To(BeNumerically(">=", 0))
		status, _, data := request(api.Address(), http.MethodGet, "/meshes/race", "")
		Expect(status).To(Equal(http.StatusOK))
		var result struct {
			Labels map[string]string `json:"labels"`
		}
		Expect(json.Unmarshal(data, &result)).To(Succeed())
		Expect(result.Labels["writer"]).To(Equal(fmt.Sprint(winner)))
	})
})
