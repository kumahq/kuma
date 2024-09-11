package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/emicklei/go-restful/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/go-logr/logr"
	"github.com/golang/protobuf/jsonpb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	observability_v1 "github.com/kumahq/kuma/api/observability/v1"
	mads_config "github.com/kumahq/kuma/pkg/config/mads"
	config_types "github.com/kumahq/kuma/pkg/config/types"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	rest_error_types "github.com/kumahq/kuma/pkg/core/rest/errors/types"
	mads_v1 "github.com/kumahq/kuma/pkg/mads/v1"
	"github.com/kumahq/kuma/pkg/mads/v1/service"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/util/test"
)

var _ = Describe("MADS http service", func() {
	var url string
	var monitoringAssignmentPath string

	pbMarshaller := &jsonpb.Marshaler{OrigName: true}

	var resManager core_manager.ResourceManager

	// the refresh timeout should be smaller than the default fetch timeout so
	// a refresh can happen during a single request

	const refreshInterval = 250 * time.Millisecond

	const defaultFetchTimeout = 1 * time.Second

	BeforeEach(func() {
		resManager = core_manager.NewResourceManager(memory.NewStore())

		cfg := mads_config.DefaultMonitoringAssignmentServerConfig()
		cfg.AssignmentRefreshInterval = config_types.Duration{Duration: refreshInterval}
		cfg.DefaultFetchTimeout = config_types.Duration{Duration: defaultFetchTimeout}

		svc := service.NewService(cfg, resManager, logr.Discard(), nil)

		ws := new(restful.WebService)
		svc.RegisterRoutes(ws)

		container := restful.NewContainer()
		container.Add(ws)

		srv := test.NewHttpServer(container)
		url = srv.Server().URL
		monitoringAssignmentPath = fmt.Sprintf("%s%s", url, service.FetchMonitoringAssignmentsPath)

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
		reqBytes, err := pbMarshaller.MarshalToString(&discoveryReq)
		Expect(err).ToNot(HaveOccurred())

		// when
		req, err := http.NewRequest("POST", monitoringAssignmentPath, strings.NewReader(reqBytes))
		Expect(err).ToNot(HaveOccurred())
		req.Header.Add("content-type", "application/json")
		resp, err := http.DefaultClient.Do(req)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusOK))

		// when
		respBody, err := io.ReadAll(resp.Body)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		discoveryRes := &envoy_v3.DiscoveryResponse{}
		err = jsonpb.Unmarshal(bytes.NewReader(respBody), discoveryRes)
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(discoveryRes.TypeUrl).To(Equal(mads_v1.MonitoringAssignmentType))
		Expect(discoveryRes.VersionInfo).ToNot(BeEmpty())
		Expect(discoveryRes.Resources).To(BeEmpty())

		// and given the same version
		discoveryReq.VersionInfo = discoveryRes.VersionInfo
		reqBytes, err = pbMarshaller.MarshalToString(&discoveryReq)
		Expect(err).ToNot(HaveOccurred())

		// when
		req, err = http.NewRequest("POST", monitoringAssignmentPath, strings.NewReader(reqBytes))
		Expect(err).ToNot(HaveOccurred())
		req.Header.Add("content-type", "application/json")
		resp, err = http.DefaultClient.Do(req)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusNotModified))

		// when
		respBody, err = io.ReadAll(resp.Body)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(respBody).To(BeEmpty())
	})

	Describe("with resources", func() {
		createMesh := func(mesh *core_mesh.MeshResource) error {
			return resManager.Create(context.Background(), mesh, store.CreateByKey(mesh.GetMeta().GetName(), model.NoMesh))
		}

		createDataPlane := func(dp *core_mesh.DataplaneResource) error {
			err := resManager.Create(context.Background(), dp, store.CreateByKey(dp.Meta.GetName(), dp.GetMeta().GetMesh()))
			return err
		}

		mesh := &core_mesh.MeshResource{
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

		dp1 := &core_mesh.DataplaneResource{
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

		dp2 := &core_mesh.DataplaneResource{
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
			reqBytes, err := pbMarshaller.MarshalToString(&discoveryReq)
			Expect(err).ToNot(HaveOccurred())

			// when
			req, err := http.NewRequest("POST", monitoringAssignmentPath, strings.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")
			resp, err := http.DefaultClient.Do(req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(HaveHTTPStatus(http.StatusOK))

			// when
			respBody, err := io.ReadAll(resp.Body)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			discoveryRes := &envoy_v3.DiscoveryResponse{}
			err = jsonpb.Unmarshal(bytes.NewReader(respBody), discoveryRes)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(discoveryRes.TypeUrl).To(Equal(mads_v1.MonitoringAssignmentType))
			Expect(discoveryRes.VersionInfo).ToNot(BeEmpty())
			Expect(discoveryRes.Resources).To(HaveLen(1))
			Expect(discoveryRes.Resources[0].TypeUrl).To(Equal(mads_v1.MonitoringAssignmentType))
			Expect(discoveryRes.Resources[0].Value).ToNot(BeEmpty())

			// when
			assignment := &observability_v1.MonitoringAssignment{}
			err = util_proto.UnmarshalAnyTo(discoveryRes.Resources[0], assignment)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(assignment.Mesh).To(Equal(dp1.GetMeta().GetMesh()))
			Expect(assignment.Targets).To(HaveLen(1))
			Expect(assignment.Targets[0].Name).To(Equal(dp1.GetMeta().GetName()))

			// given the same version
			discoveryReq.VersionInfo = discoveryRes.VersionInfo
			reqBytes, err = pbMarshaller.MarshalToString(&discoveryReq)
			Expect(err).ToNot(HaveOccurred())

			// when
			req, err = http.NewRequest("POST", monitoringAssignmentPath, strings.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")
			resp, err = http.DefaultClient.Do(req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(HaveHTTPStatus(http.StatusNotModified))

			// when
			respBody, err = io.ReadAll(resp.Body)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(respBody).To(BeEmpty())
		})

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
			reqBytes, err := pbMarshaller.MarshalToString(&discoveryReq)
			Expect(err).ToNot(HaveOccurred())

			// when
			req, err := http.NewRequest("POST", monitoringAssignmentPath, strings.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")
			resp, err := http.DefaultClient.Do(req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(HaveHTTPStatus(http.StatusOK))

			// when
			respBody, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			// when
			discoveryRes := &envoy_v3.DiscoveryResponse{}
			err = jsonpb.Unmarshal(bytes.NewReader(respBody), discoveryRes)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(discoveryRes.TypeUrl).To(Equal(mads_v1.MonitoringAssignmentType))
			Expect(discoveryRes.VersionInfo).ToNot(BeEmpty())
			Expect(discoveryRes.Resources).To(HaveLen(1))

			// given an updated mesh
			err = createDataPlane(dp2)
			Expect(err).ToNot(HaveOccurred())

			// and given the same version
			discoveryReq.VersionInfo = discoveryRes.VersionInfo
			reqBytes, err = pbMarshaller.MarshalToString(&discoveryReq)
			Expect(err).ToNot(HaveOccurred())

			// and given time to refresh
			time.Sleep(refreshInterval * 2)

			// when
			req, err = http.NewRequest("POST", monitoringAssignmentPath, strings.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")
			resp, err = http.DefaultClient.Do(req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(HaveHTTPStatus(http.StatusOK))

			// when
			respBody, err = io.ReadAll(resp.Body)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = jsonpb.Unmarshal(bytes.NewReader(respBody), discoveryRes)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(discoveryRes.TypeUrl).To(Equal(mads_v1.MonitoringAssignmentType))
			Expect(discoveryRes.VersionInfo).ToNot(BeEmpty())
			Expect(discoveryRes.Resources).To(HaveLen(2))
		})

		It("should block until there are updates", func() {
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
			reqBytes, err := pbMarshaller.MarshalToString(&discoveryReq)
			Expect(err).ToNot(HaveOccurred())

			// when
			req, err := http.NewRequest("POST", monitoringAssignmentPath, strings.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")

			resp, err := http.DefaultClient.Do(req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(HaveHTTPStatus(http.StatusOK))

			// when
			respBody, err := io.ReadAll(resp.Body)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			discoveryRes := &envoy_v3.DiscoveryResponse{}
			err = jsonpb.Unmarshal(bytes.NewReader(respBody), discoveryRes)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(discoveryRes.TypeUrl).To(Equal(mads_v1.MonitoringAssignmentType))
			Expect(discoveryRes.VersionInfo).ToNot(BeEmpty())
			Expect(discoveryRes.Resources).To(HaveLen(1))

			// and given the same version
			discoveryReq.VersionInfo = discoveryRes.VersionInfo
			// simulate restarted prometheus
			discoveryReq.Node.Id = "new-prome"
			reqBytes, err = pbMarshaller.MarshalToString(&discoveryReq)
			Expect(err).ToNot(HaveOccurred())

			// when
			req, err = http.NewRequest("POST", monitoringAssignmentPath, strings.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")

			respChan := make(chan *http.Response, 1)

			go func() {
				resp2, err := http.DefaultClient.Do(req)
				Expect(err).ToNot(HaveOccurred())

				respChan <- resp2
			}()

			// given an updated mesh while the request is in progress
			time.Sleep(defaultFetchTimeout / 2)

			err = createDataPlane(dp2)
			Expect(err).ToNot(HaveOccurred())

			resp2 := <-respChan

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp2).To(HaveHTTPStatus(http.StatusOK))

			// when
			respBody, err = io.ReadAll(resp2.Body)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = jsonpb.Unmarshal(bytes.NewReader(respBody), discoveryRes)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(discoveryRes.TypeUrl).To(Equal(mads_v1.MonitoringAssignmentType))
			Expect(discoveryRes.VersionInfo).ToNot(BeEmpty())
			Expect(discoveryRes.Resources).To(HaveLen(2))
		})

		It("should return straightaway on first request with a fetch timeout", func() {
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
			reqBytes, err := pbMarshaller.MarshalToString(&discoveryReq)
			Expect(err).ToNot(HaveOccurred())

			// when
			req, err := http.NewRequest("POST", monitoringAssignmentPath+"?fetch-timeout=1s", strings.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")

			start := time.Now()
			resp, err := http.DefaultClient.Do(req)
			// Ensure we're returning in less than 1 sec as this is the first request
			Expect(time.Now()).To(BeTemporally("<", start.Add(time.Second)))
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(HaveHTTPStatus(http.StatusOK))

			// when
			respBody, err := io.ReadAll(resp.Body)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			discoveryRes := &envoy_v3.DiscoveryResponse{}
			err = jsonpb.Unmarshal(bytes.NewReader(respBody), discoveryRes)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(discoveryRes.TypeUrl).To(Equal(mads_v1.MonitoringAssignmentType))
			Expect(discoveryRes.VersionInfo).ToNot(BeEmpty())
			// and should only contain the first resource
			Expect(discoveryRes.Resources).To(HaveLen(1))
		})

		It("should return no update if the fetch times out", func() {
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
			reqBytes, err := pbMarshaller.MarshalToString(&discoveryReq)
			Expect(err).ToNot(HaveOccurred())

			// when
			req, err := http.NewRequest("POST", monitoringAssignmentPath, strings.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")

			resp, err := http.DefaultClient.Do(req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(HaveHTTPStatus(http.StatusOK))

			// when
			respBody, err := io.ReadAll(resp.Body)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			discoveryRes := &envoy_v3.DiscoveryResponse{}
			err = jsonpb.Unmarshal(bytes.NewReader(respBody), discoveryRes)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(discoveryRes.TypeUrl).To(Equal(mads_v1.MonitoringAssignmentType))
			Expect(discoveryRes.VersionInfo).ToNot(BeEmpty())
			Expect(discoveryRes.Resources).To(HaveLen(1))

			// and given the same version
			discoveryReq.VersionInfo = discoveryRes.VersionInfo
			reqBytes, err = pbMarshaller.MarshalToString(&discoveryReq)
			Expect(err).ToNot(HaveOccurred())

			// when
			req, err = http.NewRequest("POST", monitoringAssignmentPath+"?fetch-timeout=1ms", strings.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")

			resp, err = http.DefaultClient.Do(req)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(HaveHTTPStatus(http.StatusNotModified))
		})

		It("should allow synchronous requests", func() {
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
			reqBytes, err := pbMarshaller.MarshalToString(&discoveryReq)
			Expect(err).ToNot(HaveOccurred())

			// when
			req, err := http.NewRequest("POST", monitoringAssignmentPath+"?fetch-timeout=0s", strings.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")

			respChan := make(chan *http.Response, 1)

			go func() {
				resp, err := http.DefaultClient.Do(req)
				Expect(err).ToNot(HaveOccurred())

				respChan <- resp
			}()

			// given an updated mesh while the request is in progress
			time.Sleep(defaultFetchTimeout / 2)

			err = createDataPlane(dp2)
			Expect(err).ToNot(HaveOccurred())

			resp := <-respChan

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(HaveHTTPStatus(http.StatusOK))

			// when
			respBody, err := io.ReadAll(resp.Body)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			discoveryRes := &envoy_v3.DiscoveryResponse{}
			err = jsonpb.Unmarshal(bytes.NewReader(respBody), discoveryRes)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(discoveryRes.TypeUrl).To(Equal(mads_v1.MonitoringAssignmentType))
			Expect(discoveryRes.VersionInfo).ToNot(BeEmpty())
			// and should only contain the first resource
			Expect(discoveryRes.Resources).To(HaveLen(1))
		})

		It("should return an error if the fetch timeout is unparseable", func() {
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
			reqBytes, err := pbMarshaller.MarshalToString(&discoveryReq)
			Expect(err).ToNot(HaveOccurred())

			// when
			req, err := http.NewRequest("POST", monitoringAssignmentPath+"?fetch-timeout=not-a-timeout", strings.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")

			resp, err := http.DefaultClient.Do(req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(HaveHTTPStatus(http.StatusBadRequest))

			// when
			respBody, err := io.ReadAll(resp.Body)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			resError := &rest_error_types.Error{}
			err = json.Unmarshal(respBody, resError)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resError.Title).To(Equal("Could not parse fetch-timeout"))
		})
	})
})
