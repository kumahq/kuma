package framework

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type InstallFunc func(cluster Cluster) error

func Yaml(yaml string) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "install yaml resource", DefaultRetries, DefaultTimeout,
			func() (s string, err error) {
				return "", k8s.KubectlApplyFromStringE(cluster.GetTesting(), cluster.GetKubectlOptions(), yaml)
			})
		return err
	}
}

func YamlPath(path string) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "install yaml resource by path", DefaultRetries, DefaultTimeout,
			func() (s string, err error) {
				return "", k8s.KubectlApplyE(cluster.GetTesting(), cluster.GetKubectlOptions(), path)
			})
		return err
	}
}

func Kuma() InstallFunc {
	return func(cluster Cluster) error {
		_, err := cluster.DeployKuma()
		return err
	}
}

func KumaDNS() InstallFunc {
	return func(cluster Cluster) error {
		err := cluster.InjectDNS()
		return err
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
			}, num, DefaultRetries, DefaultTimeout)
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
			err := k8s.WaitUntilPodAvailableE(c.GetTesting(), c.GetKubectlOptions(namespace), p.GetName(), DefaultRetries, DefaultTimeout)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func EchoServer() InstallFunc {
	const name = "echo-server"
	return Combine(
		YamlPath(filepath.Join("testdata", fmt.Sprintf("%s.yaml", name))),
		WaitService(TestNamespace, name),
		WaitNumPods(1, name),
		WaitPodsAvailable(TestNamespace, name),
	)
}

type IngressDesc struct {
	Port int32
	IP   string
}

func Ingress(ingress *IngressDesc) InstallFunc {
	const name = "kuma-ingress"
	return func(c Cluster) error {
		yaml, err := c.GetKumactlOptions().KumactlInstallIngress()
		if err != nil {
			return err
		}
		return Combine(
			Yaml(yaml),
			WaitService(kumaNamespace, name),
			WaitNumPods(1, name),
			WaitPodsAvailable(kumaNamespace, name),
			func(cluster Cluster) error {
				ctx := context.Background()
				cs, err := k8s.GetKubernetesClientFromOptionsE(c.GetTesting(), c.GetKubectlOptions())
				if err != nil {
					return err
				}
				ingressSvc, err := cs.CoreV1().Services(kumaNamespace).Get(ctx, name, kube_meta.GetOptions{})
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
	const name = "demo-client"
	return Combine(
		YamlPath(filepath.Join("testdata", fmt.Sprintf("%s.yaml", name))),
		WaitService(TestNamespace, name),
		WaitNumPods(1, name),
		WaitPodsAvailable(TestNamespace, name),
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
