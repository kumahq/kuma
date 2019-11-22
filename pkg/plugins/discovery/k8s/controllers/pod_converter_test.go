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
		pod           *kube_core.Pod
		services      []*kube_core.Service
		others        []string
		serviceGetter fakeReader
		expected      string
	}

	pod := &kube_core.Pod{
		ObjectMeta: kube_meta.ObjectMeta{
			Namespace: "demo",
			Name:      "example",
			Labels: map[string]string{
				"app":     "example",
				"version": "0.1",
			},
		},
		Spec: kube_core.PodSpec{
			Containers: []kube_core.Container{
				{
					Ports: []kube_core.ContainerPort{
						{ContainerPort: 8080},
						{ContainerPort: 8443},
					},
				},
				{
					Ports: []kube_core.ContainerPort{
						{ContainerPort: 7070},
						{ContainerPort: 6060, Name: "metrics"},
					},
				},
			},
		},
		Status: kube_core.PodStatus{
			PodIP: "192.168.0.1",
		},
	}

	ParseDataplanes := func(values []string) ([]*mesh_k8s.Dataplane, error) {
		dataplanes := make([]*mesh_k8s.Dataplane, 0, len(values))
		for _, value := range values {
			dataplane := &mesh_k8s.Dataplane{}
			if err := yaml.Unmarshal([]byte(value), dataplane); err != nil {
				return nil, err
			}
			dataplanes = append(dataplanes, dataplane)
		}
		return dataplanes, nil
	}

	DescribeTable("should convert Pod into a Dataplane",
		func(given testCase) {
			// given
			dataplane := &mesh_k8s.Dataplane{}

			others, err := ParseDataplanes(given.others)
			Expect(err).ToNot(HaveOccurred())

			// when
			err = PodToDataplane(dataplane, given.pod, given.services, others, given.serviceGetter, "kuma-test")
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := json.Marshal(dataplane)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("Pod without Services", testCase{
			pod:      pod,
			services: nil,
			expected: `
            mesh: default.kuma-test
            metadata:
              creationTimestamp: null
            spec:
              networking: {}
`,
		}),
		Entry("Pod with a Service but mismatching ports", testCase{
			pod: pod,
			services: []*kube_core.Service{
				{
					Spec: kube_core.ServiceSpec{
						Ports: []kube_core.ServicePort{
							{
								Protocol: "UDP", // all non-TCP ports should be ignored
								Port:     80,
								TargetPort: kube_intstr.IntOrString{
									Type:   kube_intstr.Int,
									IntVal: 8080,
								},
							},
							{
								Protocol: "SCTP", // all non-TCP ports should be ignored
								Port:     443,
								TargetPort: kube_intstr.IntOrString{
									Type:   kube_intstr.Int,
									IntVal: 8443,
								},
							},
							{
								Protocol: "TCP",
								Port:     7070,
								TargetPort: kube_intstr.IntOrString{
									Type:   kube_intstr.Int,
									IntVal: 7071,
								},
							},
							{
								Protocol: "", // defaults to TCP
								Port:     6060,
								TargetPort: kube_intstr.IntOrString{
									Type:   kube_intstr.String,
									StrVal: "diagnostics",
								},
							},
						},
					},
				},
			},
			expected: `
            mesh: default.kuma-test
            metadata:
              creationTimestamp: null
            spec:
              networking: {}
`,
		}),
		Entry("Pod with 2 Services", testCase{
			pod: pod,
			services: []*kube_core.Service{
				{
					ObjectMeta: kube_meta.ObjectMeta{
						Namespace: "demo",
						Name:      "example",
					},
					Spec: kube_core.ServiceSpec{
						Ports: []kube_core.ServicePort{
							{
								Protocol: "", // defaults to TCP
								Port:     80,
								TargetPort: kube_intstr.IntOrString{
									Type:   kube_intstr.Int,
									IntVal: 8080,
								},
							},
							{
								Protocol: "TCP",
								Port:     443,
								TargetPort: kube_intstr.IntOrString{
									Type:   kube_intstr.Int,
									IntVal: 8443,
								},
							},
						},
					},
				},
				{
					ObjectMeta: kube_meta.ObjectMeta{
						Namespace: "playground",
						Name:      "sample",
					},
					Spec: kube_core.ServiceSpec{
						Ports: []kube_core.ServicePort{
							{
								Protocol: "TCP",
								Port:     7071,
								TargetPort: kube_intstr.IntOrString{
									Type:   kube_intstr.Int,
									IntVal: 7070,
								},
							},
							{
								Protocol: "TCP",
								Port:     6061,
								TargetPort: kube_intstr.IntOrString{
									Type:   kube_intstr.String,
									StrVal: "metrics",
								},
							},
						},
					},
				},
			},
			expected: `
            mesh: default.kuma-test
            metadata:
              creationTimestamp: null
            spec:
              networking:
                inbound:
                - interface: 192.168.0.1:8080:8080
                  tags:
                    app: example
                    service: example.demo.svc:80
                    version: "0.1"
                - interface: 192.168.0.1:8443:8443
                  tags:
                    app: example
                    service: example.demo.svc:443
                    version: "0.1"
                - interface: 192.168.0.1:7070:7070
                  tags:
                    app: example
                    service: sample.playground.svc:7071
                    version: "0.1"
                - interface: 192.168.0.1:6060:6060
                  tags:
                    app: example
                    service: sample.playground.svc:6061
                    version: "0.1"
`,
		}),
		Entry("Pod with 1 Service and 1 other Dataplane", testCase{
			pod: pod,
			services: []*kube_core.Service{
				{
					ObjectMeta: kube_meta.ObjectMeta{
						Namespace: "demo",
						Name:      "example",
					},
					Spec: kube_core.ServiceSpec{
						Ports: []kube_core.ServicePort{
							{
								Protocol: "", // defaults to TCP
								Port:     80,
								TargetPort: kube_intstr.IntOrString{
									Type:   kube_intstr.Int,
									IntVal: 8080,
								},
							},
						},
					},
				},
			},
			others: []string{`
            apiVersion: kuma.io/v1alpha1
            kind: Dataplane
            mesh: default.kuma-test
            metadata:
              name: test-app-8646b8bbc8-5qbl2
              namespace: playground
            spec:
              networking:
                inbound:
                - interface: 10.244.0.25:80:80
                  tags:
                    app: test-app
                    pod-template-hash: 8646b8bbc8
                    service: test-app.playground.svc:80
                - interface: 10.244.0.25:443:443
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
            mesh: default.kuma-test
            metadata:
              creationTimestamp: null
            spec:
              networking:
                inbound:
                - interface: 192.168.0.1:8080:8080
                  tags:
                    app: example
                    service: example.demo.svc:80
                    version: "0.1"
                outbound:
                - interface: 10.108.144.24:443
                  service: test-app.playground.svc:443
                - interface: 10.108.144.24:80
                  service: test-app.playground.svc:80
`,
		}),
	)
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
			Expect(MeshFor(pod, "kuma-test")).To(Equal(given.expected))
		},
		Entry("Pod without annotations", testCase{
			podAnnotations: nil,
			expected:       "default.kuma-test",
		}),
		Entry("Pod with empty `kuma.io/mesh` annotation", testCase{
			podAnnotations: map[string]string{
				"kuma.io/mesh": "",
			},
			expected: "default.kuma-test",
		}),
		Entry("Pod with non-empty `kuma.io/mesh` annotation", testCase{
			podAnnotations: map[string]string{
				"kuma.io/mesh": "pilot",
			},
			expected: "pilot",
		}),
	)
})

var _ = Describe("InboundTagsFor(..)", func() {

	type testCase struct {
		podLabels map[string]string
		expected  map[string]string
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
			Expect(InboundTagsFor(pod, svc, &svc.Spec.Ports[0])).To(Equal(given.expected))
		},
		Entry("Pod without labels", testCase{
			podLabels: nil,
			expected: map[string]string{
				"service": "example.demo.svc:80",
			},
		}),
		Entry("Pod with labels", testCase{
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
		Entry("Pod with `service` label", testCase{
			podLabels: map[string]string{
				"service": "something",
				"app":     "example",
				"version": "0.1",
			},
			expected: map[string]string{
				"app":     "example",
				"version": "0.1",
				"service": "example.demo.svc:80",
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
		Expect(ServiceTagFor(svc, &svc.Spec.Ports[0])).To(Equal("example.demo.svc:80"))
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
