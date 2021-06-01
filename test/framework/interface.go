package framework

import (
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/testing"
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

type deployOptions struct {
	// cp specific
	globalAddress    string
	installationMode InstallationMode
	skipDefaultMesh  bool
	helmReleaseName  string
	helmChartPath    *string
	helmChartVersion string
	helmOpts         map[string]string
	noHelmOpts       []string
	ctlOpts          map[string]string
	env              map[string]string
	ingress          bool
	cni              bool
	cpReplicas       int
	proxyOnly        bool
	hdsDisabled      bool
	serviceProbe     bool
	isipv6           bool
	storeType        string

	// app specific
	namespace      string
	appname        string
	name           string
	appYaml        string
	appArgs        []string
	token          string
	transparent    bool
	builtindns     bool
	protocol       string
	serviceName    string
	serviceVersion string
	mesh           string
	dpVersion      string
}

type DeployOptionsFunc func(*deployOptions)

func WithYaml(appYaml string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.appYaml = appYaml
	}
}

func WithIPv6(isipv6 bool) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.isipv6 = isipv6
	}
}

func WithProtocol(protocol string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.protocol = protocol
	}
}

func WithArgs(appArgs []string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.appArgs = appArgs
	}
}

func WithServiceName(name string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.serviceName = name
	}
}

func WithServiceVersion(version string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.serviceVersion = version
	}
}

func ProxyOnly() DeployOptionsFunc {
	return func(o *deployOptions) {
		o.proxyOnly = true
	}
}

func ServiceProbe() DeployOptionsFunc {
	return func(o *deployOptions) {
		o.serviceProbe = true
	}
}

func WithHDS(enabled bool) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.hdsDisabled = !enabled
	}
}

func WithGlobalAddress(address string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.globalAddress = address
	}
}

// WithDPVersion only works with Universal now
func WithDPVersion(version string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.dpVersion = version
	}
}

// WithCPReplicas works only with HELM now
func WithCPReplicas(cpReplicas int) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.cpReplicas = cpReplicas
	}
}

// WithSkipDefaultMesh works only with HELM now
func WithSkipDefaultMesh(skip bool) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.skipDefaultMesh = skip
	}
}

func WithInstallationMode(mode InstallationMode) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.installationMode = mode
	}
}

func WithHelmReleaseName(name string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.helmReleaseName = name
	}
}

func WithHelmChartPath(path string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.helmChartPath = &path
	}
}

func WithHelmChartVersion(version string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.helmChartVersion = version
	}
}

func WithHelmOpt(name, value string) DeployOptionsFunc {
	return func(o *deployOptions) {
		if o.helmOpts == nil {
			o.helmOpts = map[string]string{}
		}
		o.helmOpts[name] = value
	}
}

func WithoutHelmOpt(name string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.noHelmOpts = append(o.noHelmOpts, name)
	}
}

func WithEnv(name, value string) DeployOptionsFunc {
	return func(o *deployOptions) {
		if o.env == nil {
			o.env = map[string]string{}
		}
		o.env[name] = value
	}
}

func WithIngress() DeployOptionsFunc {
	return func(o *deployOptions) {
		o.ingress = true
	}
}

func WithCNI() DeployOptionsFunc {
	return func(o *deployOptions) {
		o.cni = true
	}
}

func WithNamespace(namespace string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.namespace = namespace
	}
}

func WithMesh(mesh string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.mesh = mesh
	}
}

func WithAppname(appname string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.appname = appname
	}
}

func WithName(name string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.name = name
	}
}

func WithToken(token string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.token = token
	}
}

func WithTransparentProxy(transparent bool) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.transparent = transparent
	}
}

func WithBuiltinDNS(builtindns bool) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.builtindns = builtindns
	}
}

func WithCtlOpt(name, value string) DeployOptionsFunc {
	return func(o *deployOptions) {
		if o.ctlOpts == nil {
			o.ctlOpts = map[string]string{}
		}
		o.ctlOpts[name] = value
	}
}

type DeployPostgresOptionFunc func(map[string]string)

func WithPostgresHost(host string) DeployPostgresOptionFunc {
	return func(env map[string]string) {
		env[EnvStorePostgresHost] = host
	}
}

func WithPostgresPort(port string) DeployPostgresOptionFunc {
	return func(env map[string]string) {
		env[EnvStorePostgresPort] = port
	}
}

func WithPostgresUser(user string) DeployPostgresOptionFunc {
	return func(env map[string]string) {
		env[EnvStorePostgresUser] = user
	}
}

func WithPostgresPassword(password string) DeployPostgresOptionFunc {
	return func(env map[string]string) {
		env[EnvStorePostgresPassword] = password
	}
}

func WithPostgresDBName(dbName string) DeployPostgresOptionFunc {
	return func(env map[string]string) {
		env[EnvStorePostgresDBName] = dbName
	}
}

func WithPostgresTLSMode(mode string) DeployPostgresOptionFunc {
	return func(env map[string]string) {
		env[EnvStorePostgresTLSMode] = mode
	}
}

func WithPostgresTLSCertPath(path string) DeployPostgresOptionFunc {
	return func(env map[string]string) {
		env[EnvStorePostgresTLSCertPath] = path
	}
}

func WithPostgresTLSKeyPath(path string) DeployPostgresOptionFunc {
	return func(env map[string]string) {
		env[EnvStorePostgresTLSKeyPath] = path
	}
}

func WithPostgresTLSCAPath(path string) DeployPostgresOptionFunc {
	return func(env map[string]string) {
		env[EnvStorePostgresTLSCAPath] = path
	}
}

func WithPostgres(fs ...DeployPostgresOptionFunc) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.storeType = StoreTypePostgres

		if o.env == nil {
			o.env = map[string]string{}
		}

		o.env[EnvStoreType] = StoreTypePostgres
		o.env[EnvStorePostgresHost] = DefaultPostgresHost
		o.env[EnvStorePostgresPort] = DefaultPostgresPort
		o.env[EnvStorePostgresUser] = DefaultPostgresUser
		o.env[EnvStorePostgresPassword] = DefaultPostgresPassword
		o.env[EnvStorePostgresDBName] = DefaultPostgresDBName
		o.env[EnvStorePostgresTLSMode] = DefaultPostgresTLSMode
		o.env[EnvStorePostgresTLSCertPath] = DefaultPostgresTLSCertPath
		o.env[EnvStorePostgresTLSKeyPath] = DefaultPostgresTLSKeyPath
		o.env[EnvStorePostgresTLSCAPath] = DefaultPostgresTLSCAPath

		for _, f := range fs {
			f(o.env)
		}
	}
}

func newDeployOpt(fs ...DeployOptionsFunc) *deployOptions {
	rv := &deployOptions{
		installationMode: KumactlInstallationMode,
	}
	for _, f := range fs {
		f(rv)
	}
	return rv
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
	DeployKuma(mode string, opts ...DeployOptionsFunc) error
	GetKuma() ControlPlane
	VerifyKuma() error
	DeleteKuma(opts ...DeployOptionsFunc) error
	InjectDNS(namespace ...string) error
	GetKumactlOptions() *KumactlOptions
	Deployment(name string) Deployment
	Deploy(deployment Deployment) error
	WithTimeout(timeout time.Duration) Cluster
	WithRetries(retries int) Cluster

	// K8s
	GetKubectlOptions(namespace ...string) *k8s.KubectlOptions
	CreateNamespace(namespace string) error
	DeleteNamespace(namespace string) error
	DeployApp(fs ...DeployOptionsFunc) error
	DeleteApp(namespace, appname string) error
	Exec(namespace, podName, containerName string, cmd ...string) (string, string, error)
	ExecWithRetries(namespace, podName, containerName string, cmd ...string) (string, string, error)

	// Testing
	GetTesting() testing.TestingT
}

type ControlPlane interface {
	GetName() string
	GetKumaCPLogs() (string, error)
	GetKDSServerAddress() string
	GetGlobaStatusAPI() string
	GenerateDpToken(mesh, appname string) (string, error)
}
