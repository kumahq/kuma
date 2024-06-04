package controllers_test

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_discovery "k8s.io/api/discovery/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_intstr "k8s.io/apimachinery/pkg/util/intstr"
	kube_record "k8s.io/client-go/tools/record"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	. "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

var _ = Describe("PodReconciler", func() {
	var kubeClient kube_client.Client
	var reconciler kube_reconcile.Reconciler

	BeforeEach(func() {
		kubeClient = kube_client_fake.NewClientBuilder().WithScheme(k8sClientScheme).WithObjects(
			&kube_core.Namespace{
				ObjectMeta: kube_meta.ObjectMeta{
					Name: "demo",
				},
			},
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
					Labels: map[string]string{
						"app": "sample",
					},
					UID: "pod-with-kuma-sidecar-but-no-ip-demo",
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
					Labels: map[string]string{
						"app": "sample",
					},
					UID: "pod-with-kuma-sidecar-and-ip-demo",
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
						{
							Ports: []kube_core.ContainerPort{
								{ContainerPort: 9090},
							},
						},
					},
				},
				Status: kube_core.PodStatus{
					PodIP: "192.168.0.1",
					ContainerStatuses: []kube_core.ContainerStatus{
						{
							State: kube_core.ContainerState{},
						},
					},
				},
			},
			&kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "pod-ingress",
					Annotations: map[string]string{
						"kuma.io/sidecar-injected": "true",
						"kuma.io/ingress":          "enabled",
					},
					Labels: map[string]string{
						"app": "ingress",
					},
					UID: "pod-ingress-kuma-demo",
				},
				Status: kube_core.PodStatus{
					PodIP: "192.168.0.2",
					ContainerStatuses: []kube_core.ContainerStatus{
						{
							State: kube_core.ContainerState{},
						},
					},
				},
			},
			&kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "kuma-system",
					Name:      "pod-ingress",
					Annotations: map[string]string{
						"kuma.io/sidecar-injected": "true",
						"kuma.io/ingress":          "enabled",
					},
					Labels: map[string]string{
						"app": "ingress",
					},
					UID: "pod-ingress-kuma-system",
				},
				Status: kube_core.PodStatus{
					PodIP: "192.168.0.3",
					ContainerStatuses: []kube_core.ContainerStatus{
						{
							State: kube_core.ContainerState{},
						},
					},
				},
			},
			&kube_core.Service{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace:   "kuma-system",
					Name:        "ingress",
					Annotations: map[string]string{},
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
					},
					Selector: map[string]string{
						"app": "ingress",
					},
				},
			},
			&kube_discovery.EndpointSlice{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "kuma-system",
					Name:      "ingress-ip4",
					Labels: map[string]string{
						kube_discovery.LabelServiceName: "ingress",
					},
				},
				AddressType: kube_discovery.AddressTypeIPv4,
				Endpoints: []kube_discovery.Endpoint{{
					Addresses: []string{"192.168.0.3"},
					TargetRef: &kube_core.ObjectReference{
						Kind:      "Pod",
						Name:      "pod-ingress",
						Namespace: "kuma-system",
						UID:       "pod-ingress-kuma-system",
					},
				}},
			},
			&kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "pod-egress",
					Annotations: map[string]string{
						"kuma.io/sidecar-injected": "true",
						"kuma.io/egress":           "enabled",
					},
					Labels: map[string]string{
						"app": "egress",
					},
					UID: "pod-egress-demo",
				},
				Status: kube_core.PodStatus{
					PodIP: "192.168.0.4",
					ContainerStatuses: []kube_core.ContainerStatus{
						{
							State: kube_core.ContainerState{},
						},
					},
				},
			},
			&kube_core.Pod{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "kuma-system",
					Name:      "pod-egress",
					Annotations: map[string]string{
						"kuma.io/sidecar-injected": "true",
						"kuma.io/egress":           "enabled",
					},
					Labels: map[string]string{
						"app": "egress",
					},
					UID: "pod-egress-kuma-system",
				},
				Status: kube_core.PodStatus{
					PodIP: "192.168.0.5",
					ContainerStatuses: []kube_core.ContainerStatus{
						{
							State: kube_core.ContainerState{},
						},
					},
				},
			},
			&kube_core.Service{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace:   "kuma-system",
					Name:        "egress",
					Annotations: map[string]string{},
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
					},
					Selector: map[string]string{
						"app": "egress",
					},
				},
			},
			&kube_discovery.EndpointSlice{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "kuma-system",
					Name:      "egress-ip4",
					Labels: map[string]string{
						kube_discovery.LabelServiceName: "egress",
					},
				},
				AddressType: kube_discovery.AddressTypeIPv4,
				Endpoints: []kube_discovery.Endpoint{{
					Addresses: []string{"192.168.0.5"},
					TargetRef: &kube_core.ObjectReference{
						Kind:      "Pod",
						Name:      "pod-egress",
						Namespace: "kuma-system",
						UID:       "pod-egress-kuma-system",
					},
				}},
			},
			&kube_core.Service{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "example",
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
							AppProtocol: pointer.To("http"),
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
					Selector: map[string]string{
						"app": "sample",
					},
				},
			},
			&kube_discovery.EndpointSlice{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "example-ip4",
					Labels: map[string]string{
						kube_discovery.LabelServiceName: "example",
					},
				},
				AddressType: kube_discovery.AddressTypeIPv4,
				Endpoints: []kube_discovery.Endpoint{{
					Addresses: []string{"192.168.0.1"},
					TargetRef: &kube_core.ObjectReference{
						Kind:      "Pod",
						Name:      "pod-with-kuma-sidecar-and-ip",
						Namespace: "demo",
						UID:       "pod-with-kuma-sidecar-and-ip-demo",
					},
				}},
			},
			&kube_core.Service{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "manual-example",
				},
				Spec: kube_core.ServiceSpec{
					ClusterIP: "192.168.0.1",
					Ports: []kube_core.ServicePort{
						{
							Port: 90,
							TargetPort: kube_intstr.IntOrString{
								Type:   kube_intstr.Int,
								IntVal: 9090,
							},
							AppProtocol: pointer.To("http"),
						},
					},
				},
			},
			&kube_discovery.EndpointSlice{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "manual-example-ip4",
					Labels: map[string]string{
						kube_discovery.LabelServiceName: "manual-example",
					},
				},
				AddressType: kube_discovery.AddressTypeIPv4,
				Endpoints: []kube_discovery.Endpoint{{
					Addresses: []string{"192.168.0.1"},
					TargetRef: &kube_core.ObjectReference{
						Kind:      "Pod",
						Name:      "pod-with-kuma-sidecar-and-ip",
						Namespace: "demo",
						UID:       "pod-with-kuma-sidecar-and-ip-demo",
					},
				}},
			},
			&kube_core.Service{
				ObjectMeta: kube_meta.ObjectMeta{
					Namespace: "demo",
					Name:      "ignored-service",
					Annotations: map[string]string{
						metadata.KumaIgnoreAnnotation: metadata.AnnotationTrue,
					},
				},
				Spec: kube_core.ServiceSpec{
					ClusterIP: "192.168.0.1",
					Ports: []kube_core.ServicePort{
						{
							Port: 85,
							TargetPort: kube_intstr.IntOrString{
								Type:   kube_intstr.Int,
								IntVal: 8080,
							},
						},
					},
					Selector: map[string]string{
						"app": "sample",
					},
				},
			}).Build()

		reconciler = &PodReconciler{
			Client:        kubeClient,
			EventRecorder: kube_record.NewFakeRecorder(10),
			Scheme:        k8sClientScheme,
			Log:           core.Log.WithName("test"),
			PodConverter: PodConverter{
				ResourceConverter: k8s.NewSimpleConverter(),
			},
			SystemNamespace:   "kuma-system",
			ResourceConverter: k8s.NewSimpleConverter(),
		}
	})

	It("should ignore deleted Pods", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "non-existing-key"},
		}

		// when
		result, err := reconciler.Reconcile(context.Background(), req)
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
		Expect(dataplanes.Items).To(BeEmpty())
	})

	It("should ignore Pods without Kuma sidecar", func() {
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

		// when
		dataplanes := &mesh_k8s.DataplaneList{}
		err = kubeClient.List(context.Background(), dataplanes)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(dataplanes.Items).To(BeEmpty())
	})

	It("should not reconcile Ingress with namespace other than system", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "pod-ingress"},
		}

		// when
		_, err := reconciler.Reconcile(context.Background(), req)

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(`Ingress can only be deployed in system namespace "kuma-system"`))
	})

	It("should not reconcile Egress with namespace other than system", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "pod-egress"},
		}

		// when
		_, err := reconciler.Reconcile(context.Background(), req)

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(`Egress can only be deployed in system namespace "kuma-system"`))
	})

	It("should reconcile Ingress with system namespace", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "kuma-system", Name: "pod-ingress"},
		}

		// when
		_, err := reconciler.Reconcile(context.Background(), req)

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should reconcile Egress with system namespace", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "kuma-system", Name: "pod-egress"},
		}

		// when
		_, err := reconciler.Reconcile(context.Background(), req)

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should ignore Pods without Kuma sidecar", func() {
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

		// when
		dataplanes := &mesh_k8s.DataplaneList{}
		err = kubeClient.List(context.Background(), dataplanes)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(dataplanes.Items).To(BeEmpty())
	})

	It("should ignore Pods without IP address", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "pod-with-kuma-sidecar-but-no-ip"},
		}

		// when
		result, err := reconciler.Reconcile(context.Background(), req)

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
		Expect(dataplanes.Items).To(BeEmpty())
	})

	It("should generate Dataplane resource for every Pod that has Kuma sidecar injected", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "pod-with-kuma-sidecar-and-ip"},
		}

		// when
		result, err := reconciler.Reconcile(context.Background(), req)

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
            uid: pod-with-kuma-sidecar-and-ip-demo
          resourceVersion: "1"
        spec:
          networking:
            address: 192.168.0.1
            inbound:
            - state: NotReady
              health: {}
              port: 8080
              tags:
                app: sample
                kuma.io/protocol: http
                kuma.io/service: example_demo_svc_80
                k8s.kuma.io/service-name: example
                k8s.kuma.io/service-port: "80"
                k8s.kuma.io/namespace: demo
            - state: NotReady
              health: {}
              port: 6060
              tags:
                app: sample
                kuma.io/service: example_demo_svc_6061
                kuma.io/protocol: tcp
                k8s.kuma.io/service-name: example
                k8s.kuma.io/service-port: "6061"
                k8s.kuma.io/namespace: demo
            - state: NotReady
              health: {}
              port: 9090
              tags:
                app: sample
                kuma.io/service: manual-example_demo_svc_90
                kuma.io/protocol: http
                k8s.kuma.io/service-name: manual-example
                k8s.kuma.io/service-port: "90"
                k8s.kuma.io/namespace: demo
`))
	})

	It("should update Dataplane resource, e.g. when new Services get registered", func() {
		// setup
		err := kubeClient.Create(context.Background(), &mesh_k8s.Dataplane{
			ObjectMeta: kube_meta.ObjectMeta{
				Namespace: "demo",
				Name:      "pod-with-kuma-sidecar-and-ip",
			},
			Spec: mesh_k8s.ToSpec(&mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{},
			}),
		})
		Expect(err).NotTo(HaveOccurred())

		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "pod-with-kuma-sidecar-and-ip"},
		}

		// when
		result, err := reconciler.Reconcile(context.Background(), req)

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
            uid: "pod-with-kuma-sidecar-and-ip-demo"
          resourceVersion: "2"
        spec:
          networking:
            address: 192.168.0.1
            inbound:
            - state: NotReady 
              health: {}
              port: 8080
              tags:
                app: sample
                kuma.io/protocol: http
                kuma.io/service: example_demo_svc_80
                k8s.kuma.io/service-name: example
                k8s.kuma.io/service-port: "80"
                k8s.kuma.io/namespace: demo
            - state: NotReady
              health: {}
              port: 6060
              tags:
                app: sample
                kuma.io/service: example_demo_svc_6061
                kuma.io/protocol: tcp
                k8s.kuma.io/service-name: example
                k8s.kuma.io/service-port: "6061"
                k8s.kuma.io/namespace: demo
            - state: NotReady
              health: {}
              port: 9090
              tags:
                app: sample
                kuma.io/service: manual-example_demo_svc_90
                kuma.io/protocol: http
                k8s.kuma.io/service-name: manual-example
                k8s.kuma.io/service-port: "90"
                k8s.kuma.io/namespace: demo
`))
	})

	It("should update Pods when ExternalService updated", func() {
		err := kubeClient.Create(context.Background(), &mesh_k8s.Dataplane{
			Mesh: "mesh-1",
			ObjectMeta: kube_meta.ObjectMeta{
				Namespace: "demo",
				Name:      "dp-1",
				OwnerReferences: []kube_meta.OwnerReference{{
					Controller: pointer.To(true),
					Kind:       "Pod",
					Name:       "dp-1",
				}},
			},
			Spec: mesh_k8s.ToSpec(&mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{},
			}),
		})
		Expect(err).NotTo(HaveOccurred())

		err = kubeClient.Create(context.Background(), &mesh_k8s.Dataplane{
			Mesh: "mesh-1",
			ObjectMeta: kube_meta.ObjectMeta{
				Namespace: "demo",
				Name:      "dp-2",
				OwnerReferences: []kube_meta.OwnerReference{{
					Controller: pointer.To(true),
					Kind:       "Pod",
					Name:       "dp-2",
				}},
			},
			Spec: mesh_k8s.ToSpec(&mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{},
			}),
		})
		Expect(err).NotTo(HaveOccurred())

		err = kubeClient.Create(context.Background(), &mesh_k8s.Dataplane{
			Mesh: "mesh-2",
			ObjectMeta: kube_meta.ObjectMeta{
				Namespace: "demo",
				Name:      "dp-3",
				OwnerReferences: []kube_meta.OwnerReference{{
					Controller: pointer.To(true),
					Kind:       "Pod",
					Name:       "dp-3",
				}},
			},
			Spec: mesh_k8s.ToSpec(&mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{},
			}),
		})
		Expect(err).NotTo(HaveOccurred())
	})
})
