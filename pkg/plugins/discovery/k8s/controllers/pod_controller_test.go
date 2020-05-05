package controllers_test

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/plugins/discovery/k8s/controllers"

	"github.com/Kong/kuma/pkg/core"

	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	kube_handler "sigs.k8s.io/controller-runtime/pkg/handler"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
	kube_reconile "sigs.k8s.io/controller-runtime/pkg/reconcile"

	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_intstr "k8s.io/apimachinery/pkg/util/intstr"
	kube_record "k8s.io/client-go/tools/record"

	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

var _ = Describe("PodReconciler", func() {

	var kubeClient kube_client.Client
	var reconciler kube_reconcile.Reconciler

	BeforeEach(func() {
		kubeClient = kube_client_fake.NewFakeClientWithScheme(
			k8sClientScheme,
			&kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "pod-without-kuma-sidecar",
				},
			},
			&kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "pod-with-kuma-sidecar-but-no-ip",
					Annotations: map[string]string{
						"kuma.io/sidecar-injected": "true",
					},
				},
			},
			&kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "pod-with-kuma-sidecar-and-ip",
					Annotations: map[string]string{
						"kuma.io/mesh":             "poc",
						"kuma.io/sidecar-injected": "true",
					},
				},
				Spec: kube_core.PodSpec{
					Containers: []kube_core.Container{
						{
							Ports: []kube_core.ContainerPort{
								{ContainerPort: 8080},
							},
						},
						{
							Ports: []kube_core.ContainerPort{
								{ContainerPort: 6060, Name: "metrics"},
							},
						},
					},
				},
				Status: kube_core.PodStatus{
					PodIP: "192.168.0.1",
				},
			},
			&kube_core.Service{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "example",
					Annotations: map[string]string{
						"80.service.kuma.io/protocol": "http",
					},
				},
				Spec: kube_core.ServiceSpec{
					ClusterIP: "192.168.0.1",
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
							Port:     6061,
							TargetPort: kube_intstr.IntOrString{
								Type:   kube_intstr.String,
								StrVal: "metrics",
							},
						},
					},
				},
			})

		reconciler = &PodReconciler{
			Client:        kubeClient,
			EventRecorder: kube_record.NewFakeRecorder(10),
			Scheme:        k8sClientScheme,
			Log:           core.Log.WithName("test"),
		}
	})

	It("should ignore deleted Pods", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "non-existing-key"},
		}

		// when
		result, err := reconciler.Reconcile(req)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(result).To(BeZero())

		// when
		dataplanes := &mesh_k8s.DataplaneList{}
		err = kubeClient.List(context.Background(), dataplanes)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(dataplanes.Items).To(HaveLen(0))
	})

	It("should ignore Pods without Kuma sidecar", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "pod-without-kuma-sidecar"},
		}

		// when
		result, err := reconciler.Reconcile(req)

		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(result).To(BeZero())

		// when
		dataplanes := &mesh_k8s.DataplaneList{}
		err = kubeClient.List(context.Background(), dataplanes)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(dataplanes.Items).To(HaveLen(0))
	})

	It("should ignore Pods without Kuma sidecar", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "pod-without-kuma-sidecar"},
		}

		// when
		result, err := reconciler.Reconcile(req)

		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(result).To(BeZero())

		// when
		dataplanes := &mesh_k8s.DataplaneList{}
		err = kubeClient.List(context.Background(), dataplanes)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(dataplanes.Items).To(HaveLen(0))
	})

	It("should ignore Pods without IP address", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "pod-with-kuma-sidecar-but-no-ip"},
		}

		// when
		result, err := reconciler.Reconcile(req)

		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(result).To(BeZero())

		// when
		dataplanes := &mesh_k8s.DataplaneList{}
		err = kubeClient.List(context.Background(), dataplanes)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(dataplanes.Items).To(HaveLen(0))
	})

	It("should generate Dataplane resource for every Pod that has Kuma sidecar injected", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "pod-with-kuma-sidecar-and-ip"},
		}

		// when
		result, err := reconciler.Reconcile(req)

		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(result).To(BeZero())

		// when
		dataplanes := &mesh_k8s.DataplaneList{}
		err = kubeClient.List(context.Background(), dataplanes)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(dataplanes.Items).To(HaveLen(1))

		// when
		actual, err := json.Marshal(dataplanes.Items[0])
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(actual).To(MatchYAML(`
        mesh: poc
        metadata:
          creationTimestamp: null
          name: pod-with-kuma-sidecar-and-ip
          namespace: demo
          ownerReferences:
          - apiVersion: v1
            blockOwnerDeletion: true
            controller: true
            kind: Pod
            name: pod-with-kuma-sidecar-and-ip
            uid: ""
          resourceVersion: "1"
        spec:
          networking:
            address: 192.168.0.1
            inbound:
            - port: 8080
              tags:
                protocol: http
                service: example.demo.svc:80
            - port: 6060
              tags:
                service: example.demo.svc:6061
                protocol: tcp
`))
	})

	It("should update Dataplane resource, e.g. when new Services get registered", func() {
		// setup
		err := kubeClient.Create(context.Background(), &mesh_k8s.Dataplane{
			ObjectMeta: kube_meta.ObjectMeta{
				Namespace: "demo",
				Name:      "pod-with-kuma-sidecar-and-ip",
			},
			Spec: map[string]interface{}{
				"networking": map[string]interface{}{},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "pod-with-kuma-sidecar-and-ip"},
		}

		// when
		result, err := reconciler.Reconcile(req)

		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(result).To(BeZero())

		// when
		dataplanes := &mesh_k8s.DataplaneList{}
		err = kubeClient.List(context.Background(), dataplanes)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(dataplanes.Items).To(HaveLen(1))

		// when
		actual, err := json.Marshal(dataplanes.Items[0])
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(actual).To(MatchYAML(`
        apiVersion: kuma.io/v1alpha1
        kind: Dataplane
        mesh: poc
        metadata:
          creationTimestamp: null
          name: pod-with-kuma-sidecar-and-ip
          namespace: demo
          ownerReferences:
          - apiVersion: v1
            blockOwnerDeletion: true
            controller: true
            kind: Pod
            name: pod-with-kuma-sidecar-and-ip
            uid: ""
          resourceVersion: "2"
        spec:
          networking:
            address: 192.168.0.1
            inbound:
            - port: 8080
              tags:
                protocol: http
                service: example.demo.svc:80
            - port: 6060
              tags:
                service: example.demo.svc:6061
                protocol: tcp
`))
	})
})

var _ = Describe("DataplaneToSameMeshDataplanesMapper", func() {

	isController := true
	blockOwnerDeletion := true

	var kubeClient kube_client.Client

	BeforeEach(func() {
		kubeClient = kube_client_fake.NewFakeClientWithScheme(
			k8sClientScheme,
			&mesh_k8s.Dataplane{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "existing-app-without-other-dataplanes-in-that-mesh",
				},
				Mesh: "poc",
			},
			&mesh_k8s.Dataplane{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "client-app",
					Name:      "existing-app-with-other-dataplanes-in-that-mesh",
				},
				Mesh: "default",
			},
			&mesh_k8s.Dataplane{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "service1-app",
					Name:      "other-app-in-that-mesh",
					OwnerReferences: []kube_meta.OwnerReference{
						{
							APIVersion:         "",
							Kind:               "Pod",
							Name:               "other-app-in-that-mesh",
							UID:                kube_types.UID("abcdefgh"),
							Controller:         &isController,
							BlockOwnerDeletion: &blockOwnerDeletion,
						},
					},
				},
				Mesh: "default",
			},
			&mesh_k8s.Dataplane{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "service2-app",
					Name:      "yet-another-app-in-that-mesh",
					// should be ignored because it has no owner reference to a Pod
				},
				Mesh: "default",
			},
		)
	})

	type testCase struct {
		obj      kube_handler.MapObject
		expected []kube_reconile.Request
	}

	DescribeTable("should map a Dataplane to its peers",
		func(given testCase) {
			// setup
			mapper := &DataplaneToSameMeshDataplanesMapper{
				Client: kubeClient,
				Log:    core.Log.WithName("dataplane-to-dataplanes-mapper"),
			}

			// when
			requests := mapper.Map(given.obj)

			// then
			Expect(requests).To(Equal(given.expected))
		},
		Entry("nil object", testCase{
			obj:      kube_handler.MapObject{},
			expected: nil,
		}),
		Entry("non-Dataplane object", testCase{
			obj: kube_handler.MapObject{
				Meta:   &kube_meta.ObjectMeta{},
				Object: &kube_core.Pod{},
			},
			expected: nil,
		}),
		Entry("deleted Dataplane object with no owner", testCase{
			obj: kube_handler.MapObject{
				Meta: &kube_meta.ObjectMeta{},
				Object: &mesh_k8s.Dataplane{
					ObjectMeta: kube_meta.ObjectMeta{
						Name:      "deleted-app",
						Namespace: "demo",
					},
				},
			},
			expected: nil,
		}),
		Entry("deleted Dataplane object with owner other than Pod", testCase{
			obj: kube_handler.MapObject{
				Meta: &kube_meta.ObjectMeta{},
				Object: &mesh_k8s.Dataplane{
					ObjectMeta: kube_meta.ObjectMeta{
						Name:      "deleted-app",
						Namespace: "demo",
						OwnerReferences: []kube_meta.OwnerReference{
							{
								APIVersion:         "apps/v1",
								Kind:               "Deployment",
								Name:               "example-deployment",
								UID:                kube_types.UID("abcdefgh"),
								Controller:         &isController,
								BlockOwnerDeletion: &blockOwnerDeletion,
							},
						},
					},
				},
			},
			expected: nil,
		}),
		Entry("deleted Dataplane object with owner Pod", testCase{
			obj: kube_handler.MapObject{
				Meta: &kube_meta.ObjectMeta{},
				Object: &mesh_k8s.Dataplane{
					ObjectMeta: kube_meta.ObjectMeta{
						Name:      "deleted-app",
						Namespace: "demo",
						OwnerReferences: []kube_meta.OwnerReference{
							{
								APIVersion:         "v1",
								Kind:               "Pod",
								Name:               "deleted-app",
								UID:                kube_types.UID("abcdefgh"),
								Controller:         &isController,
								BlockOwnerDeletion: &blockOwnerDeletion,
							},
						},
					},
				},
			},
			expected: []kube_reconile.Request{
				{
					NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "deleted-app"},
				},
			},
		}),
		Entry("existing Dataplane object with no other dataplanes in that mesh", testCase{
			obj: kube_handler.MapObject{
				Meta: &kube_meta.ObjectMeta{},
				Object: &mesh_k8s.Dataplane{
					Mesh: "poc",
					ObjectMeta: kube_meta.ObjectMeta{
						Name:      "existing-app-without-other-dataplanes-in-that-mesh",
						Namespace: "demo",
					},
				},
			},
			expected: nil, // should not return Dataplane itself unles it was deleted
		}),
		Entry("existing Dataplane object with other dataplanes in that mesh", testCase{
			obj: kube_handler.MapObject{
				Meta: &kube_meta.ObjectMeta{},
				Object: &mesh_k8s.Dataplane{
					Mesh: "default",
					ObjectMeta: kube_meta.ObjectMeta{
						Name:      "existing-app-with-other-dataplanes-in-that-mesh",
						Namespace: "client-app",
					},
				},
			},
			expected: []kube_reconile.Request{
				{
					NamespacedName: kube_types.NamespacedName{Namespace: "service1-app", Name: "other-app-in-that-mesh"},
				},
			},
		}),
	)

})
