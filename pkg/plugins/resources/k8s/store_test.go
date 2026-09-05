package k8s_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_apierrs "k8s.io/apimachinery/pkg/api/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s"
)

var _ = Describe("KubernetesStore.Create", func() {
	It("should surface a namespace-terminating Forbidden error distinctly from other create failures", func() {
		// The NamespaceLifecycle admission controller returns this exact shape
		// of StatusError when a create is attempted against a namespace that is
		// being deleted.
		namespaceTerminatingErr := &kube_apierrs.StatusError{
			ErrStatus: kube_meta.Status{
				Status:  kube_meta.StatusFailure,
				Reason:  kube_meta.StatusReasonForbidden,
				Message: `unable to create new content in namespace "demo" because it is being terminated`,
				Details: &kube_meta.StatusDetails{
					Causes: []kube_meta.StatusCause{{
						Type:    kube_core.NamespaceTerminatingCause,
						Message: `namespace "demo" is being terminated`,
						Field:   "metadata.namespace",
					}},
				},
			},
		}

		fakeClient := fake.NewClientBuilder().
			WithScheme(k8sClientScheme).
			WithInterceptorFuncs(interceptor.Funcs{
				Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
					return namespaceTerminatingErr
				},
			}).
			Build()

		s := &k8s.KubernetesStore{
			Client:    fakeClient,
			Converter: k8s.NewSimpleConverter(),
			Scheme:    k8sClientScheme,
		}

		err := s.Create(context.Background(), meshexternalservice_api.NewMeshExternalServiceResource(), store.CreateByKey("demo-service.demo", "default"))

		Expect(store.IsNamespaceTerminating(err)).To(BeTrue())
		Expect(store.IsAlreadyExists(err)).To(BeFalse())
	})
})
