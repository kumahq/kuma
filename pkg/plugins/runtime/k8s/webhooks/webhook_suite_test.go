package webhooks_test

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	kube_admission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

func TestWebhook(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webhooks Suite")
}

var decoder *kube_admission.Decoder
var testEnv *envtest.Environment
var k8sClient client.Client
var scheme *kube_runtime.Scheme
var defaultMesh *mesh_k8s.Mesh

var _ = BeforeSuite(func() {
	// setup K8S with Kuma CRDs
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "resources", "k8s", "native", "config", "crd", "bases"),
		},
	}
	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	scheme = kube_runtime.NewScheme()
	Expect(kube_core.AddToScheme(scheme)).To(Succeed())
	Expect(mesh_k8s.AddToScheme(scheme)).To(Succeed())

	decoder, err = kube_admission.NewDecoder(scheme)
	Expect(err).ToNot(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	// create default mesh
	defaultMesh = &mesh_k8s.Mesh{
		ObjectMeta: kube_meta.ObjectMeta{
			Name: "default",
		},
	}
	err = k8sClient.Create(context.Background(), defaultMesh)
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
