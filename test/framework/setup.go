package framework

import (
	"bytes"
	"fmt"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"

	bootstrap_k8s "github.com/kumahq/kuma/pkg/plugins/bootstrap/k8s"
	"github.com/kumahq/kuma/pkg/tls"
)

type InstallFunc func(cluster Cluster) error

var Serializer *k8sjson.Serializer

func init() {
	K8sScheme, err := bootstrap_k8s.NewScheme()
	if err != nil {
		panic(err)
	}

	Serializer = k8sjson.NewSerializerWithOptions(
		k8sjson.DefaultMetaFactory, K8sScheme, K8sScheme,
		k8sjson.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: true,
		},
	)
}

func YamlK8sObject(obj runtime.Object) InstallFunc {
	return func(cluster Cluster) error {
		b := bytes.Buffer{}
		err := Serializer.Encode(obj, &b)
		if err != nil {
			return err
		}
		_, err = retry.DoWithRetryE(cluster.GetTesting(), "install yaml resource", DefaultRetries, DefaultTimeout,
			func() (s string, err error) {
				return "", k8s.KubectlApplyFromStringE(cluster.GetTesting(), cluster.GetKubectlOptions(), b.String())
			})
		return err
	}
}

func YamlK8s(yaml string) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "install yaml resource", DefaultRetries, DefaultTimeout,
			func() (s string, err error) {
				return "", k8s.KubectlApplyFromStringE(cluster.GetTesting(), cluster.GetKubectlOptions(), yaml)
			})
		return err
	}
}

func YamlUniversal(yaml string) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "install yaml resource", DefaultRetries, DefaultTimeout,
			func() (s string, err error) {
				kumactl := cluster.GetKumactlOptions()
				return "", kumactl.KumactlApplyFromString(yaml)
			})
		return err
	}
}

func YamlPathK8s(path string) InstallFunc {
	return func(cluster Cluster) error {
		_, err := retry.DoWithRetryE(cluster.GetTesting(), "install yaml resource by path", DefaultRetries, DefaultTimeout,
			func() (s string, err error) {
				return "", k8s.KubectlApplyE(cluster.GetTesting(), cluster.GetKubectlOptions(), path)
			})
		return err
	}
}

func Kuma(mode string, fs ...DeployOptionsFunc) InstallFunc {
	return func(cluster Cluster) error {
		fs = append(fs, WithIPv6(IsIPv6()))
		err := cluster.DeployKuma(mode, fs...)
		return err
	}
}

func KumaDNS() InstallFunc {
	return func(cluster Cluster) error {
		err := cluster.InjectDNS(KumaNamespace)
		return err
	}
}

func WaitService(namespace, service string) InstallFunc {
	return func(c Cluster) error {
		k8s.WaitUntilServiceAvailable(c.GetTesting(), c.GetKubectlOptions(namespace), service, 10, 3*time.Second)
		return nil
	}
}

func WaitNumPodsNamespace(namespace string, num int, app string) InstallFunc {
	return func(c Cluster) error {
		k8s.WaitUntilNumPodsCreated(c.GetTesting(), c.GetKubectlOptions(namespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", app),
			}, num, DefaultRetries, DefaultTimeout)
		return nil
	}
}

func WaitNumPods(num int, app string) InstallFunc {
	return func(c Cluster) error {
		k8s.WaitUntilNumPodsCreated(c.GetTesting(), c.GetKubectlOptions(),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", app),
			}, num, DefaultRetries, DefaultTimeout)
		return nil
	}
}

func WaitPodsAvailable(namespace, app string) InstallFunc {
	return func(c Cluster) error {
		pods, err := k8s.ListPodsE(c.GetTesting(), c.GetKubectlOptions(namespace),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", app)})
		if err != nil {
			return err
		}
		for _, p := range pods {
			err := WaitUntilPodReadyE(c.GetTesting(), c.GetKubectlOptions(namespace), p.GetName(), DefaultRetries, DefaultTimeout)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// WaitUntilPodReadyE waits until all of the containers within the pod are ready, retrying the check for the specified amount of times, sleeping
// for the provided duration between each try.
func WaitUntilPodReadyE(t testing.TestingT, options *k8s.KubectlOptions, podName string, retries int, sleepBetweenRetries time.Duration) error {
	statusMsg := fmt.Sprintf("Wait for pod %s to be provisioned.", podName)
	message, err := retry.DoWithRetryE(
		t,
		statusMsg,
		retries,
		sleepBetweenRetries,
		func() (string, error) {
			pod, err := k8s.GetPodE(t, options, podName)
			if err != nil {
				return "", err
			}
			if !IsPodReady(pod) {
				return "", k8s.NewPodNotAvailableError(pod)
			}
			return "Pod is now available", nil
		},
	)
	if err != nil {
		logger.Default.Logf(t, "Timeout waiting for Pod to be provisioned: %s", err)
		return err
	}
	logger.Default.Logf(t, message)
	return nil
}

// IsPodReady returns true if the all of the containers within the pod are ready and started
func IsPodReady(pod *corev1.Pod) bool {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if !containerStatus.Ready {
			return false
		}
	}
	return pod.Status.Phase == corev1.PodRunning
}

func WaitUntilPodCompleteE(t testing.TestingT, options *k8s.KubectlOptions, podName string, retries int, sleepBetweenRetries time.Duration) error {
	statusMsg := fmt.Sprintf("Wait for pod %s to be provisioned.", podName)
	message, err := retry.DoWithRetryE(
		t,
		statusMsg,
		retries,
		sleepBetweenRetries,
		func() (string, error) {
			pod, err := k8s.GetPodE(t, options, podName)
			if err != nil {
				return "", err
			}
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.State.Terminated == nil || cs.State.Terminated.ExitCode != 0 {
					return "", errors.Errorf("Pod is not complete yet")
				}
			}
			return "Pod is now complete", nil
		},
	)
	if err != nil {
		logger.Default.Logf(t, "Timedout waiting for Pod to be completed: %s", err)
		return err
	}
	logger.Default.Logf(t, message)
	return nil
}

func WaitPodsComplete(namespace, app string) InstallFunc {
	return func(c Cluster) error {
		pods, err := k8s.ListPodsE(c.GetTesting(), c.GetKubectlOptions(namespace),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", app)})
		if err != nil {
			return err
		}
		for _, p := range pods {
			err := WaitUntilPodCompleteE(c.GetTesting(), c.GetKubectlOptions(namespace), p.GetName(), DefaultRetries, DefaultTimeout)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WaitPodsNotAvailable(namespace, app string) InstallFunc {
	return func(c Cluster) error {
		pods, err := k8s.ListPodsE(c.GetTesting(), c.GetKubectlOptions(namespace),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", app)})
		if err != nil {
			return err
		}

		for _, p := range pods {
			_, _ = retry.DoWithRetryE(
				c.GetTesting(),
				"Wait pod deletion",
				DefaultRetries,
				DefaultTimeout,
				func() (string, error) {
					pod, err := k8s.GetPodE(c.GetTesting(), c.GetKubectlOptions(namespace), p.GetName())
					if err == nil {
						return "", err
					}
					if !k8s.IsPodAvailable(pod) {
						return "Pod is not available", nil
					}
					return "", errors.Errorf("Pod is still available")
				},
			)
		}
		return nil
	}
}

func IngressUniversal(token string) InstallFunc {
	return func(cluster Cluster) error {
		uniCluster := cluster.(*UniversalCluster)
		isipv6 := IsIPv6()
		verbose := false
		app, err := NewUniversalApp(cluster.GetTesting(), uniCluster.name, AppIngress, AppIngress, isipv6, verbose, []string{})
		if err != nil {
			return err
		}

		app.CreateMainApp([]string{}, []string{})

		err = app.mainApp.Start()
		if err != nil {
			return err
		}
		uniCluster.apps[AppIngress] = app

		publicAddress := uniCluster.apps[AppIngress].ip
		dpyaml := fmt.Sprintf(ZoneIngress, publicAddress, kdsPort, kdsPort)
		return uniCluster.CreateZoneIngress(app, "ingress", app.ip, dpyaml, token, false)
	}
}

func DemoClientK8s(mesh string) InstallFunc {
	const name = "demo-client"
	deployment := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-client
  namespace: kuma-test
  labels:
    app: demo-client
spec:
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: demo-client
  template:
    metadata:
      annotations:
        kuma.io/mesh: %s
      labels:
        app: demo-client
    spec:
      containers:
        - name: demo-client
          image: %s
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 3000
          command: [ "ncat" ]
          args:
            - -lk
            - -p
            - "3000"
          resources:
            limits:
              cpu: 50m
              memory: 128Mi
`
	return Combine(
		YamlK8s(fmt.Sprintf(deployment, mesh, GetUniversalImage())),
		WaitNumPods(1, name),
		WaitPodsAvailable(TestNamespace, name),
	)
}

func NamespaceWithSidecarInjection(namespace string) InstallFunc {
	return YamlK8s(fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
  annotations:
    kuma.io/sidecar-injection: "enabled"
`, namespace))
}

func DemoClientJobK8s(mesh, destination string) InstallFunc {
	const name = "demo-job-client"
	deployment := `
apiVersion: batch/v1
kind: Job
metadata:
  name: demo-job-client
  namespace: kuma-test
  labels:
    app: demo-job-client
spec:
  template:
    metadata:
      annotations:
        kuma.io/mesh: %s
      labels:
        app: demo-job-client
    spec:
      containers:
      - name: demo-job-client
        image: %s
        imagePullPolicy: IfNotPresent
        command: [ "curl" ]
        args:
          - -v
          - -m
          - "3"
          - --fail
          - %s
      restartPolicy: OnFailure
`
	return Combine(
		YamlK8s(fmt.Sprintf(deployment, mesh, GetUniversalImage(), destination)),
		WaitNumPods(1, name),
		WaitPodsComplete(TestNamespace, name),
	)
}

func DemoClientUniversal(name, mesh, token string, fs ...DeployOptionsFunc) InstallFunc {
	return func(cluster Cluster) error {
		opts := newDeployOpt(fs...)
		args := []string{"ncat", "-lvk", "-p", "3000"}
		appYaml := ""
		if opts.transparent {
			appYaml = fmt.Sprintf(DemoClientDataplaneTransparentProxy, mesh, "3000", name, redirectPortInbound, redirectPortInboundV6, redirectPortOutbound)
		} else {
			if opts.serviceProbe {
				appYaml = fmt.Sprintf(DemoClientDataplaneWithServiceProbe, mesh, "13000", "3000", name, "80", "8080")
			} else {
				appYaml = fmt.Sprintf(DemoClientDataplane, mesh, "13000", "3000", name, "80", "8080")
			}
		}
		fs = append(fs, WithName(name), WithMesh(mesh), WithAppname(AppModeDemoClient), WithToken(token), WithArgs(args), WithYaml(appYaml), WithIPv6(IsIPv6()))
		return cluster.DeployApp(fs...)
	}
}

func TestServerUniversal(name, mesh, token string, fs ...DeployOptionsFunc) InstallFunc {
	return func(cluster Cluster) error {
		opts := newDeployOpt(fs...)
		if len(opts.protocol) == 0 {
			opts.protocol = "http"
		}
		if opts.serviceVersion == "" {
			opts.serviceVersion = "v1"
		}
		args := []string{"test-server"}
		if len(opts.appArgs) > 0 {
			args = append(args, opts.appArgs...)
		}
		if opts.serviceName == "" {
			opts.serviceName = "test-server"
		}
		if opts.serviceInstance == "" {
			opts.serviceInstance = "1"
		}

		serviceProbe := ""
		if opts.serviceProbe {
			serviceProbe =
				`    serviceProbe:
      tcp: {}`
		}

		args = append(args, "--port", "8080")
		appYaml := fmt.Sprintf(`
type: Dataplane
mesh: %s
name: {{ name }}
networking:
  address:  {{ address }}
  inbound:
  - port: %s
    servicePort: %s
    tags:
      kuma.io/service: %s
      kuma.io/protocol: %s
      version: %s
      instance: '%s'
      team: server-owners
%s
  transparentProxying:
    redirectPortInbound: %s
    redirectPortInboundV6: %s
    redirectPortOutbound: %s
`, mesh, "80", "8080", opts.serviceName, opts.protocol, opts.serviceVersion, opts.serviceInstance, serviceProbe, redirectPortInbound, redirectPortInboundV6, redirectPortOutbound)

		fs = append(fs,
			WithName(name),
			WithMesh(mesh),
			WithAppname("test-server"),
			WithTransparentProxy(true), // test server is always ment to use with transparent proxy
			WithToken(token),
			WithArgs(args),
			WithYaml(appYaml),
			WithIPv6(IsIPv6()))
		return cluster.DeployApp(fs...)
	}
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

func CreateCertsFor(names []string) (cert, key string, err error) {
	keyPair, err := tls.NewSelfSignedCert("kuma", tls.ServerCertType, names...)
	if err != nil {
		return "", "", err
	}

	return string(keyPair.CertPEM), string(keyPair.KeyPEM), nil
}
