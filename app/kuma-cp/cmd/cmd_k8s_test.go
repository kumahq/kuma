package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/testing_frameworks/integration/addr"
)

// Disabling this one as there are potential issues due to https://github.com/kumahq/kuma/issues/1001
var _ = XDescribe("K8S CMD test", func() {
	var k8sClient client.Client
	var testEnv *envtest.Environment

	BeforeEach(func(done Done) {
		By("bootstrapping test environment")
		testEnv = &envtest.Environment{
			CRDDirectoryPaths:        []string{filepath.Join("..", "..", "..", "pkg", "plugins", "resources", "k8s", "native", "config", "crd", "bases")},
			ControlPlaneStartTimeout: 60 * time.Second,
			ControlPlaneStopTimeout:  60 * time.Second,
		}

		cfg, err := testEnv.Start()
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg).ToNot(BeNil())

		// +kubebuilder:scaffold:scheme

		k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
		Expect(err).ToNot(HaveOccurred())
		Expect(k8sClient).ToNot(BeNil())
		Expect(k8sClient.Create(context.Background(), &kube_core.Namespace{ObjectMeta: kube_meta.ObjectMeta{Name: "kuma-system"}})).To(Succeed())

		ctrl.GetConfigOrDie = func() *rest.Config {
			return testEnv.Config
		}

		close(done)
	}, 60)

	AfterEach(func() {
		By("tearing down the test environment")
		Expect(testEnv.Stop()).To(Succeed())
	}, 60)

	RunSmokeTest(ConfigFactoryFunc(func() string {
		admissionServerPort, _, err := addr.Suggest()
		Expect(err).NotTo(HaveOccurred())

		return fmt.Sprintf(`
apiServer:
  http:
    port: 0
  https:
    port: 0
environment: kubernetes
store:
  type: kubernetes
runtime:
  kubernetes:
    admissionServer:
      port: %d
      certDir: %s
guiServer:
  port: 0
dnsServer:
  port: 0
diagnostics:
  serverPort: %%d
`,
			admissionServerPort,
			filepath.Join("testdata"))
	}), "")
})
