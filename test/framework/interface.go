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
	GetKumaCPLogs() (string, error)
	DeleteKuma() error

	// K8s
	GetKubectlOptions(namespace ...string) *k8s.KubectlOptions
	LabelNamespaceForSidecarInjection(namespace string) error

	// Testing
	GetTesting() testing.TestingT
}
