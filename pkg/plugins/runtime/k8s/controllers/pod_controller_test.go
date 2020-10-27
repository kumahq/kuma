package controllers_test

import (
	"context"
	"encoding/json"

	"github.com/kumahq/kuma/pkg/dns/persistence"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers"

	"github.com/kumahq/kuma/pkg/core"

	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_intstr "k8s.io/apimachinery/pkg/util/intstr"
	kube_record "k8s.io/client-go/tools/record"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"

	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

type testPersistence struct {
}

func (t *testPersistence) Get() (persistence.VIPList, error) {
	return persistence.VIPList{}, nil
}

func (t *testPersistence) GetByMesh(mesh string) (persistence.VIPList, error) {
	return persistence.VIPList{}, nil
}

func (t *testPersistence) Set(mesh string, vips persistence.VIPList) error {
	return nil
}

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
					Labels: map[string]string{
						"app": "sample",
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
					Labels: map[string]string{
						"app": "sample",
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
				},
				Status: kube_core.PodStatus{
					PodIP: "192.168.0.1",
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
				},
				Status: kube_core.PodStatus{
					PodIP: "192.168.0.1",
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
					Selector: map[string]string{
						"app": "sample",
					},
				},
			})

		reconciler = &PodReconciler{
			Client:          kubeClient,
			EventRecorder:   kube_record.NewFakeRecorder(10),
			Scheme:          k8sClientScheme,
			Log:             core.Log.WithName("test"),
			SystemNamespace: "kuma-system",
			Persistence:     &testPersistence{},
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

	It("should not reconcile Ingress with namespace other than system", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "demo", Name: "pod-ingress"},
		}

		// when
		_, err := reconciler.Reconcile(req)

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(`Ingress can only be deployed in system namespace "kuma-system"`))
	})

	It("should reconcile Ingress with system namespace", func() {
		// given
		req := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{Namespace: "kuma-system", Name: "pod-ingress"},
		}

		// when
		_, err := reconciler.Reconcile(req)

		// then
		Expect(err).ToNot(HaveOccurred())
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
                app: sample
                kuma.io/protocol: http
                kuma.io/service: example_demo_svc_80
            - port: 6060
              tags:
                app: sample
                kuma.io/service: example_demo_svc_6061
                kuma.io/protocol: tcp
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
                app: sample
                kuma.io/protocol: http
                kuma.io/service: example_demo_svc_80
            - port: 6060
              tags:
                app: sample
                kuma.io/service: example_demo_svc_6061
                kuma.io/protocol: tcp
`))
	})
})
