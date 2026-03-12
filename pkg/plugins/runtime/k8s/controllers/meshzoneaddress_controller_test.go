package controllers_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_discovery "k8s.io/api/discovery/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_events "k8s.io/client-go/tools/events"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"

	meshzoneaddress_k8s "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshzoneaddress/k8s/v1alpha1"
	. "github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/controllers"
	. "github.com/kumahq/kuma/v2/pkg/test/matchers"
	yaml2 "github.com/kumahq/kuma/v2/pkg/util/yaml"
)

const (
	testZone      = "zone-1"
	testNamespace = "kuma-system"
	testSvcName   = "zone-ingress"
)

var _ = Describe("MeshZoneAddressReconciler", func() {
	type testCase struct {
		inputFile  string
		outputFile string
		svcName    string
	}

	DescribeTable("should reconcile service",
		func(given testCase) {
			// given
			bytes, err := os.ReadFile(filepath.Join("testdata", "meshzoneaddress", given.inputFile))
			Expect(err).ToNot(HaveOccurred())

			var objects []kube_client.Object
			for _, yamlObj := range yaml2.SplitYAML(string(bytes)) {
				var obj kube_client.Object
				switch {
				case strings.Contains(yamlObj, "kind: MeshZoneAddress"):
					obj = &meshzoneaddress_k8s.MeshZoneAddress{}
				case strings.Contains(yamlObj, "kind: Service"):
					obj = &kube_core.Service{}
				case strings.Contains(yamlObj, "kind: EndpointSlice"):
					obj = &kube_discovery.EndpointSlice{}
				case strings.Contains(yamlObj, "kind: Namespace"):
					obj = &kube_core.Namespace{}
				case strings.Contains(yamlObj, "kind: Node"):
					obj = &kube_core.Node{}
				default:
					continue
				}
				Expect(yaml.Unmarshal([]byte(yamlObj), obj)).To(Succeed())
				objects = append(objects, obj)
			}

			kubeClient := kube_client_fake.NewClientBuilder().
				WithObjects(objects...).
				WithScheme(k8sClientScheme).
				Build()

			recorder := kube_events.NewFakeRecorder(10)
			reconciler := &MeshZoneAddressReconciler{
				Client:        kubeClient,
				Log:           logr.Discard(),
				Scheme:        k8sClientScheme,
				EventRecorder: recorder,
				ZoneName:      testZone,
			}

			svcName := given.svcName
			if svcName == "" {
				svcName = testSvcName
			}

			// when
			_, err = reconciler.Reconcile(context.Background(), kube_ctrl.Request{
				NamespacedName: kube_types.NamespacedName{Name: svcName, Namespace: testNamespace},
			})
			Expect(err).ToNot(HaveOccurred())

			// then
			mzas := &meshzoneaddress_k8s.MeshZoneAddressList{}
			Expect(kubeClient.List(context.Background(), mzas)).To(Succeed())
			Expect(yaml.Marshal(mzas)).To(MatchGoldenYAML("testdata", "meshzoneaddress", given.outputFile))
		},
		Entry("skips service without zone-proxy label", testCase{
			inputFile:  "01.no-label.resources.yaml",
			outputFile: "01.no-label.mza.yaml",
		}),
		Entry("skips when no ready endpoints", testCase{
			inputFile:  "02.no-ready-endpoints.resources.yaml",
			outputFile: "02.no-ready-endpoints.mza.yaml",
		}),
		Entry("LoadBalancer hostname takes precedence over IP", testCase{
			inputFile:  "03.lb-hostname.resources.yaml",
			outputFile: "03.lb-hostname.mza.yaml",
		}),
		Entry("LoadBalancer IP when no hostname", testCase{
			inputFile:  "04.lb-ip.resources.yaml",
			outputFile: "04.lb-ip.mza.yaml",
		}),
		Entry("LoadBalancer not yet provisioned", testCase{
			inputFile:  "05.lb-not-ready.resources.yaml",
			outputFile: "05.lb-not-ready.mza.yaml",
		}),
		Entry("NodePort uses node ExternalIP", testCase{
			inputFile:  "06.nodeport-external.resources.yaml",
			outputFile: "06.nodeport-external.mza.yaml",
		}),
		Entry("NodePort falls back to InternalIP when no ExternalIP", testCase{
			inputFile:  "07.nodeport-internal.resources.yaml",
			outputFile: "07.nodeport-internal.mza.yaml",
		}),
		Entry("externalIPs take precedence over service type", testCase{
			inputFile:  "08.external-ips.resources.yaml",
			outputFile: "08.external-ips.mza.yaml",
		}),
		Entry("ClusterIP without externalIPs emits warning, no MeshZoneAddress", testCase{
			inputFile:  "09.clusterip-no-address.resources.yaml",
			outputFile: "09.clusterip-no-address.mza.yaml",
		}),
		Entry("mesh name read from namespace label", testCase{
			inputFile:  "10.mesh-from-namespace.resources.yaml",
			outputFile: "10.mesh-from-namespace.mza.yaml",
		}),
		Entry("falls back to default mesh", testCase{
			inputFile:  "11.default-mesh.resources.yaml",
			outputFile: "11.default-mesh.mza.yaml",
		}),
		Entry("deletes MeshZoneAddress when label removed", testCase{
			inputFile:  "12.delete-on-label-removed.resources.yaml",
			outputFile: "12.delete-on-label-removed.mza.yaml",
		}),
		Entry("updates existing MeshZoneAddress", testCase{
			inputFile:  "13.update-existing.resources.yaml",
			outputFile: "13.update-existing.mza.yaml",
		}),
	)
})
