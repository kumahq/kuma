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

func newDeployOpt(fs ...DeployOptionsFunc) *deployOptions {
	rv := &deployOptions{
		installationMode: KumactlInstallationMode,
	}
	for _, f := range fs {
		f(rv)
	}
	return rv
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
	SetLbAddress(name, lbAddress string) error
	GetKumaCPLogs() (string, error)
	GetKDSServerAddress() string
	GetIngressAddress() string
	GetGlobaStatusAPI() string
}
