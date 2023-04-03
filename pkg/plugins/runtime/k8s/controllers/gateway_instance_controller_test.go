package controllers_test

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_apps "k8s.io/api/apps/v1"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_ctrl "sigs.k8s.io/controller-runtime"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	kube_client_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	kube_reconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kumahq/kuma/pkg/core"
	kuma_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	. "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/controllers"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

var _ = Describe("GatewayInstanceReconciler", func() {
	var kubeClient kube_client.Client
	var reconciler kube_reconcile.Reconciler

	BeforeEach(func() {
		if err := kuma_k8s.AddToScheme(k8sClientScheme); err != nil {
			fmt.Printf("could not add %q to scheme", kuma_k8s.GroupVersion)
		}
		k8sClientScheme.AddKnownTypes(kuma_k8s.GroupVersion,
			&kuma_k8s.MeshGatewayInstance{},
		)
		kubeClient = kube_client_fake.NewClientBuilder().WithScheme(k8sClientScheme).WithObjects(&kube_core.Namespace{
			ObjectMeta: kube_meta.ObjectMeta{
				Name: "non-system-ns-with-sidecar-injection",
				Labels: map[string]string{
					metadata.KumaSidecarInjectionAnnotation: metadata.AnnotationEnabled,
				},
			},
		}, &kuma_k8s.MeshGatewayInstance{
			ObjectMeta: kube_meta.ObjectMeta{
				Namespace: "non-system-ns-with-sidecar-injection",
				Name:      "mesh-gateway-instance",
			},
			Spec: kuma_k8s.MeshGatewayInstanceSpec{
				MeshGatewayCommonConfig: kuma_k8s.MeshGatewayCommonConfig{
					Replicas:    1,
					ServiceType: "LoadBalancer",
					ServiceTemplate: kuma_k8s.MeshGatewayServiceTemplate{
						Metadata: kuma_k8s.MeshGatewayObjectMetadata{
							Annotations: map[string]string{
								"testServiceAnnotation": "testServiceAnnotationValue",
							},
							Labels: map[string]string{
								"testServiceLabel": "testServiceLabelValue",
							},
						},
					},
					PodTemplate: kuma_k8s.MeshGatewayPodTemplate{
						Metadata: kuma_k8s.MeshGatewayObjectMetadata{
							Annotations: map[string]string{
								"testPodAnnotation": "testPodAnnotationValue",
							},
							Labels: map[string]string{
								"testPodLabel": "testPodLabelValue",
							},
						},
						Spec: kuma_k8s.MeshGatewayPodSpec{
							ServiceAccountName: "test",
							PodSecurityContext: kuma_k8s.PodSecurityContext{
								FSGroup: pointer.To(int64(4000)),
							},
							Container: kuma_k8s.Container{
								SecurityContext: kuma_k8s.SecurityContext{
									ReadOnlyRootFilesystem: pointer.To(false),
								},
							},
						},
					},
				},
				Tags: map[string]string{
					"kuma.io/service": "demo-app-gateway",
				},
			},
		}).Build()

		reconciler = &GatewayInstanceReconciler{
			Client: kubeClient,
			Log:    core.Log.WithName("test"),
		}
	})

	It("should create correct service and deployment objects", func() {
		Skip("Currently broken") // TODO: Fix this w/@mikebeaumont
		// given
		mgReq := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{
				Namespace: "non-system-ns-with-sidecar-injection",
				Name:      "mesh-gateway-instance",
			},
		}

		// when
		result, err := reconciler.Reconcile(context.Background(), mgReq)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		// and service is created with correctly propagated fields
		svc := &kube_core.Service{}
		svcReq := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{
				Namespace: "non-system-ns-with-sidecar-injection",
				Name:      "mesh-gateway-instance",
			},
		}
		err = kubeClient.Get(context.Background(), svcReq.NamespacedName, svc)
		Expect(err).ToNot(HaveOccurred())
		Expect(svc.GetAnnotations()).To(HaveKey("testServiceAnnotation"))
		Expect(svc.GetAnnotations()[metadata.KumaGatewayAnnotation]).To(Equal(metadata.AnnotationBuiltin))
		Expect(svc.GetLabels()).To(HaveKey("testServiceLabel"))

		// and deployment is created with correctly propagated fields
		dep := &kube_apps.Deployment{}
		depReq := kube_ctrl.Request{
			NamespacedName: kube_types.NamespacedName{
				Namespace: "non-system-ns-with-sidecar-injection",
				Name:      "mesh-gateway-instance",
			},
		}
		err = kubeClient.Get(context.Background(), depReq.NamespacedName, dep)
		Expect(err).ToNot(HaveOccurred())
		Expect(dep.Spec.Template.GetAnnotations()).To(HaveKey(metadata.KumaGatewayAnnotation))
		Expect(dep.Spec.Template.GetAnnotations()).To(HaveKey("testPodAnnotation"))
		Expect(dep.Spec.Template.GetAnnotations()[metadata.KumaTagsAnnotation]).To(Equal(`'{"kuma.io/service":"demo-app_gateway"}'`))
		Expect(dep.Spec.Template.GetLabels()).To(HaveKey("testPodLabel"))
		Expect(dep.Spec.Template.Spec.ServiceAccountName).To(Equal("test"))
	})
})
