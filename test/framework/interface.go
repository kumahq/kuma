package framework

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/testing"
)

type Clusters interface {
	GetCluster(name string) Cluster
	Cluster
}

type Cluster interface {
	DeployKuma(mode ...string) (ControlPlane, error)
	VerifyKuma() error
	RestartKuma() error
	DeleteKuma() error
	InjectDNS() error

	// K8s
	GetKubectlOptions(namespace ...string) *k8s.KubectlOptions
	CreateNamespace(namespace string) error
	DeleteNamespace(namespace string) error
	LabelNamespaceForSidecarInjection(namespace string) error
	DeployApp(namespace, appname string) error
	DeleteApp(namespace, appname string) error

	// Testing
	GetTesting() testing.TestingT
}

type ControlPlane interface {
	GetName() string
	AddLocalCP(name, url string) error
	GetKumaCPLogs() (string, error)
	GetHostAPI() string
	GetGlobaStatusAPI() string
}
