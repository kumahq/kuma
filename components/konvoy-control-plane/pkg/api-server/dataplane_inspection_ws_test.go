package api_server_test

import (
	"context"
	"fmt"
	"github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/api-server"
	config "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/api-server"
	mesh2 "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"time"
)

var _ = Describe("Dataplane Inspection WS", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	var stop chan struct{}

	BeforeEach(func() {
		resourceStore = memory.NewStore()
		apiServer = createTestApiServer(resourceStore, *config.DefaultApiServerConfig())
		client := resourceApiClient{
			address: apiServer.Address(),
			path:    "/meshes/default/traffic-routes",
		}
		stop = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := apiServer.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()
		waitForServer(&client)
	}, 5)

	AfterEach(func() {
		close(stop)
	})

	BeforeEach(func() {
		// given
		dpResource := mesh2.DataplaneResource{
			Spec: v1alpha1.Dataplane{
				Networking: &v1alpha1.Dataplane_Networking{
					Inbound: []*v1alpha1.Dataplane_Networking_Inbound{
						{
							Interface: "127.0.0.1:9090:9091",
							Tags: map[string]string{
								"service": "sample",
								"version": "v1",
							},
						},
					},
				},
			},
		}
		err := resourceStore.Create(context.Background(), &dpResource, store.CreateByKey("default", "dp1", "mesh1"))
		Expect(err).ToNot(HaveOccurred())

		sampleTime, _ := time.Parse(time.RFC3339, "2019-07-01T00:00:00+00:00")
		insightResource := mesh2.DataplaneInsightResource{
			Spec: v1alpha1.DataplaneInsight{
				Subscriptions: []*v1alpha1.DiscoverySubscription{
					{
						Id:                     "stream-id-1",
						ControlPlaneInstanceId: "cp-1",
						ConnectTime:            proto.MustTimestampProto(sampleTime),
						Status:                 v1alpha1.DiscoverySubscriptionStatus{},
					},
				},
			},
		}
		err = resourceStore.Create(context.Background(), &insightResource, store.CreateByKey("default", "dp1", "mesh1"))
		Expect(err).ToNot(HaveOccurred())
	})

	sampleJson := `
{
	"type": "DataplaneInspection",
	"name": "dp1",
	"mesh": "mesh1",
	"dataplane": {
		"networking": {
			"inbound": [
				{
					"interface": "127.0.0.1:9090:9091",
					"tags": {
						"service": "sample",
						"version": "v1"
					}
				}
			]
		}
	},
	"dataplaneInsight": {
		"subscriptions": [
			{
				"id": "stream-id-1",
				"controlPlaneInstanceId": "cp-1",
				"connectTime": "2019-07-01T00:00:00Z",
				"status": {
					"total": {},
					"cds": {},
					"eds": {},
					"lds": {},
					"rds": {}
				}
			}
		]
	}
}
`

	Describe("On GET", func() {
		It("should return an existing resource", func() {
			// when
			response, err := http.Get("http://" + apiServer.Address() + "/meshes/mesh1/dataplane-inspections/dp1")
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(response.StatusCode).To(Equal(200))
			body, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(MatchJSON(sampleJson))
		})

		type testCase struct {
			url          string
			expectedJson string
		}

		DescribeTable("Listing resources filtering by tag",
			func(tc testCase) {
				// when
				response, err := http.Get("http://" + apiServer.Address() + tc.url)
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(response.StatusCode).To(Equal(200))
				body, err := ioutil.ReadAll(response.Body)
				Expect(err).ToNot(HaveOccurred())

				Expect(string(body)).To(MatchJSON(tc.expectedJson))
			},
			Entry("should list all when no tag is provided", testCase{
				url:          "/meshes/mesh1/dataplane-inspections",
				expectedJson: fmt.Sprintf(`{"items": [%s]}`, sampleJson),
			}),
			Entry("should list with only one matching tag", testCase{
				url:          "/meshes/mesh1/dataplane-inspections?tag=service:sample",
				expectedJson: fmt.Sprintf(`{"items": [%s]}`, sampleJson),
			}),
			Entry("should list all with all matching tags", testCase{
				url:          "/meshes/mesh1/dataplane-inspections?tag=service:sample&tag=version:v1",
				expectedJson: fmt.Sprintf(`{"items": [%s]}`, sampleJson),
			}),
			Entry("should not list when any tag is not matching", testCase{
				url:          "/meshes/mesh1/dataplane-inspections?tag=service:sample&tag=version:v2",
				expectedJson: `{"items": []}`,
			}),
		)
	})
})
