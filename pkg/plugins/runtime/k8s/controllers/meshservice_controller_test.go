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
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_events "k8s.io/client-go/tools/events"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	kube_event "sigs.k8s.io/controller-runtime/pkg/event"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"

	meshservice_k8s "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/k8s/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/api/v1alpha1"
	. "github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/controllers"
	. "github.com/kumahq/kuma/v3/pkg/test/matchers"
	yaml2 "github.com/kumahq/kuma/v3/pkg/util/yaml"
)

var _ = Describe("MeshServiceController", func() {
	var kubeClient kube_client.Client
	var reconciler kube_reconcile.Reconciler

	type testCase struct {
		inputFile  string
		outputFile string
	}

	DescribeTable("should reconcile service",
		func(given testCase) {
			// given
			bytes, err := os.ReadFile(filepath.Join("testdata", "meshservice", given.inputFile))
			Expect(err).ToNot(HaveOccurred())
			var objects []kube_client.Object
			for _, yamlObj := range yaml2.SplitYAML(string(bytes)) {
				var obj kube_client.Object
				switch {
				case strings.Contains(yamlObj, "kind: MeshService"):
					obj = &meshservice_k8s.MeshService{}
				case strings.Contains(yamlObj, "kind: Service"):
					obj = &kube_core.Service{}
				case strings.Contains(yamlObj, "kind: EndpointSlice"):
					obj = &kube_discovery.EndpointSlice{}
				case strings.Contains(yamlObj, "kind: Namespace"):
					obj = &kube_core.Namespace{}
				case strings.Contains(yamlObj, "kind: Pod"):
					obj = &kube_core.Pod{}
				case strings.Contains(yamlObj, "kind: Mesh"):
					obj = &v1alpha1.Mesh{}
				}
				Expect(yaml.Unmarshal([]byte(yamlObj), obj)).To(Succeed())
				objects = append(objects, obj)
			}

			kubeClient = kube_client_fake.NewClientBuilder().
				WithObjects(objects...).
				WithScheme(k8sClientScheme).
				Build()

			reconciler = &MeshServiceReconciler{
				Client:        kubeClient,
				Log:           logr.Discard(),
				Scheme:        k8sClientScheme,
				EventRecorder: kube_events.NewFakeRecorder(10),
			}

			key := kube_types.NamespacedName{
				Name:      "example",
				Namespace: "demo",
			}

			// when
			_, err = reconciler.Reconcile(context.Background(), kube_ctrl.Request{NamespacedName: key})

			// then
			Expect(err).ToNot(HaveOccurred())

			mss := &meshservice_k8s.MeshServiceList{}
			Expect(kubeClient.List(context.Background(), mss)).To(Succeed())
			Expect(yaml.Marshal(mss)).To(MatchGoldenYAML("testdata", "meshservice", given.outputFile))
		},
		Entry("with service in sidecar injection namespace", testCase{
			inputFile:  "01.resources.yaml",
			outputFile: "01.meshservice.yaml",
		}),
		Entry("with service with mesh label", testCase{
			inputFile:  "02.resources.yaml",
			outputFile: "02.meshservice.yaml",
		}),
		Entry("without mesh label and sidecar injection namespace", testCase{
			inputFile:  "03.resources.yaml",
			outputFile: "03.meshservice.yaml",
		}),
		Entry("without mesh", testCase{
			inputFile:  "04.resources.yaml",
			outputFile: "04.meshservice.yaml",
		}),
		Entry("service for kuma gateway", testCase{
			inputFile:  "05.resources.yaml",
			outputFile: "05.meshservice.yaml",
		}),
		Entry("service for delegated gateway (annotation on Pod)", testCase{
			inputFile:  "06.resources.yaml",
			outputFile: "06.meshservice.yaml",
		}),
		Entry("service for pod opting out with kuma.io/gateway: disabled", testCase{
			inputFile:  "07.resources.yaml",
			outputFile: "07.meshservice.yaml",
		}),
		Entry("service for headless Service", testCase{
			inputFile:  "headless.resources.yaml",
			outputFile: "headless.meshservice.yaml",
		}),
		Entry("with kuma.io/ignore", testCase{
			inputFile:  "ignore.resources.yaml",
			outputFile: "ignore.meshservice.yaml",
		}),
		Entry("headless gateway service is unaffected by meshServices.mode", testCase{
			inputFile:  "headless-gateway-disabled.resources.yaml",
			outputFile: "headless-gateway-disabled.meshservice.yaml",
		}),
		Entry("with Service selector matching Pod labels", testCase{
			inputFile:  "skip-inbound-tags.resources.yaml",
			outputFile: "skip-inbound-tags.meshservice.yaml",
		}),
	)

	DescribeTable("GatewayAnnotationChangedPredicate",
		func(oldValue string, newValue string, expected bool) {
			pod := func(value string) *kube_core.Pod {
				pod := &kube_core.Pod{ObjectMeta: kube_meta.ObjectMeta{Name: "pod", Namespace: "demo"}}
				if value != "" {
					pod.Annotations = map[string]string{"kuma.io/gateway": value}
				}
				return pod
			}

			changed := GatewayAnnotationChangedPredicate{}.Update(kube_event.UpdateEvent{
				ObjectOld: pod(oldValue),
				ObjectNew: pod(newValue),
			})

			Expect(changed).To(Equal(expected))
		},
		Entry("annotation added", "", "enabled", true),
		Entry("annotation removed", "enabled", "", true),
		Entry("gateway turned off", "enabled", "disabled", true),
		Entry("gateway turned on", "disabled", "true", true),
		Entry("same value", "enabled", "enabled", false),
		Entry("equivalent values", "enabled", "true", false),
		Entry("annotation added as disabled", "", "disabled", false),
	)
})
