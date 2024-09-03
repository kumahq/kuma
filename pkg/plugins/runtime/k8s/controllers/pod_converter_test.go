package controllers_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	kube_apps "k8s.io/api/apps/v1"
	kube_batch "k8s.io/api/batch/v1"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_intstr "k8s.io/apimachinery/pkg/util/intstr"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	. "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	util_yaml "github.com/kumahq/kuma/pkg/util/yaml"
)

func Parse[T any](values []string) ([]T, error) {
	l := make([]T, len(values))
	for i, value := range values {
		obj := new(T)
		if err := yaml.Unmarshal([]byte(value), obj); err != nil {
			return nil, err
		}
		l[i] = *obj
	}
	return l, nil
}

var _ = Describe("PodToDataplane(..)", func() {
	type testCase struct {
		pod               string
		servicesForPod    string
		otherDataplanes   string
		otherServices     string
		otherReplicaSets  string
		otherJobs         string
		node              string
		dataplane         string
		existingDataplane string
		nodeLabelsToCopy  []string
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
				services, err = Parse[*kube_core.Service](YAMLs)
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
				services, err := Parse[*kube_core.Service](YAMLs)
				Expect(err).ToNot(HaveOccurred())
				reader, err := newFakeServiceReader(services)
				Expect(err).ToNot(HaveOccurred())
				serviceGetter = reader
			}

			// node
			var nodeGetter kube_client.Reader
			if given.node != "" {
				bytes, err = os.ReadFile(filepath.Join("testdata", given.node))
				Expect(err).ToNot(HaveOccurred())
				nodeGetter = fakeNodeReader(bytes)
			}

			// other ReplicaSets
			var replicaSetGetter kube_client.Reader
			if given.otherReplicaSets != "" {
				replicaSetGetter = getReplicaSetsReader("testdata", given.otherReplicaSets)
			}

			var jobGetter kube_client.Reader
			if given.otherJobs != "" {
				jobGetter = getJobsReader("testdata", given.otherJobs)
			}

			// other dataplanes
			var otherDataplanes []*mesh_k8s.Dataplane
			if given.otherDataplanes != "" {
				bytes, err = os.ReadFile(filepath.Join("testdata", given.otherDataplanes))
				Expect(err).ToNot(HaveOccurred())
				YAMLs := util_yaml.SplitYAML(string(bytes))
				otherDataplanes, err = Parse[*mesh_k8s.Dataplane](YAMLs)
				Expect(err).ToNot(HaveOccurred())
			}

			converter := PodConverter{
				ServiceGetter: serviceGetter,
				InboundConverter: InboundConverter{
					NameExtractor: NameExtractor{
						ReplicaSetGetter: replicaSetGetter,
						JobGetter:        jobGetter,
					},
					NodeGetter:       nodeGetter,
					NodeLabelsToCopy: given.nodeLabelsToCopy,
				},
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
		Entry("03. Pod with gateway annotation and 1 service - legacy", testCase{
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
		Entry("18. Gateway with non tcp appProtocol", testCase{
			pod:            "18.pod.yaml",
			servicesForPod: "18.services-for-pod.yaml",
			dataplane:      "18.dataplane.yaml",
		}),
		Entry("19. Terminating pod is unhealthy", testCase{
			pod:            "19.pod.yaml",
			servicesForPod: "19.services-for-pod.yaml",
			dataplane:      "19.dataplane.yaml",
		}),
		Entry("20. Pod with gateway annotation and 1 service identified by deployment", testCase{
			pod:              "20.pod.yaml",
			servicesForPod:   "20.services-for-pod.yaml",
			otherReplicaSets: "20.replicasets-for-pod.yaml",
			dataplane:        "20.dataplane.yaml",
		}),
		Entry("21. Pod with gateway annotation and 1 service with no replicaset", testCase{
			pod:            "21.pod.yaml",
			servicesForPod: "21.services-for-pod.yaml",
			dataplane:      "21.dataplane.yaml",
		}),
		Entry("22. Pod with gateway annotation and 1 service with replicaset but no deployment", testCase{
			pod:              "22.pod.yaml",
			servicesForPod:   "22.services-for-pod.yaml",
			otherReplicaSets: "22.replicasets-for-pod.yaml",
			dataplane:        "22.dataplane.yaml",
		}),
		Entry("23. Pod with ignored listener", testCase{
			pod:            "23.pod.yaml",
			servicesForPod: "23.services-for-pod.yaml",
			dataplane:      "23.dataplane.yaml",
		}),
		Entry("24. Pod with transparent proxy enabled, with ipv6 disabled", testCase{
			pod:            "24.pod.yaml",
			servicesForPod: "08.services-for-pod.yaml",
			dataplane:      "24.dataplane.yaml",
		}),
		Entry("25. Pod with transparent proxy enabled, with no ip-family-mode and ipv6 disabled", testCase{
			pod:            "25.pod.yaml",
			servicesForPod: "08.services-for-pod.yaml",
			dataplane:      "25.dataplane.yaml",
		}),
		Entry("26. Should copy node label to the dataplane", testCase{
			pod:              "26.pod.yaml",
			node:             "26.node.yaml",
			dataplane:        "26.dataplane.yaml",
			nodeLabelsToCopy: []string{"topology.kubernetes.io/region"},
		}),
		Entry("27. Should not copy label to the dataplane if there is no node label", testCase{
			pod:              "27.pod.yaml",
			node:             "27.node.yaml",
			dataplane:        "27.dataplane.yaml",
			nodeLabelsToCopy: []string{"topology.kubernetes.io/region"},
		}),
		Entry("28. Pod with reachable backend refs", testCase{
			pod:            "28.pod.yaml",
			servicesForPod: "28.services-for-pod.yaml",
			dataplane:      "28.dataplane.yaml",
		}),
		Entry("29. Pod with empty reachable backend refs", testCase{
			pod:            "29.pod.yaml",
			servicesForPod: "29.services-for-pod.yaml",
			dataplane:      "29.dataplane.yaml",
		}),
		Entry("should create dataplane even if service ports don't match", testCase{
			pod:            "mismatch-ports.pod.yaml",
			servicesForPod: "mismatch-ports.services-for-pod.yaml",
			dataplane:      "mismatch-ports.dataplane.yaml",
		}),
		Entry("30. Pod using application probe proxy", testCase{
			pod:            "30.pod.yaml",
			servicesForPod: "30.services-for-pod.yaml",
			dataplane:      "30.dataplane.yaml",
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
			services, err := Parse[*kube_core.Service](YAMLs)
			Expect(err).ToNot(HaveOccurred())

			// node
			var nodeGetter kube_client.Reader
			if given.node != "" {
				bytes, err = os.ReadFile(filepath.Join("testdata", "ingress", given.node))
				Expect(err).ToNot(HaveOccurred())
				nodeGetter = fakeNodeReader(bytes)
			}

			converter := PodConverter{
				ServiceGetter:     nil,
				NodeGetter:        nodeGetter,
				ResourceConverter: k8s.NewSimpleConverter(),
				Zone:              "zone-1",
				InboundConverter: InboundConverter{
					NodeGetter: nodeGetter,
				},
			}

			// when
			ingress := &mesh_k8s.ZoneIngress{}
			if given.existingDataplane != "" {
				bytes, err = os.ReadFile(filepath.Join("testdata", "ingress", given.existingDataplane))
				Expect(err).ToNot(HaveOccurred())
				err = yaml.Unmarshal(bytes, ingress)
				Expect(err).ToNot(HaveOccurred())
			}

			// then
			err = converter.PodToIngress(context.Background(), ingress, pod, services)
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
		Entry("Existing ZoneIngress with load balancer and ip", testCase{
			pod:               "ingress-exists.pod.yaml",
			servicesForPod:    "ingress-exists.services-for-pod.yaml",
			existingDataplane: "ingress-exists.existing-dataplane.yaml",
			dataplane:         "ingress-exists.golden.yaml",
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
			ctx := context.Background()

			// services for pod
			bytes, err = os.ReadFile(filepath.Join("testdata", "egress", given.servicesForPod))
			Expect(err).ToNot(HaveOccurred())
			YAMLs := util_yaml.SplitYAML(string(bytes))
			services, err := Parse[*kube_core.Service](YAMLs)
			Expect(err).ToNot(HaveOccurred())

			// node
			var nodeGetter kube_client.Reader
			if given.node != "" {
				bytes, err = os.ReadFile(filepath.Join("testdata", "egress", given.node))
				Expect(err).ToNot(HaveOccurred())
				nodeGetter = fakeNodeReader(bytes)
			}

			converter := PodConverter{
				ServiceGetter:     nil,
				NodeGetter:        nodeGetter,
				ResourceConverter: k8s.NewSimpleConverter(),
				Zone:              "zone-1",
				InboundConverter: InboundConverter{
					NodeGetter: nodeGetter,
				},
			}

			// when
			egress := &mesh_k8s.ZoneEgress{}
			err = converter.PodToEgress(ctx, egress, pod, services)

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
})

var _ = Describe("InboundTagsForService(..)", func() {
	type testCase struct {
		isGateway      bool
		zone           string
		podLabels      map[string]string
		svcAnnotations map[string]string
		appProtocol    *string
		nodeLabels     map[string]string
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
				Spec: kube_core.PodSpec{
					NodeName: "test-node",
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
			nodeLabels := given.nodeLabels

			// expect
			Expect(InboundTagsForService(given.zone, pod, svc, &svc.Spec.Ports[0], nodeLabels)).To(Equal(given.expected))
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
		Entry("Pod with node's topology labels", testCase{
			isGateway: false,
			podLabels: map[string]string{
				"app":     "example",
				"version": "0.1",
			},
			nodeLabels: map[string]string{
				kube_core.LabelTopologyRegion: "east",
				kube_core.LabelTopologyZone:   "east-2a",
			},
			expected: map[string]string{
				"app":                         "example",
				"version":                     "0.1",
				"kuma.io/service":             "example_demo_svc_80",
				"kuma.io/protocol":            "tcp", // we want Kuma's default behavior to be explicit to a user
				"k8s.kuma.io/service-name":    "example",
				"k8s.kuma.io/service-port":    "80",
				"k8s.kuma.io/namespace":       "demo",
				kube_core.LabelTopologyRegion: "east",
				kube_core.LabelTopologyZone:   "east-2a",
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
			appProtocol: pointer.To("http"),
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

var _ = Describe("MetricsAggregateFor(..)", func() {
	type testCase struct {
		annotations map[string]string
		expected    []*mesh_proto.PrometheusAggregateMetricsConfig
	}

	DescribeTable("should create proper metrics configuration",
		func(given testCase) {
			// given
			pod := &kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace:   "demo",
					Annotations: given.annotations,
				},
			}

			// expect
			configuration, err := MetricsAggregateFor(pod)
			Expect(err).ToNot(HaveOccurred())
			Expect(configuration).To(HaveLen(len(given.expected)))
			Expect(configuration).To(ContainElements(given.expected))
		},
		Entry("one service with double tag", testCase{
			annotations: map[string]string{
				"prometheus.metrics.kuma.io/aggregate-my-app-path":    "/stats",
				"prometheus.metrics.kuma.io/aggregate-my-app-port":    "123",
				"prometheus.metrics.kuma.io/aggregate-my-app-address": "localhost",
				"prometheus.metrics.kuma.io/aggregate-my-app-enabled": "false",
			},
			expected: []*mesh_proto.PrometheusAggregateMetricsConfig{
				{
					Name:    "my-app",
					Path:    "/stats",
					Port:    123,
					Enabled: util_proto.Bool(false),
					Address: "localhost",
				},
			},
		}),
		Entry("few services", testCase{
			annotations: map[string]string{
				"prometheus.metrics.kuma.io/aggregate-my-app-path":       "/stats",
				"prometheus.metrics.kuma.io/aggregate-my-app-port":       "123",
				"prometheus.metrics.kuma.io/aggregate-my-app-2-path":     "/stats/2",
				"prometheus.metrics.kuma.io/aggregate-my-app-2-port":     "1234",
				"prometheus.metrics.kuma.io/aggregate-my-app-2-enabled":  "true",
				"prometheus.metrics.kuma.io/aggregate-sidecar-path":      "/metrics",
				"prometheus.metrics.kuma.io/aggregate-sidecar-port":      "12345",
				"prometheus.metrics.kuma.io/aggregate-sidecar-enabled":   "false",
				"prometheus.metrics.kuma.io/aggregate-disabled-enabled":  "false",
				"prometheus.metrics.kuma.io/aggregate-default-path-port": "11111",
			},
			expected: []*mesh_proto.PrometheusAggregateMetricsConfig{
				{
					Name:    "my-app",
					Path:    "/stats",
					Port:    123,
					Enabled: util_proto.Bool(true),
				},
				{
					Name:    "my-app-2",
					Path:    "/stats/2",
					Port:    1234,
					Enabled: util_proto.Bool(true),
				},
				{
					Name:    "sidecar",
					Path:    "/metrics",
					Port:    12345,
					Enabled: util_proto.Bool(false),
				},
				{
					Name:    "disabled",
					Enabled: util_proto.Bool(false),
					Path:    "/metrics",
				},
				{
					Name:    "default-path",
					Port:    11111,
					Path:    "/metrics",
					Enabled: util_proto.Bool(true),
				},
			},
		}),
	)
})

var _ = Describe("MetricsAggregateFor(..)", func() {
	type testCase struct {
		annotations map[string]string
		expected    string
	}

	DescribeTable("should fail when",
		func(given testCase) {
			// given
			pod := &kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace:   "demo",
					Annotations: given.annotations,
				},
			}

			// expect
			_, err := MetricsAggregateFor(pod)
			Expect(err.Error()).To(Equal(given.expected))
		},
		Entry("one parameter for each service only defined", testCase{
			annotations: map[string]string{
				"prometheus.metrics.kuma.io/aggregate-my-app-path":   "/stats",
				"prometheus.metrics.kuma.io/aggregate-my-app-2-port": "123",
			},
			expected: "port needs to be specified for metrics scraping",
		}),
		Entry("parsing integer", testCase{
			annotations: map[string]string{
				"prometheus.metrics.kuma.io/aggregate-my-app-2-path": "/stats",
				"prometheus.metrics.kuma.io/aggregate-my-app-2-port": "123a",
			},
			expected: "failed to parse annotation \"prometheus.metrics.kuma.io/aggregate-my-app-2-port\": strconv.ParseUint: parsing \"123a\": invalid syntax",
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
			appProtocol: pointer.To(""),
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
			appProtocol: pointer.To("not-yet-supported-protocol"),
			expected:    "tcp", // we want Kuma's behavior to be straightforward to a user (appProtocol is not Kuma specific)
		}),
		Entry("appProtocol has a lowercase value", testCase{
			appProtocol: pointer.To("HtTp"),
			expected:    "http", // we want Kuma's behavior to be straightforward to a user (copy appProtocol lowercase value)
		}),
		Entry("appProtocol has a known value: http", testCase{
			appProtocol: pointer.To("http"),
			expected:    "http",
		}),
		Entry("appProtocol has a known value: tcp", testCase{
			appProtocol: pointer.To("tcp"),
			expected:    "tcp",
		}),
		Entry("no appProtocol and no `<port>.service.kuma.io/protocol`", testCase{
			appProtocol: nil,
			annotations: nil,
			expected:    "tcp",
		}),
	)
})

var _ = Describe("Serviceless Name for(...)", func() {
	type testCase struct {
		pod          string
		replicaSets  string
		jobs         string
		expectedName string
		expectedKind string
	}
	DescribeTable("should infer name based on the resource type",
		func(given testCase) {
			// given
			ctx := context.Background()

			pod := &kube_core.Pod{}
			bytes, err := os.ReadFile(filepath.Join("testdata", "serviceless", given.pod))
			Expect(err).ToNot(HaveOccurred())
			err = yaml.Unmarshal(bytes, pod)
			Expect(err).ToNot(HaveOccurred())

			var replicaSetGetter kube_client.Reader
			if given.replicaSets != "" {
				replicaSetGetter = getReplicaSetsReader("testdata", "serviceless", given.replicaSets)
			}

			var jobGetter kube_client.Reader
			if given.jobs != "" {
				jobGetter = getJobsReader("testdata", "serviceless", given.jobs)
			}

			nameExtractor := NameExtractor{
				ReplicaSetGetter: replicaSetGetter,
				JobGetter:        jobGetter,
			}

			// when
			name, kind, err := nameExtractor.Name(ctx, pod)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(name).To(Equal(given.expectedName))
			Expect(kind).To(Equal(given.expectedKind))
		},
		Entry("name from deployment", testCase{
			pod:          "01.pod.yaml",
			replicaSets:  "01.replicasets-for-pod.yaml",
			expectedName: "test-server",
			expectedKind: "Deployment",
		}),
		Entry("name from cronjob", testCase{
			pod:          "02.pod.yaml",
			jobs:         "02.job-for-pod.yaml",
			expectedName: "test-job",
			expectedKind: "CronJob",
		}),
		Entry("name from replicaset", testCase{
			pod:          "03.pod.yaml",
			replicaSets:  "03.replicasets-for-pod.yaml",
			expectedName: "test-rs",
			expectedKind: "ReplicaSet",
		}),
		Entry("name from job", testCase{
			pod:          "04.pod.yaml",
			jobs:         "04.job-for-pod.yaml",
			expectedName: "test-job",
			expectedKind: "Job",
		}),
		Entry("name from pod", testCase{
			pod:          "05.pod.yaml",
			expectedName: "test-pod-1",
			expectedKind: "Pod",
		}),
		Entry("name from daemonset", testCase{
			pod:          "06.pod.yaml",
			expectedName: "test-ds",
			expectedKind: "DaemonSet",
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

func (r fakeServiceReader) Get(ctx context.Context, key kube_client.ObjectKey, obj kube_client.Object, _ ...kube_client.GetOption) error {
	fqName := fmt.Sprintf("%s/%s", key.Namespace, key.Name)
	data, ok := r[fqName]
	if !ok {
		return errors.Errorf("service not found: %s", fqName)
	}
	return yaml.Unmarshal([]byte(data), obj)
}

func (f fakeServiceReader) List(ctx context.Context, list kube_client.ObjectList, opts ...kube_client.ListOption) error {
	return errors.New("not implemented")
}

type fakeNodeReader string

func (f fakeNodeReader) Get(ctx context.Context, key kube_client.ObjectKey, obj kube_client.Object, _ ...kube_client.GetOption) error {
	err := yaml.Unmarshal([]byte(f), &obj)
	if err != nil {
		return err
	}
	return nil
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

type fakeReplicaSetReader map[string]string

func newFakeReplicaSetReader(replicaSets []*kube_apps.ReplicaSet) (fakeReplicaSetReader, error) {
	replicaSetsMap := map[string]string{}
	for _, rs := range replicaSets {
		bytes, err := yaml.Marshal(rs)
		if err != nil {
			return nil, err
		}
		replicaSetsMap[rs.GetNamespace()+"/"+rs.GetName()] = string(bytes)
	}
	return replicaSetsMap, nil
}

type fakeJobReader map[string]string

func newFakeJobReader(jobs []*kube_batch.Job) (fakeJobReader, error) {
	jobsMap := map[string]string{}
	for _, job := range jobs {
		bytes, err := yaml.Marshal(job)
		if err != nil {
			return nil, err
		}
		jobsMap[job.GetNamespace()+"/"+job.GetName()] = string(bytes)
	}
	return jobsMap, nil
}

var _ kube_client.Reader = fakeReplicaSetReader{}

func (r fakeReplicaSetReader) Get(ctx context.Context, key kube_client.ObjectKey, obj kube_client.Object, _ ...kube_client.GetOption) error {
	fqName := fmt.Sprintf("%s/%s", key.Namespace, key.Name)
	data, ok := r[fqName]
	if !ok {
		return errors.Errorf("replicaset not found: %s", fqName)
	}
	return yaml.Unmarshal([]byte(data), obj)
}

func (f fakeReplicaSetReader) List(ctx context.Context, list kube_client.ObjectList, opts ...kube_client.ListOption) error {
	return errors.New("not implemented")
}

var _ kube_client.Reader = fakeJobReader{}

func (r fakeJobReader) Get(ctx context.Context, key kube_client.ObjectKey, obj kube_client.Object, _ ...kube_client.GetOption) error {
	fqName := fmt.Sprintf("%s/%s", key.Namespace, key.Name)
	data, ok := r[fqName]
	if !ok {
		return errors.Errorf("job not found: %s", fqName)
	}
	return yaml.Unmarshal([]byte(data), obj)
}

func (f fakeJobReader) List(ctx context.Context, list kube_client.ObjectList, opts ...kube_client.ListOption) error {
	return errors.New("not implemented")
}

func getReplicaSetsReader(path ...string) fakeReplicaSetReader {
	bytes, err := os.ReadFile(filepath.Join(path...))
	Expect(err).ToNot(HaveOccurred())
	YAMLs := util_yaml.SplitYAML(string(bytes))
	rsets, err := Parse[*kube_apps.ReplicaSet](YAMLs)
	Expect(err).ToNot(HaveOccurred())
	reader, err := newFakeReplicaSetReader(rsets)
	Expect(err).ToNot(HaveOccurred())
	return reader
}

func getJobsReader(path ...string) fakeJobReader {
	bytes, err := os.ReadFile(filepath.Join(path...))
	Expect(err).ToNot(HaveOccurred())
	YAMLs := util_yaml.SplitYAML(string(bytes))
	rsets, err := Parse[*kube_batch.Job](YAMLs)
	Expect(err).ToNot(HaveOccurred())
	reader, err := newFakeJobReader(rsets)
	Expect(err).ToNot(HaveOccurred())
	return reader
}
