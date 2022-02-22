package controllers_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_intstr "k8s.io/apimachinery/pkg/util/intstr"
	utilpointer "k8s.io/utils/pointer"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	. "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	util_yaml "github.com/kumahq/kuma/pkg/util/yaml"
)

var _ = Describe("PodToDataplane(..)", func() {

	pod := `
    metadata:
      namespace: demo
      name: example
      labels:
        app: example
        version: "0.1"
    spec:
      containers:
      - ports: []
        # when a 'targetPort' in a ServicePort is a number,
        # it should not be mandatory to list container ports explicitly
        #
        # containerPort: 8080
        # containerPort: 8443
      - ports:
        - containerPort: 7070
        - containerPort: 6060
          name: metrics
    status:
      podIP: 192.168.0.1
`

	ParseServices := func(values []string) ([]*kube_core.Service, error) {
		services := make([]*kube_core.Service, len(values))
		for i, value := range values {
			service := kube_core.Service{}
			if err := yaml.Unmarshal([]byte(value), &service); err != nil {
				return nil, err
			}
			services[i] = &service
		}
		return services, nil
	}

	ParseDataplanes := func(values []string) ([]*mesh_k8s.Dataplane, error) {
		dataplanes := make([]*mesh_k8s.Dataplane, len(values))
		for i, value := range values {
			dataplane := mesh_k8s.Dataplane{}
			if err := yaml.Unmarshal([]byte(value), &dataplane); err != nil {
				return nil, err
			}
			dataplanes[i] = &dataplane
		}
		return dataplanes, nil
	}

	type testCase struct {
		pod             string
		servicesForPod  string
		otherDataplanes string
		otherServices   string
		node            string
		dataplane       string
	}
	DescribeTable("should convert Pod into a Dataplane YAML version",
		func(given testCase) {
			// given
			// pod
			pod := &kube_core.Pod{}
			bytes, err := os.ReadFile(filepath.Join("testdata", given.pod))
			Expect(err).ToNot(HaveOccurred())
			err = yaml.Unmarshal(bytes, pod)
			Expect(err).ToNot(HaveOccurred())

			// services for pod
			services := []*kube_core.Service{}
			if given.servicesForPod != "" {
				bytes, err = os.ReadFile(filepath.Join("testdata", given.servicesForPod))
				Expect(err).ToNot(HaveOccurred())
				YAMLs := util_yaml.SplitYAML(string(bytes))
				services, err = ParseServices(YAMLs)
				Expect(err).ToNot(HaveOccurred())
			}

			namespace := kube_core.Namespace{
				ObjectMeta: kube_meta.ObjectMeta{
					Name: pod.Namespace,
				},
			}

			// other services
			var serviceGetter kube_client.Reader
			if given.otherServices != "" {
				bytes, err = os.ReadFile(filepath.Join("testdata", given.otherServices))
				Expect(err).ToNot(HaveOccurred())
				YAMLs := util_yaml.SplitYAML(string(bytes))
				services, err := ParseServices(YAMLs)
				Expect(err).ToNot(HaveOccurred())
				reader, err := newFakeServiceReader(services)
				Expect(err).ToNot(HaveOccurred())
				serviceGetter = reader
			}

			// other dataplanes
			var otherDataplanes []*mesh_k8s.Dataplane
			if given.otherDataplanes != "" {
				bytes, err = os.ReadFile(filepath.Join("testdata", given.otherDataplanes))
				Expect(err).ToNot(HaveOccurred())
				YAMLs := util_yaml.SplitYAML(string(bytes))
				otherDataplanes, err = ParseDataplanes(YAMLs)
				Expect(err).ToNot(HaveOccurred())
			}

			converter := PodConverter{
				ServiceGetter:     serviceGetter,
				Zone:              "zone-1",
				ResourceConverter: k8s.NewSimpleConverter(),
			}

			// when
			dataplane := &mesh_k8s.Dataplane{}
			err = converter.PodToDataplane(context.Background(), dataplane, pod, &namespace, services, otherDataplanes)

			// then
			Expect(err).ToNot(HaveOccurred())

			actual, err := yaml.Marshal(dataplane)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchGoldenYAML("testdata", given.dataplane))
		},
		Entry("01.Pod with 2 Services", testCase{
			pod:            "01.pod.yaml",
			servicesForPod: "01.services-for-pod.yaml",
			dataplane:      "01.dataplane.yaml",
		}),
		Entry("02. Pod with 1 Service and 1 other Dataplane", testCase{
			pod:             "02.pod.yaml",
			servicesForPod:  "02.services-for-pod.yaml",
			otherDataplanes: "02.other-dataplanes.yaml",
			otherServices:   "02.other-services.yaml",
			dataplane:       "02.dataplane.yaml",
		}),
		Entry("03. Pod with gateway annotation and 1 service", testCase{
			pod:            "03.pod.yaml",
			servicesForPod: "03.services-for-pod.yaml",
			dataplane:      "03.dataplane.yaml",
		}),
		Entry("04. Pod with direct access to all services", testCase{
			pod:             "04.pod.yaml",
			servicesForPod:  "04.services-for-pod.yaml",
			otherDataplanes: "04.other-dataplanes.yaml",
			otherServices:   "04.other-services.yaml",
			dataplane:       "04.dataplane.yaml",
		}),
		Entry("05. Pod with direct access to chosen services", testCase{
			pod:             "05.pod.yaml",
			servicesForPod:  "05.services-for-pod.yaml",
			otherDataplanes: "05.other-dataplanes.yaml",
			otherServices:   "05.other-services.yaml",
			dataplane:       "05.dataplane.yaml",
		}),
		Entry("06. Pod with headless service and communication to headless services", testCase{
			pod:             "06.pod.yaml",
			servicesForPod:  "06.services-for-pod.yaml",
			otherDataplanes: "06.other-dataplanes.yaml",
			otherServices:   "06.other-services.yaml",
			dataplane:       "06.dataplane.yaml",
		}),
		Entry("07. Pod with metrics override", testCase{
			pod:            "07.pod.yaml",
			servicesForPod: "07.services-for-pod.yaml",
			dataplane:      "07.dataplane.yaml",
		}),
		Entry("08. Pod with transparent proxy enabled, without direct access servies", testCase{
			pod:            "08.pod.yaml",
			servicesForPod: "08.services-for-pod.yaml",
			dataplane:      "08.dataplane.yaml",
		}),
		Entry("09. Pod with Kuma Ingress", testCase{
			pod:            "09.pod.yaml",
			servicesForPod: "09.services-for-pod.yaml",
			dataplane:      "09.dataplane.yaml",
		}),
		Entry("10. Pod probes", testCase{
			pod:            "10.pod.yaml",
			servicesForPod: "10.services-for-pod.yaml",
			dataplane:      "10.dataplane.yaml",
		}),
		Entry("11. Pod with several containers", testCase{
			pod:            "11.pod.yaml",
			servicesForPod: "11.services-for-pod.yaml",
			dataplane:      "11.dataplane.yaml",
		}),
		Entry("11. Pod with several containers", testCase{
			pod:            "11.pod.yaml",
			servicesForPod: "11.services-for-pod.yaml",
			dataplane:      "11.dataplane.yaml",
		}),
		Entry("12. Pod with kuma-sidecar is not ready", testCase{
			pod:            "12.pod.yaml",
			servicesForPod: "12.services-for-pod.yaml",
			dataplane:      "12.dataplane.yaml",
		}),
		Entry("13. Pod without a service", testCase{
			pod:       "13.pod.yaml",
			dataplane: "13.dataplane.yaml",
		}),
		Entry("14. Gateway pod without a service", testCase{
			pod:       "14.pod.yaml",
			dataplane: "14.dataplane.yaml",
		}),
		Entry("15. Pod with transparent proxy enabled, IPv6 and without direct access servies", testCase{
			pod:            "15.pod.yaml",
			servicesForPod: "15.services-for-pod.yaml",
			dataplane:      "15.dataplane.yaml",
		}),
		Entry("16. Pod with Kuma Egress", testCase{
			pod:            "16.pod.yaml",
			servicesForPod: "16.services-for-pod.yaml",
			dataplane:      "16.dataplane.yaml",
		}),
		Entry("17. Pod with reachable services", testCase{
			pod:             "17.pod.yaml",
			servicesForPod:  "17.services-for-pod.yaml",
			otherDataplanes: "17.other-dataplanes.yaml",
			otherServices:   "17.other-services.yaml",
			dataplane:       "17.dataplane.yaml",
		}),
	)

	DescribeTable("should convert Ingress Pod into an Ingress Dataplane YAML version",
		func(given testCase) {
			// given
			// pod
			pod := &kube_core.Pod{}
			bytes, err := os.ReadFile(filepath.Join("testdata", "ingress", given.pod))
			Expect(err).ToNot(HaveOccurred())
			err = yaml.Unmarshal(bytes, pod)
			Expect(err).ToNot(HaveOccurred())

			// services for pod
			bytes, err = os.ReadFile(filepath.Join("testdata", "ingress", given.servicesForPod))
			Expect(err).ToNot(HaveOccurred())
			YAMLs := util_yaml.SplitYAML(string(bytes))
			services, err := ParseServices(YAMLs)
			Expect(err).ToNot(HaveOccurred())

			// node
			var nodeGetter kube_client.Reader
			if given.node != "" {
				bytes, err = os.ReadFile(filepath.Join("testdata", "ingress", given.node))
				Expect(err).ToNot(HaveOccurred())
				nodeGetter = fakeNodeReader(bytes)
			}

			converter := PodConverter{
				ServiceGetter: nil,
				NodeGetter:    nodeGetter,
				Zone:          "zone-1",
			}

			// when
			ingress := &mesh_k8s.ZoneIngress{}
			err = converter.PodToIngress(context.Background(), ingress, pod, services)

			// then
			Expect(err).ToNot(HaveOccurred())

			actual, err := yaml.Marshal(ingress)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "ingress", given.dataplane)))
		},
		Entry("01. Ingress with load balancer service and hostname", testCase{ // AWS use case
			pod:            "01.pod.yaml",
			servicesForPod: "01.services-for-pod.yaml",
			dataplane:      "01.dataplane.yaml",
		}),
		Entry("02. Ingress with load balancer and ip", testCase{ // GCP use case
			pod:            "02.pod.yaml",
			servicesForPod: "02.services-for-pod.yaml",
			dataplane:      "02.dataplane.yaml",
		}),
		Entry("03. Ingress with load balancer without public ip", testCase{
			pod:            "03.pod.yaml",
			servicesForPod: "03.services-for-pod.yaml",
			dataplane:      "03.dataplane.yaml",
		}),
		Entry("04. Ingress with node port external IP", testCase{ // Real deployment use case
			pod:            "04.pod.yaml",
			servicesForPod: "04.services-for-pod.yaml",
			dataplane:      "04.dataplane.yaml",
			node:           "04.node.yaml",
		}),
		Entry("05. Ingress with node port internal IP", testCase{ // KIND / Minikube use case
			pod:            "05.pod.yaml",
			servicesForPod: "05.services-for-pod.yaml",
			dataplane:      "05.dataplane.yaml",
			node:           "05.node.yaml",
		}),
		Entry("06. Ingress with annotations override", testCase{
			pod:            "06.pod.yaml",
			servicesForPod: "06.services-for-pod.yaml",
			dataplane:      "06.dataplane.yaml",
		}),
	)

	DescribeTable("should convert Egress Pod into an Egress Dataplane YAML version",
		func(given testCase) {
			// given
			// pod
			pod := &kube_core.Pod{}
			bytes, err := os.ReadFile(filepath.Join("testdata", "egress", given.pod))
			Expect(err).ToNot(HaveOccurred())
			err = yaml.Unmarshal(bytes, pod)
			Expect(err).ToNot(HaveOccurred())

			// services for pod
			bytes, err = os.ReadFile(filepath.Join("testdata", "egress", given.servicesForPod))
			Expect(err).ToNot(HaveOccurred())
			YAMLs := util_yaml.SplitYAML(string(bytes))
			services, err := ParseServices(YAMLs)
			Expect(err).ToNot(HaveOccurred())

			// node
			var nodeGetter kube_client.Reader
			if given.node != "" {
				bytes, err = os.ReadFile(filepath.Join("testdata", "egress", given.node))
				Expect(err).ToNot(HaveOccurred())
				nodeGetter = fakeNodeReader(bytes)
			}

			converter := PodConverter{
				ServiceGetter: nil,
				NodeGetter:    nodeGetter,
				Zone:          "zone-1",
			}

			// when
			egress := &mesh_k8s.ZoneEgress{}
			err = converter.PodToEgress(egress, pod, services)

			// then
			Expect(err).ToNot(HaveOccurred())

			actual, err := yaml.Marshal(egress)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "egress", given.dataplane)))
		},
		Entry("01. Egress with load balancer service and hostname", testCase{ // AWS use case
			pod:            "01.pod.yaml",
			servicesForPod: "01.services-for-pod.yaml",
			dataplane:      "01.dataplane.yaml",
		}),
		Entry("02. Egress with load balancer and ip", testCase{ // GCP use case
			pod:            "02.pod.yaml",
			servicesForPod: "02.services-for-pod.yaml",
			dataplane:      "02.dataplane.yaml",
		}),
		Entry("03. Egress with load balancer without public ip", testCase{
			pod:            "03.pod.yaml",
			servicesForPod: "03.services-for-pod.yaml",
			dataplane:      "03.dataplane.yaml",
		}),
		Entry("04. Egress with node port external IP", testCase{ // Real deployment use case
			pod:            "04.pod.yaml",
			servicesForPod: "04.services-for-pod.yaml",
			dataplane:      "04.dataplane.yaml",
			node:           "04.node.yaml",
		}),
		Entry("05. Egress with node port internal IP", testCase{ // KIND / Minikube use case
			pod:            "05.pod.yaml",
			servicesForPod: "05.services-for-pod.yaml",
			dataplane:      "05.dataplane.yaml",
			node:           "05.node.yaml",
		}),
	)

	Context("when Dataplane cannot be generated", func() {
		type testCase struct {
			pod         string
			services    []string
			expectedErr string
		}

		DescribeTable("should return a descriptive error",
			func(given testCase) {
				// given
				converter := PodConverter{}

				pod := &kube_core.Pod{}
				err := yaml.Unmarshal([]byte(given.pod), pod)
				Expect(err).ToNot(HaveOccurred())

				namespace := kube_core.Namespace{
					ObjectMeta: kube_meta.ObjectMeta{
						Name: pod.Namespace,
					},
				}

				services, err := ParseServices(given.services)
				Expect(err).ToNot(HaveOccurred())

				dataplane := &mesh_k8s.Dataplane{}

				// when
				err = converter.PodToDataplane(context.Background(), dataplane, pod, &namespace, services, nil)

				// then
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(given.expectedErr))
			},
			Entry("Pod with a Service but mismatching ports", testCase{
				pod: pod,
				services: []string{`
                spec:
                  clusterIP: 192.168.0.1
                  ports:
                  - protocol: UDP    # all non-TCP ports should be ignored
                    port: 80
                    targetPort: 8080
                  - protocol: SCTP   # all non-TCP ports should be ignored
                    port: 443
                    targetPort: 8443
                  - protocol: TCP
                    port: 7070
                    targetPort: api
                  - # defaults to TCP protocol
                    port: 6060
                    targetPort: diagnostics
`},
				expectedErr: `A service that selects pod example was found, but it doesn't match any container ports.`,
			}),
		)
	})
})

var _ = Describe("InboundTagsForService(..)", func() {

	type testCase struct {
		isGateway      bool
		zone           string
		podLabels      map[string]string
		svcAnnotations map[string]string
		appProtocol    *string
		expected       map[string]string
	}

	DescribeTable("should combine Pod's labels with Service's FQDN and port",
		func(given testCase) {
			// given
			pod := &kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Labels:    given.podLabels,
				},
			}
			// and
			svc := &kube_core.Service{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "example",
					Labels: map[string]string{
						"more": "labels",
					},
					Annotations: given.svcAnnotations,
				},
				Spec: kube_core.ServiceSpec{
					Ports: []kube_core.ServicePort{
						{
							Name:        "http",
							Port:        80,
							AppProtocol: given.appProtocol,
							TargetPort: kube_intstr.IntOrString{
								Type:   kube_intstr.Int,
								IntVal: 8080,
							},
						},
					},
				},
			}

			// expect
			Expect(InboundTagsForService(given.zone, pod, svc, &svc.Spec.Ports[0])).To(Equal(given.expected))
		},
		Entry("Pod without labels", testCase{
			isGateway: false,
			podLabels: nil,
			expected: map[string]string{
				"kuma.io/service":          "example_demo_svc_80",
				"kuma.io/protocol":         "tcp", // we want Kuma's default behavior to be explicit to a user
				"k8s.kuma.io/service-name": "example",
				"k8s.kuma.io/service-port": "80",
				"k8s.kuma.io/namespace":    "demo",
			},
		}),
		Entry("Pod with labels", testCase{
			isGateway: false,
			podLabels: map[string]string{
				"app":     "example",
				"version": "0.1",
			},
			expected: map[string]string{
				"app":                      "example",
				"version":                  "0.1",
				"kuma.io/service":          "example_demo_svc_80",
				"kuma.io/protocol":         "tcp", // we want Kuma's default behavior to be explicit to a user
				"k8s.kuma.io/service-name": "example",
				"k8s.kuma.io/service-port": "80",
				"k8s.kuma.io/namespace":    "demo",
			},
		}),
		Entry("Pod with `service` label", testCase{
			isGateway: false,
			podLabels: map[string]string{
				"kuma.io/service": "something",
				"app":             "example",
				"version":         "0.1",
			},
			expected: map[string]string{
				"app":                      "example",
				"version":                  "0.1",
				"kuma.io/service":          "example_demo_svc_80",
				"kuma.io/protocol":         "tcp", // we want Kuma's default behavior to be explicit to a user
				"k8s.kuma.io/service-name": "example",
				"k8s.kuma.io/service-port": "80",
				"k8s.kuma.io/namespace":    "demo",
			},
		}),
		Entry("Service with a `<port>.service.kuma.io/protocol` annotation and an unknown value", testCase{
			isGateway: false,
			podLabels: map[string]string{
				"app":     "example",
				"version": "0.1",
			},
			svcAnnotations: map[string]string{
				"80.service.kuma.io/protocol": "not-yet-supported-protocol",
			},
			expected: map[string]string{
				"app":                      "example",
				"version":                  "0.1",
				"kuma.io/service":          "example_demo_svc_80",
				"kuma.io/protocol":         "not-yet-supported-protocol", // we want Kuma's behavior to be straightforward to a user (just copy annotation value "as is")
				"k8s.kuma.io/service-name": "example",
				"k8s.kuma.io/service-port": "80",
				"k8s.kuma.io/namespace":    "demo",
			},
		}),
		Entry("Service with a `<port>.service.kuma.io/protocol` annotation and a known value", testCase{
			isGateway: false,
			podLabels: map[string]string{
				"app":     "example",
				"version": "0.1",
			},
			svcAnnotations: map[string]string{
				"80.service.kuma.io/protocol": "http",
			},
			expected: map[string]string{
				"app":                      "example",
				"version":                  "0.1",
				"kuma.io/service":          "example_demo_svc_80",
				"kuma.io/protocol":         "http",
				"k8s.kuma.io/service-name": "example",
				"k8s.kuma.io/service-port": "80",
				"k8s.kuma.io/namespace":    "demo",
			},
		}),
		Entry("Service with appProtocol and a known value", testCase{
			isGateway: false,
			podLabels: map[string]string{
				"app":     "example",
				"version": "0.1",
			},
			appProtocol: utilpointer.StringPtr("http"),
			expected: map[string]string{
				"app":                      "example",
				"version":                  "0.1",
				"kuma.io/service":          "example_demo_svc_80",
				"kuma.io/protocol":         "http",
				"k8s.kuma.io/service-name": "example",
				"k8s.kuma.io/service-port": "80",
				"k8s.kuma.io/namespace":    "demo",
			},
		}),
		Entry("Inject a zone tag if Zone is set", testCase{
			isGateway: false,
			zone:      "zone-1",
			podLabels: map[string]string{
				"app":     "example",
				"version": "0.1",
			},
			expected: map[string]string{
				"app":                      "example",
				"version":                  "0.1",
				mesh_proto.ServiceTag:      "example_demo_svc_80",
				mesh_proto.ZoneTag:         "zone-1",
				mesh_proto.ProtocolTag:     "tcp",
				"k8s.kuma.io/service-name": "example",
				"k8s.kuma.io/service-port": "80",
				"k8s.kuma.io/namespace":    "demo",
			},
		}),
		Entry("Pod with empty labels", testCase{
			isGateway: true,
			podLabels: map[string]string{
				"app":     "example",
				"version": "",
			},
			expected: map[string]string{
				"app":                      "example",
				"kuma.io/service":          "example_demo_svc_80",
				"kuma.io/protocol":         "tcp",
				"k8s.kuma.io/service-name": "example",
				"k8s.kuma.io/service-port": "80",
				"k8s.kuma.io/namespace":    "demo",
			},
		}),
	)
})

var _ = Describe("ProtocolTagFor(..)", func() {

	type testCase struct {
		appProtocol *string
		annotations map[string]string
		expected    string
	}

	DescribeTable("should infer protocol from `appProtocol` or `<port>.service.kuma.io/protocol` field",
		func(given testCase) {
			// given
			svc := &kube_core.Service{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace:   "demo",
					Name:        "example",
					Annotations: given.annotations,
				},
				Spec: kube_core.ServiceSpec{
					Ports: []kube_core.ServicePort{
						{
							Name: "http",
							Port: 80,
							TargetPort: kube_intstr.IntOrString{
								Type:   kube_intstr.Int,
								IntVal: 8080,
							},
							AppProtocol: given.appProtocol,
						},
					},
				},
			}

			// expect
			Expect(ProtocolTagFor(svc, &svc.Spec.Ports[0])).To(Equal(given.expected))
		},
		Entry("no appProtocol", testCase{
			appProtocol: nil,
			expected:    "tcp", // we want Kuma's default behavior to be explicit to a user
		}),
		Entry("appProtocol has an empty value", testCase{
			appProtocol: utilpointer.StringPtr(""),
			expected:    "tcp", // we want Kuma's default behavior to be explicit to a user
		}),
		Entry("no appProtocol but with `<port>.service.kuma.io/protocol` annotation", testCase{
			appProtocol: nil,
			annotations: map[string]string{
				"80.service.kuma.io/protocol": "http",
			},
			expected: "http", // we want to support both ways of providing protocol
		}),
		Entry("appProtocol has an unknown value", testCase{
			appProtocol: utilpointer.StringPtr("not-yet-supported-protocol"),
			expected:    "not-yet-supported-protocol", // we want Kuma's behavior to be straightforward to a user (just copy appProtocol value "as is")
		}),
		Entry("appProtocol has a non-lowercase value", testCase{
			appProtocol: utilpointer.StringPtr("HtTp"),
			expected:    "HtTp", // we want Kuma's behavior to be straightforward to a user (just copy appProtocol value "as is")
		}),
		Entry("appProtocol has a known value: http", testCase{
			appProtocol: utilpointer.StringPtr("http"),
			expected:    "http",
		}),
		Entry("appProtocol has a known value: tcp", testCase{
			appProtocol: utilpointer.StringPtr("tcp"),
			expected:    "tcp",
		}),
		Entry("no appProtocol and no `<port>.service.kuma.io/protocol`", testCase{
			appProtocol: nil,
			annotations: nil,
			expected:    "tcp",
		}),
	)
})

type fakeServiceReader map[string]string

func newFakeServiceReader(services []*kube_core.Service) (fakeServiceReader, error) {
	servicesMap := map[string]string{}
	for _, service := range services {
		bytes, err := yaml.Marshal(service)
		if err != nil {
			return nil, err
		}
		servicesMap[service.GetNamespace()+"/"+service.GetName()] = string(bytes)
	}
	return servicesMap, nil
}

var _ kube_client.Reader = fakeServiceReader{}

func (r fakeServiceReader) Get(ctx context.Context, key kube_client.ObjectKey, obj kube_client.Object) error {
	data, ok := r[fmt.Sprintf("%s/%s", key.Namespace, key.Name)]
	if !ok {
		return errors.New("not found")
	}
	return yaml.Unmarshal([]byte(data), obj)
}

func (f fakeServiceReader) List(ctx context.Context, list kube_client.ObjectList, opts ...kube_client.ListOption) error {
	return errors.New("not implemented")
}

type fakeNodeReader string

func (r fakeNodeReader) Get(ctx context.Context, key kube_client.ObjectKey, obj kube_client.Object) error {
	return errors.New("not implemented")
}

func (f fakeNodeReader) List(ctx context.Context, list kube_client.ObjectList, opts ...kube_client.ListOption) error {
	node := kube_core.Node{}
	err := yaml.Unmarshal([]byte(f), &node)
	if err != nil {
		return err
	}
	l := list.(*kube_core.NodeList)
	l.Items = append(l.Items, node)
	return nil
}
