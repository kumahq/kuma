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
	DeployKuma(mode ...string) error
	VerifyKuma() error
	DeleteKuma() error
	GetKumaCPLogs() (string, error)
	InjectDNS() error
	GetKumactlOptions() *KumactlOptions

	// K8s
	GetKubectlOptions(namespace ...string) *k8s.KubectlOptions
	CreateNamespace(namespace string) error
	DeleteNamespace(namespace string) error
	LabelNamespaceForSidecarInjection(namespace string) error
	DeployApp(namespace, appname string) error
	DeleteApp(namespace, appname string) error
	ExecCommandInContainerWithFullOutput(namespace, podName, containerName string, cmd ...string) (string, string, error)

	// Testing
	GetTesting() testing.TestingT
}
