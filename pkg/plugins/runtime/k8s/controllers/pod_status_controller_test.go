package controllers_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_record "k8s.io/client-go/tools/record"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	. "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers"
	util_k8s "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
	"github.com/kumahq/kuma/pkg/test/runtime"
)

var _ = Describe("PodStatusReconciler", func() {

	var kubeClient kube_client.Client
	var reconciler *PodStatusReconciler
	var postQuitCalled int

	BeforeEach(func() {
		kubeClient = kube_client_fake.NewClientBuilder().WithScheme(k8sClientScheme).WithObjects(
			&kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "pod-without-kuma-sidecar",
				},
				Status: kube_core.PodStatus{
					ContainerStatuses: []kube_core.ContainerStatus{
						{
							Name: "another-side-car",
							State: kube_core.ContainerState{
								Terminated: nil,
							},
						},
						{
							Name: "workload",
							State: kube_core.ContainerState{
								Terminated: &kube_core.ContainerStateTerminated{},
							},
						},
					},
				},
			},
			&kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "pod-with-kuma-sidecar-workload-not-owned-by-a-job",
					OwnerReferences: []kube_meta.OwnerReference{
						{
							Kind: "ReplicaSet",
						},
					},
				},
				Status: kube_core.PodStatus{
					ContainerStatuses: []kube_core.ContainerStatus{
						{
							Name: util_k8s.KumaSidecarContainerName,
							State: kube_core.ContainerState{
								Terminated: nil,
							},
						},
						{
							Name: "workload",
							State: kube_core.ContainerState{
								Terminated: nil,
							},
						},
					},
				},
			},
			&kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "pod-with-kuma-sidecar-workload-not-terminated",
					OwnerReferences: []kube_meta.OwnerReference{
						{
							Kind: "Job",
						},
					},
				},
				Status: kube_core.PodStatus{
					ContainerStatuses: []kube_core.ContainerStatus{
						{
							Name: util_k8s.KumaSidecarContainerName,
							State: kube_core.ContainerState{
								Terminated: nil,
							},
						},
						{
							Name: "workload",
							State: kube_core.ContainerState{
								Terminated: nil,
							},
						},
					},
				},
			},
			&kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "pod-with-kuma-sidecar-workload-not-terminated-successfully",
					OwnerReferences: []kube_meta.OwnerReference{
						{
							Kind: "Job",
						},
					},
				},
				Status: kube_core.PodStatus{
					ContainerStatuses: []kube_core.ContainerStatus{
						{
							Name: util_k8s.KumaSidecarContainerName,
							State: kube_core.ContainerState{
								Terminated: nil,
							},
						},
						{
							Name: "workload",
							State: kube_core.ContainerState{
								Terminated: &kube_core.ContainerStateTerminated{
									ExitCode: -1,
								},
							},
						},
					},
				},
			},
			&kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "pod-with-kuma-sidecar-workload-terminated",
					OwnerReferences: []kube_meta.OwnerReference{
						{
							Kind: "Job",
						},
					},
				},
				Status: kube_core.PodStatus{
					ContainerStatuses: []kube_core.ContainerStatus{
						{
							Name: util_k8s.KumaSidecarContainerName,
							State: kube_core.ContainerState{
								Terminated: nil,
							},
						},
						{
							Name: "workload",
							State: kube_core.ContainerState{
								Terminated: &kube_core.ContainerStateTerminated{
									ExitCode: 0,
								},
							},
						},
					},
				},
			},
			&mesh_k8s.Dataplane{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "pod-with-kuma-sidecar-workload-terminated",
				},
			},
		).Build()

		postQuitCalled = 0
		reconciler = &PodStatusReconciler{
			Client:            kubeClient,
			EventRecorder:     kube_record.NewFakeRecorder(10),
			Scheme:            k8sClientScheme,
			Log:               core.Log.WithName("test"),
			ResourceConverter: k8s.NewSimpleConverter(),
			EnvoyAdminClient: &runtime.DummyEnvoyAdminClient{
				PostQuitCalled: &postQuitCalled,
			},
		}
	})

	It("should ignore non-existing Pods", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "non-existing-pod"},
		}

		// when
		result, err := reconciler.Reconcile(context.Background(), req)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(result).To(BeZero())
		// and
		Expect(postQuitCalled).To(Equal(0))
	})

	It("should ignore Pods without kuma sidecar", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "pod-without-kuma-sidecar"},
		}

		// when
		result, err := reconciler.Reconcile(context.Background(), req)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(result).To(BeZero())
		// and
		Expect(postQuitCalled).To(Equal(0))
	})

	It("should ignore Pods with kuma sidecar, not owned by a job", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "pod-with-kuma-sidecar-workload-not-owned-by-a-job"},
		}

		// when
		result, err := reconciler.Reconcile(context.Background(), req)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(result).To(BeZero())
		// and
		Expect(postQuitCalled).To(Equal(0))
	})

	It("should ignore Pods with kuma sidecar terminated", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "pod-with-kuma-sidecar-terminated"},
		}

		// when
		result, err := reconciler.Reconcile(context.Background(), req)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(result).To(BeZero())
		// and
		Expect(postQuitCalled).To(Equal(0))
	})

	It("should ignore Pods with workload not terminated", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "pod-with-kuma-sidecar-workload-not-terminated"},
		}

		// when
		result, err := reconciler.Reconcile(context.Background(), req)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(result).To(BeZero())
		// and
		Expect(postQuitCalled).To(Equal(0))
	})

	It("should ignore Pods with workload not terminated successfully", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "pod-with-kuma-sidecar-workload-not-terminated-successfully"},
		}

		// when
		result, err := reconciler.Reconcile(context.Background(), req)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(result).To(BeZero())
		// and
		Expect(postQuitCalled).To(Equal(0))
	})

	It("should call envoy quit for Pods with workload terminated", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "pod-with-kuma-sidecar-workload-terminated"},
		}

		// when
		result, err := reconciler.Reconcile(context.Background(), req)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(result).To(BeZero())
		// and
		Expect(postQuitCalled).To(Equal(1))
	})
})
