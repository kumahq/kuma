package framework

import (
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/testing"

	"github.com/kumahq/kuma/pkg/config/core"
)

type Clusters interface {
	GetCluster(name string) Cluster
	Cluster
}

type InstallationMode string

var (
	HelmInstallationMode    = InstallationMode("helm")
	KumactlInstallationMode = InstallationMode("kumactl")
)

// kumaDeploymentOptions are the options used for installing the Kuma
// control plane.
type kumaDeploymentOptions struct {
	// common
	isipv6  bool
	verbose *bool

	// cp specific
	globalAddress        string
	installationMode     InstallationMode
	skipDefaultMesh      bool
	helmReleaseName      string
	helmChartPath        *string
	helmChartVersion     string
	helmOpts             map[string]string
	noHelmOpts           []string
	env                  map[string]string
	ingress              bool
	cni                  bool
	cpReplicas           int
	hdsDisabled          bool
	runPostgresMigration bool
}

func (k *kumaDeploymentOptions) apply(opts ...KumaDeploymentOption) {
	// Set defaults.
	k.installationMode = KumactlInstallationMode
	k.env = map[string]string{}

	// Apply options.
	for _, o := range opts {
		o.ApplyKuma(k)
	}
}

type KumaDeploymentOption interface {
	ApplyKuma(options *kumaDeploymentOptions)
}

type KumaOptionFunc func(options *kumaDeploymentOptions)

func (f KumaOptionFunc) ApplyKuma(opts *kumaDeploymentOptions) {
	if f != nil {
		f(opts)
	}
}

type appDeploymentOptions struct {
	// common
	isipv6  bool
	verbose *bool

	// app specific
	namespace       string
	appname         string
	name            string
	appYaml         string
	appArgs         []string
	token           string
	transparent     bool
	builtindns      *bool // true by default
	protocol        string
	serviceName     string
	serviceVersion  string
	serviceInstance string
	mesh            string
	dpVersion       string
	kumactlFlow     bool
	concurrency     int
	omitDataplane   bool
	proxyOnly       bool
	serviceProbe    bool
}

func (d *appDeploymentOptions) apply(opts ...AppDeploymentOption) {
	// Set defaults.

	// Apply options.
	for _, o := range opts {
		o.ApplyApp(d)
	}
}

type AppDeploymentOption interface {
	ApplyApp(*appDeploymentOptions)
}

type AppOptionFunc func(*appDeploymentOptions)

func (f AppOptionFunc) ApplyApp(opts *appDeploymentOptions) {
	if f != nil {
		f(opts)
	}
}

type GenericOption struct {
	Kuma KumaOptionFunc
	App  AppOptionFunc
}

var _ KumaDeploymentOption = &GenericOption{}
var _ AppDeploymentOption = &GenericOption{}

func (o *GenericOption) ApplyApp(opts *appDeploymentOptions) {
	o.App(opts)
}

func (o *GenericOption) ApplyKuma(opts *kumaDeploymentOptions) {
	o.Kuma(opts)
}

// WithVerbose enables verbose mode on the deployment.
func WithVerbose() *GenericOption {
	enabled := true

	return &GenericOption{
		App: func(o *appDeploymentOptions) {
			o.verbose = &enabled
		},
		Kuma: func(o *kumaDeploymentOptions) {
			o.verbose = &enabled
		},
	}
}

func WithIPv6(enabled bool) *GenericOption {
	return &GenericOption{
		Kuma: func(o *kumaDeploymentOptions) {
			o.isipv6 = enabled
		},
		App: func(o *appDeploymentOptions) {
			o.isipv6 = enabled
		},
	}
}

func WithPostgres(envVars map[string]string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *kumaDeploymentOptions) {
		o.runPostgresMigration = true

		if o.env == nil {
			o.env = map[string]string{}
		}

		for key, value := range envVars {
			o.env[key] = value
		}
	})
}

func WithHDS(enabled bool) KumaDeploymentOption {
	return KumaOptionFunc(func(o *kumaDeploymentOptions) {
		o.hdsDisabled = !enabled
	})
}

// WithCPReplicas works only with HELM for now.
func WithCPReplicas(cpReplicas int) KumaDeploymentOption {
	return KumaOptionFunc(func(o *kumaDeploymentOptions) {
		o.cpReplicas = cpReplicas
	})
}

// WithSkipDefaultMesh works only with HELM now.
func WithSkipDefaultMesh(skip bool) KumaDeploymentOption {
	return KumaOptionFunc(func(o *kumaDeploymentOptions) {
		o.skipDefaultMesh = skip
	})
}

func WithInstallationMode(mode InstallationMode) KumaDeploymentOption {
	return KumaOptionFunc(func(o *kumaDeploymentOptions) {
		o.installationMode = mode
	})
}

func WithHelmReleaseName(name string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *kumaDeploymentOptions) {
		o.helmReleaseName = name
	})
}

func WithHelmChartPath(path string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *kumaDeploymentOptions) {
		o.helmChartPath = &path
	})
}

func WithHelmChartVersion(version string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *kumaDeploymentOptions) {
		o.helmChartVersion = version
	})
}

func WithHelmOpt(name, value string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *kumaDeploymentOptions) {
		if o.helmOpts == nil {
			o.helmOpts = map[string]string{}
		}
		o.helmOpts[name] = value
	})
}

func WithoutHelmOpt(name string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *kumaDeploymentOptions) {
		o.noHelmOpts = append(o.noHelmOpts, name)
	})
}

func WithEnv(name, value string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *kumaDeploymentOptions) {
		o.env[name] = value
	})
}

func WithIngress() KumaDeploymentOption {
	return KumaOptionFunc(func(o *kumaDeploymentOptions) {
		o.ingress = true
	})
}

func WithCNI() KumaDeploymentOption {
	return KumaOptionFunc(func(o *kumaDeploymentOptions) {
		o.cni = true
	})
}

func WithGlobalAddress(address string) KumaDeploymentOption {
	return KumaOptionFunc(func(o *kumaDeploymentOptions) {
		o.globalAddress = address
	})
}

// WithoutDataplane suppresses the automatic configuration of kuma-dp
// in the application container. This is useful when the test requires a
// container that is not bound to the mesh.
func WithoutDataplane() AppDeploymentOption {
	return AppOptionFunc(func(o *appDeploymentOptions) {
		o.omitDataplane = true
	})
}

// WithKumactlFlow creates the Dataplane resource in the control plane before the actual data plane proxy start.
// If proxy is disconnected, resource won't be deleted from the storage and will be displayed as Offline.
func WithKumactlFlow() AppDeploymentOption {
	return AppOptionFunc(func(options *appDeploymentOptions) {
		options.kumactlFlow = true
	})
}

func WithYaml(appYaml string) AppDeploymentOption {
	return AppOptionFunc(func(o *appDeploymentOptions) {
		o.appYaml = appYaml
	})
}

func WithProtocol(protocol string) AppDeploymentOption {
	return AppOptionFunc(func(o *appDeploymentOptions) {
		o.protocol = protocol
	})
}

func WithArgs(appArgs []string) AppDeploymentOption {
	return AppOptionFunc(func(o *appDeploymentOptions) {
		o.appArgs = appArgs
	})
}

func WithServiceName(name string) AppDeploymentOption {
	return AppOptionFunc(func(o *appDeploymentOptions) {
		o.serviceName = name
	})
}

func WithServiceVersion(version string) AppDeploymentOption {
	return AppOptionFunc(func(o *appDeploymentOptions) {
		o.serviceVersion = version
	})
}

func WithServiceInstance(instance string) AppDeploymentOption {
	return AppOptionFunc(func(o *appDeploymentOptions) {
		o.serviceInstance = instance
	})
}

func ProxyOnly() AppDeploymentOption {
	return AppOptionFunc(func(o *appDeploymentOptions) {
		o.proxyOnly = true
	})
}

func ServiceProbe() AppDeploymentOption {
	return AppOptionFunc(func(o *appDeploymentOptions) {
		o.serviceProbe = true
	})
}

// WithDPVersion only works with Universal now
func WithDPVersion(version string) AppDeploymentOption {
	return AppOptionFunc(func(o *appDeploymentOptions) {
		o.dpVersion = version
	})
}

func WithNamespace(namespace string) AppDeploymentOption {
	return AppOptionFunc(func(o *appDeploymentOptions) {
		o.namespace = namespace
	})
}

func WithMesh(mesh string) AppDeploymentOption {
	return AppOptionFunc(func(o *appDeploymentOptions) {
		o.mesh = mesh
	})
}

func WithAppname(appname string) AppDeploymentOption {
	return AppOptionFunc(func(o *appDeploymentOptions) {
		o.appname = appname
	})
}

func WithName(name string) AppDeploymentOption {
	return AppOptionFunc(func(o *appDeploymentOptions) {
		o.name = name
	})
}

func WithToken(token string) AppDeploymentOption {
	return AppOptionFunc(func(o *appDeploymentOptions) {
		o.token = token
	})
}

func WithTransparentProxy(transparent bool) AppDeploymentOption {
	return AppOptionFunc(func(o *appDeploymentOptions) {
		o.transparent = transparent
	})
}

func WithBuiltinDNS(builtindns bool) AppDeploymentOption {
	return AppOptionFunc(func(o *appDeploymentOptions) {
		o.builtindns = &builtindns
	})
}

func WithConcurrency(concurrency int) AppDeploymentOption {
	return AppOptionFunc(func(o *appDeploymentOptions) {
		o.concurrency = concurrency
	})
}

type Deployment interface {
	Name() string
	Deploy(cluster Cluster) error
	Delete(cluster Cluster) error
}

type Cluster interface {
	// Cluster
	Name() string
	DismissCluster() error
	// Generic
	DeployKuma(mode core.CpMode, opts ...KumaDeploymentOption) error
	GetKuma() ControlPlane
	VerifyKuma() error
	DeleteKuma(opts ...KumaDeploymentOption) error
	InjectDNS(namespace ...string) error
	GetKumactlOptions() *KumactlOptions
	Deployment(name string) Deployment
	Deploy(deployment Deployment) error
	DeleteDeployment(name string) error
	WithTimeout(timeout time.Duration) Cluster
	WithRetries(retries int) Cluster
	Verbose() bool

	// K8s
	GetKubectlOptions(namespace ...string) *k8s.KubectlOptions
	CreateNamespace(namespace string) error
	DeleteNamespace(namespace string) error
	DeployApp(fs ...AppDeploymentOption) error
	DeleteApp(namespace, appname string) error
	Exec(namespace, podName, containerName string, cmd ...string) (string, string, error)
	ExecWithRetries(namespace, podName, containerName string, cmd ...string) (string, string, error)

	// Testing
	GetTesting() testing.TestingT
}

type ControlPlane interface {
	GetName() string
	GetKumaCPLogs() (string, error)
	GetMetrics() (string, error)
	GetKDSServerAddress() string
	GetGlobaStatusAPI() string
	GenerateDpToken(mesh, serviceName string) (string, error)
	GenerateZoneIngressToken(zone string) (string, error)
}
