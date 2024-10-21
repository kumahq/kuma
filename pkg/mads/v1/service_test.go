package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"google.golang.org/protobuf/types/known/anypb"

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
		DeferCleanup(func() {
			srv.Server().Close()
		})

		// wait for the server
		Eventually(srv.Ready).Should(Succeed())
	})

	Context("with resources", func() {
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

			discoveryRes := &envoy_v3.DiscoveryResponse{}
			// then
			Expect(http.DefaultClient.Do(req)).
				Should(And(
					HaveHTTPStatus(http.StatusOK),
					HaveHTTPBody(WithTransform(func(b []byte) (*envoy_v3.DiscoveryResponse, error) {
						err := jsonpb.Unmarshal(bytes.NewReader(b), discoveryRes)
						return discoveryRes, err
					}, And(
						HaveField("TypeUrl", mads_v1.MonitoringAssignmentType),
						HaveField("VersionInfo", Not(BeEmpty())),
						HaveField("Resources", BeEmpty()),
					))),
				))

			// and given the same version
			discoveryReq.VersionInfo = discoveryRes.VersionInfo
			reqBytes, err = pbMarshaller.MarshalToString(&discoveryReq)
			Expect(err).ToNot(HaveOccurred())

			// when
			req, err = http.NewRequest("POST", monitoringAssignmentPath, strings.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")

			// then
			By(fmt.Sprintf("Doing other discovery request with clientId %q and version %q", discoveryReq.Node.Id, discoveryReq.VersionInfo))
			Expect(http.DefaultClient.Do(req)).
				Should(And(
					HaveHTTPStatus(http.StatusNotModified),
					HaveHTTPBody(BeEmpty()),
				))
		})
	})

	Context("with resources", func() {
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
			Expect(createMesh(mesh)).To(Succeed())
			Expect(createDataPlane(dp1)).To(Succeed())
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

			// then
			discoveryRes := &envoy_v3.DiscoveryResponse{}
			Expect(http.DefaultClient.Do(req)).
				Should(And(
					HaveHTTPStatus(http.StatusOK),
					HaveHTTPBody(WithTransform(func(b []byte) (*envoy_v3.DiscoveryResponse, error) {
						GinkgoLogr.Info("Got response", "response", string(b))
						err := jsonpb.Unmarshal(bytes.NewReader(b), discoveryRes)
						return discoveryRes, err
					}, And(
						HaveField("TypeUrl", mads_v1.MonitoringAssignmentType),
						HaveField("VersionInfo", Not(BeEmpty())),
						HaveField("Resources", HaveExactElements(And(
							HaveField("TypeUrl", mads_v1.MonitoringAssignmentType),
							HaveField("Value", Not(BeEmpty())),
							WithTransform(func(entry *anypb.Any) (*observability_v1.MonitoringAssignment, error) {
								assignment := &observability_v1.MonitoringAssignment{}
								err = util_proto.UnmarshalAnyTo(entry, assignment)
								return assignment, err
							}, And(
								HaveField("Mesh", dp1.GetMeta().GetMesh()),
								HaveField("Targets",
									HaveExactElements(HaveField("Name", dp1.GetMeta().GetName())),
								),
							)),
						)))),
					))),
				)

			// given the same version
			discoveryReq.VersionInfo = discoveryRes.VersionInfo
			reqBytes, err = pbMarshaller.MarshalToString(&discoveryReq)
			Expect(err).ToNot(HaveOccurred())

			// when
			req, err = http.NewRequest("POST", monitoringAssignmentPath, strings.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")

			// then
			By(fmt.Sprintf("Doing other discovery request with clientId %q and version %q", discoveryReq.Node.Id, discoveryReq.VersionInfo))
			Expect(http.DefaultClient.Do(req)).
				Should(And(
					HaveHTTPStatus(http.StatusNotModified),
					HaveHTTPBody(BeEmpty()),
				))
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

			// then
			discoveryRes := &envoy_v3.DiscoveryResponse{}
			Expect(http.DefaultClient.Do(req)).
				Should(And(
					HaveHTTPStatus(http.StatusOK),
					HaveHTTPBody(WithTransform(func(b []byte) (*envoy_v3.DiscoveryResponse, error) {
						err = jsonpb.Unmarshal(bytes.NewReader(b), discoveryRes)
						return discoveryRes, err
					}, And(
						HaveField("TypeUrl", mads_v1.MonitoringAssignmentType),
						HaveField("VersionInfo", Not(BeEmpty())),
						HaveField("Resources", HaveLen(1)),
					))),
				))

			// given an updated mesh (adding a DP)
			Expect(createDataPlane(dp2)).To(Succeed())

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

			// then
			Expect(http.DefaultClient.Do(req)).
				Should(And(
					HaveHTTPStatus(http.StatusOK),
					HaveHTTPBody(WithTransform(func(b []byte) (*envoy_v3.DiscoveryResponse, error) {
						err = jsonpb.Unmarshal(bytes.NewReader(b), discoveryRes)
						return discoveryRes, err
					}, And(
						HaveField("TypeUrl", mads_v1.MonitoringAssignmentType),
						HaveField("VersionInfo", Not(BeEmpty())),
						HaveField("Resources", HaveLen(2)),
					))),
				))
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

			discoveryRes := &envoy_v3.DiscoveryResponse{}
			Expect(http.DefaultClient.Do(req)).
				Should(And(
					HaveHTTPStatus(http.StatusOK),
					HaveHTTPBody(WithTransform(func(b []byte) (*envoy_v3.DiscoveryResponse, error) {
						err := jsonpb.Unmarshal(bytes.NewReader(b), discoveryRes)
						return discoveryRes, err
					}, And(
						HaveField("TypeUrl", mads_v1.MonitoringAssignmentType),
						HaveField("VersionInfo", Not(BeEmpty())),
						HaveField("Resources", HaveLen(1)),
					))),
				))

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
				localResp, err := http.DefaultClient.Do(req)
				Expect(err).ToNot(HaveOccurred())

				respChan <- localResp
			}()
			// Request is pending because there isn't any change
			Consistently(respChan, defaultFetchTimeout/2).ShouldNot(Receive())

			// given an updated mesh while the request is in progress
			Expect(createDataPlane(dp2)).To(Succeed())

			// then
			Eventually(respChan).Should(Receive(And(
				HaveHTTPStatus(http.StatusOK),
				HaveHTTPBody(WithTransform(func(b []byte) (*envoy_v3.DiscoveryResponse, error) {
					err = jsonpb.Unmarshal(bytes.NewReader(b), discoveryRes)
					return discoveryRes, err
				}, And(
					HaveField("TypeUrl", mads_v1.MonitoringAssignmentType),
					HaveField("VersionInfo", Not(BeEmpty())),
					HaveField("Resources", HaveLen(2)),
				)),
				),
			)))
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
			discoveryRes := &envoy_v3.DiscoveryResponse{}
			Expect(http.DefaultClient.Do(req)).
				Should(And(
					HaveHTTPStatus(http.StatusOK),
					HaveHTTPBody(WithTransform(func(b []byte) (*envoy_v3.DiscoveryResponse, error) {
						err = jsonpb.Unmarshal(bytes.NewReader(b), discoveryRes)
						return discoveryRes, err
					}, And(
						HaveField("TypeUrl", mads_v1.MonitoringAssignmentType),
						HaveField("VersionInfo", Not(BeEmpty())),
						HaveField("Resources", HaveLen(1)),
					))),
				))
			// Ensure we're returning in less than 1 sec as this is the first request
			Expect(time.Now()).To(BeTemporally("<", start.Add(time.Second)))
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

			// then
			discoveryRes := &envoy_v3.DiscoveryResponse{}
			Expect(http.DefaultClient.Do(req)).
				Should(And(
					HaveHTTPStatus(http.StatusOK),
					HaveHTTPBody(WithTransform(func(b []byte) (*envoy_v3.DiscoveryResponse, error) {
						err = jsonpb.Unmarshal(bytes.NewReader(b), discoveryRes)
						return discoveryRes, err
					}, And(
						HaveField("TypeUrl", mads_v1.MonitoringAssignmentType),
						HaveField("VersionInfo", Not(BeEmpty())),
						HaveField("Resources", HaveLen(1)),
					))),
				))

			// and given the same version
			discoveryReq.VersionInfo = discoveryRes.VersionInfo
			reqBytes, err = pbMarshaller.MarshalToString(&discoveryReq)
			Expect(err).ToNot(HaveOccurred())

			// when
			req, err = http.NewRequest("POST", monitoringAssignmentPath+"?fetch-timeout=1ms", strings.NewReader(reqBytes))
			Expect(err).ToNot(HaveOccurred())
			req.Header.Add("content-type", "application/json")

			By(fmt.Sprintf("Doing other discovery request with clientId %q and version %q", discoveryReq.Node.Id, discoveryReq.VersionInfo))
			Expect(http.DefaultClient.Do(req)).
				Should(HaveHTTPStatus(http.StatusNotModified))
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

			discoveryRes := &envoy_v3.DiscoveryResponse{}
			// Will resolve very quickly
			Eventually(respChan).WithTimeout(defaultFetchTimeout / 2).Should(Receive(And(
				HaveHTTPStatus(http.StatusOK),
				HaveHTTPBody(WithTransform(func(b []byte) (*envoy_v3.DiscoveryResponse, error) {
					err = jsonpb.Unmarshal(bytes.NewReader(b), discoveryRes)
					return discoveryRes, err
				}, And(
					HaveField("TypeUrl", mads_v1.MonitoringAssignmentType),
					HaveField("VersionInfo", Not(BeEmpty())),
					HaveField("Resources", HaveLen(1)),
				))),
			)))
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

			Expect(http.DefaultClient.Do(req)).
				Should(And(
					HaveHTTPStatus(http.StatusBadRequest),
					HaveHTTPBody(WithTransform(func(b []byte) (rest_error_types.Error, error) {
						e := rest_error_types.Error{}
						err := json.Unmarshal(b, &e)
						return e, err
					}, HaveField("Title", "Could not parse fetch-timeout"))),
				))
		})
	})
})
