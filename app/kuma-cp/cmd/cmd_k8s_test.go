package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/testing_frameworks/integration/addr"

	"github.com/kumahq/kuma/pkg/test"
)

// Disabling this one as there are potential issues due to https://github.com/kumahq/kuma/issues/1001
var _ = XDescribe("K8S CMD test", func() {
	var k8sClient client.Client
	var testEnv *envtest.Environment

	BeforeEach(test.Within(time.Minute, func() {
		By("bootstrapping test environment")
		testEnv = &envtest.Environment{
			CRDDirectoryPaths:        []string{filepath.Join("..", "..", "..", test.CustomResourceDir)},
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
	}))

	AfterEach(func() {
		By("tearing down the test environment")
		Eventually(testEnv.Stop, 60*time.Second).Should(Succeed())
	})

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
			"testdata")
	}), "")
})
