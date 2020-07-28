package framework

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/testing"
)

type Clusters interface {
	GetCluster(name string) Cluster
	Cluster
}

type deployOptions struct {
	globalAddress string
}

type DeployOptionsFunc func(*deployOptions)

func WithGlobalAddress(address string) DeployOptionsFunc {
	return func(o *deployOptions) {
		o.globalAddress = address
	}
}

func newDeployOpt(fs ...DeployOptionsFunc) *deployOptions {
	rv := &deployOptions{}
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
	DeleteKuma() error
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
	GetKumaCPLogs() (string, error)
	GetKDSServerAddress() string
	GetIngressAddress() string
	GetGlobaStatusAPI() string
}
