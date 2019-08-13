package controllers_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/discovery/k8s/controllers"
	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"

	mesh_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/api/v1alpha1"

	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_intstr "k8s.io/apimachinery/pkg/util/intstr"
)

var _ = Describe("PodToDataplane(..)", func() {

	type testCase struct {
		pod      *kube_core.Pod
		services []*kube_core.Service
		expected string
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

	DescribeTable("should create Inbound item for every Service port",
		func(given testCase) {
			// given
			dataplane := &mesh_k8s.Dataplane{}

			// when
			err := PodToDataplane(given.pod, given.services, dataplane)
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
            mesh: default
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
								Protocol: "UDP",
								Port:     80,
								TargetPort: kube_intstr.IntOrString{
									Type:   kube_intstr.Int,
									IntVal: 8080,
								},
							},
							{
								Protocol: "SCTP",
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
								Port: 6060,
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
            mesh: default
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
								Port: 80,
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
            mesh: default
            metadata:
              creationTimestamp: null
            spec:
              networking:
                inbound:
                - interface: 192.168.0.1:80:8080
                  tags:
                    app: example
                    service: example.demo.svc
                    version: "0.1"
                - interface: 192.168.0.1:443:8443
                  tags:
                    app: example
                    service: example.demo.svc
                    version: "0.1"
                - interface: 192.168.0.1:7071:7070
                  tags:
                    app: example
                    service: sample.playground.svc
                    version: "0.1"
                - interface: 192.168.0.1:6061:6060
                  tags:
                    app: example
                    service: sample.playground.svc
                    version: "0.1"
`,
		}),
	)
})

var _ = Describe("MeshFor(..)", func() {

	type testCase struct {
		podLabels map[string]string
		expected  string
	}

	DescribeTable("should use value of `getkonvoy.io/mesh` label on a Pod or fallback to the `default` Mesh",
		func(given testCase) {
			// given
			pod := &kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Labels: given.podLabels,
				},
			}

			// then
			Expect(MeshFor(pod)).To(Equal(given.expected))
		},
		Entry("Pod without labels", testCase{
			podLabels: nil,
			expected:  "default",
		}),
		Entry("Pod with empty `getkonvoy.io/mesh` label", testCase{
			podLabels: map[string]string{
				"getkonvoy.io/mesh": "",
			},
			expected: "default",
		}),
		Entry("Pod with non-empty `getkonvoy.io/mesh` label", testCase{
			podLabels: map[string]string{
				"getkonvoy.io/mesh": "pilot",
			},
			expected: "pilot",
		}),
	)
})

var _ = Describe("DataplaneFor(..)", func() {

	type testCase struct {
		pod      *kube_core.Pod
		services []*kube_core.Service
		expected string
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

	DescribeTable("should create Inbound item for every Service port",
		func(given testCase) {
			// when
			dataplane, err := DataplaneFor(given.pod, given.services)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(dataplane)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("Pod without Services", testCase{
			pod:      pod,
			services: nil,
			expected: `
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
								Protocol: "UDP",
								Port:     80,
								TargetPort: kube_intstr.IntOrString{
									Type:   kube_intstr.Int,
									IntVal: 8080,
								},
							},
							{
								Protocol: "SCTP",
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
								Port: 6061,
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
								Port: 80,
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
            networking:
              inbound:
              - interface: 192.168.0.1:80:8080
                tags:
                  app: example
                  service: example.demo.svc
                  version: "0.1"
              - interface: 192.168.0.1:443:8443
                tags:
                  app: example
                  service: example.demo.svc
                  version: "0.1"
              - interface: 192.168.0.1:7071:7070
                tags:
                  app: example
                  service: sample.playground.svc
                  version: "0.1"
              - interface: 192.168.0.1:6061:6060
                tags:
                  app: example
                  service: sample.playground.svc
                  version: "0.1"
`,
		}),
	)
})

var _ = Describe("InboundTagsFor(..)", func() {

	type testCase struct {
		podLabels map[string]string
		expected  map[string]string
	}

	DescribeTable("should combine Pod's labels with Service's FQDN",
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
			}

			// then
			Expect(InboundTagsFor(pod, svc)).To(Equal(given.expected))
		},
		Entry("Pod without labels", testCase{
			podLabels: nil,
			expected: map[string]string{
				"service": "example.demo.svc",
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
				"service": "example.demo.svc",
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
				"service": "example.demo.svc",
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
		}

		// then
		Expect(ServiceTagFor(svc)).To(Equal("example.demo.svc"))
	})
})
