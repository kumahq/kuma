package controllers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	. "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers"

	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	util_yaml "github.com/kumahq/kuma/pkg/util/yaml"

	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_intstr "k8s.io/apimachinery/pkg/util/intstr"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
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

	gatewayPod := `
    metadata:
      namespace: demo
      name: example
      labels:
        app: example
        version: "0.1"
      annotations:
        kuma.io/gateway: enabled
    spec:
      containers:
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
		dataplane       string
	}
	DescribeTable("should convert Pod into a Dataplane YAML version",
		func(given testCase) {
			// given
			// pod
			pod := &kube_core.Pod{}
			bytes, err := ioutil.ReadFile(filepath.Join("testdata", given.pod))
			Expect(err).ToNot(HaveOccurred())
			err = yaml.Unmarshal(bytes, pod)
			Expect(err).ToNot(HaveOccurred())

			// services for pod
			bytes, err = ioutil.ReadFile(filepath.Join("testdata", given.servicesForPod))
			Expect(err).ToNot(HaveOccurred())
			YAMLs := util_yaml.SplitYAML(string(bytes))
			services, err := ParseServices(YAMLs)
			Expect(err).ToNot(HaveOccurred())

			// other services
			var serviceGetter kube_client.Reader
			if given.otherServices != "" {
				bytes, err = ioutil.ReadFile(filepath.Join("testdata", given.otherServices))
				Expect(err).ToNot(HaveOccurred())
				YAMLs := util_yaml.SplitYAML(string(bytes))
				services, err := ParseServices(YAMLs)
				Expect(err).ToNot(HaveOccurred())
				reader, err := newFakeReader(services)
				Expect(err).ToNot(HaveOccurred())
				serviceGetter = reader
			}

			// other dataplanes
			var otherDataplanes []*mesh_k8s.Dataplane
			if given.otherDataplanes != "" {
				bytes, err = ioutil.ReadFile(filepath.Join("testdata", given.otherDataplanes))
				Expect(err).ToNot(HaveOccurred())
				YAMLs := util_yaml.SplitYAML(string(bytes))
				otherDataplanes, err = ParseDataplanes(YAMLs)
				Expect(err).ToNot(HaveOccurred())
			}

			converter := PodConverter{
				ServiceGetter: serviceGetter,
				Zone:          "zone-1",
			}

			// when
			dataplane := &mesh_k8s.Dataplane{}
			err = converter.PodToDataplane(dataplane, pod, services, otherDataplanes)

			// then
			Expect(err).ToNot(HaveOccurred())

			actual, err := json.Marshal(dataplane)
			Expect(err).ToNot(HaveOccurred())
			expected, err := ioutil.ReadFile(filepath.Join("testdata", given.dataplane))
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(expected))
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

				services, err := ParseServices(given.services)
				Expect(err).ToNot(HaveOccurred())

				dataplane := &mesh_k8s.Dataplane{}

				// when
				err = converter.PodToDataplane(dataplane, pod, services, nil)

				// then
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(given.expectedErr))
			},
			Entry("regular Pod without Services", testCase{
				pod:         pod,
				services:    nil,
				expectedErr: `Kuma requires every Pod in a Mesh to be a part of at least one Service. However, there are no Services that select this Pod.`,
			}),
			Entry("gateway Pod without Services", testCase{
				pod:         gatewayPod,
				services:    nil,
				expectedErr: `Kuma requires every Pod in a Mesh to be a part of at least one Service. However, there are no Services that select this Pod.`,
			}),
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
				expectedErr: `Kuma requires every Pod in a Mesh to be a part of at least one Service. However, this Pod doesn't have any container ports that would satisfy matching Service(s).`,
			}),
		)
	})
})

var _ = Describe("MeshFor(..)", func() {

	type testCase struct {
		podAnnotations map[string]string
		expected       string
	}

	DescribeTable("should use value of `kuma.io/mesh` annotation on a Pod or fallback to the `default` Mesh",
		func(given testCase) {
			// given
			pod := &kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Annotations: given.podAnnotations,
				},
			}

			// then
			Expect(MeshFor(pod)).To(Equal(given.expected))
		},
		Entry("Pod without annotations", testCase{
			podAnnotations: nil,
			expected:       "default",
		}),
		Entry("Pod with empty `kuma.io/mesh` annotation", testCase{
			podAnnotations: map[string]string{
				"kuma.io/mesh": "",
			},
			expected: "default",
		}),
		Entry("Pod with non-empty `kuma.io/mesh` annotation", testCase{
			podAnnotations: map[string]string{
				"kuma.io/mesh": "demo",
			},
			expected: "demo",
		}),
	)
})

var _ = Describe("InboundTagsFor(..)", func() {

	type testCase struct {
		isGateway      bool
		zone           string
		podLabels      map[string]string
		svcAnnotations map[string]string
		expected       map[string]string
	}

	DescribeTable("should combine Pod's labels with Service's FQDN and port",
		func(given testCase) {
			// given
			pod := &kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Labels: given.podLabels,
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
							Name: "http",
							Port: 80,
							TargetPort: kube_intstr.IntOrString{
								Type:   kube_intstr.Int,
								IntVal: 8080,
							},
						},
					},
				},
			}

			// expect
			Expect(InboundTagsFor(given.zone, pod, svc, &svc.Spec.Ports[0])).To(Equal(given.expected))
		},
		Entry("Pod without labels", testCase{
			isGateway: false,
			podLabels: nil,
			expected: map[string]string{
				"kuma.io/service":  "example_demo_svc_80",
				"kuma.io/protocol": "tcp", // we want Kuma's default behaviour to be explicit to a user
			},
		}),
		Entry("Pod with labels", testCase{
			isGateway: false,
			podLabels: map[string]string{
				"app":     "example",
				"version": "0.1",
			},
			expected: map[string]string{
				"app":              "example",
				"version":          "0.1",
				"kuma.io/service":  "example_demo_svc_80",
				"kuma.io/protocol": "tcp", // we want Kuma's default behaviour to be explicit to a user
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
				"app":              "example",
				"version":          "0.1",
				"kuma.io/service":  "example_demo_svc_80",
				"kuma.io/protocol": "tcp", // we want Kuma's default behaviour to be explicit to a user
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
				"app":              "example",
				"version":          "0.1",
				"kuma.io/service":  "example_demo_svc_80",
				"kuma.io/protocol": "not-yet-supported-protocol", // we want Kuma's behaviour to be straightforward to a user (just copy annotation value "as is")
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
				"app":              "example",
				"version":          "0.1",
				"kuma.io/service":  "example_demo_svc_80",
				"kuma.io/protocol": "http",
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
				"app":                  "example",
				"version":              "0.1",
				mesh_proto.ServiceTag:  "example_demo_svc_80",
				mesh_proto.ZoneTag:     "zone-1",
				mesh_proto.ProtocolTag: "tcp",
			},
		}),
		Entry("Pod with empty labels", testCase{
			isGateway: true,
			podLabels: map[string]string{
				"app":     "example",
				"version": "",
			},
			expected: map[string]string{
				"app":              "example",
				"kuma.io/service":  "example_demo_svc_80",
				"kuma.io/protocol": "tcp",
			},
		}),
	)
})

var _ = Describe("ServiceTagFor(..)", func() {
	It("should use Service FQDN", func() {
		// given
		svc := &kube_core.Service{
			ObjectMeta: kube_meta.ObjectMeta{
				Namespace: "demo",
				Name:      "example",
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
					},
				},
			},
		}

		// then
		Expect(ServiceTagFor(svc, &svc.Spec.Ports[0])).To(Equal("example_demo_svc_80"))
	})
})

var _ = Describe("ProtocolTagFor(..)", func() {

	type testCase struct {
		annotations map[string]string
		expected    string
	}

	DescribeTable("should infer service protocol from a `<port>.service.kuma.io/protocol` annotation",
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
						},
					},
				},
			}

			// expect
			Expect(ProtocolTagFor(svc, &svc.Spec.Ports[0])).To(Equal(given.expected))
		},
		Entry("no `<port>.service.kuma.io/protocol` annotation", testCase{
			annotations: nil,
			expected:    "tcp", // we want Kuma's default behaviour to be explicit to a user
		}),
		Entry("`<port>.service.kuma.io/protocol` annotation has an empty value", testCase{
			annotations: map[string]string{
				"80.service.kuma.io/protocol": "",
			},
			expected: "tcp", // we want Kuma's default behaviour to be explicit to a user
		}),
		Entry("`<port>.service.kuma.io/protocol` annotation is for a different port", testCase{
			annotations: map[string]string{
				"8080.service.kuma.io/protocol": "http",
			},
			expected: "tcp", // we want Kuma's default behaviour to be explicit to a user
		}),
		Entry("`<port>.service.kuma.io/protocol` annotation has an unknown value", testCase{
			annotations: map[string]string{
				"80.service.kuma.io/protocol": "not-yet-supported-protocol",
			},
			expected: "not-yet-supported-protocol", // we want Kuma's behaviour to be straightforward to a user (just copy annotation value "as is")
		}),
		Entry("`<port>.service.kuma.io/protocol` annotation has a non-lowercase value", testCase{
			annotations: map[string]string{
				"80.service.kuma.io/protocol": "HtTp",
			},
			expected: "HtTp", // we want Kuma's behaviour to be straightforward to a user (just copy annotation value "as is")
		}),
		Entry("`<port>.service.kuma.io/protocol` annotation has a known value: http", testCase{
			annotations: map[string]string{
				"80.service.kuma.io/protocol": "http",
			},
			expected: "http",
		}),
		Entry("`<port>.service.kuma.io/protocol` annotation has a known value: tcp", testCase{
			annotations: map[string]string{
				"80.service.kuma.io/protocol": "tcp",
			},
			expected: "tcp",
		}),
	)
})

type fakeReader map[string]string

func newFakeReader(services []*kube_core.Service) (fakeReader, error) {
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

var _ kube_client.Reader = fakeReader{}

func (r fakeReader) Get(ctx context.Context, key kube_client.ObjectKey, obj kube_runtime.Object) error {
	data, ok := r[fmt.Sprintf("%s/%s", key.Namespace, key.Name)]
	if !ok {
		return errors.New("not found")
	}
	return yaml.Unmarshal([]byte(data), obj)
}

func (f fakeReader) List(ctx context.Context, list kube_runtime.Object, opts ...kube_client.ListOption) error {
	return errors.New("not implemented")
}
