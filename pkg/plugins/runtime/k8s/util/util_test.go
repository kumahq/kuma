package util_test

import (
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_intstr "k8s.io/apimachinery/pkg/util/intstr"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

var _ = Describe("Util", func() {
	Describe("MatchServiceThatSelectsPod", func() {
		It("should match", func() {
			// given
			pod := &kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Labels: map[string]string{
						"app":               "demo-app",
						"pod-template-hash": "7cbbd658d5",
					},
				},
			}
			// and
			svc := &kube_core.Service{
				Spec: kube_core.ServiceSpec{
					Selector: map[string]string{
						"app": "demo-app",
					},
				},
			}

			// when
			predicate := util.MatchServiceThatSelectsPod(pod, nil)
			// then
			Expect(predicate(svc)).To(BeTrue())
		})

		It("should not match", func() {
			// given
			pod := &kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Labels: map[string]string{
						"app":               "demo-app",
						"pod-template-hash": "7cbbd658d5",
					},
				},
			}
			// and
			svc := &kube_core.Service{
				Spec: kube_core.ServiceSpec{
					Selector: map[string]string{
						"app": "nginx",
					},
				},
			}

			// when
			predicate := util.MatchServiceThatSelectsPod(pod, nil)
			// then
			Expect(predicate(svc)).To(BeFalse())
		})

		It("should match with ignored labels", func() {
			// given
			pod := &kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Labels: map[string]string{
						"app":               "demo-app",
						"pod-template-hash": "7cbbd658d5",
					},
				},
			}
			// and
			svc := &kube_core.Service{
				Spec: kube_core.ServiceSpec{
					Selector: map[string]string{
						"app":               "demo-app",
						"pod-template-hash": "xxxxxx",
					},
				},
			}

			// when
			predicate := util.MatchServiceThatSelectsPod(pod, []string{"pod-template-hash"})
			// then
			Expect(predicate(svc)).To(BeTrue())
		})
	})

	exampleTime := time.Date(2020, 0o1, 0o1, 0o1, 12, 0o0, 0o0, time.UTC)
	DescribeTable("MatchService",
		func(pod *kube_core.Pod, svcs *kube_core.ServiceList, matchSvcNames []string) {
			// when
			var svcNames []string
			for _, svc := range svcs.Items {
				s := svc
				if util.MatchService(&s, util.AnySelector(), util.MatchServiceThatSelectsPod(pod, nil)) {
					svcNames = append(svcNames, svc.Name)
				}
			}
			// then
			Expect(svcNames).To(ConsistOf(matchSvcNames))
		},
		Entry("should match services by a predicate",
			&kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Labels: map[string]string{
						"app":               "demo-app",
						"pod-template-hash": "7cbbd658d5",
					},
				},
			},
			// and
			&kube_core.ServiceList{
				Items: []kube_core.Service{
					{
						ObjectMeta: kube_meta.ObjectMeta{
							Name: "demo-app",
						},
						Spec: kube_core.ServiceSpec{
							Selector: map[string]string{
								"app": "demo-app",
							},
						},
					},
					{
						ObjectMeta: kube_meta.ObjectMeta{
							Name: "nginx",
						},
						Spec: kube_core.ServiceSpec{
							Selector: map[string]string{
								"app": "nginx",
							},
						},
					},
					{
						ObjectMeta: kube_meta.ObjectMeta{
							Name:      "kubernetes",
							Namespace: "default",
						},
						Spec: kube_core.ServiceSpec{},
					},
				},
			},
			[]string{"demo-app"},
		),
		Entry("should match multiple services in order",
			&kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Labels: map[string]string{
						"app":               "demo-app",
						"pod-template-hash": "7cbbd658d5",
					},
				},
			},
			// and
			&kube_core.ServiceList{
				Items: []kube_core.Service{
					{
						ObjectMeta: kube_meta.ObjectMeta{
							CreationTimestamp: kube_meta.NewTime(exampleTime),
							Name:              "demo-app2",
						},
						Spec: kube_core.ServiceSpec{
							Selector: map[string]string{
								"app": "demo-app",
							},
						},
					},
					{
						ObjectMeta: kube_meta.ObjectMeta{
							CreationTimestamp: kube_meta.NewTime(exampleTime),
							Name:              "demo-app",
						},
						Spec: kube_core.ServiceSpec{
							Selector: map[string]string{
								"app": "demo-app",
							},
						},
					},
					{
						ObjectMeta: kube_meta.ObjectMeta{
							CreationTimestamp: kube_meta.NewTime(exampleTime.Add(-time.Hour)),
							Name:              "nginx",
						},
						Spec: kube_core.ServiceSpec{
							Selector: map[string]string{
								"app": "demo-app",
							},
						},
					},
					{
						ObjectMeta: kube_meta.ObjectMeta{
							CreationTimestamp: kube_meta.NewTime(exampleTime),
							Name:              "kubernetes",
							Namespace:         "default",
						},
						Spec: kube_core.ServiceSpec{},
					},
				},
			},
			[]string{"demo-app", "demo-app2", "nginx"},
		),
	)

	Describe("MeshOf(..)", func() {
		type testCase struct {
			podAnnotations map[string]string
			nsAnnotations  map[string]string
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
				ns := &kube_core.Namespace{
					ObjectMeta: kube_meta.ObjectMeta{
						Annotations: given.nsAnnotations,
					},
				}

				// then
				Expect(util.MeshOfByLabelOrAnnotation(logr.Discard(), pod, ns)).To(Equal(given.expected))
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
			Entry("Pod with empty `kuma.io/mesh` annotation, Namespace with annotation", testCase{
				podAnnotations: map[string]string{
					"kuma.io/mesh": "",
				},
				nsAnnotations: map[string]string{
					"kuma.io/mesh": "demo",
				},
				expected: "demo",
			}),
		)
	})
	Describe("MeshOfByLabelOrAnnotation(..)", func() {
		type testCase struct {
			podLabels      map[string]string
			podAnnotations map[string]string
			nsLabels       map[string]string
			nsAnnotations  map[string]string
			expected       string
		}

		DescribeTable("should use value of `kuma.io/mesh` annotation on a Pod or fallback to the `default` Mesh",
			func(given testCase) {
				// given
				pod := &kube_core.Pod{
					ObjectMeta: kube_meta.ObjectMeta{
						Annotations: given.podAnnotations,
						Labels:      given.podLabels,
					},
				}
				ns := &kube_core.Namespace{
					ObjectMeta: kube_meta.ObjectMeta{
						Annotations: given.nsAnnotations,
						Labels:      given.nsLabels,
					},
				}

				// then
				Expect(util.MeshOfByLabelOrAnnotation(logr.Discard(), pod, ns)).To(Equal(given.expected))
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
			Entry("Pod with non-empty `kuma.io/mesh` label", testCase{
				podLabels: map[string]string{
					"kuma.io/mesh": "demo",
				},
				expected: "demo",
			}),
			Entry("Pod with non-empty `kuma.io/mesh` annotation", testCase{
				podAnnotations: map[string]string{
					"kuma.io/mesh": "demo",
				},
				expected: "demo",
			}),
			Entry("Pod with empty `kuma.io/mesh` annotation, Namespace with label", testCase{
				podAnnotations: map[string]string{
					"kuma.io/mesh": "",
				},
				nsLabels: map[string]string{
					"kuma.io/mesh": "demo",
				},
				expected: "demo",
			}),
			Entry("Pod with empty `kuma.io/mesh` annotation, Namespace with annotation", testCase{
				podAnnotations: map[string]string{
					"kuma.io/mesh": "",
				},
				nsAnnotations: map[string]string{
					"kuma.io/mesh": "demo",
				},
				expected: "demo",
			}),
		)
	})

	Describe("CopyStringMap", func() {
		It("should return nil if input is nil", func() {
			Expect(util.CopyStringMap(nil)).To(BeNil())
		})
		It("should return empty map if input is empty map", func() {
			Expect(util.CopyStringMap(map[string]string{})).To(Equal(map[string]string{}))
		})
		It("should return a copy if input map is not empty", func() {
			// given
			original := map[string]string{
				"a": "b",
				"c": "d",
			}

			// when
			copy := util.CopyStringMap(original)
			// then
			Expect(copy).To(Equal(original))
			Expect(copy).ToNot(BeIdenticalTo(original))
		})
	})

	Describe("FindPort()", func() {
		Describe("should return a correct port number", func() {
			type testCase struct {
				pod      string
				svcPort  string
				expected int
			}

			DescribeTable("should correctly find a matching port in a given Pod",
				func(given testCase) {
					// setup
					pod := kube_core.Pod{}
					svcPort := kube_core.ServicePort{}

					// when
					err := yaml.Unmarshal([]byte(given.pod), &pod)
					// then
					Expect(err).ToNot(HaveOccurred())

					// when
					err = yaml.Unmarshal([]byte(given.svcPort), &svcPort)
					// then
					Expect(err).ToNot(HaveOccurred())

					// when
					actual, _, err := util.FindPort(&pod, &svcPort)
					// then
					Expect(err).ToNot(HaveOccurred())
					// and
					Expect(actual).To(Equal(given.expected))
				},
				Entry("Service with `targetPort` as a number (TCP)", testCase{
					pod: `
                    spec:
                      containers:
                      - name: container-1
                        ports: [] # notice that actual container ports become irrelevant when Service has 'targetPort' as a number
`,
					svcPort: `
                    name: http
                    port: 8080
                    protocol: TCP
                    targetPort: 8080
`,
					expected: 8080,
				}),
				Entry("Service with `targetPort` as a number (UDP)", testCase{
					pod: `
                    spec:
                      containers:
                      - name: container-1
                        ports: [] # notice that actual container ports become irrelevant when Service has 'targetPort' as a number
`,
					svcPort: `
                    name: dns
                    port: 53
                    protocol: UDP
                    targetPort: 53
`,
					expected: 53,
				}),
				Entry("Service with `targetPort` as a name (container port protocol is omitted)", testCase{
					pod: `
                    spec:
                      containers:
                      - name: container-1
                        ports: []
                      - name: container-2
                        ports:
                        - containerPort: 8080 # should be ignored
                          protocol: TCP
                      - name: container-3
                        ports:
                        - containerPort: 7070
                          name: http-api      # should match
                          # no protocol should default to 'TCP'
`,
					svcPort: `
                    name: http
                    port: 8080
                    protocol: TCP
                    targetPort: http-api
`,
					expected: 7070,
				}),
				Entry("Service with `targetPort` as a name (container port protocol is set to TCP)", testCase{
					pod: `
                    spec:
                      containers:
                      - name: container-1
                        ports: []
                      - name: container-2
                        ports:
                        - containerPort: 8080 # should be ignored
                          protocol: TCP
                      - name: container-3
                        ports:
                        - containerPort: 7070
                          name: http-api      # should match
                          protocol: TCP
`,
					svcPort: `
                    name: http
                    port: 8080
                    protocol: TCP
                    targetPort: http-api
`,
					expected: 7070,
				}),
				Entry("Service with `targetPort` as a name (container port protocol is set to UDP)", testCase{
					pod: `
                    spec:
                      containers:
                      - name: container-1
                        ports: []
                      - name: container-2
                        ports:
                        - containerPort: 53   # should be ignored
                          protocol: UDP
                      - name: container-3
                        ports:
                        - containerPort: 1053
                          name: dns-port      # should match
                          protocol: UDP
`,
					svcPort: `
                    name: dns
                    port: 53
                    protocol: UDP
                    targetPort: dns-port
`,
					expected: 1053,
				}),
				Entry("Service with `targetPort` as a name (service port protocol is omitted and container port protocol is omitted)", testCase{
					pod: `
                    spec:
                      containers:
                      - name: container-1
                        ports: []
                      - name: container-2
                        ports:
                        - containerPort: 8080 # should be ignored
                          protocol: TCP
                      - name: container-3
                        ports:
                        - containerPort: 7070
                          name: http-api      # should match
                          # no protocol should default to 'TCP'
`,
					svcPort: `
                    name: http
                    port: 8080
                    targetPort: http-api
`,
					expected: 7070,
				}),
				Entry("Service with `targetPort` as a name (service port protocol is omitted while container port protocol set to TCP)", testCase{
					pod: `
                    spec:
                      containers:
                      - name: container-1
                        ports: []
                      - name: container-2
                        ports:
                        - containerPort: 8080 # should be ignored
                          protocol: TCP
                      - name: container-3
                        ports:
                        - containerPort: 7070
                          name: http-api      # should match
                          protocol: TCP
`,
					svcPort: `
                    name: http
                    port: 8080
                    targetPort: http-api
`,
					expected: 7070,
				}),
			)
		})

		Describe("should return an error if a Pod doesn't have a matching container port", func() {
			type testCase struct {
				pod         string
				svcPort     string
				expectedErr string
			}

			DescribeTable("should return a proper error",
				func(given testCase) {
					// setup
					pod := kube_core.Pod{}
					svcPort := kube_core.ServicePort{}

					// when
					err := yaml.Unmarshal([]byte(given.pod), &pod)
					// then
					Expect(err).ToNot(HaveOccurred())

					// when
					err = yaml.Unmarshal([]byte(given.svcPort), &svcPort)
					// then
					Expect(err).ToNot(HaveOccurred())

					// when
					actual, _, err := util.FindPort(&pod, &svcPort)
					// then
					Expect(err).To(HaveOccurred())
					// and
					Expect(err.Error()).To(Equal(given.expectedErr))
					// and
					Expect(actual).To(Equal(0))
				},
				Entry("Pod has no container port with such name", testCase{
					pod: `
                    metadata:
                      uid: 8648e081-576d-4a23-861b-8f2d94d28d34
                    spec:
                      containers:
                      - name: container-1
                        ports: []
                      - name: container-2
                        ports:
                        - containerPort: 8080 # should be ignored
                          protocol: TCP
`,
					svcPort: `
                    name: http
                    port: 8080
                    protocol: TCP
                    targetPort: http-api
`,
					expectedErr: `no suitable port for manifest: 8648e081-576d-4a23-861b-8f2d94d28d34`,
				}),
				Entry("Pod has no container port with such name and protocol (TCP)", testCase{
					pod: `
                    metadata:
                      uid: 8648e081-576d-4a23-861b-8f2d94d28d34
                    spec:
                      containers:
                      - name: container-1
                        ports: []
                      - name: container-2
                        ports:
                        - containerPort: 8080 # should be ignored
                          protocol: TCP
                      - name: container-3
                        ports:
                        - containerPort: 7070
                          name: http-api      # should be ignored
                          protocol: UDP
`,
					svcPort: `
                    name: http
                    port: 8080
                    protocol: TCP
                    targetPort: http-api
`,
					expectedErr: `no suitable port for manifest: 8648e081-576d-4a23-861b-8f2d94d28d34`,
				}),
				Entry("Pod has no container port with such name and protocol (UDP)", testCase{
					pod: `
                    metadata:
                      uid: 8648e081-576d-4a23-861b-8f2d94d28d34
                    spec:
                      containers:
                      - name: container-1
                        ports: []
                      - name: container-2
                        ports:
                        - containerPort: 53 # should be ignored
                          protocol: UDP
                      - name: container-3
                        ports:
                        - containerPort: 1053
                          name: dns-port      # should be ignored
                          protocol: TCP
`,
					svcPort: `
                    name: dns
                    port: 53
                    protocol: UDP
                    targetPort: dns-port
`,
					expectedErr: `no suitable port for manifest: 8648e081-576d-4a23-861b-8f2d94d28d34`,
				}),
				Entry("Pod has no container port with such name and protocol (TCP)", testCase{
					pod: `
                    metadata:
                      uid: 8648e081-576d-4a23-861b-8f2d94d28d34
                    spec:
                      containers:
                      - name: container-1
                        ports: []
                      - name: container-2
                        ports:
                        - containerPort: 8080 # should be ignored
                          protocol: TCP
                      - name: container-3
                        ports:
                        - containerPort: 7070
                          name: http-api      # should be ignored
                          protocol: UDP
`,
					svcPort: `
                    name: http
                    port: 8080
                    # no protocol defaults to TCP
                    targetPort: http-api
`,
					expectedErr: `no suitable port for manifest: 8648e081-576d-4a23-861b-8f2d94d28d34`,
				}),
			)
		})
	})
	Describe("ServiceTagFor", func() {
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
			Expect(util.ServiceTag(kube_client.ObjectKeyFromObject(svc), &svc.Spec.Ports[0].Port)).To(Equal("example_demo_svc_80"))
		})
	})
})
