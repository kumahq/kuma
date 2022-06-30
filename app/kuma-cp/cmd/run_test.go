package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
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

	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	"github.com/kumahq/kuma/pkg/config"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
	"github.com/kumahq/kuma/pkg/test"
	test_postgres "github.com/kumahq/kuma/pkg/test/store/postgres"
)

type state interface {
	setup() error
	cleanup() error
	workdir() string
	generateConfig() string
}

type k8sState struct {
	k8sClient client.Client
	testEnv   *envtest.Environment
}

var _ state = &k8sState{}

func (k *k8sState) setup() error {
	By("bootstrapping test environment")
	k.testEnv = &envtest.Environment{
		CRDDirectoryPaths:        []string{filepath.Join("..", "..", "..", test.CustomResourceDir)},
		ControlPlaneStartTimeout: 60 * time.Second,
		ControlPlaneStopTimeout:  60 * time.Second,
	}

	cfg, err := k.testEnv.Start()
	if err != nil {
		return err
	}

	// +kubebuilder:scaffold:scheme

	k.k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		return err
	}
	if err = k.k8sClient.Create(context.Background(), &kube_core.Namespace{ObjectMeta: kube_meta.ObjectMeta{Name: "kuma-system"}}); err != nil {
		return err
	}

	ctrl.GetConfigOrDie = func() *rest.Config {
		return k.testEnv.Config
	}
	return nil
}

func (k *k8sState) cleanup() error {
	By("tearing down the test environment")
	return k.testEnv.Stop()
}

func (k *k8sState) workdir() string {
	return ""
}

func (k *k8sState) generateConfig() string {
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
}

type pgState struct {
	c     test_postgres.PostgresContainer
	pgCfg *postgres.PostgresStoreConfig
}

var _ state = &pgState{}

func (p *pgState) setup() error {
	// setup migrate DB
	var err error
	p.c = test_postgres.PostgresContainer{}
	if err = p.c.Start(); err != nil {
		return err
	}
	if p.pgCfg, err = p.c.Config(); err != nil {
		return err
	}
	cfg := kuma_cp.DefaultConfig()
	cfg.Store.Type = store.PostgresStore
	cfg.Store.Postgres = p.pgCfg
	if err = config.Load("", &cfg); err != nil {
		return err
	}
	return migrate(cfg)
}

func (p *pgState) cleanup() error {
	return p.c.Stop()
}

func (p *pgState) workdir() string {
	return "./kuma-workdir"
}

func (p *pgState) generateConfig() string {
	return fmt.Sprintf(`
general:
  workDir: ./kuma-workdir
apiServer:
  http:
    port: 0
  https:
    port: 0
dnsServer:
  port: 0
environment: universal
store:
  type: postgres
  postgres:
    host: %s
    port: %d
diagnostics:
  serverPort: %%d
`, p.pgCfg.Host, p.pgCfg.Port)
}

type memState struct {
}

var _ state = &memState{}

func (m memState) setup() error {
	return nil
}

func (m memState) cleanup() error {
	return nil
}

func (m memState) workdir() string {
	return "./kuma-workdir"
}

func (m memState) generateConfig() string {
	return `
general:
  workDir: ./kuma-workdir
apiServer:
  http:
    port: 0
  https:
    port: 0
dnsServer:
  port: 0
environment: universal
store:
  type: memory
diagnostics:
  serverPort: %d
`
}

var _ = DescribeTable("should be possible to run `kuma-cp run with default mode`", func(s state) {
	Expect(s.setup()).To(Succeed())
	var errCh chan error
	var configFile *os.File

	var diagnosticsPort int
	var ctx context.Context
	var cancel func()
	var opts = kuma_cmd.RunCmdOpts{
		SetupSignalHandler: func() (context.Context, context.Context) {
			return ctx, ctx
		},
	}

	ctx, cancel = context.WithCancel(context.Background())
	errCh = make(chan error)

	freePort, _, err := addr.Suggest()
	Expect(err).NotTo(HaveOccurred())
	diagnosticsPort = freePort

	file, err := os.CreateTemp("", "*")
	Expect(err).ToNot(HaveOccurred())
	configFile = file
	defer func() {
		if configFile != nil {
			err := os.Remove(configFile.Name())
			Expect(err).ToNot(HaveOccurred())
		}
		if s.workdir() != "" {
			err := os.RemoveAll(s.workdir())
			Expect(err).ToNot(HaveOccurred())
		}
		Expect(s.cleanup()).To(Succeed())
	}()

	// given
	config := fmt.Sprintf(s.generateConfig(), diagnosticsPort)
	_, err = configFile.WriteString(config)
	Expect(err).ToNot(HaveOccurred())
	cmd := newRunCmdWithOpts(opts)
	cmd.SetArgs([]string{"--config-file=" + configFile.Name()})

	// when
	By("starting the Control Plane")
	go func() {
		defer close(errCh)
		errCh <- cmd.Execute()
	}()

	// then
	By("waiting for Control Plane to become healthy")
	Eventually(func() bool {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/healthy", diagnosticsPort))
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, "10s", "10ms").Should(BeTrue())

	// then
	By("waiting for Control Plane to become ready")
	Eventually(func() bool {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/ready", diagnosticsPort))
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, "10s", "10ms").Should(BeTrue())

	// when
	By("signaling Control Plane to stop")
	cancel()

	// then
	err = <-errCh
	Expect(err).ToNot(HaveOccurred())
},
	Entry("memory", &memState{}),
	Entry("postgres", &pgState{}),
	// Disabling this one as there are potential issues due to https://github.com/kumahq/kuma/issues/1001
	XEntry("k8s", &k8sState{}),
)
