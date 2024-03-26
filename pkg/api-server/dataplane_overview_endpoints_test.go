package api_server_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	api_server "github.com/kumahq/kuma/pkg/api-server"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Dataplane Overview Endpoints", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	stop := func() {}
	t1, _ := time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
	BeforeAll(func() {
		resourceStore = memory.NewStore()
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer().WithStore(store.NewPaginationStore(resourceStore)))
	})

	AfterAll(func() {
		stop()
	})

	createDpWithInsights := func(name string, dp *v1alpha1.Dataplane) {
		dpResource := core_mesh.DataplaneResource{
			Spec: dp,
		}
		err := resourceStore.Create(context.Background(), &dpResource, store.CreateByKey(name, "mesh1"), store.CreatedAt(t1))
		Expect(err).ToNot(HaveOccurred())

		sampleTime, _ := time.Parse(time.RFC3339, "2019-07-01T00:00:00+00:00")
		insightResource := core_mesh.DataplaneInsightResource{
			Spec: &v1alpha1.DataplaneInsight{
				Subscriptions: []*v1alpha1.DiscoverySubscription{
					{
						Id:                     "stream-id-1",
						ControlPlaneInstanceId: "cp-1",
						ConnectTime:            proto.MustTimestampProto(sampleTime),
						Status:                 v1alpha1.NewSubscriptionStatus(sampleTime),
					},
				},
			},
		}
		err = resourceStore.Create(context.Background(), &insightResource, store.CreateByKey(name, "mesh1"))
		Expect(err).ToNot(HaveOccurred())
	}

	BeforeEach(func() {
		memory.ClearStore(resourceStore)
		err := resourceStore.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("mesh1", model.NoMesh), store.CreatedAt(t1))
		Expect(err).ToNot(HaveOccurred())
		// given
		createDpWithInsights("gateway-delegated", &v1alpha1.Dataplane{
			Networking: &v1alpha1.Dataplane_Networking{
				Address: "127.0.0.1",
				Gateway: &v1alpha1.Dataplane_Networking_Gateway{
					Tags: map[string]string{
						"service": "gateway",
					},
				},
			},
		})
		createDpWithInsights("gateway-builtin", &v1alpha1.Dataplane{
			Networking: &v1alpha1.Dataplane_Networking{
				Address: "127.0.0.1",
				Gateway: &v1alpha1.Dataplane_Networking_Gateway{
					Type: v1alpha1.Dataplane_Networking_Gateway_BUILTIN,
					Tags: map[string]string{
						"service": "gateway",
					},
				},
			},
		})

		createDpWithInsights("dp-1", &v1alpha1.Dataplane{
			Networking: &v1alpha1.Dataplane_Networking{
				Address: "127.0.0.1",
				Inbound: []*v1alpha1.Dataplane_Networking_Inbound{
					{
						Port: 1234,
						Tags: map[string]string{
							"service":   "backend",
							"version":   "v1",
							"tagcolumn": "tag:v",
						},
					},
				},
			},
		})

		dpResource := core_mesh.DataplaneResource{
			Spec: &v1alpha1.Dataplane{
				Networking: &v1alpha1.Dataplane_Networking{
					Address: "127.0.0.1",
					Inbound: []*v1alpha1.Dataplane_Networking_Inbound{
						{
							Port: 1234,
							Tags: map[string]string{
								"service":   "backend",
								"version":   "v1",
								"tagcolumn": "tag:v",
							},
						},
					},
				},
			},
		}
		err = resourceStore.Create(context.Background(), &dpResource, store.CreateByKey("dp-no-insights", "mesh1"), store.CreatedAt(t1))
		Expect(err).ToNot(HaveOccurred())
	})

	dp1Json := `
{
	"type": "DataplaneOverview",
	"name": "dp-1",
	"mesh": "mesh1",
	"creationTime": "2018-07-17T16:05:36.995Z",
	"modificationTime": "2018-07-17T16:05:36.995Z",
	"dataplane": {
		"networking": {
			"address": "127.0.0.1",
			"inbound": [
				{
					"port": 1234,
					"tags": {
						"service": "backend",
						"version": "v1",
						"tagcolumn": "tag:v"
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
					"lastUpdateTime": "2019-07-01T00:00:00Z",
					"lds": {},
					"rds": {}
				}
			}
		]
	}
}`

	It("should return an existing resource", func() {
		// when
		response, err := http.Get("http://" + apiServer.Address() + "/meshes/mesh1/dataplanes+insights/dp-1")
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(response.StatusCode).To(Equal(200))
		body, err := io.ReadAll(response.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(body).To(MatchJSON(dp1Json))
	})

	type testCase struct {
		url string
	}

	DescribeTable("Listing resources filtering by tag",
		func(tc testCase) {
			// given
			url := fmt.Sprintf("http://%s/%s", apiServer.Address(), tc.url)
			goldenFileName := fmt.Sprintf(
				"%s.json",
				regexp.MustCompile(`[^a-zA-Z0-9 ]+`).ReplaceAllString(tc.url, "_"),
			)

			// when
			response, err := http.Get(url) // #nosec G107 -- these are just tests
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(response.StatusCode).To(Equal(200))
			body, err := io.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(matchers.MatchGoldenJSON("testdata", goldenFileName))
		},
		Entry("should list all when no tag is provided", testCase{
			url: "meshes/mesh1/dataplanes+insights",
		}),
		Entry("should list with only one matching tag", testCase{
			url: "meshes/mesh1/dataplanes+insights?tag=service:backend",
		}),
		Entry("should list with only subset tag", testCase{
			url: "meshes/mesh1/dataplanes+insights?tag=service:ck",
		}),
		Entry("should list all with all matching tags", testCase{
			url: "meshes/mesh1/dataplanes+insights?tag=service:backend&tag=version:v1",
		}),
		Entry("should list all with all matching tags with value with a column", testCase{
			url: "meshes/mesh1/dataplanes+insights?tag=tagcolumn:tag:v",
		}),
		Entry("should not list when any tag is not matching", testCase{
			url: "meshes/mesh1/dataplanes+insights?tag=service:backend&tag=version:v2",
		}),
		Entry("should list only gateway dataplanes", testCase{
			url: "meshes/mesh1/dataplanes+insights?gateway=true",
		}),
		Entry("should list only gateway builtin", testCase{
			url: "meshes/mesh1/dataplanes+insights?gateway=builtin",
		}),
		Entry("should list only gateway delegated", testCase{
			url: "meshes/mesh1/dataplanes+insights?gateway=delegated",
		}),
		Entry("should list only dataplanes that starts with gateway", testCase{
			url: "meshes/mesh1/dataplanes+insights?name=gateway",
		}),
		Entry("should list only dataplanes that contains with tew", testCase{
			url: "meshes/mesh1/dataplanes+insights?name=tew",
		}),
		Entry("should list only dataplanes that contains with tew using _overview", testCase{
			url: "meshes/mesh1/dataplanes/_overview?name=tew",
		}),
	)

	It("should paginate correctly", func() {
		// when
		response, err := http.Get("http://" + apiServer.Address() + "/meshes/mesh1/dataplanes+insights?size=1")
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(response.StatusCode).To(Equal(200))
		body, err := io.ReadAll(response.Body)
		Expect(err).ToNot(HaveOccurred())
		overviewList := rest.ResourceListReceiver{
			NewResource: func() model.Resource {
				return core_mesh.NewDataplaneOverviewResource()
			},
		}
		Expect(overviewList.UnmarshalJSON(body)).To(Succeed())
		Expect(*overviewList.Next).To(Equal(fmt.Sprintf(`http://%s/meshes/mesh1/dataplanes+insights?offset=1&size=1`, apiServer.Address())))
	})
}, Ordered)
