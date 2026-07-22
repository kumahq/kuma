package api_server_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/emicklei/go-restful/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	api_common "github.com/kumahq/kuma/v3/api/openapi/types/common"
	api_server "github.com/kumahq/kuma/v3/pkg/api-server"
	"github.com/kumahq/kuma/v3/pkg/core"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v3/pkg/test/matchers"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	samples2 "github.com/kumahq/kuma/v3/pkg/test/resources/samples"
)

var _ = Describe("Inspect WS", func() {
	type testCase struct {
		path        string
		matcher     types.GomegaMatcher
		resources   []core_model.Resource
		global      bool
		contentType string
		query       string
	}
	AfterEach(func() {
		core.Now = time.Now
	})

	DescribeTable(
		"should return valid response",
		func(given testCase) {
			// setup
			core.Now = func() time.Time { return time.Time{} }

			resourceStore := memory.NewStore()
			rm := manager.NewResourceManager(resourceStore)
			for _, resource := range given.resources {
				err := rm.Create(context.Background(), resource,
					store.CreateBy(core_model.MetaToResourceKey(resource.GetMeta())))
				Expect(err).ToNot(HaveOccurred())
			}

			var apiServer *api_server.ApiServer
			var stop func()
			conf := NewTestApiServerConfigurer().WithStore(resourceStore)
			if given.global {
				conf = conf.WithGlobal()
			} else {
				conf = conf.WithZone("local")
			}
			apiServer, _, stop = StartApiServer(conf)
			defer stop()

			// when
			resp, err := http.Get((&url.URL{
				Scheme:   "http",
				Host:     apiServer.Address(),
				Path:     given.path,
				RawQuery: given.query,
			}).String())
			Expect(err).ToNot(HaveOccurred())

			// then
			bytes, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(given.matcher)

			Expect(resp.Header.Get("content-type")).To(Equal(given.contentType))
		},
		Entry("inspect dataplane", testCase{
			path:    "/meshes/default/dataplanes/backend-1/policies",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_dataplane.json")),
			resources: []core_model.Resource{
				builders.Mesh().Build(),
				builders.Dataplane().
					WithName("backend-1").
					WithHttpServices("backend").
					AddOutboundsToServices("redis", "elastic", "postgres", "web").
					Build(),
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefMesh()).
					AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
					Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect dataplane, empty response", testCase{
			path:    "/meshes/default/dataplanes/backend-1/policies",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_dataplane_empty-response.json")),
			resources: []core_model.Resource{
				builders.Mesh().Build(),
				builders.Dataplane().
					WithName("backend-1").
					WithServices("backend").
					AddOutboundsToServices("redis", "elastic", "postgres", "web").
					Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect meshtrafficpermission", testCase{
			path:    "/meshes/mesh-1/meshtrafficpermissions/mtp-1/dataplanes",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_meshtrafficpermission.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				builders.Dataplane().WithName("backend-1").WithMesh("mesh-1").WithServices("backend").Build(),
				builders.MeshTrafficPermission().
					WithMesh("mesh-1").
					WithTargetRef(builders.TargetRefDataplaneName("backend-1")).
					AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
					Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect xds for dataplane", testCase{
			path:    "/meshes/mesh-1/dataplanes/backend-1/xds",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_xds_dataplane.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				builders.Dataplane().WithName("backend-1").WithAddress("1.1.1.1").WithMesh("mesh-1").WithAdminPort(3301).WithServices("backend").AddOutboundsToServices("redis", "elastic", "web").Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect xds for dataplane with eds", testCase{
			path:    "/meshes/mesh-1/dataplanes/backend-1/xds",
			query:   "include_eds=true",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_xds_dataplane_with_eds.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				builders.Dataplane().WithName("backend-1").WithAddress("1.1.1.1").WithMesh("mesh-1").WithAdminPort(3301).WithServices("backend").AddOutboundsToServices("redis", "elastic", "web").Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect xds for local zone ingress", testCase{
			path:    "/zoneingresses/zi-1/xds",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_xds_local_zoneingress.json")),
			resources: []core_model.Resource{
				builders.ZoneIngress().
					WithName("zi-1").
					WithZone("").
					WithAdminPort(2201).
					WithAddress("2.2.2.2").
					WithPort(8080).
					WithAdvertisedAddress("3.3.3.3").
					WithAdvertisedPort(80).
					Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect xds for zone ingress from another zone", testCase{
			path:    "/zoneingresses/zi-1/xds",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_xds_remote_zoneingress.json")),
			resources: []core_model.Resource{
				builders.ZoneIngress().
					WithName("zi-1").
					WithZone("not-local-zone").
					WithAdminPort(2201).
					WithAddress("2.2.2.2").
					WithPort(8080).
					WithAdvertisedAddress("3.3.3.3").
					WithAdvertisedPort(80).
					Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect xds for zone ingress on global", testCase{
			global:  true,
			path:    "/zoneingresses/zi-1/xds",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_xds_local_zoneingress.json")),
			resources: []core_model.Resource{
				builders.ZoneIngress().
					WithName("zi-1").
					WithZone(""). // local zone ingress has empty "zone" field
					WithAdminPort(2201).
					WithAddress("2.2.2.2").
					WithPort(8080).
					WithAdvertisedAddress("3.3.3.3").
					WithAdvertisedPort(80).
					Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect xds for dataplane on global", testCase{
			global:  true,
			path:    "/meshes/mesh-1/dataplanes/backend-1/xds",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_xds_dataplane.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				builders.Dataplane().WithName("backend-1").WithMesh("mesh-1").WithAdminPort(3301).WithAddress("1.1.1.1").WithServices("backend").Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect xds for zone egress", testCase{
			path:    "/zoneegresses/ze-1/xds",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_xds_zoneegress.json")),
			resources: []core_model.Resource{
				builders.ZoneEgress().WithName("ze-1").WithAddress("4.4.4.4").WithPort(8080).WithAdminPort(4321).Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect stats for dataplane", testCase{
			path:    "/meshes/mesh-1/dataplanes/backend-1/stats",
			matcher: matchers.MatchGoldenEqual(path.Join("testdata", "inspect_stats_dataplane.out")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				builders.Dataplane().WithName("backend-1").WithMesh("mesh-1").WithAdminPort(3301).WithServices("backend").AddOutboundsToServices("redis", "elastic", "web").Build(),
			},
			contentType: "text/plain",
		}),
		Entry("inspect stats for dataplane as json", testCase{
			path:    "/meshes/mesh-1/dataplanes/backend-1/stats",
			query:   "format=json",
			matcher: matchers.MatchGoldenEqual(path.Join("testdata", "inspect_stats_dataplane.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				builders.Dataplane().WithName("backend-1").WithMesh("mesh-1").WithAdminPort(3301).WithServices("backend").AddOutboundsToServices("redis", "elastic", "web").Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect clusters for dataplane", testCase{
			path:    "/meshes/mesh-1/dataplanes/backend-1/clusters",
			matcher: matchers.MatchGoldenEqual(path.Join("testdata", "inspect_clusters_dataplane.out")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				builders.Dataplane().WithName("backend-1").WithMesh("mesh-1").WithAdminPort(3301).WithServices("backend").AddOutboundsToServices("redis", "elastic", "web").Build(),
			},
			contentType: "text/plain",
		}),
		Entry("inspect clusters for dataplane as json", testCase{
			path:    "/meshes/mesh-1/dataplanes/backend-1/clusters",
			query:   "format=json",
			matcher: matchers.MatchGoldenEqual(path.Join("testdata", "inspect_clusters_dataplane.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				builders.Dataplane().WithName("backend-1").WithMesh("mesh-1").WithAdminPort(3301).WithServices("backend").AddOutboundsToServices("redis", "elastic", "web").Build(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect invalid admin type for dataplane", testCase{
			path:    "/meshes/mesh-1/dataplanes/backend-1/notAType",
			query:   "format=json",
			matcher: matchers.MatchGoldenEqual(path.Join("testdata", "inspect_dataplanes.invalid_admin.json")),
			resources: []core_model.Resource{
				builders.Mesh().WithName("mesh-1").Build(),
				builders.Dataplane().WithName("backend-1").WithMesh("mesh-1").WithAdminPort(3301).WithServices("backend").AddOutboundsToServices("redis", "elastic", "web").Build(),
			},
			contentType: restful.MIME_JSON,
		}),

		Entry("inspect rules empty", testCase{
			path:    "/meshes/default/dataplanes/web-01/rules",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_dataplane_rules_empty.golden.json")),
			resources: []core_model.Resource{
				samples2.MeshDefault(),
				samples2.DataplaneWeb(),
			},
			contentType: restful.MIME_JSON,
		}),
		Entry("inspect rules basic", testCase{
			path:    "/meshes/default/dataplanes/web-01/rules",
			matcher: matchers.MatchGoldenJSON(path.Join("testdata", "inspect_dataplane_rules.golden.json")),
			resources: []core_model.Resource{
				samples2.MeshDefault(),
				samples2.DataplaneWeb(),
				builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefDataplaneName("web-01")).
					AddFrom(builders.TargetRefService("client"), v1alpha1.Deny).
					Build(),
				builders.MeshAccessLog().
					WithTargetRef(builders.TargetRefDataplaneName("web-01")).
					AddTo(builders.TargetRefMesh(), samples2.MeshAccessLogFileConf()).
					Build(),
				builders.MeshTrace().
					WithTargetRef(builders.TargetRefDataplaneName("web-01")).
					WithZipkinBackend(samples2.ZipkinBackend()).
					Build(),
			},
			contentType: restful.MIME_JSON,
		}),
	)

	It("marshals empty meshgateway inspect rule slices as arrays", func() {
		toRules := []api_common.Rule{}
		fromRules := []api_common.FromRule{}
		inboundRules := []api_common.InboundRulesEntry{}
		toResourceRules := []api_common.ResourceRule{}
		warnings := []string{"warning"}

		bytes, err := json.Marshal(api_common.InspectRule{
			Type:            "MeshRetry",
			ToRules:         &toRules,
			FromRules:       &fromRules,
			InboundRules:    &inboundRules,
			ToResourceRules: &toResourceRules,
			Warnings:        &warnings,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(bytes).To(MatchJSON(`{
			"type": "MeshRetry",
			"toRules": [],
			"fromRules": [],
			"inboundRules": [],
			"toResourceRules": [],
			"warnings": ["warning"]
		}`))
	})

	It("should change response if state changed", func() {
		// setup
		var apiServer *api_server.ApiServer
		var stop func()
		resourceStore := memory.NewStore()
		rm := manager.NewResourceManager(resourceStore)
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer().WithStore(resourceStore))
		defer stop()

		// when init the state
		// TrafficPermission that selects 2 DPPs
		initState := []core_model.Resource{
			builders.Mesh().Build(),
			builders.Dataplane().WithName("backend-1").WithHttpServices("backend").AddOutboundsToServices("redis", "elastic").Build(),
			builders.Dataplane().WithName("redis-1").WithHttpServices("redis").AddOutboundsToServices("redis", "backend", "elastic").Build(),
		}
		for _, resource := range initState {
			err := rm.Create(context.Background(), resource,
				store.CreateBy(core_model.MetaToResourceKey(resource.GetMeta())))
			Expect(err).ToNot(HaveOccurred())
		}

		// then
		var resp *http.Response
		Eventually(func() error {
			r, err := http.Get((&url.URL{
				Scheme: "http",
				Host:   apiServer.Address(),
				Path:   "/meshes/default/meshtrafficpermissions/tp-1/dataplanes",
			}).String())
			resp = r
			return err
		}, "3s").ShouldNot(HaveOccurred())
		bytes, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "inspect_changed_state_before.json")))

		// when change the state
		err = rm.Delete(context.Background(), core_mesh.NewDataplaneResource(), store.DeleteByKey("backend-1", "default"))
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func() error {
			r, err := http.Get((&url.URL{
				Scheme: "http",
				Host:   apiServer.Address(),
				Path:   "/meshes/default/meshtrafficpermissions/tp-1/dataplanes",
			}).String())
			resp = r
			return err
		}, "3s").ShouldNot(HaveOccurred())
		bytes, err = io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "inspect_changed_state_after.json")))
	})
})
