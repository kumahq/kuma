package api_server_test

import (
	"context"
	"io"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	config "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("Global Insights Endpoints", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
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

		client := resourceApiClient{
			address: apiServer.Address(),
			path:    "/global-insights",
		}

		stop = make(chan struct{})

		go func() {
			defer GinkgoRecover()
			Expect(apiServer.Start(stop)).To(Succeed())
		}()

		waitForServer(&client)
	})

	AfterEach(func() {
		close(stop)
		core.Now = time.Now
	})

	BeforeEach(func() {
		resources := map[string]core_model.Resource{
			"zone-1":         system.NewZoneResource(),
			"zone-2":         system.NewZoneResource(),
			"zone-ingress-1": core_mesh.NewZoneIngressResource(),
			"zone-egress-1":  core_mesh.NewZoneEgressResource(),
			"zone-egress-2":  core_mesh.NewZoneEgressResource(),
			"mesh-1":         core_mesh.NewMeshResource(),
			"mesh-2":         core_mesh.NewMeshResource(),
			"mesh-3":         core_mesh.NewMeshResource(),
		}

		for name, resource := range resources {
			Expect(resourceStore.Create(
				context.Background(),
				resource,
				store.CreateByKey(name, core_model.NoMesh),
			)).To(Succeed())
		}

	})

	globalInsightsJSON := `
{
  "type": "GlobalInsights",
  "creationTime": "2018-07-17T16:05:36.995Z",
  "resources": {
    "GlobalSecret": {
      "total": 0
    },
    "Mesh": {
      "total": 3
    },
    "Zone": {
      "total": 2
    },
    "ZoneEgress": {
      "total": 2
    },
    "ZoneIngress": {
      "total": 1
    }
  }
}
`

	Describe("On GET", func() {
		It("should return an existing resource", func() {
			// when
			response, err := http.Get("http://" + apiServer.Address() + "/global-insights")
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(response.StatusCode).To(Equal(200))
			body, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(MatchJSON(globalInsightsJSON))
		})
	})
})
