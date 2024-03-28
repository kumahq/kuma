package api_server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/emicklei/go-restful/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	api_server "github.com/kumahq/kuma/pkg/api-server"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/unversioned"
	rest_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/matchers"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

var _ = Describe("Resource Endpoints", func() {
	var apiServer *api_server.ApiServer
	var resourceStore core_store.ResourceStore
	stop := func() {}
	var metrics core_metrics.Metrics

	const mesh = "default"

	BeforeEach(func() {
		resourceStore = core_store.NewPaginationStore(memory.NewStore())
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer().WithStore(resourceStore).WithMetrics(func() core_metrics.Metrics {
			m, _ := core_metrics.NewMetrics("Zone")
			metrics = m
			return m
		}))
	})

	AfterEach(func() {
		stop()
	})

	BeforeEach(func() {
		// create default mesh
		err := resourceStore.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(mesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should support CORS", func() {
		// given
		req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/meshes/%s/traffic-routes", apiServer.Address(), mesh), nil)
		Expect(err).NotTo(HaveOccurred())
		req.Header.Add(restful.HEADER_Origin, "test")

		// when
		response, err := http.DefaultClient.Do(req)

		// then
		Expect(err).NotTo(HaveOccurred())

		// when
		value := response.Header.Get(restful.HEADER_AccessControlAllowOrigin)

		// then server returns that the domain is allowed
		Expect(value).To(Equal("test"))
	})

	It("should expose metrics", func() {
		// given
		req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/meshes/%s/traffic-routes", apiServer.Address(), mesh), nil)
		Expect(err).NotTo(HaveOccurred())

		// when
		_, err = http.DefaultClient.Do(req)

		// then
		Expect(err).NotTo(HaveOccurred())
		Expect(test_metrics.FindMetric(metrics, "api_server_http_request_duration_seconds")).ToNot(BeNil())
		Expect(test_metrics.FindMetric(metrics, "api_server_http_requests_inflight")).ToNot(BeNil())
		Expect(test_metrics.FindMetric(metrics, "api_server_http_response_size_bytes")).ToNot(BeNil())
	})
})

var _ = Describe("Resource Endpoints on Zone, label origin", func() {
	createServer := func(federatedZone, validateOriginLabel bool) (*api_server.ApiServer, core_store.ResourceStore, func()) {
		store := core_store.NewPaginationStore(memory.NewStore())
		zone := ""
		if federatedZone {
			zone = "zone-1"
		}
		apiServer, _, stop := StartApiServer(
			NewTestApiServerConfigurer().
				WithStore(store).
				WithDisableOriginLabelValidation(!validateOriginLabel).
				WithZone(zone),
		)
		return apiServer, store, stop
	}
	put := func(address string, mesh string, resType model.ResourceTypeDescriptor, name string, res rest.Resource) (*http.Response, error) {
		GinkgoHelper()
		jsonBytes, err := json.Marshal(res)
		Expect(err).ToNot(HaveOccurred())

		request, err := http.NewRequest(
			"PUT",
			fmt.Sprintf("http://%s/meshes/%s/%s/%s", address, mesh, resType.WsPath, name),
			bytes.NewBuffer(jsonBytes),
		)
		Expect(err).ToNot(HaveOccurred())
		request.Header.Add("content-type", "application/json")
		return http.DefaultClient.Do(request)
	}

	createMesh := func(s core_store.ResourceStore, mesh string) {
		// create default mesh
		err := s.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(mesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	}

	It("should return 400 when origin validation is enabled and origin label is not set", func() {
		// given
		apiServer, store, stop := createServer(true, true)
		defer stop()
		createMesh(store, "mesh-1")

		// when
		res := &rest_v1alpha1.Resource{
			ResourceMeta: rest_v1alpha1.ResourceMeta{
				Name: "mtp-1",
				Mesh: "mesh-1",
				Type: string(v1alpha1.MeshTrafficPermissionType),
			},
			Spec: builders.MeshTrafficPermission().
				WithTargetRef(builders.TargetRefMesh()).
				AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
				Build().Spec,
		}
		resp, err := put(apiServer.Address(), "mesh-1", v1alpha1.MeshTrafficPermissionResourceTypeDescriptor, "mtp-1", res)

		// and then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		bytes, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(bytes).To(matchers.MatchGoldenJSON(path.Join("testdata", "resource_400onNoOriginLabel.golden.json")))
	})

	DescribeTable("should set origin label automatically when origin validation is disabled",
		func(federatedZone bool) {
			// given
			apiServer, store, stop := createServer(federatedZone, false)
			defer stop()
			createMesh(store, "mesh-1")

			// when
			res := &rest_v1alpha1.Resource{
				ResourceMeta: rest_v1alpha1.ResourceMeta{
					Name: "mtp-1",
					Mesh: "mesh-1",
					Type: string(v1alpha1.MeshTrafficPermissionType),
				},
				Spec: builders.MeshTrafficPermission().
					WithTargetRef(builders.TargetRefMesh()).
					AddFrom(builders.TargetRefMesh(), v1alpha1.Allow).
					Build().Spec,
			}
			resp, err := put(apiServer.Address(), "mesh-1", v1alpha1.MeshTrafficPermissionResourceTypeDescriptor, "mtp-1", res)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			// and then
			actualMtp := v1alpha1.NewMeshTrafficPermissionResource()
			Expect(store.Get(context.Background(), actualMtp, core_store.GetByKey("mtp-1", "mesh-1"))).To(Succeed())
			Expect(actualMtp.Meta.GetLabels()).To(Equal(map[string]string{
				mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
			}))
		},
		Entry("non-federated zone", false),
		Entry("federated zone", true),
	)

	It("should set origin label automatically for DPPs", func() {
		// given
		apiServer, store, stop := createServer(false, false)
		defer stop()
		createMesh(store, "mesh-1")
		name := "dpp-1"
		mesh := "mesh-1"

		// when
		res := &unversioned.Resource{
			Meta: rest_v1alpha1.ResourceMeta{
				Name: name,
				Mesh: mesh,
				Type: string(core_mesh.DataplaneType),
			},
			Spec: builders.Dataplane().
				WithName("backend-1").
				WithHttpServices("backend").
				AddOutboundsToServices("redis", "elastic", "postgres").
				Build().Spec,
		}
		resp, err := put(apiServer.Address(), mesh, core_mesh.DataplaneResourceTypeDescriptor, name, res)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))
		// and then
		actualDpp := core_mesh.NewDataplaneResource()
		Expect(store.Get(context.Background(), actualDpp, core_store.GetByKey("dpp-1", "mesh-1"))).To(Succeed())
		Expect(actualDpp.Meta.GetLabels()).To(Equal(map[string]string{
			mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
		}))
	})
})
