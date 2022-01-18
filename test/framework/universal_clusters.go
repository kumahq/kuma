package framework

import (
	"fmt"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/core"
)

type UniversalClusters struct {
	t        testing.TestingT
	clusters map[string]*UniversalCluster
	verbose  bool
}

var _ Clusters = &UniversalClusters{}

func NewUniversalClusters(clusterNames []string, verbose bool) (Clusters, error) {
	if len(clusterNames) < 1 || len(clusterNames) > maxClusters {
		return nil, errors.Errorf("Invalid cluster number. Should be in the range [1,3], but it is %d", len(clusterNames))
	}

	t := NewTestingT()

	clusters := map[string]*UniversalCluster{}

	for _, name := range clusterNames {
		if name == Kuma4 {
			clusters[name] = NewUniversalCluster(t, name, true)
		} else {
			clusters[name] = NewUniversalCluster(t, name, false)
		}
	}

	return &UniversalClusters{
		t:        t,
		clusters: clusters,
		verbose:  verbose,
	}, nil
}

func (cs *UniversalClusters) WithTimeout(timeout time.Duration) Cluster {
	for _, c := range cs.clusters {
		c.WithTimeout(timeout)
	}

	return cs
}

func (cs *UniversalClusters) Verbose() bool {
	return cs.verbose
}

func (cs *UniversalClusters) WithRetries(retries int) Cluster {
	for _, c := range cs.clusters {
		c.WithRetries(retries)
	}

	return cs
}

func (cs *UniversalClusters) Name() string {
	panic("not implemented")
}

func (cs *UniversalClusters) DismissCluster() error {
	for name, c := range cs.clusters {
		if err := c.DismissCluster(); err != nil {
			return errors.Wrapf(err, "Deploy Kuma on %s failed: %v", name, err)
		}
	}

	return nil
}

func (cs *UniversalClusters) GetCluster(name string) Cluster {
	c, found := cs.clusters[name]
	if !found {
		return nil
	}

	return c
}

func (cs *UniversalClusters) DeployKuma(mode core.CpMode, opt ...KumaDeploymentOption) error {
	for name, c := range cs.clusters {
		if err := c.DeployKuma(mode, opt...); err != nil {
			return errors.Wrapf(err, "Deploy Kuma on %s failed: %v", name, err)
		}
	}

	return nil
}

func (cs *UniversalClusters) GetKuma() ControlPlane {
	panic("Not supported at this level.")
}

func (cs *UniversalClusters) VerifyKuma() error {
	for name, c := range cs.clusters {
		if err := c.VerifyKuma(); err != nil {
			return errors.Wrapf(err, "Verify Kuma on %s failed: %v", name, err)
		}
	}

	return nil
}

func (cs *UniversalClusters) DeleteKuma() error {
	failed := []string{}

	for name, c := range cs.clusters {
		if err := c.DeleteKuma(); err != nil {
			fmt.Printf("Delete Kuma on %s failed", name)
			failed = append(failed, name)
		}
	}

	if len(failed) > 0 {
		return errors.Errorf("Clusters failed to delete %v", failed)
	}

	return nil
}

func (cs *UniversalClusters) GetKubectlOptions(namespace ...string) *k8s.KubectlOptions {
	panic("Not supported at this level.")
}

func (cs *UniversalClusters) CreateNamespace(namespace string) error {
	for name, c := range cs.clusters {
		if err := c.CreateNamespace(namespace); err != nil {
			return errors.Wrapf(err, "Creating Namespace %s on %s failed: %v", namespace, name, err)
		}
	}

	return nil
}

func (cs *UniversalClusters) DeleteNamespace(namespace string) error {
	for name, c := range cs.clusters {
		if err := c.DeleteNamespace(namespace); err != nil {
			return errors.Wrapf(err, "Creating Namespace %s on %s failed: %v", namespace, name, err)
		}
	}

	return nil
}

func (c *UniversalClusters) GetKumactlOptions() *KumactlOptions {
	fmt.Println("Not supported at this level.")
	return nil
}

func (cs *UniversalClusters) DeployApp(opt ...AppDeploymentOption) error {
	for name, c := range cs.clusters {
		if err := c.DeployApp(opt...); err != nil {
			return errors.Wrapf(err, "unable to deploy on %s", name)
		}
	}

	return nil
}

func (cs *UniversalClusters) DeleteApp(namespace, appname string) error {
	for name, c := range cs.clusters {
		if err := c.DeleteApp(namespace, appname); err != nil {
			return errors.Wrapf(err, "Labeling Namespace %s on %s failed: %v", namespace, name, err)
		}
	}

	return nil
}

func (cs *UniversalClusters) InjectDNS(namespace ...string) error {
	for name, c := range cs.clusters {
		if err := c.InjectDNS(namespace...); err != nil {
			return errors.Wrapf(err, "Injecting DNS on %s failed: %v", name, err)
		}
	}

	return nil
}

func (cs *UniversalClusters) GetTesting() testing.TestingT {
	return cs.t
}
func (cs *UniversalClusters) Exec(namespace, podName, containerName string, cmd ...string) (string, string, error) {
	panic("implement me")
}

func (cs *UniversalClusters) ExecWithRetries(namespace, podName, containerName string, cmd ...string) (string, string, error) {
	panic("implement me")
}

func (cs *UniversalClusters) Deployment(name string) Deployment {
	panic("not supported")
}

func (cs *UniversalClusters) Deploy(deployment Deployment) error {
	for name, c := range cs.clusters {
		if err := c.Deploy(deployment); err != nil {
			return errors.Wrapf(err, "deployment %s failed on %s cluster", deployment.Name(), name)
		}
	}
	return nil
}

func (cs *UniversalClusters) DeleteDeployment(deploymentName string) error {
	for name, c := range cs.clusters {
		if err := c.DeleteDeployment(deploymentName); err != nil {
			return errors.Wrapf(err, "delete deployment %s failed on %s cluster", deploymentName, name)
		}
	}
	return nil
}
