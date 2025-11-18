package controllers_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_record "k8s.io/client-go/tools/record"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"

	workload_k8s "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/k8s/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s/native/api/v1alpha1"
	. "github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/controllers"
	. "github.com/kumahq/kuma/v2/pkg/test/matchers"
	yaml2 "github.com/kumahq/kuma/v2/pkg/util/yaml"
)

var _ = Describe("WorkloadController", func() {
	var kubeClient kube_client.Client
	var reconciler kube_reconcile.Reconciler

	type testCase struct {
		inputFile    string
		outputFile   string
		workloadName string
		namespace    string
	}

	DescribeTable("should reconcile workload",
		func(given testCase) {
			// given
			bytes, err := os.ReadFile(filepath.Join("testdata", "workload", given.inputFile))
			Expect(err).ToNot(HaveOccurred())
			var objects []kube_client.Object
			for _, yamlObj := range yaml2.SplitYAML(string(bytes)) {
				var obj kube_client.Object
				switch {
				case strings.Contains(yamlObj, "kind: Workload"):
					obj = &workload_k8s.Workload{}
				case strings.Contains(yamlObj, "kind: Dataplane"):
					obj = &mesh_k8s.Dataplane{}
				case strings.Contains(yamlObj, "kind: Mesh"):
					obj = &mesh_k8s.Mesh{}
				}
				Expect(yaml.Unmarshal([]byte(yamlObj), obj)).To(Succeed())
				objects = append(objects, obj)
			}

			kubeClient = kube_client_fake.NewClientBuilder().
				WithObjects(objects...).
				WithScheme(k8sClientScheme).
				Build()

			reconciler = &WorkloadReconciler{
				Client:            kubeClient,
				Log:               logr.Discard(),
				Scheme:            k8sClientScheme,
				EventRecorder:     kube_record.NewFakeRecorder(10),
				ResourceConverter: k8s.NewSimpleConverter(),
			}

			key := kube_types.NamespacedName{
				Name:      given.workloadName,
				Namespace: given.namespace,
			}

			// when
			_, err = reconciler.Reconcile(context.Background(), kube_ctrl.Request{NamespacedName: key})

			// then
			Expect(err).ToNot(HaveOccurred())

			workloads := &workload_k8s.WorkloadList{}
			Expect(kubeClient.List(context.Background(), workloads)).To(Succeed())
			Expect(yaml.Marshal(workloads)).To(MatchGoldenYAML("testdata", "workload", given.outputFile))
		},
		Entry("should create Workload when Dataplane with workload label exists", testCase{
			inputFile:    "01.resources.yaml",
			outputFile:   "01.workload.yaml",
			workloadName: "test-workload",
			namespace:    "demo",
		}),
		Entry("should not create Workload when Dataplane has no workload label", testCase{
			inputFile:    "02.resources.yaml",
			outputFile:   "02.workload.yaml",
			workloadName: "nonexistent-workload",
			namespace:    "demo",
		}),
		Entry("should keep Workload when multiple Dataplanes reference it", testCase{
			inputFile:    "03.resources.yaml",
			outputFile:   "03.workload.yaml",
			workloadName: "shared-workload",
			namespace:    "demo",
		}),
		Entry("should delete Workload when no Dataplanes reference it", testCase{
			inputFile:    "04.resources.yaml",
			outputFile:   "04.workload.yaml",
			workloadName: "orphaned-workload",
			namespace:    "demo",
		}),
		Entry("should create Workload in correct namespace and mesh", testCase{
			inputFile:    "05.resources.yaml",
			outputFile:   "05.workload.yaml",
			workloadName: "custom-workload",
			namespace:    "demo",
		}),
		Entry("should not delete Workload when one Dataplane still references it", testCase{
			inputFile:    "06.resources.yaml",
			outputFile:   "06.workload.yaml",
			workloadName: "persistent-workload",
			namespace:    "demo",
		}),
	)
})
