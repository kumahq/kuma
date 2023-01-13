package framework

import (
	"fmt"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/test/framework/envoy_admin"
)

type K8sClusters struct {
	t        testing.TestingT
	clusters map[string]*K8sCluster
	verbose  bool
}

var _ Clusters = &K8sClusters{}

func NewK8sClusters(clusterNames []string, verbose bool) (Clusters, error) {
	if len(clusterNames) < 1 || len(clusterNames) > maxClusters {
		return nil, errors.Errorf("Invalid cluster number. Should be in the range [1,3], but it is %d", len(clusterNames))
	}

	t := NewTestingT()

	clusters := map[string]*K8sCluster{}

	for _, name := range clusterNames {
		clusters[name] = NewK8sCluster(t, name, verbose)
	}

	return &K8sClusters{
		t:        t,
		clusters: clusters,
		verbose:  verbose,
	}, nil
}

func (cs *K8sClusters) WithTimeout(timeout time.Duration) Cluster {
	for _, c := range cs.clusters {
		c.WithTimeout(timeout)
	}

	return cs
}

func (cs *K8sClusters) Verbose() bool {
	return cs.verbose
}

func (c *K8sClusters) Install(fn InstallFunc) error {
	panic("not implemented")
}

func (cs *K8sClusters) WithRetries(retries int) Cluster {
	for _, c := range cs.clusters {
		c.WithRetries(retries)
	}

	return cs
}

func (cs *K8sClusters) Name() string {
	panic("not implemented")
}

func (cs *K8sClusters) GetKumaCPLogs() (string, error) {
	panic("not implemented")
}

func (cs *K8sClusters) DismissCluster() error {
	for name, c := range cs.clusters {
		if err := c.DismissCluster(); err != nil {
			return errors.Wrapf(err, "Deploy Kuma on %s failed: %v", name, err)
		}
	}

	return nil
}

func (cs *K8sClusters) Exec(namespace, podName, containerName string, cmd ...string) (string, string, error) {
	panic("not supported")
}

func (cs *K8sClusters) GetCluster(name string) Cluster {
	c, found := cs.clusters[name]
	if !found {
		return nil
	}

	return c
}

func (cs *K8sClusters) DeployKuma(mode core.CpMode, opt ...KumaDeploymentOption) error {
	for name, c := range cs.clusters {
		if err := c.DeployKuma(mode, opt...); err != nil {
			return errors.Wrapf(err, "Deploy Kuma on %s failed: %v", name, err)
		}
	}

	return nil
}

func (cs *K8sClusters) UpgradeKuma(mode string, opt ...KumaDeploymentOption) error {
	for name, c := range cs.clusters {
		if err := c.UpgradeKuma(mode, opt...); err != nil {
			return errors.Wrapf(err, "Upgrade Kuma on %s failed: %v", name, err)
		}
	}

	return nil
}

func (cs *K8sClusters) GetKuma() ControlPlane {
	panic("Not supported at this level.")
}

func (cs *K8sClusters) VerifyKuma() error {
	for name, c := range cs.clusters {
		if err := c.VerifyKuma(); err != nil {
			return errors.Wrapf(err, "Verify Kuma on %s failed: %v", name, err)
		}
	}

	return nil
}

func (cs *K8sClusters) DeleteKuma() error {
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

func (cs *K8sClusters) GetKubectlOptions(namespace ...string) *k8s.KubectlOptions {
	panic("Not supported at this level.")
}

func (cs *K8sClusters) CreateNamespace(namespace string) error {
	for name, c := range cs.clusters {
		if err := c.CreateNamespace(namespace); err != nil {
			return errors.Wrapf(err, "Creating Namespace %s on %s failed: %v", namespace, name, err)
		}
	}

	return nil
}

func (cs *K8sClusters) DeleteNamespace(namespace string) error {
	for name, c := range cs.clusters {
		if err := c.DeleteNamespace(namespace); err != nil {
			return errors.Wrapf(err, "Delete Namespace %s on %s failed: %v", namespace, name, err)
		}
	}

	return nil
}

func (cs *K8sClusters) GetKumactlOptions() *KumactlOptions {
	fmt.Println("Not supported at this level.")
	return nil
}

func (cs *K8sClusters) DeployApp(opt ...AppDeploymentOption) error {
	for name, c := range cs.clusters {
		if err := c.DeployApp(opt...); err != nil {
			return errors.Wrapf(err, "unable to deploy on %s", name)
		}
	}

	return nil
}

func (cs *K8sClusters) DeleteApp(namespace, appname string) error {
	for name, c := range cs.clusters {
		if err := c.DeleteApp(namespace, appname); err != nil {
			return errors.Wrapf(err, "Labeling Namespace %s on %s failed: %v", namespace, name, err)
		}
	}

	return nil
}

func (cs *K8sClusters) GetTesting() testing.TestingT {
	return cs.t
}

func (cs *K8sClusters) Deployment(name string) Deployment {
	panic("not supported")
}

func (cs *K8sClusters) Deploy(deployment Deployment) error {
	for name, c := range cs.clusters {
		if err := c.Deploy(deployment); err != nil {
			return errors.Wrapf(err, "deployment %s failed on %s cluster", deployment.Name(), name)
		}
	}
	return nil
}

func (cs *K8sClusters) DeleteDeployment(deploymentName string) error {
	for name, c := range cs.clusters {
		if err := c.DeleteDeployment(deploymentName); err != nil {
			return errors.Wrapf(err, "delete deployment %s failed on %s cluster", deploymentName, name)
		}
	}
	return nil
}

func (cs *K8sClusters) GetZoneEgressEnvoyTunnel() envoy_admin.Tunnel {
	panic("not supported")
}

func (cs *K8sClusters) GetZoneIngressEnvoyTunnel() envoy_admin.Tunnel {
	panic("not supported")
}

func (cs *K8sClusters) CreateNode(name string, label string) error {
	allErrors := errors.New("combined create node errors")
	for _, cluster := range cs.clusters {
		allErrors = multierr.Append(allErrors, cluster.CreateNode(name, label))
	}
	return allErrors
}

func (cs *K8sClusters) DeleteNode(name string) error {
	allErrors := errors.New("combined delete node errors")
	for _, cluster := range cs.clusters {
		allErrors = multierr.Append(allErrors, cluster.DeleteNode(name))
	}
	return allErrors
}

func (cs *K8sClusters) LoadImages(names ...string) error {
	allErrors := errors.New("combined load images to node errors")
	for _, cluster := range cs.clusters {
		allErrors = multierr.Append(allErrors, cluster.LoadImages(names...))
	}
	return allErrors
}
