package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/plugins/discovery/k8s/util"

	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ghodss/yaml"
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
			predicate := MatchServiceThatSelectsPod(pod)
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
			predicate := MatchServiceThatSelectsPod(pod)
			// then
			Expect(predicate(svc)).To(BeFalse())
		})
	})

	Describe("FindServices", func() {
		It("should match Services by a predicate", func() {
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
			svcs := &kube_core.ServiceList{
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
				},
			}

			// when
			matchingServices := FindServices(svcs, MatchServiceThatSelectsPod(pod))
			// then
			Expect(matchingServices).To(HaveLen(1))
			Expect(matchingServices).To(ConsistOf(&svcs.Items[0]))
		})
	})

	Describe("CopyStringMap", func() {
		It("should return nil if input is nil", func() {
			Expect(CopyStringMap(nil)).To(BeNil())
		})
		It("should return empty map if input is empty map", func() {
			Expect(CopyStringMap(map[string]string{})).To(Equal(map[string]string{}))
		})
		It("should return a copy if input map is not empty", func() {
			// given
			original := map[string]string{
				"a": "b",
				"c": "d",
			}

			// when
			copy := CopyStringMap(original)
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
					actual, err := FindPort(&pod, &svcPort)
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
					actual, err := FindPort(&pod, &svcPort)
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
})
