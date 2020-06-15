package cmd

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/testing_frameworks/integration/addr"
)

var _ = Describe("K8S CMD test", func() {
	var k8sClient client.Client
	var testEnv *envtest.Environment

	BeforeEach(func(done Done) {
		By("bootstrapping test environment")
		testEnv = &envtest.Environment{
			CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "pkg", "plugins", "resources", "k8s", "native", "config", "crd", "bases")},
		}

		cfg, err := testEnv.Start()
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg).ToNot(BeNil())

		// +kubebuilder:scaffold:scheme

		k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
		Expect(err).ToNot(HaveOccurred())
		Expect(k8sClient).ToNot(BeNil())

		ctrl.GetConfigOrDie = func() *rest.Config {
			return testEnv.Config
		}

		close(done)
	}, 60)

	AfterEach(func() {
		By("tearing down the test environment")
		err := testEnv.Stop()
		Expect(err).ToNot(HaveOccurred())
	})

	RunSmokeTest(ConfigFactoryFunc(func() string {
		admissionServerPort, _, err := addr.Suggest()
		Expect(err).NotTo(HaveOccurred())

		return fmt.Sprintf(`
xdsServer:
  grpcPort: 0
  diagnosticsPort: %%d
bootstrapServer:
  port: 0
apiServer:
  port: 0
sdsServer:
  grpcPort: 0
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
`,
			admissionServerPort,
			filepath.Join("testdata"))
	}))
})
