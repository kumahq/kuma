package controllers_test

import (
	"context"
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	. "github.com/Kong/kuma/pkg/plugins/discovery/k8s/controllers"

	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"

	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	kube_intstr "k8s.io/apimachinery/pkg/util/intstr"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("PodToDataplane(..)", func() {

	type testCase struct {
		pod           string
		services      []string
		others        []string
		serviceGetter fakeReader
		expected      string
	}

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

	DescribeTable("should convert Pod into a Dataplane",
		func(given testCase) {
			// given
			pod := &kube_core.Pod{}
			// when
			err := yaml.Unmarshal([]byte(given.pod), pod)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			services, err := ParseServices(given.services)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			others, err := ParseDataplanes(given.others)
			// then
			Expect(err).ToNot(HaveOccurred())

			// given
			dataplane := &mesh_k8s.Dataplane{}

			// when
			err = PodToDataplane(dataplane, pod, services, others, given.serviceGetter)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := json.Marshal(dataplane)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("Pod with 2 Services", testCase{
			pod: pod,
			services: []string{
				`
                metadata:
                  namespace: demo
                  name: example
                  annotations:
                    80.service.kuma.io/protocol: http
                spec:
                  ports:
                  - # protocol defaults to TCP
                    port: 80
                    targetPort: 8080
                  - protocol: TCP
                    port: 443
                    targetPort: 8443
`,
				`
                metadata:
                  namespace: playground
                  name: sample
                  annotations:
                    7071.service.kuma.io/protocol: TCP
                spec:
                  ports:
                  - protocol: TCP
                    port: 7071
                    targetPort: 7070
                  - protocol: TCP
                    port: 6061
                    targetPort: metrics
`,
			},
			expected: `
            mesh: default
            metadata:
              creationTimestamp: null
            spec:
              networking:
                address: 192.168.0.1
                inbound:
                - port: 8080
                  tags:
                    app: example
                    protocol: http
                    service: example.demo.svc:80
                    version: "0.1"
                - port: 8443
                  tags:
                    app: example
                    protocol: tcp
                    service: example.demo.svc:443
                    version: "0.1"
                - port: 7070
                  tags:
                    app: example
                    protocol: tcp
                    service: sample.playground.svc:7071
                    version: "0.1"
                - port: 6060
                  tags:
                    app: example
                    protocol: tcp
                    service: sample.playground.svc:6061
                    version: "0.1"
`,
		}),
		Entry("Pod with 1 Service and 1 other Dataplane", testCase{
			pod: pod,
			services: []string{`
            metadata:
              namespace: demo
              name: example
            spec:
              ports:
              - # protocol defaults to TCP
                port: 80
                targetPort: 8080
`,
			},
			others: []string{`
            apiVersion: kuma.io/v1alpha1
            kind: Dataplane
            mesh: default
            metadata:
              name: test-app-8646b8bbc8-5qbl2
              namespace: playground
            spec:
              networking:
                address: 10.244.0.25
                inbound:
                - port: 80
                  tags:
                    app: test-app
                    pod-template-hash: 8646b8bbc8
                    service: test-app.playground.svc:80
                - port: 443
                  tags:
                    app: test-app
                    pod-template-hash: 8646b8bbc8
                    service: test-app.playground.svc:443
                transparentProxying:
                  redirectPort: 15001
`,
			},
			serviceGetter: fakeReader{
				"playground/test-app": `
                    apiVersion: v1
                    kind: Service
                    metadata:
                      name: test-app
                      namespace: playground
                    spec:
                      clusterIP: 10.108.144.24
                      ports:
                      - name: http
                        port: 80
                        protocol: TCP
                        targetPort: 80
                      - name: https
                        port: 443
                        protocol: TCP
                        targetPort: 80
                      selector:
                        app: test-app
                      sessionAffinity: None
                      type: ClusterIP
                    status:
                      loadBalancer: {}
`,
			},
			expected: `
            mesh: default
            metadata:
              creationTimestamp: null
            spec:
              networking:
                address: 192.168.0.1
                inbound:
                - port: 8080
                  tags:
                    app: example
                    protocol: tcp
                    service: example.demo.svc:80
                    version: "0.1"
                outbound:
                - address: 10.108.144.24
                  port: 443
                  service: test-app.playground.svc:443
                - address: 10.108.144.24
                  port: 80
                  service: test-app.playground.svc:80
`,
		}),
		Entry("pod with gateway annotation and 1 service", testCase{
			pod: gatewayPod,
			services: []string{`
            metadata:
              namespace: demo
              name: example
              annotations:
                80.service.kuma.io/protocol: http # should be ignored in case of a gateway
            spec:
              ports:
              - # protocol defaults to TCP
                port: 80
                targetPort: 8080
              - protocol: TCP
                port: 443
                targetPort: 8443
`},
			expected: `
            mesh: default
            metadata:
              creationTimestamp: null
            spec:
              networking:
                address: 192.168.0.1
                gateway:
                  tags:
                    app: example
                    service: example.demo.svc:80
                    version: "0.1"
`,
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
				pod := &kube_core.Pod{}
				// when
				err := yaml.Unmarshal([]byte(given.pod), pod)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				services, err := ParseServices(given.services)
				// then
				Expect(err).ToNot(HaveOccurred())

				// given
				dataplane := &mesh_k8s.Dataplane{}

				// when
				err = PodToDataplane(dataplane, pod, services, nil, nil)
				// then
				Expect(err).To(HaveOccurred())
				// and
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
			Entry("Pod with a Service and invalid `<port>.service.kuma.io/protocol` annotation", testCase{
				pod: pod,
				services: []string{`
                metadata:
                  namespace: demo
                  name: example
                  annotations:
                    80.service.kuma.io/protocol: MONGO
                spec:
                  ports:
                  - # protocol defaults to TCP
                    port: 80
                    targetPort: 8080
                  - protocol: TCP
                    port: 443
                    targetPort: 8443
`},
				expectedErr: `metadata.annotations["80.service.kuma.io/protocol"]: value "MONGO" is not valid. Allowed values: http, tcp`,
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

			// when
			actual, err := InboundTagsFor(pod, svc, &svc.Spec.Ports[0], given.isGateway)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(Equal(given.expected))
		},
		Entry("Pod without labels", testCase{
			isGateway: false,
			podLabels: nil,
			expected: map[string]string{
				"service":  "example.demo.svc:80",
				"protocol": "tcp", // we want Kuma's default behaviour to be explicit to a user
			},
		}),
		Entry("Pod with labels", testCase{
			isGateway: false,
			podLabels: map[string]string{
				"app":     "example",
				"version": "0.1",
			},
			expected: map[string]string{
				"app":      "example",
				"version":  "0.1",
				"service":  "example.demo.svc:80",
				"protocol": "tcp", // we want Kuma's default behaviour to be explicit to a user
			},
		}),
		Entry("Pod with `service` label", testCase{
			isGateway: false,
			podLabels: map[string]string{
				"service": "something",
				"app":     "example",
				"version": "0.1",
			},
			expected: map[string]string{
				"app":      "example",
				"version":  "0.1",
				"service":  "example.demo.svc:80",
				"protocol": "tcp", // we want Kuma's default behaviour to be explicit to a user
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
				"app":      "example",
				"version":  "0.1",
				"service":  "example.demo.svc:80",
				"protocol": "http",
			},
		}),
		Entry("`gateway` Pod should not have a `protocol` tag", testCase{
			isGateway: true,
			podLabels: map[string]string{
				"app":     "example",
				"version": "0.1",
			},
			expected: map[string]string{
				"app":     "example",
				"version": "0.1",
				"service": "example.demo.svc:80",
			},
		}),
		Entry("`gateway` Pod should not have a `protocol` tag even if `<port>.service.kuma.io/protocol` annotation is present", testCase{
			isGateway: true,
			podLabels: map[string]string{
				"app":     "example",
				"version": "0.1",
			},
			svcAnnotations: map[string]string{
				"80.service.kuma.io/protocol": "http",
			},
			expected: map[string]string{
				"app":     "example",
				"version": "0.1",
				"service": "example.demo.svc:80",
			},
		}),
	)

	Context("when InboundTagsFor() cannot be generated", func() {

		type testCase struct {
			isGateway      bool
			podLabels      map[string]string
			svcAnnotations map[string]string
			expectedErr    string
		}

		DescribeTable("should return a descriptive error",
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

				// when
				actual, err := InboundTagsFor(pod, svc, &svc.Spec.Ports[0], given.isGateway)
				// then
				Expect(err).To(HaveOccurred())
				// and
				Expect(err.Error()).To(Equal(given.expectedErr))
				// and
				Expect(actual).To(BeNil())
			},
			Entry("Service with a `<port>.service.kuma.io/protocol` annotation and an empty value", testCase{
				isGateway: false,
				podLabels: map[string]string{
					"app":     "example",
					"version": "0.1",
				},
				svcAnnotations: map[string]string{
					"80.service.kuma.io/protocol": "",
				},
				expectedErr: `metadata.annotations["80.service.kuma.io/protocol"]: value "" is not valid. Allowed values: http, tcp`,
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
				expectedErr: `metadata.annotations["80.service.kuma.io/protocol"]: value "not-yet-supported-protocol" is not valid. Allowed values: http, tcp`,
			}),
		)
	})
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
		Expect(ServiceTagFor(svc, &svc.Spec.Ports[0])).To(Equal("example.demo.svc:80"))
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

			// when
			actual, err := ProtocolTagFor(svc, &svc.Spec.Ports[0])
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(Equal(given.expected))
		},
		Entry("no `<port>.service.kuma.io/protocol` annotation", testCase{
			annotations: nil,
			expected:    "tcp", // we want Kuma's default behaviour to be explicit to a user
		}),
		Entry("`<port>.service.kuma.io/protocol` annotation is for a different port", testCase{
			annotations: map[string]string{
				"8080.service.kuma.io/protocol": "http",
			},
			expected: "tcp", // we want Kuma's default behaviour to be explicit to a user
		}),
		Entry("`<port>.service.kuma.io/protocol` annotation has a non-lowercase value", testCase{
			annotations: map[string]string{
				"80.service.kuma.io/protocol": "HtTp",
			},
			expected: "http", // we want Kuma's behaviour to be user-friendly
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

	Context("when ProtocolTagFor() cannot be generated", func() {

		type testCase struct {
			annotations map[string]string
			expectedErr string
		}

		DescribeTable("should return a descriptive error",
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

				// when
				actual, err := ProtocolTagFor(svc, &svc.Spec.Ports[0])
				// then
				Expect(err).To(HaveOccurred())
				// and
				Expect(err.Error()).To(Equal(given.expectedErr))
				// and
				Expect(actual).To(Equal(""))
			},
			Entry("`<port>.service.kuma.io/protocol` annotation has an empty value", testCase{
				annotations: map[string]string{
					"80.service.kuma.io/protocol": "",
				},
				expectedErr: `metadata.annotations["80.service.kuma.io/protocol"]: value "" is not valid. Allowed values: http, tcp`,
			}),
			Entry("`<port>.service.kuma.io/protocol` annotation has an unknown value", testCase{
				annotations: map[string]string{
					"80.service.kuma.io/protocol": "not-yet-supported-protocol",
				},
				expectedErr: `metadata.annotations["80.service.kuma.io/protocol"]: value "not-yet-supported-protocol" is not valid. Allowed values: http, tcp`,
			}),
		)
	})
})

type fakeReader map[string]string

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
