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
	globalAddress    string
	installationMode InstallationMode
	helmReleaseName  string
	helmOpts         map[string]string
	ctlOpts          map[string]string
}

type DeployOptionsFunc func(*deployOptions)

func WithGlobalAddress(address string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.globalAddress = address
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
	DismissCluster() error
	// Generic
	DeployKuma(mode string, opts ...DeployOptionsFunc) error
	GetKuma() ControlPlane
	VerifyKuma() error
	RestartKuma() error
	DeleteKuma(opts ...DeployOptionsFunc) error
	InjectDNS() error
	GetKumactlOptions() *KumactlOptions
	Deployment(name string) Deployment
	Deploy(deployment Deployment) error

	// K8s
	GetKubectlOptions(namespace ...string) *k8s.KubectlOptions
	CreateNamespace(namespace string) error
	DeleteNamespace(namespace string) error
	DeployApp(namespace, appname string) error
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
	GetIngressAddress() string
	GetGlobaStatusAPI() string
}
