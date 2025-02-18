package webhooks_test

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	bootstrap_scheme "github.com/kumahq/kuma/pkg/plugins/bootstrap/k8s/scheme"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test"
)

func TestWebhook(t *testing.T) {
	test.RunSpecs(t, "Webhooks Suite")
}

var (
	decoder     kube_admission.Decoder
	testEnv     *envtest.Environment
	k8sClient   client.Client
	scheme      *kube_runtime.Scheme
	defaultMesh = &mesh_k8s.Mesh{
		ObjectMeta: kube_meta.ObjectMeta{
			Name: "default",
		},
	}
	dp1 = &mesh_k8s.Dataplane{
		ObjectMeta: kube_meta.ObjectMeta{
			Name:      "dp-1",
			Namespace: "default",
		},
		Mesh: "default",
	}
)

var _ = BeforeSuite(func() {
	// setup K8S with Kuma CRDs
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "..", "..", test.CustomResourceDir),
		},
	}
	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	scheme, err = bootstrap_scheme.NewScheme()
	Expect(err).ToNot(HaveOccurred())
	Expect(mesh_k8s.AddToScheme(scheme)).To(Succeed())

	decoder = kube_admission.NewDecoder(scheme)

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	// create default mesh
	err = k8sClient.Create(context.Background(), defaultMesh)
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient.Create(context.Background(), dp1)).To(Succeed())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
