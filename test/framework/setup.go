package framework

import (
	"context"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"
	"time"
)

type InstallFunc func(cluster Cluster) error

func Yaml(yaml string) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "install yaml resource", defaultRetries, defaultTimeout,
			func() (s string, err error) {
				return "", k8s.KubectlApplyFromStringE(cluster.GetTesting(), cluster.GetKubectlOptions(), yaml)
			})
		return err
	}
}

func YamlPath(path string) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "install yaml resource by path", defaultRetries, defaultTimeout,
			func() (s string, err error) {
				return "", k8s.KubectlApplyE(cluster.GetTesting(), cluster.GetKubectlOptions(), path)
			})
		return err
	}
}

func Kuma() InstallFunc {
	return func(cluster Cluster) error {
		return cluster.DeployKuma()
	}
}

func WaitService(namespace, service string) InstallFunc {
	return func(c Cluster) error {
		k8s.WaitUntilServiceAvailable(c.GetTesting(), c.GetKubectlOptions(namespace), service, 10, 3*time.Second)
		return nil
	}
}

func WaitNumPods(num int, app string) InstallFunc {
	return func(c Cluster) error {
		k8s.WaitUntilNumPodsCreated(c.GetTesting(), c.GetKubectlOptions(),
			kube_meta.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", app),
			}, num, defaultRetries, defaultTimeout)
		return nil
	}
}

func WaitPodsAvailable(namespace, app string) InstallFunc {
	return func(c Cluster) error {
		pods, err := k8s.ListPodsE(c.GetTesting(), c.GetKubectlOptions(namespace),
			kube_meta.ListOptions{LabelSelector: fmt.Sprintf("app=%s", app)})
		if err != nil {
			return err
		}
		for _, p := range pods {
			err := k8s.WaitUntilPodAvailableE(c.GetTesting(), c.GetKubectlOptions(namespace), p.GetName(), defaultRetries, defaultTimeout)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func HttpBin() InstallFunc {
	return Combine(
		YamlPath(filepath.Join("testdata", "httpbin.yaml")),
		WaitService("kuma-test", "httpbin"),
		WaitNumPods(1, "httpbin"),
		WaitPodsAvailable("kuma-test", "httpbin"),
	)
}

type IngressDesc struct {
	Port int32
	IP   string
}

func Ingress(ingress *IngressDesc) InstallFunc {
	return func(c Cluster) error {
		yaml, err := c.GetKumactlOptions().KumactlInstallIngress()
		if err != nil {
			return err
		}
		return Combine(
			Yaml(yaml),
			WaitService("kuma-system", "kuma-ingress"),
			WaitNumPods(1, "kuma-ingress"),
			WaitPodsAvailable("kuma-system", "kuma-ingress"),
			func(cluster Cluster) error {
				ctx := context.Background()
				cs, err := k8s.GetKubernetesClientFromOptionsE(c.GetTesting(), c.GetKubectlOptions())
				if err != nil {
					return err
				}
				ingressSvc, err := cs.CoreV1().Services("kuma-system").Get(ctx, "kuma-ingress", kube_meta.GetOptions{})
				if err != nil {
					return nil
				}
				ingress.Port = ingressSvc.Spec.Ports[0].NodePort

				nodes, err := cs.CoreV1().Nodes().List(ctx, kube_meta.ListOptions{})
				if err != nil {
					return err
				}
				// assume that we have single node cluster
				for _, addr := range nodes.Items[0].Status.Addresses {
					if addr.Type == kube_core.NodeInternalIP {
						ingress.IP = addr.Address
					}
				}
				return nil
			},
		)(c)
	}
}

func DemoClient() InstallFunc {
	return Combine(
		YamlPath(filepath.Join("testdata", "demo-client.yaml")),
		WaitService("kuma-test", "demo-client"),
		WaitNumPods(1, "demo-client"),
		WaitPodsAvailable("kuma-test", "demo-client"),
	)
}

func Combine(fs ...InstallFunc) InstallFunc {
	return func(cluster Cluster) error {
		for _, f := range fs {
			if err := f(cluster); err != nil {
				return err
			}
		}
		return nil
	}
}

func Namespace(name string) InstallFunc {
	return func(cluster Cluster) error {
		return k8s.CreateNamespaceE(cluster.GetTesting(), cluster.GetKubectlOptions(), name)
	}
}

type ClusterSetup struct {
	installFuncs []InstallFunc
}

func NewClusterSetup() *ClusterSetup {
	return &ClusterSetup{}
}

func (cs *ClusterSetup) Install(fn InstallFunc) *ClusterSetup {
	cs.installFuncs = append(cs.installFuncs, fn)
	return cs
}

func (cs *ClusterSetup) Setup(cluster Cluster) error {
	return Combine(cs.installFuncs...)(cluster)
}
