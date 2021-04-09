package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/emicklei/go-restful"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/go-logr/logr/testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	mads_config "github.com/kumahq/kuma/pkg/config/mads"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	mads_v1 "github.com/kumahq/kuma/pkg/mads/v1"
	"github.com/kumahq/kuma/pkg/mads/v1/service"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/util/test"
)

var _ = Describe("MADS http service", func() {

	var url string

	var resManager core_manager.ResourceManager

	BeforeEach(func() {
		resManager = core_manager.NewResourceManager(memory.NewStore())

		cfg := mads_config.DefaultMonitoringAssignmentServerConfig()
		svc := service.NewService(cfg, resManager, testing.NullLogger{})

		ws := new(restful.WebService)
		svc.RegisterRoutes(ws)

		container := restful.NewContainer()
		container.Add(ws)

		srv := test.NewHttpServer(container)
		url = srv.Server().URL

		// wait for the server
		Eventually(srv.Ready).ShouldNot(HaveOccurred())
	})

	It("should respond with an empty discovery response", func() {
		// given
		discoveryReq := envoy_v3.DiscoveryRequest{
			VersionInfo:   "",
			ResponseNonce: "",
			TypeUrl:       mads_v1.MonitoringAssignmentType,
			ResourceNames: []string{},
			Node: &envoy_core.Node{
				Id: "test",
			},
		}
		reqBytes, err := json.Marshal(&discoveryReq)
		Expect(err).ToNot(HaveOccurred())

		// when
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/discovery:monitoringassignment", url), bytes.NewReader(reqBytes))
		Expect(err).ToNot(HaveOccurred())
		req.Header.Add("content-type", "application/json")
		resp, err := http.DefaultClient.Do(req)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(200))

		// when
		respBody, err := ioutil.ReadAll(resp.Body)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		discoveryRes := &envoy_v3.DiscoveryResponse{}
		err = json.Unmarshal(respBody, discoveryRes)
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(discoveryRes.TypeUrl).To(Equal(mads_v1.MonitoringAssignmentType))
		Expect(discoveryRes.VersionInfo).ToNot(BeEmpty())
		Expect(discoveryRes.Resources).To(BeEmpty())

		// and given the same version
		discoveryReq.VersionInfo = discoveryRes.VersionInfo
		reqBytes, err = json.Marshal(&discoveryReq)
		Expect(err).ToNot(HaveOccurred())

		// when
		req, err = http.NewRequest("POST", fmt.Sprintf("%s/v3/discovery:monitoringassignment", url), bytes.NewReader(reqBytes))
		Expect(err).ToNot(HaveOccurred())
		req.Header.Add("content-type", "application/json")
		resp, err = http.DefaultClient.Do(req)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(304))

		// when
		respBody, err = ioutil.ReadAll(resp.Body)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(respBody).To(BeEmpty())
	})

	Describe("with resources", func() {
		createMesh := func(mesh *mesh_core.MeshResource) error {
			return resManager.Create(context.Background(), mesh, store.CreateByKey(mesh.GetMeta().GetName(), model.NoMesh))
		}

		createDataPlane := func(dp *mesh_core.DataplaneResource) error {
			err := resManager.Create(context.Background(), dp, store.CreateByKey(dp.Meta.GetName(), dp.GetMeta().GetMesh()))
			return err
		}

		var mesh = &mesh_core.MeshResource{
			Meta: &test_model.ResourceMeta{
				Name: "test",
			},
			Spec: &v1alpha1.Mesh{
				Metrics: &v1alpha1.Metrics{
					EnabledBackend: "prometheus-1",
					Backends: []*v1alpha1.MetricsBackend{
						{
							Name: "prometheus-1",
							Type: v1alpha1.MetricsPrometheusType,
						},
					},
				},
			},
		}

		var dp1 = &mesh_core.DataplaneResource{
			Meta: &test_model.ResourceMeta{
				Name: "dp-1",
				Mesh: mesh.GetMeta().GetName(),
			},
			Spec: &v1alpha1.Dataplane{
				Networking: &v1alpha1.Dataplane_Networking{
					Address: "192.168.0.1",
					Gateway: &v1alpha1.Dataplane_Networking_Gateway{
						Tags: map[string]string{
							"kuma.io/service": "gateway",
							"region":          "eu",
						},
					},
				},
			},
		}

		var dp2 = &mesh_core.DataplaneResource{
			Meta: &test_model.ResourceMeta{
				Name: "dp-2",
				Mesh: mesh.GetMeta().GetName(),
			},
			Spec: &v1alpha1.Dataplane{
				Networking: &v1alpha1.Dataplane_Networking{
					Address: "192.168.0.1",
					Inbound: []*v1alpha1.Dataplane_Networking_Inbound{
						{
							Port:        80,
							ServicePort: 8080,
							Tags: map[string]string{
								"kuma.io/service": "backend",
								"env":             "prod",
								"version":         "v1",
							},
						},
						{
							Address:     "192.168.0.2",
							Port:        443,
							ServicePort: 8443,
							Tags: map[string]string{
								"kuma.io/service": "backend-https",
								"env":             "prod",
								"version":         "v2",
							},
						},
					},
				},
			},
		}

		BeforeEach(func() {
			// given
			err := createMesh(mesh)
			Expect(err).ToNot(HaveOccurred())

			err = createDataPlane(dp1)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return the monitoring assignments", func() {
			// given
			discoveryReq := envoy_v3.DiscoveryRequest{
				VersionInfo:   "",
				ResponseNonce: "",
				TypeUrl:       mads_v1.MonitoringAssignmentType,
				ResourceNames: []string{},
				Node: &envoy_core.Node{
					Id: "test",
				},
			}
			reqBytes, err := json.Marshal(&discoveryReq)
			Expect(err).ToNot(HaveOccurred())

			// when
			req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/discovery:monitoringassignment", url), bytes.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")
			resp, err := http.DefaultClient.Do(req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			// when
			respBody, err := ioutil.ReadAll(resp.Body)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			discoveryRes := &envoy_v3.DiscoveryResponse{}
			err = json.Unmarshal(respBody, discoveryRes)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(discoveryRes.TypeUrl).To(Equal(mads_v1.MonitoringAssignmentType))
			Expect(discoveryRes.VersionInfo).ToNot(BeEmpty())
			Expect(discoveryRes.Resources).To(HaveLen(1))

			// and given the same version
			discoveryReq.VersionInfo = discoveryRes.VersionInfo
			reqBytes, err = json.Marshal(&discoveryReq)
			Expect(err).ToNot(HaveOccurred())

			// when
			req, err = http.NewRequest("POST", fmt.Sprintf("%s/v3/discovery:monitoringassignment", url), bytes.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")
			resp, err = http.DefaultClient.Do(req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(304))

			// when
			respBody, err = ioutil.ReadAll(resp.Body)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(respBody).To(BeEmpty())
		})

		// TODO: add ticking refresh to rest-based methods
		It("should return the refreshed monitoring assignments when there are updates", func() {
			// given
			discoveryReq := envoy_v3.DiscoveryRequest{
				VersionInfo:   "",
				ResponseNonce: "",
				TypeUrl:       mads_v1.MonitoringAssignmentType,
				ResourceNames: []string{},
				Node: &envoy_core.Node{
					Id: "test",
				},
			}
			reqBytes, err := json.Marshal(&discoveryReq)
			Expect(err).ToNot(HaveOccurred())

			// when
			req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/discovery:monitoringassignment", url), bytes.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")
			resp, err := http.DefaultClient.Do(req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			// when
			respBody, err := ioutil.ReadAll(resp.Body)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			discoveryRes := &envoy_v3.DiscoveryResponse{}
			err = json.Unmarshal(respBody, discoveryRes)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(discoveryRes.TypeUrl).To(Equal(mads_v1.MonitoringAssignmentType))
			Expect(discoveryRes.VersionInfo).ToNot(BeEmpty())
			Expect(discoveryRes.Resources).To(HaveLen(1))

			// TODO: force a tick here
			// given an updated mesh
			err = createDataPlane(dp2)
			Expect(err).ToNot(HaveOccurred())

			// and given the same version
			discoveryReq.VersionInfo = discoveryRes.VersionInfo
			reqBytes, err = json.Marshal(&discoveryReq)
			Expect(err).ToNot(HaveOccurred())

			// when
			req, err = http.NewRequest("POST", fmt.Sprintf("%s/v3/discovery:monitoringassignment", url), bytes.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")
			resp, err = http.DefaultClient.Do(req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			// when
			respBody, err = ioutil.ReadAll(resp.Body)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = json.Unmarshal(respBody, discoveryRes)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(discoveryRes.TypeUrl).To(Equal(mads_v1.MonitoringAssignmentType))
			Expect(discoveryRes.VersionInfo).ToNot(BeEmpty())
			Expect(discoveryRes.Resources).To(HaveLen(2))
		})
	})
})
