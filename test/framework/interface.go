package framework

import (
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
	helmOpts         map[string]string
	ctlOpts          map[string]string
	env              map[string]string
	ingress          bool
	cni              bool
	cpReplicas       int

	// app specific
	namespace   string
	appname     string
	id          string
	token       string
	transparent bool
	mesh        string
}

type DeployOptionsFunc func(*deployOptions)

func WithGlobalAddress(address string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.globalAddress = address
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

func WithHelmOpt(name, value string) DeployOptionsFunc {
	return func(o *deployOptions) {
		if o.helmOpts == nil {
			o.helmOpts = map[string]string{}
		}
		o.helmOpts[name] = value
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

func WithId(id string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.id = id
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

func WithCtlOpt(name, value string) DeployOptionsFunc {
	return func(o *deployOptions) {
		if o.ctlOpts == nil {
			o.ctlOpts = map[string]string{}
		}
		o.ctlOpts[name] = value
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
