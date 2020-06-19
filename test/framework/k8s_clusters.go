package framework

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/testing"
)

type K8sClusters struct {
	t        testing.TestingT
	clusters map[string]*K8sCluster
	verbose  bool
}

func NewK8sClusters(clusterNames []string, verbose bool) (Clusters, error) {
	if len(clusterNames) < 1 || len(clusterNames) > maxClusters {
		return nil, errors.Errorf("Invalid cluster number. Should be in the range [1,3], but it is %d", len(clusterNames))
	}

	t := NewTestingT()

	clusters := map[string]*K8sCluster{}

	for i, name := range clusterNames {
		clusters[name] = &K8sCluster{
			t:                   t,
			name:                name,
			kubeconfig:          os.ExpandEnv(fmt.Sprintf(defaultKubeConfigPathPattern, name)),
			loPort:              uint32(kumaCPAPIPortFwdBase + i*1000),
			hiPort:              uint32(kumaCPAPIPortFwdBase + (i+1)*1000 - 1),
			forwardedPortsChans: map[uint32]chan struct{}{},
			verbose:             verbose,
		}

		var err error
		clusters[name].clientset, err = k8s.GetKubernetesClientFromOptionsE(t, clusters[name].GetKubectlOptions())
		if err != nil {
			return nil, errors.Wrapf(err, "error in getting access to K8S")
		}
	}

	return &K8sClusters{
		t:        t,
		clusters: clusters,
		verbose:  verbose,
	}, nil
}

func (cs *K8sClusters) Exec(namespace, podName, containerName string, cmd ...string) (string, string, error) {
	panic("not supported")
}

func (cs *K8sClusters) ExecWithRetries(namespace, podName, containerName string, cmd ...string) (string, string, error) {
	panic("not supported")
}

func (cs *K8sClusters) GetCluster(name string) Cluster {
	c, found := cs.clusters[name]
	if !found {
		return nil
	}

	return c
}

func (cs *K8sClusters) DeployKuma(mode ...string) (ControlPlane, error) {
	for name, c := range cs.clusters {
		if _, err := c.DeployKuma(mode...); err != nil {
			return nil, errors.Wrapf(err, "Deploy Kuma on %s failed: %v", name, err)
		}
	}

	return nil, nil
}

func (cs *K8sClusters) RestartKuma() error {
	for name, c := range cs.clusters {
		if err := c.RestartKuma(); err != nil {
			return errors.Wrapf(err, "Restart Kuma on %s failed: %v", name, err)
		}
	}

	return nil
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
			return errors.Wrapf(err, "Creating Namespace %s on %s failed: %v", namespace, name, err)
		}
	}

	return nil
}

func (c *K8sClusters) GetKumactlOptions() *KumactlOptions {
	fmt.Println("Not supported at this level.")
	return nil
}

func (cs *K8sClusters) LabelNamespaceForSidecarInjection(namespace string) error {
	for name, c := range cs.clusters {
		if err := c.LabelNamespaceForSidecarInjection(namespace); err != nil {
			return errors.Wrapf(err, "Labeling Namespace %s on %s failed: %v", namespace, name, err)
		}
	}

	return nil
}

func (cs *K8sClusters) DeployApp(namespace, appname string) error {
	for name, c := range cs.clusters {
		if err := c.DeployApp(namespace, appname); err != nil {
			return errors.Wrapf(err, "Labeling Namespace %s on %s failed: %v", namespace, name, err)
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
func (cs *K8sClusters) InjectDNS() error {
	for name, c := range cs.clusters {
		if err := c.InjectDNS(); err != nil {
			return errors.Wrapf(err, "Injecting DNS on %s failed: %v", name, err)
		}
	}

	return nil
}

func (cs *K8sClusters) GetTesting() testing.TestingT {
	return cs.t
}

func IsK8sClustersStarted() bool {
	_, found := os.LookupEnv(envK8SCLUSTERS)
	return found
}
