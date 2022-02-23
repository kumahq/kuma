package api_server_test

import (
	"context"
	"io"
	"time"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	api_server "github.com/kumahq/kuma/pkg/api-server"
	config "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("GlobalSecret Endpoints", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	var client resourceApiClient
	var stop chan struct{}

	BeforeEach(func() {
		core.Now = func() time.Time {
			now, _ := time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
			return now
		}
		resourceStore = memory.NewStore()
		metrics, err := metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())
		apiServer = createTestApiServer(resourceStore, config.DefaultApiServerConfig(), true, metrics)
		client = resourceApiClient{
			apiServer.Address(),
			"/global-secrets",
		}
		stop = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := apiServer.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()
		waitForServer(&client)
	})

	AfterEach(func() {
		close(stop)
		core.Now = time.Now
	})

	BeforeEach(func() {
		// when
		err := resourceStore.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey(model.DefaultMesh, model.NoMesh))
		// then
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("PUT => GET", func() {

		given := `
        type: GlobalSecret
        name: sec-1
        creationTime: "2018-07-17T16:05:36.995Z"
        modificationTime: "2018-07-17T16:05:36.995Z"
        data: "dGVzdAo="
`
		It("GET should return data saved by PUT", func() {
			// given
			resource := rest.Resource{
				Spec: &system_proto.Secret{},
			}

			// when
			err := yaml.Unmarshal([]byte(given), &resource)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			response := client.put(resource)
			// then
			Expect(response.StatusCode).To(Equal(201))

			// when
			response = client.get("sec-1")
			// then
			Expect(response.StatusCode).To(Equal(200))
			// when
			body, err := io.ReadAll(response.Body)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := yaml.JSONToYAML(body)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given))
		})
	})
})
